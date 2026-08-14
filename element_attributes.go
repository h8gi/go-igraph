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

func attributeScopeLabel(scope AttributeScope) string {
	switch scope {
	case AttributeVertex:
		return "vertex"
	case AttributeEdge:
		return "edge"
	default:
		return "graph"
	}
}

func elementCountLocked(graph *C.igraph_t, scope AttributeScope) (int, error) {
	switch scope {
	case AttributeVertex:
		return igraphIntToInt(C.igraph_vcount(graph), "vertex count")
	case AttributeEdge:
		return igraphIntToInt(C.igraph_ecount(graph), "edge count")
	default:
		return 0, fmt.Errorf("igraph: invalid element attribute scope: %d", scope)
	}
}

func validateElementIDLocked(graph *C.igraph_t, scope AttributeScope, id int) error {
	count, err := elementCountLocked(graph, scope)
	if err != nil {
		return err
	}
	if id < 0 || id >= count {
		return fmt.Errorf(
			"igraph: %s ID %d out of range [0, %d)",
			attributeScopeLabel(scope),
			id,
			count,
		)
	}
	return nil
}

func elementAttributeTypeLocked(
	graph *C.igraph_t,
	scope AttributeScope,
	name string,
) (AttributeType, error) {
	metadata, err := attributeMetadataLocked(graph, scope)
	if err != nil {
		return 0, err
	}
	index := sort.Search(len(metadata), func(i int) bool { return metadata[i].Name >= name })
	if index == len(metadata) || metadata[index].Name != name {
		return 0, fmt.Errorf(
			"%w: %s attribute %q",
			ErrAttributeNotFound,
			attributeScopeLabel(scope),
			name,
		)
	}
	return metadata[index].Type, nil
}

func requireElementAttributeTypeLocked(
	graph *C.igraph_t,
	scope AttributeScope,
	name string,
	want AttributeType,
) error {
	got, err := elementAttributeTypeLocked(graph, scope, name)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf(
			"%w: %s attribute %q has type %d, want %d",
			ErrAttributeTypeMismatch,
			attributeScopeLabel(scope),
			name,
			got,
			want,
		)
	}
	return nil
}

func validateElementAttributeOverwriteLocked(
	graph *C.igraph_t,
	scope AttributeScope,
	name string,
	want AttributeType,
) error {
	got, err := elementAttributeTypeLocked(graph, scope, name)
	if errors.Is(err, ErrAttributeNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf(
			"%w: %s attribute %q has type %d, cannot overwrite with %d",
			ErrAttributeTypeMismatch,
			attributeScopeLabel(scope),
			name,
			got,
			want,
		)
	}
	return nil
}

func (g *Graph) elementAttributes(scope AttributeScope) ([]AttributeMetadata, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	return attributeMetadataLocked(&g.graph, scope)
}

// VertexAttributes returns vertex attribute metadata sorted by name. The
// returned non-nil slice and names are Go-owned.
func (g *Graph) VertexAttributes() ([]AttributeMetadata, error) {
	return g.elementAttributes(AttributeVertex)
}

// EdgeAttributes returns edge attribute metadata sorted by name. The returned
// non-nil slice and names are Go-owned.
func (g *Graph) EdgeAttributes() ([]AttributeMetadata, error) {
	return g.elementAttributes(AttributeEdge)
}

