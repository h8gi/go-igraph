package igraph

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDisjointUnionPreservesOrderMultiplicityAndMappings(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {0, 1}, {2, 2}}, true)
	right := testGraphFromEdges(t, 2, []Edge{{1, 0}, {1, 1}}, true)
	result, err := left.DisjointUnion(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 5, 5, true)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 1}, {0, 1}, {2, 2}, {4, 3}, {4, 4}})
	wantLeft := GraphIDMapping{
		Vertices: IDMapping{OldToNew: []int{0, 1, 2}, NewToOld: []int{0, 1, 2, RemovedID, RemovedID}},
		Edges:    IDMapping{OldToNew: []int{0, 1, 2}, NewToOld: []int{0, 1, 2, RemovedID, RemovedID}},
	}
	wantRight := GraphIDMapping{
		Vertices: IDMapping{OldToNew: []int{3, 4}, NewToOld: []int{RemovedID, RemovedID, RemovedID, 0, 1}},
		Edges:    IDMapping{OldToNew: []int{3, 4}, NewToOld: []int{RemovedID, RemovedID, RemovedID, 0, 1}},
	}
	if !reflect.DeepEqual(result.Left, wantLeft) || !reflect.DeepEqual(result.Right, wantRight) {
		t.Errorf("mappings = left %#v right %#v, want left %#v right %#v", result.Left, result.Right, wantLeft, wantRight)
	}
}

func TestUnionAndIntersectionMappingsLoopsParallelAndOperandOrder(t *testing.T) {
	left := testGraphFromEdges(t, 4, []Edge{{0, 1}, {0, 1}, {1, 1}, {3, 0}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}, {1, 1}, {2, 0}}, true)

	union, err := left.Union(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = union.Graph.Close() })
	assertGraphShape(t, union.Graph, 4, 6, true)
	assertEdgesEqual(t, union.Graph, []Edge{{0, 1}, {0, 1}, {1, 1}, {1, 1}, {2, 0}, {3, 0}})
	assertOperatorMappingConsistent(t, left, union.Graph, union.Left)
	assertOperatorMappingConsistent(t, right, union.Graph, union.Right)

	intersection, err := left.Intersection(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = intersection.Graph.Close() })
	assertGraphShape(t, intersection.Graph, 4, 2, true)
	assertEdgesEqual(t, intersection.Graph, []Edge{{0, 1}, {1, 1}})
	assertOperatorMappingConsistent(t, left, intersection.Graph, intersection.Left)
	assertOperatorMappingConsistent(t, right, intersection.Graph, intersection.Right)
	if !reflect.DeepEqual(intersection.Left.Edges.NewToOld, []int{0, 2}) ||
		!reflect.DeepEqual(intersection.Right.Edges.NewToOld, []int{0, 1}) {
		t.Errorf("intersection exact inverse maps = %v / %v", intersection.Left.Edges.NewToOld, intersection.Right.Edges.NewToOld)
	}

	reversed, err := right.Union(left)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reversed.Graph.Close() })
	assertEdgesEqual(t, reversed.Graph, []Edge{{0, 1}, {0, 1}, {1, 1}, {1, 1}, {2, 0}, {3, 0}})
	assertOperatorMappingConsistent(t, right, reversed.Graph, reversed.Left)
	assertOperatorMappingConsistent(t, left, reversed.Graph, reversed.Right)
}

