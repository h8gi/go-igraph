package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "operators_cgo.h"
import "C"

import (
	"errors"
	"fmt"
)

// BinaryGraphOperatorResult contains an independently owned result graph and
// exact source-to-result provenance for both operands. Left and Right vertex
// and edge mappings are indexed by IDs in the corresponding operand. All
// mapping slices are non-nil, Go-owned, and survive closure of every graph.
type BinaryGraphOperatorResult struct {
	Graph *Graph
	Left  GraphIDMapping
	Right GraphIDMapping
}

// DifferenceResult contains an independently owned difference graph and exact
// left-operand vertex provenance. Left.Edges is a deterministic structural
// source-to-result mapping: endpoint-equivalent parallel edges are paired by
// ascending source and result edge IDs. It is not attribute provenance. Graph
// must be closed by the caller; mapping slices are non-nil, Go-owned, and
// survive closure of both operands and the result.
type DifferenceResult struct {
	Graph *Graph
	Left  GraphIDMapping
}

// CompositionEdgeProvenance identifies the left and right operand edges whose
// two-step path produced one result edge. CompositionResult.Edges is indexed by
// result edge ID; repeated source edge IDs are preserved because one operand
// edge can contribute to several result edges.
type CompositionEdgeProvenance struct {
	LeftEdge  int
	RightEdge int
}

// CompositionResult contains an independently owned composition graph, exact
// operand vertex mappings, and one exact provenance pair per result edge.
// Every slice is non-nil, Go-owned, and survives closure of all graphs.
type CompositionResult struct {
	Graph         *Graph
	LeftVertices  IDMapping
	RightVertices IDMapping
	Edges         []CompositionEdgeProvenance
}

// DisjointUnion returns the disjoint union of g followed by other. Vertex and
// edge order within each operand is preserved; IDs from other are offset by
// the vertex and edge counts of g. Operands are borrowed only for the call.
// They must have the same directedness. Loops and parallel edges are preserved.
// The result graph is independently owned and must be closed by the caller.
// Same-typed vertex and edge attributes are copied in operand order. A policy
// is required for same-name graph attributes and for a same-name element type
// conflict, which can only be resolved by dropping that attribute.
//
//igraph:bind igraph_disjoint_union
func (g *Graph) DisjointUnion(other *Graph, attributes *GraphOperatorAttributePolicy) (BinaryGraphOperatorResult, error) {
	return binaryDerivedGraph(g, other, func(left, right *C.igraph_t) (BinaryGraphOperatorResult, error) {
		leftAttributes, err := snapshotGraphAttributesLocked(left)
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		rightAttributes, err := snapshotGraphAttributesLocked(right)
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		leftVertices, leftEdges, err := graphCounts(left, "left disjoint-union operand")
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		rightVertices, rightEdges, err := graphCounts(right, "right disjoint-union operand")
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		result, err := collectBinaryGraphResult(func() (*Graph, error) {
			var value C.igraph_t
			if code := C.go_igraph_disjoint_union(&value, left, right); code != C.IGRAPH_SUCCESS {
				return nil, igraphError("create disjoint union", int(code))
			}
			return adoptInitializedGraph(&value), nil
		}, func(graph *Graph) (GraphIDMapping, GraphIDMapping, error) {
			leftMapping, err := offsetGraphIDMapping(leftVertices, leftEdges, leftVertices+rightVertices, leftEdges+rightEdges, 0, 0)
			if err != nil {
				return GraphIDMapping{}, GraphIDMapping{}, err
			}
			rightMapping, err := offsetGraphIDMapping(rightVertices, rightEdges, leftVertices+rightVertices, leftEdges+rightEdges, leftVertices, leftEdges)
			return leftMapping, rightMapping, err
		})
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		if err := restoreBinaryOperatorAttributes(result.Graph, leftAttributes, rightAttributes, result.Left, result.Right, attributes); err != nil {
			_ = result.Graph.Close()
			return BinaryGraphOperatorResult{}, err
		}
		return result, nil
	})
}