func (g *Graph) elementNumericAttribute(
	scope AttributeScope,
	name string,
	id int,
) (float64, error) {
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
	if err := validateElementIDLocked(&g.graph, scope, id); err != nil {
		return 0, err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeNumeric); err != nil {
		return 0, err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	switch scope {
	case AttributeVertex:
		return float64(C.igraph_cattribute_VAN(&g.graph, cName, C.igraph_int_t(id))), nil
	case AttributeEdge:
		return float64(C.igraph_cattribute_EAN(&g.graph, cName, C.igraph_int_t(id))), nil
	default:
		return 0, fmt.Errorf("igraph: invalid element attribute scope: %d", scope)
	}
}

// VertexNumericAttribute returns one numeric value by checked vertex ID.
//
//igraph:bind igraph_cattribute_VAN
func (g *Graph) VertexNumericAttribute(name string, vertexID int) (float64, error) {
	return g.elementNumericAttribute(AttributeVertex, name, vertexID)
}

// EdgeNumericAttribute returns one numeric value by checked edge ID.
//
//igraph:bind igraph_cattribute_EAN
func (g *Graph) EdgeNumericAttribute(name string, edgeID int) (float64, error) {
	return g.elementNumericAttribute(AttributeEdge, name, edgeID)
}

type numericElementReadHooks struct {
	newResult   func() (*realVector, error)
	read        func() error
	resultClose func()
}

func numericElementAttributesLocked(
	graph *C.igraph_t,
	scope AttributeScope,
	name string,
	hooks numericElementReadHooks,
) ([]float64, error) {
	newResult := hooks.newResult
	if newResult == nil {
		newResult = func() (*realVector, error) { return newRealVectorSize(0) }
	}
	result, err := newResult()
	if err != nil {
		return nil, err
	}
	defer func() {
		result.close()
		if hooks.resultClose != nil {
			hooks.resultClose()
		}
	}()
	if hooks.read != nil {
		if err := hooks.read(); err != nil {
			return nil, err
		}
	} else {
		cScope, err := scope.cValue()
		if err != nil {
			return nil, err
		}
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		if code := C.go_igraph_cattribute_numeric_values(graph, cScope, cName, &result.value); code != C.IGRAPH_SUCCESS {
			return nil, igraphError("get numeric element attributes", int(code))
		}
	}
	return result.slice()
}

func (g *Graph) elementNumericAttributes(scope AttributeScope, name string) ([]float64, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return nil, err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeNumeric); err != nil {
		return nil, err
	}
	return numericElementAttributesLocked(&g.graph, scope, name, numericElementReadHooks{})
}

// VertexNumericAttributes returns a non-nil Go-owned slice in vertex-ID order.
//
//igraph:bind igraph_cattribute_VANV
func (g *Graph) VertexNumericAttributes(name string) ([]float64, error) {
	return g.elementNumericAttributes(AttributeVertex, name)
}

// EdgeNumericAttributes returns a non-nil Go-owned slice in edge-ID order.
//
//igraph:bind igraph_cattribute_EANV
func (g *Graph) EdgeNumericAttributes(name string) ([]float64, error) {
	return g.elementNumericAttributes(AttributeEdge, name)
}