func TestDifferenceIsOrderedAndComplementOptions(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {0, 1}, {1, 1}, {2, 0}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}, {1, 1}}, true)
	difference, err := left.Difference(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = difference.Graph.Close() })
	assertGraphShape(t, difference.Graph, 3, 2, true)
	assertEdgesEqual(t, difference.Graph, []Edge{{0, 1}, {2, 0}})
	wantIdentity, _ := identityIDMapping(3)
	if !reflect.DeepEqual(difference.Left.Vertices, wantIdentity) {
		t.Errorf("difference vertices = %#v, want %#v", difference.Left.Vertices, wantIdentity)
	}
	wantEdges := IDMapping{
		OldToNew: []int{0, RemovedID, RemovedID, 1},
		NewToOld: []int{0, 3},
	}
	if !reflect.DeepEqual(difference.Left.Edges, wantEdges) {
		t.Errorf("difference edges = %#v, want %#v", difference.Left.Edges, wantEdges)
	}
	reversed, err := right.Difference(left)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reversed.Graph.Close() })
	assertGraphShape(t, reversed.Graph, 3, 1, true)
	assertEdgesEqual(t, reversed.Graph, []Edge{{1, 1}})
	wantReversedEdges := IDMapping{
		OldToNew: []int{RemovedID, 0, RemovedID},
		NewToOld: []int{1},
	}
	if !reflect.DeepEqual(reversed.Left.Edges, wantReversedEdges) {
		t.Errorf("reversed difference edges = %#v, want %#v", reversed.Left.Edges, wantReversedEdges)
	}

	undirected := testGraphFromEdges(t, 3, []Edge{{0, 1}, {2, 2}}, false)
	withoutLoops, err := undirected.Complement(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = withoutLoops.Graph.Close() })
	assertEdgesEqual(t, withoutLoops.Graph, []Edge{{0, 2}, {1, 2}})
	withLoops, err := undirected.Complement(true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = withLoops.Graph.Close() })
	assertEdgesEqual(t, withLoops.Graph, []Edge{{0, 2}, {0, 0}, {1, 2}, {1, 1}})

	directedParallel := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}}, true)
	if result, err := directedParallel.Complement(true); err == nil || result.Graph != nil {
		t.Errorf("parallel Complement() = %#v, %v, want zero, error", result, err)
	}
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	directedComplement, err := directed.Complement(true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = directedComplement.Graph.Close() })
	assertEdgesEqual(t, directedComplement.Graph, []Edge{{0, 0}, {1, 1}, {1, 0}})
}

func TestComposeReturnsPerResultEdgeProvenance(t *testing.T) {
	left := testGraphFromEdges(t, 4, []Edge{{0, 1}, {0, 2}, {0, 2}, {2, 2}}, true)
	right := testGraphFromEdges(t, 4, []Edge{{1, 3}, {2, 3}, {2, 3}, {2, 2}}, true)
	result, err := left.Compose(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 4, 10, true)
	if len(result.Edges) != 10 {
		t.Fatalf("composition provenance length = %d, want 10", len(result.Edges))
	}
	resultEdges, err := result.Graph.Edges()
	if err != nil {
		t.Fatal(err)
	}
	leftEdges, _ := left.Edges()
	rightEdges, _ := right.Edges()
	for edgeID, provenance := range result.Edges {
		leftEdge := leftEdges[provenance.LeftEdge]
		rightEdge := rightEdges[provenance.RightEdge]
		if leftEdge.To != rightEdge.From {
			t.Errorf("edge %d provenance does not join: %#v then %#v", edgeID, leftEdge, rightEdge)
		}
		if want := (Edge{From: leftEdge.From, To: rightEdge.To}); resultEdges[edgeID] != want {
			t.Errorf("edge %d = %#v, provenance produces %#v", edgeID, resultEdges[edgeID], want)
		}
	}
	wantIdentity, _ := identityIDMapping(4)
	if !reflect.DeepEqual(result.LeftVertices, wantIdentity) || !reflect.DeepEqual(result.RightVertices, wantIdentity) {
		t.Errorf("composition vertex mappings = %#v / %#v", result.LeftVertices, result.RightVertices)
	}
}

