package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "cohesive_blocks_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// CohesiveBlocksResult is a cohesive block hierarchy. The shared slice index
// is the block ID. Blocks contains source-graph vertex IDs, Cohesion contains
// each block's vertex connectivity, and Parents contains the parent block ID;
// the root at index zero has parent -1. All slices, including every inner block
// slice, are non-nil and Go-owned. Parent blocks precede their children; no
// other block or vertex ordering is promised. BlockTree is an independently
// owned directed tree whose vertex IDs are block IDs and whose edges point from
// parent to child. The caller must close BlockTree; repeated Close calls are
// safe.
type CohesiveBlocksResult struct {
	Blocks    [][]int
	Cohesion  []int
	Parents   []int
	BlockTree *Graph
}

// CohesiveBlocks computes the complete cohesive block hierarchy. The source
// graph must be undirected and simple: loops and parallel edges are rejected.
// Disconnected, empty, singleton, and other small simple graphs are accepted;
// the root block represents the complete source vertex set and may have
// cohesion zero. Results remain valid after the source graph is closed.
//
//igraph:bind igraph_cohesive_blocks
func (g *Graph) CohesiveBlocks() (CohesiveBlocksResult, error) {
	if g == nil {
		return CohesiveBlocksResult{}, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return CohesiveBlocksResult{}, ErrClosed
	}
	if C.igraph_is_directed(&g.graph) != booltoint(false) {
		return CohesiveBlocksResult{}, errors.New("igraph: cohesive blocks require an undirected graph")
	}
	var simple C.igraph_bool_t
	if code := C.go_igraph_is_simple_for_cohesive(&g.graph, &simple); code != C.IGRAPH_SUCCESS {
		return CohesiveBlocksResult{}, igraphError("check cohesive block input simplicity", int(code))
	}
	if simple == booltoint(false) {
		return CohesiveBlocksResult{}, errors.New("igraph: cohesive blocks require a simple graph without loops or parallel edges")
	}
	return collectCohesiveBlocks(int(C.igraph_vcount(&g.graph)), cohesiveBlockOperations{
		newList:     newIntVectorList,
		newVector:   func() (*intVector, error) { return newIntVector(nil) },
		closeList:   (*intVectorList).close,
		closeVector: (*intVector).close,
		query: func(blocks *intVectorList, cohesion, parents *intVector, tree *cohesiveBlockTree) (bool, error) {
			if code := C.go_igraph_cohesive_blocks(&g.graph, &blocks.value, &cohesion.value, &parents.value, &tree.value); code != C.IGRAPH_SUCCESS {
				return false, igraphError("calculate cohesive blocks", int(code))
			}
			return true, nil
		},
		listSlices:  func(list *intVectorList) ([][]int, error) { return list.slices() },
		vectorSlice: func(vector *intVector) ([]int, error) { return vector.slice() },
		treeInfo:    readCohesiveBlockTreeInfo,
		destroyTree: func(tree *cohesiveBlockTree) { C.igraph_destroy(&tree.value) },
		adoptTree:   func(tree *cohesiveBlockTree) *Graph { return adoptInitializedGraph(&tree.value) },
	})
}

type cohesiveBlockTree struct {
	value C.igraph_t
}

type cohesiveBlockTreeInfo struct {
	vertexCount int
	directed    bool
	edges       []Edge
}

type cohesiveBlockOperations struct {
	newList     func() (*intVectorList, error)
	newVector   func() (*intVector, error)
	closeList   func(*intVectorList)
	closeVector func(*intVector)
	query       func(*intVectorList, *intVector, *intVector, *cohesiveBlockTree) (bool, error)
	listSlices  func(*intVectorList) ([][]int, error)
	vectorSlice func(*intVector) ([]int, error)
	treeInfo    func(*cohesiveBlockTree) (cohesiveBlockTreeInfo, error)
	destroyTree func(*cohesiveBlockTree)
	adoptTree   func(*cohesiveBlockTree) *Graph
}

func collectCohesiveBlocks(vertexCount int, operations cohesiveBlockOperations) (CohesiveBlocksResult, error) {
	blocks, err := operations.newList()
	if err != nil {
		return CohesiveBlocksResult{}, err
	}
	defer operations.closeList(blocks)
	cohesion, err := operations.newVector()
	if err != nil {
		return CohesiveBlocksResult{}, err
	}
	defer operations.closeVector(cohesion)
	parents, err := operations.newVector()
	if err != nil {
		return CohesiveBlocksResult{}, err
	}
	defer operations.closeVector(parents)

	var tree cohesiveBlockTree
	initialized, err := operations.query(blocks, cohesion, parents, &tree)
	if initialized {
		defer func() {
			if initialized {
				operations.destroyTree(&tree)
			}
		}()
	}
	if err != nil {
		return CohesiveBlocksResult{}, err
	}
	if !initialized {
		return CohesiveBlocksResult{}, errors.New("igraph: cohesive blocks returned an uninitialized block tree")
	}

	result := CohesiveBlocksResult{}
	result.Blocks, err = operations.listSlices(blocks)
	if err == nil {
		result.Cohesion, err = operations.vectorSlice(cohesion)
	}
	if err == nil {
		result.Parents, err = operations.vectorSlice(parents)
	}
	if err != nil {
		return CohesiveBlocksResult{}, err
	}
	info, err := operations.treeInfo(&tree)
	if err != nil {
		return CohesiveBlocksResult{}, err
	}
	if err := validateCohesiveBlocks(result, vertexCount, info); err != nil {
		return CohesiveBlocksResult{}, err
	}
	result.BlockTree = operations.adoptTree(&tree)
	initialized = false
	return result, nil
}

