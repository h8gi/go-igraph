package igraph

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestGraphPowerDistanceDirectionAndSimplicity(t *testing.T) {
	undirected := testGraphFromEdges(t, 4, []Edge{{0, 1}, {1, 2}, {2, 3}}, false)
	defer undirected.Close()
	power, err := undirected.GraphPower(2, true)
	if err != nil {
		t.Fatal(err)
	}
	defer power.Graph.Close()
	assertGraphShape(t, power.Graph, 4, 5, false)
	assertEdgesEqual(t, power.Graph, []Edge{{0, 1}, {1, 2}, {2, 3}, {0, 2}, {1, 3}})
	wantVertices, _ := identityIDMapping(4)
	if !reflect.DeepEqual(power.Vertices, wantVertices) {
		t.Errorf("vertex mapping = %#v, want %#v", power.Vertices, wantVertices)
	}

	directed := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer directed.Close()
	directedPower, err := directed.GraphPower(2, true)
	if err != nil {
		t.Fatal(err)
	}
	defer directedPower.Graph.Close()
	assertGraphShape(t, directedPower.Graph, 3, 3, true)
	assertEdgesEqual(t, directedPower.Graph, []Edge{{0, 1}, {1, 2}, {0, 2}})
	undirectedPower, err := directed.GraphPower(2, false)
	if err != nil {
		t.Fatal(err)
	}
	defer undirectedPower.Graph.Close()
	assertGraphShape(t, undirectedPower.Graph, 3, 3, false)
	assertEdgesEqual(t, undirectedPower.Graph, []Edge{{0, 1}, {1, 2}, {0, 2}})

	multiple := testGraphFromEdges(t, 2, []Edge{{0, 0}, {0, 1}, {0, 1}}, false)
	defer multiple.Close()
	first, err := multiple.GraphPower(1, true)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Graph.Close()
	assertEdgesEqual(t, first.Graph, []Edge{{0, 1}})
	zero, err := multiple.GraphPower(0, true)
	if err != nil {
		t.Fatal(err)
	}
	defer zero.Graph.Close()
	assertGraphShape(t, zero.Graph, 2, 0, false)
}

func TestGraphPowerAttributesOwnershipEmptyAndClosed(t *testing.T) {
	graph := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	if err := graph.SetGraphStringAttribute("name", "path"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexNumericAttributes("score", []float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeBooleanAttributes("old", []bool{true, false}); err != nil {
		t.Fatal(err)
	}
	result, err := graph.GraphPower(2, true)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Graph.GraphStringAttribute("name"); got != "path" {
		t.Errorf("graph attribute = %q", got)
	}
	if got, _ := result.Graph.VertexNumericAttributes("score"); !reflect.DeepEqual(got, []float64{1, 2, 3}) {
		t.Errorf("vertex attributes = %v", got)
	}
	if metadata, _ := result.Graph.EdgeAttributes(); len(metadata) != 0 {
		t.Errorf("power retained edge attributes: %v", metadata)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, result.Graph, 3, 3, true)
	if err := result.Graph.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Vertices.OldToNew, []int{0, 1, 2}) {
		t.Errorf("mapping after graph closure = %#v", result.Vertices)
	}

	empty := testGraphFromEdges(t, 0, nil, false)
	emptyPower, err := empty.GraphPower(3, true)
	if err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, emptyPower.Graph, 0, 0, false)
	if emptyPower.Vertices.OldToNew == nil || emptyPower.Vertices.NewToOld == nil {
		t.Errorf("empty mapping = %#v, want non-nil slices", emptyPower.Vertices)
	}
	_ = emptyPower.Graph.Close()
	_ = empty.Close()
	if _, err := empty.GraphPower(1, true); !errors.Is(err, ErrClosed) {
		t.Errorf("closed GraphPower error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.GraphPower(1, true); !errors.Is(err, ErrClosed) {
		t.Errorf("nil GraphPower error = %v", err)
	}
	open := testGraphFromEdges(t, 1, nil, false)
	defer open.Close()
	if result, err := open.GraphPower(-1, true); err == nil || result.Graph != nil {
		t.Errorf("negative GraphPower = %#v, %v", result, err)
	}
}

func TestGraphPowerDisconnectedSingletonConcurrentAndCloseRace(t *testing.T) {
	disconnected := testGraphFromEdges(t, 5, []Edge{{0, 1}, {1, 2}}, false)
	result, err := disconnected.GraphPower(3, true)
	if err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, result.Graph, 5, 3, false)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 1}, {1, 2}, {0, 2}})
	_ = result.Graph.Close()

	singleton := testGraphFromEdges(t, 1, nil, true)
	singlePower, err := singleton.GraphPower(4, true)
	if err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, singlePower.Graph, 1, 0, true)
	_ = singlePower.Graph.Close()
	_ = singleton.Close()

	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			power, err := disconnected.GraphPower(2, true)
			if power.Graph != nil {
				_ = power.Graph.Close()
			}
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("GraphPower close-race error = %v", err)
			}
		}()
	}
	wait.Add(1)
	go func() { defer wait.Done(); _ = disconnected.Close() }()
	wait.Wait()
}