func TestOperatorsEmptySelfOwnershipAndClosed(t *testing.T) {
	empty := testGraphFromEdges(t, 0, nil, false)
	self := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}, {1, 1}}, false)

	disjoint, err := empty.DisjointUnion(empty)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = disjoint.Graph.Close() })
	assertGraphShape(t, disjoint.Graph, 0, 0, false)
	assertNonNilBinaryMappings(t, disjoint)
	for name, call := range map[string]func() (*Graph, error){
		"union": func() (*Graph, error) {
			result, err := empty.Union(empty)
			return result.Graph, err
		},
		"intersection": func() (*Graph, error) {
			result, err := empty.Intersection(empty)
			return result.Graph, err
		},
		"difference": func() (*Graph, error) {
			result, err := empty.Difference(empty)
			return result.Graph, err
		},
		"composition": func() (*Graph, error) {
			result, err := empty.Compose(empty)
			if result.Edges == nil {
				t.Error("empty composition provenance is nil")
			}
			return result.Graph, err
		},
		"complement": func() (*Graph, error) {
			result, err := empty.Complement(true)
			return result.Graph, err
		},
	} {
		graph, err := call()
		if err != nil {
			t.Errorf("empty %s error = %v", name, err)
			continue
		}
		t.Cleanup(func() { _ = graph.Close() })
		assertGraphShape(t, graph, 0, 0, false)
	}

	union, err := self.Union(self)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = union.Graph.Close() })
	assertGraphShape(t, union.Graph, 2, 3, false)
	intersection, err := self.Intersection(self)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = intersection.Graph.Close() })
	assertGraphShape(t, intersection.Graph, 2, 3, false)
	difference, err := self.Difference(self)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = difference.Graph.Close() })
	assertGraphShape(t, difference.Graph, 2, 0, false)
	composition, err := self.Compose(self)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = composition.Graph.Close() })
	if composition.Edges == nil {
		t.Fatal("composition provenance is nil")
	}

	if err := self.Close(); err != nil {
		t.Fatal(err)
	}
	assertGraphShape(t, union.Graph, 2, 3, false)
	union.Left.Edges.OldToNew[0] = 99
	for name, call := range map[string]func() error{
		"disjoint":   func() error { _, err := self.DisjointUnion(empty); return err },
		"union":      func() error { _, err := self.Union(empty); return err },
		"intersect":  func() error { _, err := self.Intersection(empty); return err },
		"difference": func() error { _, err := self.Difference(empty); return err },
		"compose":    func() error { _, err := self.Compose(empty); return err },
		"complement": func() error { _, err := self.Complement(false); return err },
	} {
		if err := call(); !errors.Is(err, ErrClosed) {
			t.Errorf("closed %s error = %v, want %v", name, err, ErrClosed)
		}
	}
	var nilGraph *Graph
	if _, err := nilGraph.Union(empty); !errors.Is(err, ErrClosed) {
		t.Errorf("nil Union error = %v", err)
	}
}

func TestComposeDifferentVertexCounts(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{1, 2}}, true)
	result, err := left.Compose(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	assertGraphShape(t, result.Graph, 3, 1, true)
	assertEdgesEqual(t, result.Graph, []Edge{{0, 2}})
	wantLeft := IDMapping{OldToNew: []int{0, 1}, NewToOld: []int{0, 1, RemovedID}}
	wantRight, _ := identityIDMapping(3)
	if !reflect.DeepEqual(result.LeftVertices, wantLeft) || !reflect.DeepEqual(result.RightVertices, wantRight) {
		t.Errorf("composition mappings = %#v / %#v", result.LeftVertices, result.RightVertices)
	}
}

func TestComposeUndirectedLoopsPreservesExactProvenance(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 1}}, false)
	right := testGraphFromEdges(t, 3, []Edge{{1, 2}, {1, 1}}, false)
	result, err := left.Compose(right)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Graph.Close() })
	resultEdges, _ := result.Graph.Edges()
	leftEdges, _ := left.Edges()
	rightEdges, _ := right.Edges()
	if len(resultEdges) == 0 || len(result.Edges) != len(resultEdges) {
		t.Fatalf("undirected result/provenance lengths = %d/%d", len(resultEdges), len(result.Edges))
	}
	usedLoop := false
	for edgeID, provenance := range result.Edges {
		leftEdge := leftEdges[provenance.LeftEdge]
		rightEdge := rightEdges[provenance.RightEdge]
		if leftEdge.From == leftEdge.To || rightEdge.From == rightEdge.To {
			usedLoop = true
		}
		if !undirectedCompositionMatches(leftEdge, rightEdge, resultEdges[edgeID]) {
			t.Errorf("edge %d = %#v is inconsistent with provenance %#v then %#v", edgeID, resultEdges[edgeID], leftEdge, rightEdge)
		}
	}
	if !usedLoop {
		t.Error("undirected composition provenance did not exercise a loop")
	}
}

func TestOperatorsRejectMixedDirectednessWithoutResult(t *testing.T) {
	directed := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	undirected := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	if result, err := directed.DisjointUnion(undirected); err == nil || result.Graph != nil {
		t.Errorf("mixed DisjointUnion = %#v, %v", result, err)
	}
	if result, err := directed.Union(undirected); err == nil || result.Graph != nil {
		t.Errorf("mixed Union = %#v, %v", result, err)
	}
	if result, err := directed.Intersection(undirected); err == nil || result.Graph != nil {
		t.Errorf("mixed Intersection = %#v, %v", result, err)
	}
	if result, err := directed.Difference(undirected); err == nil || result.Graph != nil {
		t.Errorf("mixed Difference = %#v, %v", result, err)
	}
	if result, err := directed.Compose(undirected); err == nil || result.Graph != nil {
		t.Errorf("mixed Compose = %#v, %v", result, err)
	}
}