func (g *Graph) elementStringAttribute(
	scope AttributeScope,
	name string,
	id int,
) (string, error) {
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
	if err := validateElementIDLocked(&g.graph, scope, id); err != nil {
		return "", err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeString); err != nil {
		return "", err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	switch scope {
	case AttributeVertex:
		return C.GoString(C.igraph_cattribute_VAS(&g.graph, cName, C.igraph_int_t(id))), nil
	case AttributeEdge:
		return C.GoString(C.igraph_cattribute_EAS(&g.graph, cName, C.igraph_int_t(id))), nil
	default:
		return "", fmt.Errorf("igraph: invalid element attribute scope: %d", scope)
	}
}

// VertexStringAttribute returns one Go-owned string by checked vertex ID.
//
//igraph:bind igraph_cattribute_VAS
func (g *Graph) VertexStringAttribute(name string, vertexID int) (string, error) {
	return g.elementStringAttribute(AttributeVertex, name, vertexID)
}

// EdgeStringAttribute returns one Go-owned string by checked edge ID.
//
//igraph:bind igraph_cattribute_EAS
func (g *Graph) EdgeStringAttribute(name string, edgeID int) (string, error) {
	return g.elementStringAttribute(AttributeEdge, name, edgeID)
}

func (g *Graph) elementStringAttributes(scope AttributeScope, name string) ([]string, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return nil, err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeString); err != nil {
		return nil, err
	}
	result, err := newStringVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	cScope, err := scope.cValue()
	if err != nil {
		return nil, err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_string_values(&g.graph, cScope, cName, &result.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("get string element attributes", int(code))
	}
	return result.slice()
}

// VertexStringAttributes returns non-nil Go-owned strings in vertex-ID order.
//
//igraph:bind igraph_cattribute_VASV
func (g *Graph) VertexStringAttributes(name string) ([]string, error) {
	return g.elementStringAttributes(AttributeVertex, name)
}

// EdgeStringAttributes returns non-nil Go-owned strings in edge-ID order.
//
//igraph:bind igraph_cattribute_EASV
func (g *Graph) EdgeStringAttributes(name string) ([]string, error) {
	return g.elementStringAttributes(AttributeEdge, name)
}

func (g *Graph) elementBooleanAttribute(
	scope AttributeScope,
	name string,
	id int,
) (bool, error) {
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
	if err := validateElementIDLocked(&g.graph, scope, id); err != nil {
		return false, err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeBoolean); err != nil {
		return false, err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	switch scope {
	case AttributeVertex:
		return C.igraph_cattribute_VAB(&g.graph, cName, C.igraph_int_t(id)) != booltoint(false), nil
	case AttributeEdge:
		return C.igraph_cattribute_EAB(&g.graph, cName, C.igraph_int_t(id)) != booltoint(false), nil
	default:
		return false, fmt.Errorf("igraph: invalid element attribute scope: %d", scope)
	}
}

// VertexBooleanAttribute returns one Boolean value by checked vertex ID.
//
//igraph:bind igraph_cattribute_VAB
func (g *Graph) VertexBooleanAttribute(name string, vertexID int) (bool, error) {
	return g.elementBooleanAttribute(AttributeVertex, name, vertexID)
}

// EdgeBooleanAttribute returns one Boolean value by checked edge ID.
//
//igraph:bind igraph_cattribute_EAB
func (g *Graph) EdgeBooleanAttribute(name string, edgeID int) (bool, error) {
	return g.elementBooleanAttribute(AttributeEdge, name, edgeID)
}

func (g *Graph) elementBooleanAttributes(scope AttributeScope, name string) ([]bool, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return nil, ErrClosed
	}
	if err := validateAttributeName(name); err != nil {
		return nil, err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeBoolean); err != nil {
		return nil, err
	}
	result, err := newBoolVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	cScope, err := scope.cValue()
	if err != nil {
		return nil, err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_boolean_values(&g.graph, cScope, cName, &result.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("get Boolean element attributes", int(code))
	}
	return result.slice()
}

// VertexBooleanAttributes returns a non-nil Go-owned slice in vertex-ID order.
//
//igraph:bind igraph_cattribute_VABV
func (g *Graph) VertexBooleanAttributes(name string) ([]bool, error) {
	return g.elementBooleanAttributes(AttributeVertex, name)
}

// EdgeBooleanAttributes returns a non-nil Go-owned slice in edge-ID order.
//
//igraph:bind igraph_cattribute_EABV
func (g *Graph) EdgeBooleanAttributes(name string) ([]bool, error) {
	return g.elementBooleanAttributes(AttributeEdge, name)
}

func (g *Graph) setElementNumericAttribute(
	scope AttributeScope,
	name string,
	id int,
	value float64,
	beforeSet func() error,
) error {
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
	if err := validateElementIDLocked(&g.graph, scope, id); err != nil {
		return err
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("igraph: numeric %s attribute %q must be finite", attributeScopeLabel(scope), name)
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeNumeric); err != nil {
		return err
	}
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_numeric_set(&g.graph, cScope, cName, C.igraph_int_t(id), C.igraph_real_t(value)); code != C.IGRAPH_SUCCESS {
		return igraphError("set numeric element attribute", int(code))
	}
	return nil
}

// SetVertexNumericAttribute updates an existing attribute at one vertex.
//
//igraph:bind igraph_cattribute_VAN_set
func (g *Graph) SetVertexNumericAttribute(name string, vertexID int, value float64) error {
	return g.setElementNumericAttribute(AttributeVertex, name, vertexID, value, nil)
}

// SetEdgeNumericAttribute updates an existing attribute at one edge.
//
//igraph:bind igraph_cattribute_EAN_set
func (g *Graph) SetEdgeNumericAttribute(name string, edgeID int, value float64) error {
	return g.setElementNumericAttribute(AttributeEdge, name, edgeID, value, nil)
}

func (g *Graph) setElementNumericAttributes(
	scope AttributeScope,
	name string,
	values []float64,
	beforeSet func() error,
) error {
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
	count, err := elementCountLocked(&g.graph, scope)
	if err != nil {
		return err
	}
	if len(values) != count {
		return fmt.Errorf("igraph: %s attribute %q length %d does not match %s count %d", attributeScopeLabel(scope), name, len(values), attributeScopeLabel(scope), count)
	}
	for i, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("igraph: numeric %s attribute %q value at index %d must be finite", attributeScopeLabel(scope), name, i)
		}
	}
	if err := validateElementAttributeOverwriteLocked(&g.graph, scope, name, AttributeNumeric); err != nil {
		return err
	}
	cValues, err := newRealVector(values)
	if err != nil {
		return err
	}
	defer cValues.close()
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_numeric_setv(&g.graph, cScope, cName, &cValues.value); code != C.IGRAPH_SUCCESS {
		return igraphError("set numeric element attributes", int(code))
	}
	return nil
}