// Union returns the graph union. Result vertex count is the larger operand
// vertex count, and parallel-edge multiplicity is the larger operand
// multiplicity for each endpoint pair. Upstream exact edge mappings are
// returned for both operands. Directedness must match. Operands are borrowed
// only for the call; the independently owned result must be closed. Attributes
// with one contributing value are preserved; same-name values that map to one
// result element require an explicit scope-specific policy.
//
//igraph:bind igraph_union
func (g *Graph) Union(other *Graph, attributes *GraphOperatorAttributePolicy) (BinaryGraphOperatorResult, error) {
	return g.mappedMerge(other, mappedMergeUnion, attributes)
}

// Intersection returns the common edges of both graphs. Result vertex count is
// the larger operand vertex count, and parallel-edge multiplicity is the
// smaller operand multiplicity for each endpoint pair. Upstream exact
// provenance is converted to source-to-result mappings. Directedness must
// match. Operands are borrowed only for the call; the independently owned
// result must be closed. Attributes follow the same explicit conflict-policy
// contract as Union.
//
//igraph:bind igraph_intersection
func (g *Graph) Intersection(other *Graph, attributes *GraphOperatorAttributePolicy) (BinaryGraphOperatorResult, error) {
	return g.mappedMerge(other, mappedMergeIntersection, attributes)
}

type mappedMergeMode uint8

const (
	mappedMergeUnion mappedMergeMode = iota
	mappedMergeIntersection
)

func (g *Graph) mappedMerge(other *Graph, mode mappedMergeMode, attributes *GraphOperatorAttributePolicy) (BinaryGraphOperatorResult, error) {
	return binaryDerivedGraph(g, other, func(left, right *C.igraph_t) (BinaryGraphOperatorResult, error) {
		leftAttributes, err := snapshotGraphAttributesLocked(left)
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		rightAttributes, err := snapshotGraphAttributesLocked(right)
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		leftVertices, leftEdges, err := graphCounts(left, "left merge operand")
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		rightVertices, rightEdges, err := graphCounts(right, "right merge operand")
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		result, err := collectMappedMerge(leftVertices, leftEdges, rightVertices, rightEdges, mode, mappedMergeOperations{
			newVector:   func() (*intVector, error) { return newIntVector(nil) },
			closeVector: (*intVector).close,
			query: func(leftMap, rightMap *intVector) (*Graph, int, int, error) {
				var value C.igraph_t
				var code C.igraph_error_t
				if mode == mappedMergeUnion {
					code = C.go_igraph_union(&value, left, right, &leftMap.value, &rightMap.value)
				} else {
					code = C.go_igraph_intersection(&value, left, right, &leftMap.value, &rightMap.value)
				}
				if code != C.IGRAPH_SUCCESS {
					return nil, 0, 0, igraphError("merge graphs", int(code))
				}
				vertices, edges, err := graphCounts(&value, "merge result")
				if err != nil {
					C.igraph_destroy(&value)
					return nil, 0, 0, err
				}
				return adoptInitializedGraph(&value), vertices, edges, nil
			},
			vectorSlice: func(vector *intVector) ([]int, error) { return vector.slice() },
			closeGraph:  func(graph *Graph) { _ = graph.Close() },
		})
		if err != nil {
			return BinaryGraphOperatorResult{}, err
		}
		if err := restoreBinaryOperatorAttributes(result.Graph, leftAttributes, rightAttributes, result.Left, result.Right, attributes); err != nil {
			_ = result.Graph.Close()
			return BinaryGraphOperatorResult{}, err
		}
		return result, nil
	})
}