func TestOperatorLockOrderingHandlesReversedAndRepeatedOperands(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{0, 2}, {2, 1}}, true)
	var wait sync.WaitGroup
	errorsCh := make(chan error, 60)
	for index := 0; index < 20; index++ {
		wait.Add(3)
		go func() {
			defer wait.Done()
			result, err := left.Union(right)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
		go func() {
			defer wait.Done()
			result, err := right.Intersection(left)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
		go func() {
			defer wait.Done()
			result, err := left.Compose(left)
			if result.Graph != nil {
				_ = result.Graph.Close()
			}
			errorsCh <- err
		}()
	}
	done := make(chan struct{})
	go func() { wait.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reversed/repeated operator calls deadlocked")
	}
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Errorf("concurrent operator error = %v", err)
		}
	}
}

func TestCollectMappedMergeCleansInitializationQueryAndConversionFailures(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{
		"first init", "second init", "query", "query with graph", "nil graph",
		"left conversion", "right conversion", "mapping length", "mapping value",
	} {
		t.Run(stage, func(t *testing.T) {
			initCalls := 0
			sliceCalls := 0
			closedVectors := 0
			closedGraphs := 0
			operations := mappedMergeOperations{
				newVector: func() (*intVector, error) {
					initCalls++
					if (stage == "first init" && initCalls == 1) || (stage == "second init" && initCalls == 2) {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVector, *intVector) (*Graph, int, int, error) {
					if stage == "query" {
						return nil, 0, 0, forced
					}
					if stage == "query with graph" {
						return &Graph{}, 1, 1, forced
					}
					if stage == "nil graph" {
						return nil, 0, 0, nil
					}
					return &Graph{}, 1, 1, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					sliceCalls++
					if (stage == "left conversion" && sliceCalls == 1) || (stage == "right conversion" && sliceCalls == 2) {
						return nil, forced
					}
					if stage == "mapping length" && sliceCalls == 1 {
						return []int{}, nil
					}
					if stage == "mapping value" && sliceCalls == 1 {
						return []int{2}, nil
					}
					return []int{0}, nil
				},
				closeGraph: func(*Graph) { closedGraphs++ },
			}
			result, err := collectMappedMerge(1, 1, 1, 1, mappedMergeUnion, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectMappedMerge() = %#v, %v, want zero, error", result, err)
			}
			wantVectors := 2
			if stage == "first init" {
				wantVectors = 0
			} else if stage == "second init" {
				wantVectors = 1
			}
			wantGraphs := 1
			if stage == "first init" || stage == "second init" || stage == "query" || stage == "nil graph" {
				wantGraphs = 0
			}
			if closedVectors != wantVectors || closedGraphs != wantGraphs {
				t.Errorf("closed vectors/graphs = %d/%d, want %d/%d", closedVectors, closedGraphs, wantVectors, wantGraphs)
			}
		})
	}

	closed := 0
	result, err := collectMappedMerge(1, 1, 1, 1, mappedMergeIntersection, mappedMergeOperations{
		newVector:   func() (*intVector, error) { return &intVector{}, nil },
		closeVector: func(*intVector) {},
		query:       func(*intVector, *intVector) (*Graph, int, int, error) { return &Graph{}, 1, 1, nil },
		vectorSlice: func(*intVector) ([]int, error) { return []int{0}, nil },
		closeGraph:  func(*Graph) { closed++ },
	})
	if err != nil || result.Graph == nil || closed != 0 {
		t.Errorf("intersection collector success = %#v, %v, closed=%d", result, err, closed)
	}
	result, err = collectMappedMerge(1, 1, 1, 1, mappedMergeIntersection, mappedMergeOperations{
		newVector:   func() (*intVector, error) { return &intVector{}, nil },
		closeVector: func(*intVector) {},
		query:       func(*intVector, *intVector) (*Graph, int, int, error) { return &Graph{}, 1, 1, nil },
		vectorSlice: func(*intVector) ([]int, error) { return []int{}, nil },
		closeGraph:  func(*Graph) { closed++ },
	})
	if result.Graph != nil || err == nil || closed != 1 {
		t.Errorf("intersection length failure = %#v, %v, closed=%d", result, err, closed)
	}
}

func TestCollectCompositionCleansFailuresAndValidatesProvenance(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{
		"first init", "second init", "query", "query with graph", "nil graph", "left conversion",
		"right conversion", "length", "left provenance", "right provenance",
	} {
		t.Run(stage, func(t *testing.T) {
			initCalls := 0
			sliceCalls := 0
			closedVectors := 0
			closedGraphs := 0
			operations := compositionOperations{
				newVector: func() (*intVector, error) {
					initCalls++
					if (stage == "first init" && initCalls == 1) || (stage == "second init" && initCalls == 2) {
						return nil, forced
					}
					return &intVector{}, nil
				},
				closeVector: func(*intVector) { closedVectors++ },
				query: func(*intVector, *intVector) (*Graph, int, int, error) {
					if stage == "query" {
						return nil, 0, 0, forced
					}
					if stage == "query with graph" {
						return &Graph{}, 1, 1, forced
					}
					if stage == "nil graph" {
						return nil, 0, 0, nil
					}
					return &Graph{}, 1, 1, nil
				},
				vectorSlice: func(*intVector) ([]int, error) {
					sliceCalls++
					if (stage == "left conversion" && sliceCalls == 1) || (stage == "right conversion" && sliceCalls == 2) {
						return nil, forced
					}
					if stage == "length" && sliceCalls == 1 {
						return []int{}, nil
					}
					if stage == "left provenance" && sliceCalls == 1 {
						return []int{1}, nil
					}
					if stage == "right provenance" && sliceCalls == 2 {
						return []int{1}, nil
					}
					return []int{0}, nil
				},
				closeGraph: func(*Graph) { closedGraphs++ },
			}
			result, err := collectComposition(1, 1, 1, 1, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectComposition() = %#v, %v, want zero, error", result, err)
			}
			wantVectors := 2
			if stage == "first init" {
				wantVectors = 0
			} else if stage == "second init" {
				wantVectors = 1
			}
			wantGraphs := 1
			if stage == "first init" || stage == "second init" || stage == "query" || stage == "nil graph" {
				wantGraphs = 0
			}
			if closedVectors != wantVectors || closedGraphs != wantGraphs {
				t.Errorf("closed vectors/graphs = %d/%d, want %d/%d", closedVectors, closedGraphs, wantVectors, wantGraphs)
			}
		})
	}
}

func TestOperatorCollectionHelpersRejectNilAndCloseFailures(t *testing.T) {
	if graph, err := collectOwnedOperatorGraph(func() (*Graph, error) { return nil, nil }); graph != nil || err == nil {
		t.Errorf("nil collectOwnedOperatorGraph() = %v, %v", graph, err)
	}
	forced := errors.New("forced failure")
	graph := testGraphFromEdges(t, 1, nil, false)
	if result, err := collectOwnedOperatorGraph(func() (*Graph, error) { return graph, forced }); result != nil || !errors.Is(err, forced) {
		t.Errorf("failed collectOwnedOperatorGraph() = %v, %v", result, err)
	}
	if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("failed query graph error = %v, want %v", err, ErrClosed)
	}

	graph = testGraphFromEdges(t, 1, nil, false)
	result, err := collectBinaryGraphResult(
		func() (*Graph, error) { return graph, nil },
		func(*Graph) (GraphIDMapping, GraphIDMapping, error) {
			return GraphIDMapping{}, GraphIDMapping{}, forced
		},
	)
	if result.Graph != nil || !errors.Is(err, forced) {
		t.Errorf("mapping failure = %#v, %v", result, err)
	}
	if _, err := graph.VertexCount(); !errors.Is(err, ErrClosed) {
		t.Errorf("mapping failure graph error = %v, want %v", err, ErrClosed)
	}
	if _, err := mappingFromExactInverse(1, []int{1}); err == nil {
		t.Error("out-of-range inverse provenance returned nil error")
	}
	if _, err := mappingFromExactInverse(1, []int{0, 0}); err == nil {
		t.Error("repeated inverse provenance returned nil error")
	}
}