// SetVertexNumericAttributes sets a complete vertex-ID-aligned vector. The
// input is borrowed only for the call and may create a new attribute.
//
//igraph:bind igraph_cattribute_VAN_setv
func (g *Graph) SetVertexNumericAttributes(name string, values []float64) error {
	return g.setElementNumericAttributes(AttributeVertex, name, values, nil)
}

// SetEdgeNumericAttributes sets a complete edge-ID-aligned vector. The input
// is borrowed only for the call and may create a new attribute.
//
//igraph:bind igraph_cattribute_EAN_setv
func (g *Graph) SetEdgeNumericAttributes(name string, values []float64) error {
	return g.setElementNumericAttributes(AttributeEdge, name, values, nil)
}

func (g *Graph) setElementStringAttribute(
	scope AttributeScope,
	name string,
	id int,
	value string,
	beforeSet func() error,
) error {
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
	if err := validateElementIDLocked(&g.graph, scope, id); err != nil {
		return err
	}
	if err := validateAttributeString("string element attribute value", value); err != nil {
		return err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeString); err != nil {
		return err
	}
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))
	if code := C.go_igraph_cattribute_string_set(&g.graph, cScope, cName, C.igraph_int_t(id), cValue); code != C.IGRAPH_SUCCESS {
		return igraphError("set string element attribute", int(code))
	}
	return nil
}

// SetVertexStringAttribute updates an existing attribute at one vertex.
//
//igraph:bind igraph_cattribute_VAS_set
func (g *Graph) SetVertexStringAttribute(name string, vertexID int, value string) error {
	return g.setElementStringAttribute(AttributeVertex, name, vertexID, value, nil)
}

// SetEdgeStringAttribute updates an existing attribute at one edge.
//
//igraph:bind igraph_cattribute_EAS_set
func (g *Graph) SetEdgeStringAttribute(name string, edgeID int, value string) error {
	return g.setElementStringAttribute(AttributeEdge, name, edgeID, value, nil)
}

func (g *Graph) setElementStringAttributes(
	scope AttributeScope,
	name string,
	values []string,
	beforeSet func() error,
) error {
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
	count, err := elementCountLocked(&g.graph, scope)
	if err != nil {
		return err
	}
	if len(values) != count {
		return fmt.Errorf("igraph: %s attribute %q length %d does not match %s count %d", attributeScopeLabel(scope), name, len(values), attributeScopeLabel(scope), count)
	}
	if err := validateElementAttributeOverwriteLocked(&g.graph, scope, name, AttributeString); err != nil {
		return err
	}
	cValues, err := newStringVector(values)
	if err != nil {
		return err
	}
	defer cValues.close()
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_string_setv(&g.graph, cScope, cName, &cValues.value); code != C.IGRAPH_SUCCESS {
		return igraphError("set string element attributes", int(code))
	}
	return nil
}

// SetVertexStringAttributes sets complete vertex-ID-aligned strings. The input
// is borrowed only for the call and may create a new attribute.
//
//igraph:bind igraph_cattribute_VAS_setv
func (g *Graph) SetVertexStringAttributes(name string, values []string) error {
	return g.setElementStringAttributes(AttributeVertex, name, values, nil)
}

// SetEdgeStringAttributes sets complete edge-ID-aligned strings. The input is
// borrowed only for the call and may create a new attribute.
//
//igraph:bind igraph_cattribute_EAS_setv
func (g *Graph) SetEdgeStringAttributes(name string, values []string) error {
	return g.setElementStringAttributes(AttributeEdge, name, values, nil)
}

func (g *Graph) setElementBooleanAttribute(
	scope AttributeScope,
	name string,
	id int,
	value bool,
	beforeSet func() error,
) error {
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
	if err := validateElementIDLocked(&g.graph, scope, id); err != nil {
		return err
	}
	if err := requireElementAttributeTypeLocked(&g.graph, scope, name, AttributeBoolean); err != nil {
		return err
	}
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_boolean_set(&g.graph, cScope, cName, C.igraph_int_t(id), booltoint(value)); code != C.IGRAPH_SUCCESS {
		return igraphError("set Boolean element attribute", int(code))
	}
	return nil
}