// Difference returns edges of g that are not present in other. Vertex count
// and IDs come from g, and directedness must match. Parallel multiplicities are
// subtracted pairwise. Left vertex provenance is identity. Because upstream
// exposes no edge map, endpoint-equivalent source and result edges are paired
// in ascending edge-ID order; excess left parallel edges map to RemovedID.
// This is deterministic structural correspondence, not attribute provenance.
// Operands are borrowed only for the call; the independently owned result must
// be closed. Graph and vertex attributes and retained left-edge attributes are
// preserved from g; attributes from other do not contribute to the result.
//
//igraph:bind igraph_difference
func (g *Graph) Difference(other *Graph) (DifferenceResult, error) {
	var result DifferenceResult
	err := withLockedGraphs([]*Graph{g, other}, func(values []*C.igraph_t) error {
		leftVertices, _, err := graphCounts(values[0], "difference left operand")
		if err != nil {
			return err
		}
		leftEdges, err := edgeSlice(values[0])
		if err != nil {
			return err
		}
		directed := C.igraph_is_directed(values[0]) != booltoint(false)
		result, err = collectDifference(leftVertices, leftEdges, directed, differenceOperations{
			query: func() (*Graph, error) {
				var value C.igraph_t
				if code := C.go_igraph_difference(&value, values[0], values[1]); code != C.IGRAPH_SUCCESS {
					return nil, igraphError("create graph difference", int(code))
				}
				return adoptInitializedGraph(&value), nil
			},
			resultEdges:   func(graph *Graph) ([]Edge, error) { return graph.Edges() },
			vertexMapping: identityIDMapping,
			edgeMapping:   structuralDifferenceEdgeMapping,
			closeGraph:    func(graph *Graph) { _ = graph.Close() },
		})
		return err
	})
	return result, err
}

// Compose returns the relational composition of g followed by other. A result
// edge i->j is produced for every pair g(i,k), other(k,j), preserving loop and
// parallel multiplicity. Directedness must match. Edges records upstream's
// exact contributing operand edge IDs for every result edge; it is not reduced
// to IDMapping because operand-to-result provenance can be one-to-many.
// Operands are borrowed only for the call; the independently owned result must
// be closed. Non-conflicting attributes are preserved; graph, vertex, or edge
// values contributed by both operands require the corresponding policy.
//
//igraph:bind igraph_compose
func (g *Graph) Compose(other *Graph, attributes *GraphOperatorAttributePolicy) (CompositionResult, error) {
	var result CompositionResult
	err := withLockedGraphs([]*Graph{g, other}, func(values []*C.igraph_t) error {
		leftAttributes, err := snapshotGraphAttributesLocked(values[0])
		if err != nil {
			return err
		}
		rightAttributes, err := snapshotGraphAttributesLocked(values[1])
		if err != nil {
			return err
		}
		leftVertices, leftEdges, err := graphCounts(values[0], "composition left operand")
		if err != nil {
			return err
		}
		rightVertices, rightEdges, err := graphCounts(values[1], "composition right operand")
		if err != nil {
			return err
		}
		result, err = collectComposition(leftVertices, leftEdges, rightVertices, rightEdges, compositionOperations{
			newVector:   func() (*intVector, error) { return newIntVector(nil) },
			closeVector: (*intVector).close,
			query: func(leftMap, rightMap *intVector) (*Graph, int, int, error) {
				var value C.igraph_t
				if code := C.go_igraph_compose(&value, values[0], values[1], &leftMap.value, &rightMap.value); code != C.IGRAPH_SUCCESS {
					return nil, 0, 0, igraphError("compose graphs", int(code))
				}
				vertices, edges, err := graphCounts(&value, "composition result")
				if err != nil {
					C.igraph_destroy(&value)
					return nil, 0, 0, err
				}
				return adoptInitializedGraph(&value), vertices, edges, nil
			},
			vectorSlice: func(vector *intVector) ([]int, error) { return vector.slice() },
			closeGraph:  func(graph *Graph) { _ = graph.Close() },
		})
		if err != nil {
			return err
		}
		if err := restoreCompositionAttributes(result.Graph, leftAttributes, rightAttributes, result, attributes); err != nil {
			_ = result.Graph.Close()
			result = CompositionResult{}
			return err
		}
		return nil
	})
	return result, err
}