func TestCollectDifferenceCleansFailuresAndStructuralMappingIsDeterministic(t *testing.T) {
	forced := errors.New("forced failure")
	for _, stage := range []string{
		"query", "query with graph", "nil graph", "result edges", "vertices", "edges",
	} {
		t.Run(stage, func(t *testing.T) {
			closed := 0
			operations := differenceOperations{
				query:         func() (*Graph, error) { return &Graph{}, nil },
				resultEdges:   func(*Graph) ([]Edge, error) { return []Edge{{0, 1}}, nil },
				vertexMapping: identityIDMapping,
				edgeMapping:   structuralDifferenceEdgeMapping,
				closeGraph:    func(*Graph) { closed++ },
			}
			switch stage {
			case "query":
				operations.query = func() (*Graph, error) { return nil, forced }
			case "query with graph":
				operations.query = func() (*Graph, error) { return &Graph{}, forced }
			case "nil graph":
				operations.query = func() (*Graph, error) { return nil, nil }
			case "result edges":
				operations.resultEdges = func(*Graph) ([]Edge, error) { return nil, forced }
			case "vertices":
				operations.vertexMapping = func(int) (IDMapping, error) { return IDMapping{}, forced }
			case "edges":
				operations.edgeMapping = func([]Edge, []Edge, bool) (IDMapping, error) { return IDMapping{}, forced }
			}
			result, err := collectDifference(2, []Edge{{0, 1}}, true, operations)
			if result.Graph != nil || err == nil {
				t.Errorf("collectDifference() = %#v, %v, want zero, error", result, err)
			}
			wantClosed := 1
			if stage == "query" || stage == "nil graph" {
				wantClosed = 0
			}
			if closed != wantClosed {
				t.Errorf("close count = %d, want %d", closed, wantClosed)
			}
		})
	}

	mapping, err := structuralDifferenceEdgeMapping(
		[]Edge{{3, 1}, {1, 3}, {0, 2}},
		[]Edge{{1, 3}, {0, 2}},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := IDMapping{OldToNew: []int{0, RemovedID, 1}, NewToOld: []int{0, 2}}
	if !reflect.DeepEqual(mapping, want) {
		t.Errorf("parallel structural mapping = %#v, want %#v", mapping, want)
	}
	if _, err := structuralDifferenceEdgeMapping([]Edge{{0, 1}}, []Edge{{1, 1}}, true); err == nil {
		t.Error("result edge absent from source returned nil error")
	}
}

func assertOperatorMappingConsistent(t *testing.T, source, result *Graph, mapping GraphIDMapping) {
	t.Helper()
	sourceEdges, err := source.Edges()
	if err != nil {
		t.Fatal(err)
	}
	resultEdges, err := result.Edges()
	if err != nil {
		t.Fatal(err)
	}
	for oldID, newID := range mapping.Edges.OldToNew {
		if newID == RemovedID {
			continue
		}
		if newID < 0 || newID >= len(resultEdges) || sourceEdges[oldID] != resultEdges[newID] {
			t.Errorf("edge mapping %d -> %d does not preserve endpoints", oldID, newID)
		}
	}
}

func assertNonNilBinaryMappings(t *testing.T, result BinaryGraphOperatorResult) {
	t.Helper()
	for name, mapping := range map[string]IDMapping{
		"left vertices":  result.Left.Vertices,
		"left edges":     result.Left.Edges,
		"right vertices": result.Right.Vertices,
		"right edges":    result.Right.Edges,
	} {
		if mapping.OldToNew == nil || mapping.NewToOld == nil {
			t.Errorf("%s mapping contains nil slice: %#v", name, mapping)
		}
	}
}

func undirectedCompositionMatches(left, right, result Edge) bool {
	leftDirections := []Edge{left}
	if left.From != left.To {
		leftDirections = append(leftDirections, Edge{From: left.To, To: left.From})
	}
	rightDirections := []Edge{right}
	if right.From != right.To {
		rightDirections = append(rightDirections, Edge{From: right.To, To: right.From})
	}
	resultKey := endpointKey(result, false)
	for _, leftDirection := range leftDirections {
		for _, rightDirection := range rightDirections {
			if leftDirection.To != rightDirection.From {
				continue
			}
			candidate := endpointKey(Edge{From: leftDirection.From, To: rightDirection.To}, false)
			if candidate == resultKey {
				return true
			}
		}
	}
	return false
}

func TestComplementRejectsParallelEdges(t *testing.T) {
	g, err := NewGraphFromEdges(2, []Edge{{0, 1}, {0, 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if _, err := g.Complement(false); err == nil {
		t.Error("expected error when calling Complement on graph with parallel edges")
	}
}