// SetVertexBooleanAttribute updates an existing attribute at one vertex.
//
//igraph:bind igraph_cattribute_VAB_set
func (g *Graph) SetVertexBooleanAttribute(name string, vertexID int, value bool) error {
	return g.setElementBooleanAttribute(AttributeVertex, name, vertexID, value, nil)
}

// SetEdgeBooleanAttribute updates an existing attribute at one edge.
//
//igraph:bind igraph_cattribute_EAB_set
func (g *Graph) SetEdgeBooleanAttribute(name string, edgeID int, value bool) error {
	return g.setElementBooleanAttribute(AttributeEdge, name, edgeID, value, nil)
}

func (g *Graph) setElementBooleanAttributes(
	scope AttributeScope,
	name string,
	values []bool,
	beforeSet func() error,
) error {
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
	count, err := elementCountLocked(&g.graph, scope)
	if err != nil {
		return err
	}
	if len(values) != count {
		return fmt.Errorf("igraph: %s attribute %q length %d does not match %s count %d", attributeScopeLabel(scope), name, len(values), attributeScopeLabel(scope), count)
	}
	if err := validateElementAttributeOverwriteLocked(&g.graph, scope, name, AttributeBoolean); err != nil {
		return err
	}
	cValues, err := newBoolVector(values)
	if err != nil {
		return err
	}
	defer cValues.close()
	if beforeSet != nil {
		if err := beforeSet(); err != nil {
			return err
		}
	}
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_boolean_setv(&g.graph, cScope, cName, &cValues.value); code != C.IGRAPH_SUCCESS {
		return igraphError("set Boolean element attributes", int(code))
	}
	return nil
}

// SetVertexBooleanAttributes sets a complete vertex-ID-aligned vector. The
// input is borrowed only for the call and may create a new attribute.
//
//igraph:bind igraph_cattribute_VAB_setv
func (g *Graph) SetVertexBooleanAttributes(name string, values []bool) error {
	return g.setElementBooleanAttributes(AttributeVertex, name, values, nil)
}

// SetEdgeBooleanAttributes sets a complete edge-ID-aligned vector. The input
// is borrowed only for the call and may create a new attribute.
//
//igraph:bind igraph_cattribute_EAB_setv
func (g *Graph) SetEdgeBooleanAttributes(name string, values []bool) error {
	return g.setElementBooleanAttributes(AttributeEdge, name, values, nil)
}

func (g *Graph) removeElementAttribute(scope AttributeScope, name string) error {
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
	if _, err := elementAttributeTypeLocked(&g.graph, scope, name); err != nil {
		return err
	}
	cScope, err := scope.cValue()
	if err != nil {
		return err
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.go_igraph_cattribute_remove(&g.graph, cScope, cName)
	return nil
}

// RemoveVertexAttribute removes one complete vertex attribute.
//
//igraph:bind igraph_cattribute_remove_v
func (g *Graph) RemoveVertexAttribute(name string) error {
	return g.removeElementAttribute(AttributeVertex, name)
}

// RemoveEdgeAttribute removes one complete edge attribute.
//
//igraph:bind igraph_cattribute_remove_e
func (g *Graph) RemoveEdgeAttribute(name string) error {
	return g.removeElementAttribute(AttributeEdge, name)
}

func (g *Graph) removeAllElementAttributes(scope AttributeScope) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}
	switch scope {
	case AttributeVertex:
		C.igraph_cattribute_remove_all(&g.graph, booltoint(false), booltoint(true), booltoint(false))
	case AttributeEdge:
		C.igraph_cattribute_remove_all(&g.graph, booltoint(false), booltoint(false), booltoint(true))
	default:
		return fmt.Errorf("igraph: invalid element attribute scope: %d", scope)
	}
	return nil
}

// RemoveAllVertexAttributes removes every vertex attribute without changing
// graph or edge attributes.
func (g *Graph) RemoveAllVertexAttributes() error {
	return g.removeAllElementAttributes(AttributeVertex)
}

// RemoveAllEdgeAttributes removes every edge attribute without changing graph
// or vertex attributes.
func (g *Graph) RemoveAllEdgeAttributes() error {
	return g.removeAllElementAttributes(AttributeEdge)
}