// Complement returns the graph complement. IncludeLoops controls whether
// missing self-loops are included. Vertex IDs are unchanged. This binding is
// intentionally limited to simple-graph inputs: parallel edges are rejected
// because upstream complement traversal does not provide reliable multigraph
// presence semantics. Loops in an otherwise simple input are supported.
// No edge mapping is exposed because complement edges have no source edge
// correspondence. The input is borrowed only for the call; the independently
// owned result must be closed.
//
//igraph:bind igraph_complementer
func (g *Graph) Complement(includeLoops bool) (VertexMappedGraphResult, error) {
	var result VertexMappedGraphResult
	err := withLockedGraphs([]*Graph{g}, func(values []*C.igraph_t) error {
		vertices, _, err := graphCounts(values[0], "complement operand")
		if err != nil {
			return err
		}
		hasMultiple, err := operatorGraphHasMultiple(values[0])
		if err != nil {
			return err
		}
		if hasMultiple {
			return errors.New("igraph: complement requires a graph without parallel edges")
		}
		graph, err := collectOwnedOperatorGraph(func() (*Graph, error) {
			var value C.igraph_t
			if code := C.go_igraph_complementer(&value, values[0], booltoint(includeLoops)); code != C.IGRAPH_SUCCESS {
				return nil, igraphError("create graph complement", int(code))
			}
			return adoptInitializedGraph(&value), nil
		})
		if err != nil {
			return err
		}
		mapping, err := identityIDMapping(vertices)
		if err != nil {
			_ = graph.Close()
			return err
		}
		result = VertexMappedGraphResult{Graph: graph, Vertices: mapping}
		return nil
	})
	return result, err
}

//igraph:internal igraph_has_multiple
func operatorGraphHasMultiple(graph *C.igraph_t) (bool, error) {
	var result C.igraph_bool_t
	if code := C.go_igraph_has_multiple(graph, &result); code != C.IGRAPH_SUCCESS {
		return false, igraphError("check complement parallel edges", int(code))
	}
	return result != booltoint(false), nil
}

func binaryDerivedGraph(
	left, right *Graph,
	operation func(*C.igraph_t, *C.igraph_t) (BinaryGraphOperatorResult, error),
) (BinaryGraphOperatorResult, error) {
	var result BinaryGraphOperatorResult
	err := withLockedGraphs([]*Graph{left, right}, func(values []*C.igraph_t) error {
		var err error
		result, err = operation(values[0], values[1])
		return err
	})
	return result, err
}

func graphCounts(graph *C.igraph_t, context string) (int, int, error) {
	vertices, err := igraphIntToInt(C.igraph_vcount(graph), context+" vertex count")
	if err != nil {
		return 0, 0, err
	}
	edges, err := igraphIntToInt(C.igraph_ecount(graph), context+" edge count")
	return vertices, edges, err
}

func collectOwnedOperatorGraph(query func() (*Graph, error)) (*Graph, error) {
	graph, err := query()
	if err != nil {
		if graph != nil {
			_ = graph.Close()
		}
		return nil, err
	}
	if graph == nil {
		return nil, errors.New("igraph: graph operator returned a nil graph")
	}
	return graph, nil
}

func collectBinaryGraphResult(
	query func() (*Graph, error),
	mappings func(*Graph) (GraphIDMapping, GraphIDMapping, error),
) (BinaryGraphOperatorResult, error) {
	graph, err := collectOwnedOperatorGraph(query)
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	left, right, err := mappings(graph)
	if err != nil {
		_ = graph.Close()
		return BinaryGraphOperatorResult{}, err
	}
	return BinaryGraphOperatorResult{Graph: graph, Left: left, Right: right}, nil
}

type differenceOperations struct {
	query         func() (*Graph, error)
	resultEdges   func(*Graph) ([]Edge, error)
	vertexMapping func(int) (IDMapping, error)
	edgeMapping   func([]Edge, []Edge, bool) (IDMapping, error)
	closeGraph    func(*Graph)
}

func collectDifference(
	leftVertices int,
	leftEdges []Edge,
	directed bool,
	operations differenceOperations,
) (result DifferenceResult, err error) {
	graph, err := operations.query()
	if err != nil {
		if graph != nil {
			operations.closeGraph(graph)
		}
		return DifferenceResult{}, err
	}
	if graph == nil {
		return DifferenceResult{}, errors.New("igraph: difference returned a nil graph")
	}
	succeeded := false
	defer func() {
		if !succeeded {
			operations.closeGraph(graph)
		}
	}()
	resultEdges, err := operations.resultEdges(graph)
	if err != nil {
		return DifferenceResult{}, err
	}
	vertices, err := operations.vertexMapping(leftVertices)
	if err != nil {
		return DifferenceResult{}, err
	}
	edges, err := operations.edgeMapping(leftEdges, resultEdges, directed)
	if err != nil {
		return DifferenceResult{}, err
	}
	result = DifferenceResult{
		Graph: graph,
		Left:  GraphIDMapping{Vertices: vertices, Edges: edges},
	}
	succeeded = true
	return result, nil
}

