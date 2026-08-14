package igraph

// #cgo pkg-config: igraph
// #include <stdlib.h>
// #include <igraph.h>
// #include "attributes_cgo.h"
import "C"

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"unsafe"
)

// ErrAttributeNotFound reports that a requested attribute name does not exist
// in the requested scope.
var ErrAttributeNotFound = errors.New("igraph: attribute not found")

// ErrAttributeTypeMismatch reports that an attribute exists with a different
// type than the typed operation requires.
var ErrAttributeTypeMismatch = errors.New("igraph: attribute type mismatch")

// GraphAttributes returns graph-level attribute metadata sorted by name. The
// returned non-nil slice and all names in it are Go-owned and remain valid
// after the graph is closed.
//
//igraph:bind igraph_cattribute_list
func (g *Graph) GraphAttributes() ([]AttributeMetadata, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	return graphAttributesLocked(&g.graph)
}

func graphAttributesLocked(graph *C.igraph_t) ([]AttributeMetadata, error) {
	return attributeMetadataLocked(graph, AttributeGraph)
}

type graphAttributeListHooks struct {
	newNames   func() (*stringVector, error)
	newTypes   func() (*intVector, error)
	list       func() error
	namesClose func()
	typesClose func()
}

func graphAttributesLockedWithHooks(
	graph *C.igraph_t,
	hooks graphAttributeListHooks,
) ([]AttributeMetadata, error) {
	return attributeMetadataLockedWithHooks(graph, AttributeGraph, hooks)
}

func attributeMetadataLocked(
	graph *C.igraph_t,
	scope AttributeScope,
) ([]AttributeMetadata, error) {
	return attributeMetadataLockedWithHooks(graph, scope, graphAttributeListHooks{})
}

func attributeMetadataLockedWithHooks(
	graph *C.igraph_t,
	scope AttributeScope,
	hooks graphAttributeListHooks,
) ([]AttributeMetadata, error) {
	cScope, err := scope.cValue()
	if err != nil {
		return nil, err
	}
	newNames := hooks.newNames
	if newNames == nil {
		newNames = func() (*stringVector, error) { return newStringVector(nil) }
	}
	names, err := newNames()
	if err != nil {
		return nil, err
	}
	defer func() {
		names.close()
		if hooks.namesClose != nil {
			hooks.namesClose()
		}
	}()

	newTypes := hooks.newTypes
	if newTypes == nil {
		newTypes = func() (*intVector, error) { return newIntVector(nil) }
	}
	types, err := newTypes()
	if err != nil {
		return nil, err
	}
	defer func() {
		types.close()
		if hooks.typesClose != nil {
			hooks.typesClose()
		}
	}()

	if hooks.list != nil {
		if err := hooks.list(); err != nil {
			return nil, err
		}
	} else if code := C.go_igraph_cattribute_list_scope(
		graph,
		cScope,
		&names.value,
		&types.value,
	); code != C.IGRAPH_SUCCESS {
		return nil, igraphError(fmt.Sprintf("list %s attributes", attributeScopeLabel(scope)), int(code))
	}
	goNames, err := names.slice()
	if err != nil {
		return nil, err
	}
	goTypes, err := types.slice()
	if err != nil {
		return nil, err
	}
	metadata, err := attributeMetadataFromSlices(scope, goNames, goTypes)
	if err != nil {
		return nil, err
	}
	sort.Slice(metadata, func(i, j int) bool { return metadata[i].Name < metadata[j].Name })
	return metadata, nil
}

func graphAttributeTypeLocked(graph *C.igraph_t, name string) (AttributeType, error) {
	metadata, err := graphAttributesLocked(graph)
	if err != nil {
		return 0, err
	}
	index := sort.Search(len(metadata), func(i int) bool { return metadata[i].Name >= name })
	if index == len(metadata) || metadata[index].Name != name {
		return 0, fmt.Errorf("%w: graph attribute %q", ErrAttributeNotFound, name)
	}
	return metadata[index].Type, nil
}

func requireGraphAttributeTypeLocked(
	graph *C.igraph_t,
	name string,
	want AttributeType,
) error {
	got, err := graphAttributeTypeLocked(graph, name)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf(
			"%w: graph attribute %q has type %d, want %d",
			ErrAttributeTypeMismatch,
			name,
			got,
			want,
		)
	}
	return nil
}

func validateGraphAttributeOverwriteLocked(
	graph *C.igraph_t,
	name string,
	want AttributeType,
) error {
	got, err := graphAttributeTypeLocked(graph, name)
	if errors.Is(err, ErrAttributeNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf(
			"%w: graph attribute %q has type %d, cannot overwrite with %d",
			ErrAttributeTypeMismatch,
			name,
			got,
			want,
		)
	}
	return nil
}

