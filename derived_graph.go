package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
)

// RemovedID is stored in an IDMapping when an element has no corresponding
// element in the other graph.
const RemovedID = -1

// IDMapping describes how IDs of one element kind change between a source and
// a derived graph. OldToNew is indexed by source ID and contains a derived ID
// or RemovedID. NewToOld is indexed by derived ID and contains the lowest
// source ID mapped to it, or RemovedID when the derived element has no source.
// Choosing the lowest source ID makes NewToOld deterministic for many-to-one
// transformations such as edge simplification.
//
// Both slices are non-nil, Go-owned values. An empty-to-empty mapping contains
// two non-nil empty slices. An identity mapping stores ID i at index i in both
// directions. The slices remain valid and mutable after either graph is closed.
type IDMapping struct {
	OldToNew []int
	NewToOld []int
}

// GraphIDMapping contains the vertex and edge ID mappings between a source and
// a derived graph. Each field follows the indexing and ownership contract of
// IDMapping.
type GraphIDMapping struct {
	Vertices IDMapping
	Edges    IDMapping
}

func newIDMapping(oldToNew []int, newCount int) (IDMapping, error) {
	if newCount < 0 {
		return IDMapping{}, fmt.Errorf("igraph: new ID count must be non-negative: %d", newCount)
	}
	mapping := IDMapping{
		OldToNew: append([]int{}, oldToNew...),
		NewToOld: make([]int, newCount),
	}
	for i := range mapping.NewToOld {
		mapping.NewToOld[i] = RemovedID
	}
	for oldID, newID := range mapping.OldToNew {
		if newID == RemovedID {
			continue
		}
		if newID < 0 || newID >= newCount {
			return IDMapping{}, fmt.Errorf(
				"igraph: mapped ID %d for old ID %d out of range [0, %d)",
				newID, oldID, newCount,
			)
		}
		if mapping.NewToOld[newID] == RemovedID {
			mapping.NewToOld[newID] = oldID
		}
	}
	return mapping, nil
}

func identityIDMapping(count int) (IDMapping, error) {
	if count < 0 {
		return IDMapping{}, fmt.Errorf("igraph: identity ID count must be non-negative: %d", count)
	}
	ids := make([]int, count)
	for id := range ids {
		ids[id] = id
	}
	return newIDMapping(ids, count)
}

// adoptInitializedGraph moves exactly one successfully initialized igraph_t
// into a public Graph. On return value is cleared and must not be destroyed by
// the caller. Failed initializers must never be passed to this function.
func adoptInitializedGraph(value *C.igraph_t) *Graph {
	graph := &Graph{graph: *value}
	*value = C.igraph_t{}
	runtime.SetFinalizer(graph, (*Graph).finalize)
	return graph
}

// graphList owns an initialized igraph_graph_list_t and every graph still in
// it. Removing a graph transfers that one graph's ownership to the receiver.
type graphList struct {
	value       C.igraph_graph_list_t
	initialized bool
}

//igraph:internal igraph_graph_list_init
func newGraphList() (*graphList, error) {
	list := &graphList{}
	return initializeGraphList(list, func() int {
		return int(C.go_igraph_graph_list_init(&list.value, 0))
	})
}

func initializeGraphList(list *graphList, initialize func() int) (*graphList, error) {
	if code := initialize(); code != int(C.IGRAPH_SUCCESS) {
		return nil, igraphError("initialize graph list", code)
	}
	list.initialized = true
	return list, nil
}

//igraph:internal igraph_graph_list_push_back_copy
func (list *graphList) appendCopy(graph *C.igraph_t) error {
	if list == nil || !list.initialized {
		return errors.New("igraph: graph list is not initialized")
	}
	if code := C.go_igraph_graph_list_push_back_copy(&list.value, graph); code != C.IGRAPH_SUCCESS {
		return igraphError("append graph-list copy", int(code))
	}
	return nil
}

// takeGraphs consumes the list. Each returned Graph independently owns one C
// graph. The list container is destroyed on success; on any error it also
// destroys every graph not yet removed and closes every graph already adopted.
//
//igraph:internal igraph_graph_list_size
//igraph:internal igraph_graph_list_remove
func (list *graphList) takeGraphs() ([]*Graph, error) {
	return list.takeGraphsWithHooks(graphListExtractionHooks{})
}