func structuralDifferenceEdgeMapping(source, result []Edge, directed bool) (IDMapping, error) {
	resultBuckets := make(map[edgeEndpointKey][]int)
	for resultID, edge := range result {
		resultBuckets[endpointKey(edge, directed)] = append(
			resultBuckets[endpointKey(edge, directed)], resultID,
		)
	}
	positions := make(map[edgeEndpointKey]int, len(resultBuckets))
	oldToNew := make([]int, len(source))
	matched := 0
	for sourceID, edge := range source {
		oldToNew[sourceID] = RemovedID
		key := endpointKey(edge, directed)
		position := positions[key]
		if position < len(resultBuckets[key]) {
			oldToNew[sourceID] = resultBuckets[key][position]
			positions[key] = position + 1
			matched++
		}
	}
	if matched != len(result) {
		return IDMapping{}, errors.New("igraph: difference result contains an edge absent from the left operand")
	}
	return newIDMapping(oldToNew, len(result))
}

func offsetGraphIDMapping(oldVertices, oldEdges, newVertices, newEdges, vertexOffset, edgeOffset int) (GraphIDMapping, error) {
	vertices, err := offsetIDMapping(oldVertices, newVertices, vertexOffset)
	if err != nil {
		return GraphIDMapping{}, err
	}
	edges, err := offsetIDMapping(oldEdges, newEdges, edgeOffset)
	return GraphIDMapping{Vertices: vertices, Edges: edges}, err
}

func offsetIDMapping(oldCount, newCount, offset int) (IDMapping, error) {
	oldToNew := make([]int, oldCount)
	for oldID := range oldToNew {
		oldToNew[oldID] = oldID + offset
	}
	return newIDMapping(oldToNew, newCount)
}

func mappingFromExactInverse(oldCount int, inverse []int) (IDMapping, error) {
	oldToNew := make([]int, oldCount)
	for oldID := range oldToNew {
		oldToNew[oldID] = RemovedID
	}
	for newID, oldID := range inverse {
		if oldID < 0 || oldID >= oldCount {
			return IDMapping{}, fmt.Errorf("igraph: provenance ID %d out of range [0, %d)", oldID, oldCount)
		}
		if oldToNew[oldID] != RemovedID {
			return IDMapping{}, fmt.Errorf("igraph: provenance repeats source ID %d", oldID)
		}
		oldToNew[oldID] = newID
	}
	mapping, err := newIDMapping(oldToNew, len(inverse))
	if err != nil {
		return IDMapping{}, err
	}
	if !equalIntSlices(mapping.NewToOld, inverse) {
		return IDMapping{}, errors.New("igraph: inverse provenance is inconsistent")
	}
	return mapping, nil
}

type mappedMergeOperations struct {
	newVector   func() (*intVector, error)
	closeVector func(*intVector)
	query       func(*intVector, *intVector) (*Graph, int, int, error)
	vectorSlice func(*intVector) ([]int, error)
	closeGraph  func(*Graph)
}

