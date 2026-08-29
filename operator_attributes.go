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

// GraphOperatorAttributePolicy controls conflicts independently in each
// attribute scope. The outer policy is borrowed only for the call. A nil
// outer policy is accepted when no result element receives multiple values
// for the same attribute name and same-name operand attributes have matching
// types. A zero-valued scope policy explicitly drops conflicts in that scope.
type GraphOperatorAttributePolicy struct {
	Graph    AttributeCombinationPolicy
	Vertices AttributeCombinationPolicy
	Edges    AttributeCombinationPolicy
}

func (policy *GraphOperatorAttributePolicy) scope(scope AttributeScope) *AttributeCombinationPolicy {
	if policy == nil {
		return nil
	}
	switch scope {
	case AttributeGraph:
		return &policy.Graph
	case AttributeVertex:
		return &policy.Vertices
	case AttributeEdge:
		return &policy.Edges
	default:
		return nil
	}
}

type attributeValues struct {
	typeOf   AttributeType
	numeric  []float64
	strings  []string
	booleans []bool
}

type graphAttributeSnapshot struct {
	graph    map[string]attributeValues
	vertices map[string]attributeValues
	edges    map[string]attributeValues
}

func snapshotGraphAttributesLocked(graph *C.igraph_t) (graphAttributeSnapshot, error) {
	result := graphAttributeSnapshot{}
	var err error
	if result.graph, err = snapshotAttributeScopeLocked(graph, AttributeGraph); err != nil {
		return graphAttributeSnapshot{}, err
	}
	if result.vertices, err = snapshotAttributeScopeLocked(graph, AttributeVertex); err != nil {
		return graphAttributeSnapshot{}, err
	}
	if result.edges, err = snapshotAttributeScopeLocked(graph, AttributeEdge); err != nil {
		return graphAttributeSnapshot{}, err
	}
	return result, nil
}

func snapshotAttributeScopeLocked(graph *C.igraph_t, scope AttributeScope) (map[string]attributeValues, error) {
	metadata, err := attributeMetadataLocked(graph, scope)
	if err != nil {
		return nil, err
	}
	result := make(map[string]attributeValues, len(metadata))
	for _, attribute := range metadata {
		value := attributeValues{typeOf: attribute.Type}
		if scope == AttributeGraph {
			cName := C.CString(attribute.Name)
			switch attribute.Type {
			case AttributeNumeric:
				value.numeric = []float64{float64(C.igraph_cattribute_GAN(graph, cName))}
			case AttributeString:
				value.strings = []string{C.GoString(C.igraph_cattribute_GAS(graph, cName))}
			case AttributeBoolean:
				value.booleans = []bool{C.igraph_cattribute_GAB(graph, cName) != booltoint(false)}
			}
			C.free(unsafe.Pointer(cName))
		} else {
			switch attribute.Type {
			case AttributeNumeric:
				value.numeric, err = numericElementAttributesLocked(graph, scope, attribute.Name, numericElementReadHooks{})
			case AttributeString:
				value.strings, err = stringElementAttributesLocked(graph, scope, attribute.Name)
			case AttributeBoolean:
				value.booleans, err = booleanElementAttributesLocked(graph, scope, attribute.Name)
			}
			if err != nil {
				return nil, err
			}
		}
		result[attribute.Name] = value
	}
	return result, nil
}

func stringElementAttributesLocked(graph *C.igraph_t, scope AttributeScope, name string) ([]string, error) {
	result, err := newStringVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_string_values(graph, cScope, cName, &result.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("snapshot string attributes", int(code))
	}
	return result.slice()
}

func booleanElementAttributesLocked(graph *C.igraph_t, scope AttributeScope, name string) ([]bool, error) {
	result, err := newBoolVector(nil)
	if err != nil {
		return nil, err
	}
	defer result.close()
	cScope, _ := scope.cValue()
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if code := C.go_igraph_cattribute_boolean_values(graph, cScope, cName, &result.value); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("snapshot Boolean attributes", int(code))
	}
	return result.slice()
}

func (snapshot graphAttributeSnapshot) scope(scope AttributeScope) map[string]attributeValues {
	switch scope {
	case AttributeGraph:
		return snapshot.graph
	case AttributeVertex:
		return snapshot.vertices
	case AttributeEdge:
		return snapshot.edges
	default:
		return nil
	}
}