type graphListExtractionHooks struct {
	beforeRemove func(index int) error
	beforeAdopt  func(index int) error
	afterAdopt   func(index int, graph *Graph) error
}

// takeGraphsWithHooks provides failure-injection seams for upstream removal
// and fallible Go conversion around ownership adoption.
func (list *graphList) takeGraphsWithHooks(
	hooks graphListExtractionHooks,
) (result []*Graph, err error) {
	if list == nil || !list.initialized {
		return nil, errors.New("igraph: graph list is not initialized")
	}

	size, err := igraphIntToInt(C.igraph_graph_list_size(&list.value), "graph list length")
	if err != nil {
		list.close()
		return nil, err
	}
	adopted := make([]*Graph, 0, size)
	succeeded := false
	defer func() {
		list.close()
		if !succeeded {
			for _, graph := range adopted {
				_ = graph.Close()
			}
			result = nil
		}
	}()

	for index := 0; index < size; index++ {
		if hooks.beforeRemove != nil {
			if hookErr := hooks.beforeRemove(index); hookErr != nil {
				return nil, fmt.Errorf("igraph: extract graph-list element %d: %w", index, hookErr)
			}
		}
		var value C.igraph_t
		if code := C.go_igraph_graph_list_remove(&list.value, 0, &value); code != C.IGRAPH_SUCCESS {
			return nil, igraphError("extract graph from list", int(code))
		}
		if hooks.beforeAdopt != nil {
			if hookErr := hooks.beforeAdopt(index); hookErr != nil {
				C.igraph_destroy(&value)
				return nil, fmt.Errorf("igraph: convert graph-list element %d: %w", index, hookErr)
			}
		}
		graph := adoptInitializedGraph(&value)
		adopted = append(adopted, graph)
		if hooks.afterAdopt != nil {
			if hookErr := hooks.afterAdopt(index, graph); hookErr != nil {
				return nil, fmt.Errorf("igraph: finish graph-list element %d: %w", index, hookErr)
			}
		}
	}
	succeeded = true
	result = adopted
	return result, nil
}

// newGraphListFromCopies borrows source graphs only while their deterministic
// lock set is held. The returned list owns independent C copies.
func newGraphListFromCopies(graphs []*Graph) (*graphList, error) {
	list, err := newGraphList()
	if err != nil {
		return nil, err
	}
	if err := withLockedGraphs(graphs, func(values []*C.igraph_t) error {
		for _, value := range values {
			if err := list.appendCopy(value); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		list.close()
		return nil, err
	}
	return list, nil
}

//igraph:internal igraph_graph_list_destroy
func (list *graphList) close() {
	if list == nil || !list.initialized {
		return
	}
	C.igraph_graph_list_destroy(&list.value)
	list.initialized = false
}

// withLockedGraphs borrows every graph for one synchronous operation. Locks
// are acquired once per distinct graph in stable address order, so opposite
// argument orders and repeated graph arguments cannot deadlock. The callback
// receives graph pointers in the original argument order; they are valid only
// until the callback returns and must not be retained or destroyed.
func withLockedGraphs(graphs []*Graph, operation func([]*C.igraph_t) error) error {
	if operation == nil {
		return errors.New("igraph: graph operation is nil")
	}
	unique := make(map[*Graph]struct{}, len(graphs))
	ordered := make([]*Graph, 0, len(graphs))
	for _, graph := range graphs {
		if graph == nil {
			return ErrClosed
		}
		if _, exists := unique[graph]; !exists {
			unique[graph] = struct{}{}
			ordered = append(ordered, graph)
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		return reflect.ValueOf(ordered[i]).Pointer() < reflect.ValueOf(ordered[j]).Pointer()
	})
	for _, graph := range ordered {
		graph.mu.Lock()
	}
	defer func() {
		for index := len(ordered) - 1; index >= 0; index-- {
			ordered[index].mu.Unlock()
		}
	}()
	for _, graph := range ordered {
		if graph.closed {
			return ErrClosed
		}
	}
	values := make([]*C.igraph_t, len(graphs))
	for index, graph := range graphs {
		values[index] = &graph.graph
	}
	return operation(values)
}

func withGraphsLocked(graphs []*Graph, operation func() error) error {
	if operation == nil {
		return errors.New("igraph: graph operation is nil")
	}
	return withLockedGraphs(graphs, func(_ []*C.igraph_t) error {
		return operation()
	})
}