func collectMappedMerge(
	leftVertices, leftEdges, rightVertices, rightEdges int,
	mode mappedMergeMode,
	operations mappedMergeOperations,
) (result BinaryGraphOperatorResult, err error) {
	leftVector, err := operations.newVector()
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	defer operations.closeVector(leftVector)
	rightVector, err := operations.newVector()
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	defer operations.closeVector(rightVector)
	graph, resultVertices, resultEdges, err := operations.query(leftVector, rightVector)
	if err != nil {
		if graph != nil {
			operations.closeGraph(graph)
		}
		return BinaryGraphOperatorResult{}, err
	}
	if graph == nil {
		return BinaryGraphOperatorResult{}, errors.New("igraph: mapped graph operator returned a nil graph")
	}
	succeeded := false
	defer func() {
		if !succeeded {
			operations.closeGraph(graph)
		}
	}()
	leftValues, err := operations.vectorSlice(leftVector)
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	rightValues, err := operations.vectorSlice(rightVector)
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	leftVertexMap, err := offsetIDMapping(leftVertices, resultVertices, 0)
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	rightVertexMap, err := offsetIDMapping(rightVertices, resultVertices, 0)
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	var leftEdgeMap, rightEdgeMap IDMapping
	if mode == mappedMergeUnion {
		if len(leftValues) != leftEdges || len(rightValues) != rightEdges {
			return BinaryGraphOperatorResult{}, errors.New("igraph: union edge mapping length is inconsistent")
		}
		leftEdgeMap, err = newIDMapping(leftValues, resultEdges)
		if err == nil {
			rightEdgeMap, err = newIDMapping(rightValues, resultEdges)
		}
	} else {
		if len(leftValues) != resultEdges || len(rightValues) != resultEdges {
			return BinaryGraphOperatorResult{}, errors.New("igraph: intersection inverse mapping length is inconsistent")
		}
		leftEdgeMap, err = mappingFromExactInverse(leftEdges, leftValues)
		if err == nil {
			rightEdgeMap, err = mappingFromExactInverse(rightEdges, rightValues)
		}
	}
	if err != nil {
		return BinaryGraphOperatorResult{}, err
	}
	result = BinaryGraphOperatorResult{
		Graph: graph,
		Left:  GraphIDMapping{Vertices: leftVertexMap, Edges: leftEdgeMap},
		Right: GraphIDMapping{Vertices: rightVertexMap, Edges: rightEdgeMap},
	}
	succeeded = true
	return result, nil
}

type compositionOperations struct {
	newVector   func() (*intVector, error)
	closeVector func(*intVector)
	query       func(*intVector, *intVector) (*Graph, int, int, error)
	vectorSlice func(*intVector) ([]int, error)
	closeGraph  func(*Graph)
}

func collectComposition(
	leftVertices, leftEdges, rightVertices, rightEdges int,
	operations compositionOperations,
) (result CompositionResult, err error) {
	leftVector, err := operations.newVector()
	if err != nil {
		return CompositionResult{}, err
	}
	defer operations.closeVector(leftVector)
	rightVector, err := operations.newVector()
	if err != nil {
		return CompositionResult{}, err
	}
	defer operations.closeVector(rightVector)
	graph, resultVertices, resultEdges, err := operations.query(leftVector, rightVector)
	if err != nil {
		if graph != nil {
			operations.closeGraph(graph)
		}
		return CompositionResult{}, err
	}
	if graph == nil {
		return CompositionResult{}, errors.New("igraph: composition returned a nil graph")
	}
	succeeded := false
	defer func() {
		if !succeeded {
			operations.closeGraph(graph)
		}
	}()
	leftValues, err := operations.vectorSlice(leftVector)
	if err != nil {
		return CompositionResult{}, err
	}
	rightValues, err := operations.vectorSlice(rightVector)
	if err != nil {
		return CompositionResult{}, err
	}
	if len(leftValues) != resultEdges || len(rightValues) != resultEdges {
		return CompositionResult{}, errors.New("igraph: composition provenance length is inconsistent")
	}
	leftVertexMap, err := offsetIDMapping(leftVertices, resultVertices, 0)
	if err != nil {
		return CompositionResult{}, err
	}
	rightVertexMap, err := offsetIDMapping(rightVertices, resultVertices, 0)
	if err != nil {
		return CompositionResult{}, err
	}
	edges := make([]CompositionEdgeProvenance, resultEdges)
	for edgeID := range edges {
		if leftValues[edgeID] < 0 || leftValues[edgeID] >= leftEdges {
			return CompositionResult{}, fmt.Errorf(
				"igraph: composition left edge provenance %d out of range [0, %d)",
				leftValues[edgeID], leftEdges,
			)
		}
		if rightValues[edgeID] < 0 || rightValues[edgeID] >= rightEdges {
			return CompositionResult{}, fmt.Errorf(
				"igraph: composition right edge provenance %d out of range [0, %d)",
				rightValues[edgeID], rightEdges,
			)
		}
		edges[edgeID] = CompositionEdgeProvenance{LeftEdge: leftValues[edgeID], RightEdge: rightValues[edgeID]}
	}
	result = CompositionResult{Graph: graph, LeftVertices: leftVertexMap, RightVertices: rightVertexMap, Edges: edges}
	succeeded = true
	return result, nil
}