func TestConnectNeighborhoodDirectionMappingAndAttributes(t *testing.T) {
	graph := testGraphFromEdges(t, 3, []Edge{{0, 1}, {0, 1}, {1, 2}}, true)
	defer graph.Close()
	if err := graph.SetGraphStringAttribute("name", "path"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttributes("active", []bool{true, false, true}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("weight", []float64{2, 3, 5}); err != nil {
		t.Fatal(err)
	}
	result, err := graph.ConnectNeighborhoodInPlace(2, DirectionOut)
	if err != nil {
		t.Fatal(err)
	}
	assertEdgesEqual(t, graph, []Edge{{0, 1}, {0, 1}, {1, 2}, {0, 2}})
	if !reflect.DeepEqual(result.Mapping.Edges.OldToNew, []int{0, 1, 2}) ||
		!reflect.DeepEqual(result.Mapping.Edges.NewToOld, []int{0, 1, 2, RemovedID}) {
		t.Errorf("edge mapping = %#v", result.Mapping.Edges)
	}
	if !result.EdgeMappingAvailable {
		t.Error("edge mapping marked unavailable")
	}
	if got, _ := graph.GraphStringAttribute("name"); got != "path" {
		t.Errorf("graph attribute = %q", got)
	}
	if got, _ := graph.VertexBooleanAttributes("active"); !reflect.DeepEqual(got, []bool{true, false, true}) {
		t.Errorf("vertex attributes = %v", got)
	}
	weights, err := graph.EdgeNumericAttributes("weight")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(weights[:3], []float64{2, 3, 5}) || !math.IsNaN(weights[3]) {
		t.Errorf("edge weights = %v, want [2 3 5 NaN]", weights)
	}

	incoming := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer incoming.Close()
	if _, err := incoming.ConnectNeighborhoodInPlace(2, DirectionIn); err != nil {
		t.Fatal(err)
	}
	assertEdgesEqual(t, incoming, []Edge{{0, 1}, {1, 2}, {0, 2}})
	all := testGraphFromEdges(t, 3, []Edge{{0, 1}, {2, 1}}, true)
	defer all.Close()
	if _, err := all.ConnectNeighborhoodInPlace(2, DirectionAll); err != nil {
		t.Fatal(err)
	}
	assertEdgesEqual(t, all, []Edge{{0, 1}, {2, 1}, {0, 2}})
}

func TestConnectNeighborhoodNoOpValidationAtomicityAndCloseRace(t *testing.T) {
	graph := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	defer graph.Close()
	identity, err := graph.ConnectNeighborhoodInPlace(0, DirectionAll)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(identity.Mapping.Edges.OldToNew, []int{0, 1}) {
		t.Errorf("zero-order mapping = %#v", identity.Mapping)
	}
	if _, err := graph.ConnectNeighborhoodInPlace(-1, DirectionAll); err == nil {
		t.Error("negative order was accepted")
	}
	if _, err := graph.ConnectNeighborhoodInPlace(1, DirectionMode(255)); err == nil {
		t.Error("invalid direction was accepted")
	}
	if _, err := graph.ConnectNeighborhoodInPlace(2, DirectionAll); err != nil {
		t.Fatal(err)
	}
	assertEdgesEqual(t, graph, []Edge{{0, 1}, {1, 2}, {0, 2}})
	singleton := testGraphFromEdges(t, 1, nil, false)
	if _, err := singleton.ConnectNeighborhoodInPlace(3, DirectionAll); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, singleton, 1, 0, false)
	_ = singleton.Close()

	atomic := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, false)
	defer atomic.Close()
	before, _ := atomic.Edges()
	for _, stage := range []graphTransformationStage{
		graphTransformationAtClone,
		graphTransformationAtTransform,
		graphTransformationAfterTransform,
	} {
		_, err := atomic.connectNeighborhoodInPlace(2, DirectionAll, func(actual graphTransformationStage) error {
			if actual == stage {
				return errors.New("forced failure")
			}
			return nil
		})
		if err == nil {
			t.Errorf("stage %d failure returned nil", stage)
		}
		after, _ := atomic.Edges()
		if !reflect.DeepEqual(after, before) {
			t.Fatalf("stage %d failure changed graph: %v", stage, after)
		}
	}

	racing := testGraphFromEdges(t, 5, []Edge{{0, 1}, {1, 2}, {2, 3}, {3, 4}}, false)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := racing.ConnectNeighborhoodInPlace(2, DirectionAll)
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("close-race error = %v", err)
			}
		}()
	}
	wait.Add(1)
	go func() { defer wait.Done(); _ = racing.Close() }()
	wait.Wait()
	if _, err := racing.ConnectNeighborhoodInPlace(1, DirectionAll); !errors.Is(err, ErrClosed) {
		t.Errorf("closed ConnectNeighborhoodInPlace error = %v", err)
	}
	var nilGraph *Graph
	if _, err := nilGraph.ConnectNeighborhoodInPlace(1, DirectionAll); !errors.Is(err, ErrClosed) {
		t.Errorf("nil ConnectNeighborhoodInPlace error = %v", err)
	}
}

func TestPrefixEdgeMappingRejectsChangedPrefix(t *testing.T) {
	if _, err := prefixEdgeMapping([]Edge{{0, 1}}, nil, true); err == nil {
		t.Error("decreased edge count was accepted")
	}
	if _, err := prefixEdgeMapping([]Edge{{0, 1}}, []Edge{{1, 0}}, true); err == nil {
		t.Error("changed edge prefix was accepted")
	}
}