type mappedAttributeSource struct {
	attributes map[string]attributeValues
	oldToNew   []int
}

func restoreMappedAttributes(result *Graph, scope AttributeScope, count int, sources []mappedAttributeSource, policy *AttributeCombinationPolicy) error {
	names := make(map[string]struct{})
	for _, source := range sources {
		for name := range source.attributes {
			names[name] = struct{}{}
		}
	}
	orderedNames := make([]string, 0, len(names))
	for name := range names {
		orderedNames = append(orderedNames, name)
	}
	sort.Strings(orderedNames)
	if policy != nil {
		if _, err := policy.Default.cValue(); err != nil {
			return err
		}
		for name, rule := range policy.Overrides {
			if err := validateAttributeName(name); err != nil {
				return err
			}
			if _, ok := names[name]; !ok {
				return fmt.Errorf("%w: %s attribute combination override %q", ErrAttributeNotFound, attributeScopeLabel(scope), name)
			}
			if _, err := rule.cValue(); err != nil {
				return err
			}
		}
	}

	for _, name := range orderedNames {
		var attributeType AttributeType
		typeSet := false
		typeConflict := false
		contributions := make([][]attributeScalar, count)
		for _, source := range sources {
			values, ok := source.attributes[name]
			if !ok {
				continue
			}
			if typeSet && attributeType != values.typeOf {
				typeConflict = true
			}
			attributeType, typeSet = values.typeOf, true
			for oldID, newID := range source.oldToNew {
				if newID == RemovedID {
					continue
				}
				contributions[newID] = append(contributions[newID], values.scalar(oldID))
			}
		}
		needsCombination := typeConflict
		for _, values := range contributions {
			if len(values) > 1 {
				needsCombination = true
				break
			}
		}
		rule := AttributeCombineFirst
		if needsCombination {
			if policy == nil {
				return fmt.Errorf("igraph: an explicit %s attribute combination policy is required for %q", attributeScopeLabel(scope), name)
			}
			rule = policy.Default
			if override, ok := policy.Overrides[name]; ok {
				rule = override
			}
			if _, err := rule.cValue(); err != nil {
				return err
			}
			if rule == AttributeCombineDrop {
				continue
			}
			if typeConflict {
				return fmt.Errorf("%w: %s attribute %q has different operand types", ErrAttributeTypeMismatch, attributeScopeLabel(scope), name)
			}
			if err := validateCombinationType(rule, attributeType); err != nil {
				return fmt.Errorf("igraph: %s attribute %q: %w", attributeScopeLabel(scope), name, err)
			}
		}
		combined, err := combineAttributeValues(attributeType, contributions, rule)
		if err != nil {
			return err
		}
		if err := setRawAttributeValues(result, scope, name, combined); err != nil {
			return err
		}
	}
	return nil
}

type attributeScalar struct {
	numeric float64
	stringV string
	boolean bool
}

func (values attributeValues) scalar(index int) attributeScalar {
	switch values.typeOf {
	case AttributeNumeric:
		return attributeScalar{numeric: values.numeric[index]}
	case AttributeString:
		return attributeScalar{stringV: values.strings[index]}
	default:
		return attributeScalar{boolean: values.booleans[index]}
	}
}

func combineAttributeValues(attributeType AttributeType, contributions [][]attributeScalar, rule AttributeCombination) (attributeValues, error) {
	result := attributeValues{typeOf: attributeType}
	switch attributeType {
	case AttributeNumeric:
		result.numeric = make([]float64, len(contributions))
		for i, values := range contributions {
			if len(values) == 0 {
				result.numeric[i] = math.NaN()
				continue
			}
			numbers := make([]float64, len(values))
			for j := range values {
				numbers[j] = values[j].numeric
			}
			result.numeric[i] = combineNumbers(numbers, rule)
		}
	case AttributeString:
		result.strings = make([]string, len(contributions))
		for i, values := range contributions {
			if len(values) == 0 {
				continue
			}
			if rule == AttributeCombineLast {
				result.strings[i] = values[len(values)-1].stringV
			} else if rule == AttributeCombineConcat {
				for _, value := range values {
					result.strings[i] += value.stringV
				}
			} else {
				result.strings[i] = values[0].stringV
			}
		}
	case AttributeBoolean:
		result.booleans = make([]bool, len(contributions))
		for i, values := range contributions {
			if len(values) == 0 {
				continue
			}
			if rule == AttributeCombineLast {
				result.booleans[i] = values[len(values)-1].boolean
			} else {
				result.booleans[i] = values[0].boolean
			}
		}
	default:
		return attributeValues{}, fmt.Errorf("igraph: unsupported attribute type: %d", attributeType)
	}
	return result, nil
}