// GraphNumericAttribute returns a numeric graph attribute. The name is
// borrowed only for the synchronous call.
//
//igraph:bind igraph_cattribute_GAN
func (g *Graph) GraphNumericAttribute(name string) (float64, error) {
	if g == nil {
		return 0, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return 0, ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return 0, err
	}
	if err := requireGraphAttributeTypeLocked(&g.graph, name, AttributeNumeric); err != nil {
		return 0, err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return float64(C.igraph_cattribute_GAN(&g.graph, cName)), nil
}

// GraphStringAttribute returns a Go-owned string graph attribute. The name is
// borrowed only for the synchronous call.
//
//igraph:bind igraph_cattribute_GAS
func (g *Graph) GraphStringAttribute(name string) (string, error) {
	if g == nil {
		return "", ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return "", ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return "", err
	}
	if err := requireGraphAttributeTypeLocked(&g.graph, name, AttributeString); err != nil {
		return "", err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.GoString(C.igraph_cattribute_GAS(&g.graph, cName)), nil
}

// GraphBooleanAttribute returns a Boolean graph attribute. The name is
// borrowed only for the synchronous call.
//
//igraph:bind igraph_cattribute_GAB
func (g *Graph) GraphBooleanAttribute(name string) (bool, error) {
	if g == nil {
		return false, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return false, ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return false, err
	}
	if err := requireGraphAttributeTypeLocked(&g.graph, name, AttributeBoolean); err != nil {
		return false, err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.igraph_cattribute_GAB(&g.graph, cName) != booltoint(false), nil
}

// SetGraphNumericAttribute sets or overwrites a numeric graph attribute. The
// name is borrowed only for the synchronous call. NaN and infinities are
// rejected before the graph is modified.
//
//igraph:bind igraph_cattribute_GAN_set
func (g *Graph) SetGraphNumericAttribute(name string, value float64) error {
	return g.setGraphNumericAttribute(name, value, nil)
}

func (g *Graph) setGraphNumericAttribute(name string, value float64, beforeSet func() error) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return err
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("igraph: numeric graph attribute %q must be finite", name)
	}
	if err := validateGraphAttributeOverwriteLocked(&g.graph, name, AttributeNumeric); err != nil {
		return err
	}
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_GAN_set(&g.graph, cName, C.igraph_real_t(value)); code != C.IGRAPH_SUCCESS {
		return igraphError("set numeric graph attribute", int(code))
	}
	return nil
}

// SetGraphStringAttribute sets or overwrites a string graph attribute. The
// name and value are borrowed only for the synchronous call and copied by
// igraph before return. Empty strings are valid.
//
//igraph:bind igraph_cattribute_GAS_set
func (g *Graph) SetGraphStringAttribute(name, value string) error {
	return g.setGraphStringAttribute(name, value, nil)
}

func (g *Graph) setGraphStringAttribute(name, value string, beforeSet func() error) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return err
	}
	if err := validateAttributeString("string graph attribute value", value); err != nil {
		return err
	}
	if err := validateGraphAttributeOverwriteLocked(&g.graph, name, AttributeString); err != nil {
		return err
	}
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cName := C.CString(name)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))
	if code := C.go_igraph_cattribute_GAS_set(&g.graph, cName, cValue); code != C.IGRAPH_SUCCESS {
		return igraphError("set string graph attribute", int(code))
	}
	return nil
}

// SetGraphBooleanAttribute sets or overwrites a Boolean graph attribute. The
// name is borrowed only for the synchronous call.
//
//igraph:bind igraph_cattribute_GAB_set
func (g *Graph) SetGraphBooleanAttribute(name string, value bool) error {
	return g.setGraphBooleanAttribute(name, value, nil)
}

func (g *Graph) setGraphBooleanAttribute(name string, value bool, beforeSet func() error) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return err
	}
	if err := validateGraphAttributeOverwriteLocked(&g.graph, name, AttributeBoolean); err != nil {
		return err
	}
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_GAB_set(&g.graph, cName, booltoint(value)); code != C.IGRAPH_SUCCESS {
		return igraphError("set Boolean graph attribute", int(code))
	}
	return nil
}

// RemoveGraphAttribute removes one graph attribute. A missing name returns
// ErrAttributeNotFound and leaves the graph unchanged. The name is borrowed
// only for the synchronous call.
//
//igraph:bind igraph_cattribute_remove_g
func (g *Graph) RemoveGraphAttribute(name string) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return err
	}
	if _, err := graphAttributeTypeLocked(&g.graph, name); err != nil {
		return err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.igraph_cattribute_remove_g(&g.graph, cName)
	return nil
}

// RemoveAllGraphAttributes removes every graph-level attribute. Vertex and
// edge attributes are not changed.
//
//igraph:bind igraph_cattribute_remove_all
func (g *Graph) RemoveAllGraphAttributes() error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	C.igraph_cattribute_remove_all(&g.graph, booltoint(true), booltoint(false), booltoint(false))
	return nil
}