func readCohesiveBlockTreeInfo(tree *cohesiveBlockTree) (cohesiveBlockTreeInfo, error) {
	vertices, err := igraphIntToInt(C.igraph_vcount(&tree.value), "cohesive block tree vertex count")
	if err != nil {
		return cohesiveBlockTreeInfo{}, err
	}
	edgeCount, err := igraphIntToInt(C.igraph_ecount(&tree.value), "cohesive block tree edge count")
	if err != nil {
		return cohesiveBlockTreeInfo{}, err
	}
	edges := make([]Edge, edgeCount)
	for id := range edges {
		var from, to C.igraph_int_t
		if code := C.igraph_edge(&tree.value, C.igraph_int_t(id), &from, &to); code != C.IGRAPH_SUCCESS {
			return cohesiveBlockTreeInfo{}, igraphError("read cohesive block tree edge", int(code))
		}
		edges[id].From, err = igraphIntToInt(from, "cohesive block tree edge source")
		if err == nil {
			edges[id].To, err = igraphIntToInt(to, "cohesive block tree edge target")
		}
		if err != nil {
			return cohesiveBlockTreeInfo{}, err
		}
	}
	return cohesiveBlockTreeInfo{vertexCount: vertices, directed: C.igraph_is_directed(&tree.value) != booltoint(false), edges: edges}, nil
}

func validateCohesiveBlocks(result CohesiveBlocksResult, vertexCount int, tree cohesiveBlockTreeInfo) error {
	if result.Blocks == nil || result.Cohesion == nil || result.Parents == nil {
		return errors.New("igraph: cohesive blocks returned nil output storage")
	}
	n := len(result.Blocks)
	if n == 0 || len(result.Cohesion) != n || len(result.Parents) != n {
		return fmt.Errorf("igraph: cohesive block output lengths %d, %d, and %d are not aligned", n, len(result.Cohesion), len(result.Parents))
	}
	if tree.vertexCount != n || !tree.directed || len(tree.edges) != n-1 {
		return fmt.Errorf("igraph: cohesive block tree does not match %d blocks", n)
	}
	sets := make([]map[int]struct{}, n)
	for blockID, block := range result.Blocks {
		if block == nil {
			return fmt.Errorf("igraph: cohesive block %d is nil", blockID)
		}
		sets[blockID] = make(map[int]struct{}, len(block))
		for _, vertexID := range block {
			if vertexID < 0 || vertexID >= vertexCount {
				return fmt.Errorf("igraph: cohesive block %d vertex ID %d out of range", blockID, vertexID)
			}
			if _, duplicate := sets[blockID][vertexID]; duplicate {
				return fmt.Errorf("igraph: cohesive block %d repeats vertex ID %d", blockID, vertexID)
			}
			sets[blockID][vertexID] = struct{}{}
		}
		if result.Cohesion[blockID] < 0 {
			return fmt.Errorf("igraph: cohesive block %d has negative cohesion", blockID)
		}
	}
	if len(sets[0]) != vertexCount || result.Parents[0] != -1 {
		return errors.New("igraph: cohesive block root does not represent the source graph")
	}
	for vertexID := range vertexCount {
		if _, ok := sets[0][vertexID]; !ok {
			return errors.New("igraph: cohesive block root omits a source vertex")
		}
	}
	expectedEdges := make(map[Edge]struct{}, n-1)
	for blockID := 1; blockID < n; blockID++ {
		parent := result.Parents[blockID]
		if parent < 0 || parent >= blockID {
			return fmt.Errorf("igraph: cohesive block %d has invalid parent %d", blockID, parent)
		}
		if result.Cohesion[blockID] <= result.Cohesion[parent] {
			return fmt.Errorf("igraph: cohesive block %d cohesion does not exceed its parent", blockID)
		}
		for vertexID := range sets[blockID] {
			if _, ok := sets[parent][vertexID]; !ok {
				return fmt.Errorf("igraph: cohesive block %d is not contained in parent %d", blockID, parent)
			}
		}
		expectedEdges[Edge{From: parent, To: blockID}] = struct{}{}
	}
	for _, edge := range tree.edges {
		if _, ok := expectedEdges[edge]; !ok {
			return fmt.Errorf("igraph: unexpected cohesive block tree edge %d -> %d", edge.From, edge.To)
		}
		delete(expectedEdges, edge)
	}
	if len(expectedEdges) != 0 {
		return errors.New("igraph: cohesive block tree omits a parent edge")
	}
	return nil
}