func combineNumbers(values []float64, rule AttributeCombination) float64 {
	switch rule {
	case AttributeCombineLast:
		return values[len(values)-1]
	case AttributeCombineSum, AttributeCombineMean:
		value := 0.0
		for _, item := range values {
			value += item
		}
		if rule == AttributeCombineMean {
			value /= float64(len(values))
		}
		return value
	case AttributeCombineProduct:
		value := 1.0
		for _, item := range values {
			value *= item
		}
		return value
	case AttributeCombineMinimum, AttributeCombineMaximum:
		value := values[0]
		for _, item := range values[1:] {
			if (rule == AttributeCombineMinimum && item < value) || (rule == AttributeCombineMaximum && item > value) {
				value = item
			}
		}
		return value
	default:
		return values[0]
	}
}

func setRawAttributeValues(graph *Graph, scope AttributeScope, name string, values attributeValues) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	if scope == AttributeGraph {
		switch values.typeOf {
		case AttributeNumeric:
			if code := C.go_igraph_cattribute_GAN_set(&graph.graph, cName, C.igraph_real_t(values.numeric[0])); code != C.IGRAPH_SUCCESS {
				return igraphError("restore numeric graph attribute", int(code))
			}
		case AttributeString:
			cValue := C.CString(values.strings[0])
			defer C.free(unsafe.Pointer(cValue))
			if code := C.go_igraph_cattribute_GAS_set(&graph.graph, cName, cValue); code != C.IGRAPH_SUCCESS {
				return igraphError("restore string graph attribute", int(code))
			}
		case AttributeBoolean:
			if code := C.go_igraph_cattribute_GAB_set(&graph.graph, cName, booltoint(values.booleans[0])); code != C.IGRAPH_SUCCESS {
				return igraphError("restore Boolean graph attribute", int(code))
			}
		}
		return nil
	}
	cScope, _ := scope.cValue()
	switch values.typeOf {
	case AttributeNumeric:
		vector, err := newRealVector(values.numeric)
		if err != nil {
			return err
		}
		defer vector.close()
		if code := C.go_igraph_cattribute_numeric_setv(&graph.graph, cScope, cName, &vector.value); code != C.IGRAPH_SUCCESS {
			return igraphError("restore numeric element attribute", int(code))
		}
	case AttributeString:
		vector, err := newStringVector(values.strings)
		if err != nil {
			return err
		}
		defer vector.close()
		if code := C.go_igraph_cattribute_string_setv(&graph.graph, cScope, cName, &vector.value); code != C.IGRAPH_SUCCESS {
			return igraphError("restore string element attribute", int(code))
		}
	case AttributeBoolean:
		vector, err := newBoolVector(values.booleans)
		if err != nil {
			return err
		}
		defer vector.close()
		if code := C.go_igraph_cattribute_boolean_setv(&graph.graph, cScope, cName, &vector.value); code != C.IGRAPH_SUCCESS {
			return igraphError("restore Boolean element attribute", int(code))
		}
	}
	return nil
}

func restoreBinaryOperatorAttributes(result *Graph, left, right graphAttributeSnapshot, leftMap, rightMap GraphIDMapping, policy *GraphOperatorAttributePolicy) error {
	for _, item := range []struct {
		scope AttributeScope
		count int
		left  []int
		right []int
	}{
		{AttributeGraph, 1, []int{0}, []int{0}},
		{AttributeVertex, len(leftMap.Vertices.NewToOld), leftMap.Vertices.OldToNew, rightMap.Vertices.OldToNew},
		{AttributeEdge, len(leftMap.Edges.NewToOld), leftMap.Edges.OldToNew, rightMap.Edges.OldToNew},
	} {
		if err := restoreMappedAttributes(result, item.scope, item.count, []mappedAttributeSource{
			{left.scope(item.scope), item.left},
			{right.scope(item.scope), item.right},
		}, policy.scope(item.scope)); err != nil {
			return err
		}
	}
	return nil
}

func restoreManyOperatorAttributes(result *Graph, snapshots []graphAttributeSnapshot, mappings []GraphIDMapping, policy *GraphOperatorAttributePolicy) error {
	if len(snapshots) != len(mappings) {
		return errors.New("igraph: many-graph attribute snapshots and mappings are misaligned")
	}
	for _, item := range []struct {
		scope AttributeScope
		count int
		maps  func(GraphIDMapping) []int
	}{
		{AttributeGraph, 1, func(GraphIDMapping) []int { return []int{0} }},
		{AttributeVertex, lenResultMapping(mappings, true), func(mapping GraphIDMapping) []int { return mapping.Vertices.OldToNew }},
		{AttributeEdge, lenResultMapping(mappings, false), func(mapping GraphIDMapping) []int { return mapping.Edges.OldToNew }},
	} {
		sources := make([]mappedAttributeSource, len(snapshots))
		for index := range snapshots {
			sources[index] = mappedAttributeSource{snapshots[index].scope(item.scope), item.maps(mappings[index])}
		}
		if err := restoreMappedAttributes(result, item.scope, item.count, sources, policy.scope(item.scope)); err != nil {
			return err
		}
	}
	return nil
}

func lenResultMapping(mappings []GraphIDMapping, vertices bool) int {
	if len(mappings) == 0 {
		return 0
	}
	if vertices {
		return len(mappings[0].Vertices.NewToOld)
	}
	return len(mappings[0].Edges.NewToOld)
}

func expandAttributeValues(attributes map[string]attributeValues, sourceIDs []int) map[string]attributeValues {
	result := make(map[string]attributeValues, len(attributes))
	for name, values := range attributes {
		expanded := attributeValues{typeOf: values.typeOf}
		switch values.typeOf {
		case AttributeNumeric:
			expanded.numeric = make([]float64, len(sourceIDs))
			for i, sourceID := range sourceIDs {
				expanded.numeric[i] = values.numeric[sourceID]
			}
		case AttributeString:
			expanded.strings = make([]string, len(sourceIDs))
			for i, sourceID := range sourceIDs {
				expanded.strings[i] = values.strings[sourceID]
			}
		case AttributeBoolean:
			expanded.booleans = make([]bool, len(sourceIDs))
			for i, sourceID := range sourceIDs {
				expanded.booleans[i] = values.booleans[sourceID]
			}
		}
		result[name] = expanded
	}
	return result
}

func restoreCompositionAttributes(result *Graph, left, right graphAttributeSnapshot, provenance CompositionResult, policy *GraphOperatorAttributePolicy) error {
	if err := restoreMappedAttributes(result, AttributeGraph, 1, []mappedAttributeSource{
		{left.graph, []int{0}}, {right.graph, []int{0}},
	}, policy.scope(AttributeGraph)); err != nil {
		return err
	}
	if err := restoreMappedAttributes(result, AttributeVertex, len(provenance.LeftVertices.NewToOld), []mappedAttributeSource{
		{left.vertices, provenance.LeftVertices.OldToNew},
		{right.vertices, provenance.RightVertices.OldToNew},
	}, policy.scope(AttributeVertex)); err != nil {
		return err
	}
	leftIDs := make([]int, len(provenance.Edges))
	rightIDs := make([]int, len(provenance.Edges))
	identity := make([]int, len(provenance.Edges))
	for i, edge := range provenance.Edges {
		leftIDs[i], rightIDs[i], identity[i] = edge.LeftEdge, edge.RightEdge, i
	}
	return restoreMappedAttributes(result, AttributeEdge, len(provenance.Edges), []mappedAttributeSource{
		{expandAttributeValues(left.edges, leftIDs), identity},
		{expandAttributeValues(right.edges, rightIDs), identity},
	}, policy.scope(AttributeEdge))
}
