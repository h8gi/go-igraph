package igraph_test

import (
	"reflect"
	"sort"
	"sync"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone21SeparatorAndBiconnectivityPipeline(t *testing.T) {
	edges := []igraph.Edge{{0, 1}, {1, 2}, {2, 0}, {1, 3}, {3, 4}, {4, 1}}
	graph := newMilestone21Graph(t, 5, edges)
	defer graph.Close()
	candidate, err := igraph.VertexIDs(1)
	if err != nil {
		t.Fatal(err)
	}
	separator, err := graph.IsSeparator(candidate)
	if err != nil || !separator {
		t.Fatalf("separator = %v, %v", separator, err)
	}
	minimal, err := graph.IsMinimalSeparator(candidate)
	if err != nil || !minimal {
		t.Fatalf("minimal separator = %v, %v", minimal, err)
	}

	deleted := newMilestone21Graph(t, 5, edges)
	defer deleted.Close()
	if _, err := deleted.DeleteVertices(candidate); err != nil {
		t.Fatal(err)
	}
	components, err := deleted.ConnectedComponents(igraph.ConnectednessWeak)
	if err != nil || components.Count != 2 {
		t.Fatalf("components after deletion = %#v, %v", components, err)
	}

	biconnected, err := graph.IsBiconnected()
	if err != nil || biconnected {
		t.Fatalf("biconnected = %v, %v", biconnected, err)
	}
	points, err := graph.ArticulationPoints()
	if err != nil || !reflect.DeepEqual(points, []int{1}) {
		t.Fatalf("articulation points = %v, %v", points, err)
	}
	decomposition, err := graph.BiconnectedComponents()
	if err != nil || decomposition.Count != 2 {
		t.Fatalf("decomposition = %#v, %v", decomposition, err)
	}
}

func TestMilestone21CohesiveHierarchyPipeline(t *testing.T) {
	edges := []igraph.Edge{
		{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},
		{0, 4}, {0, 5}, {1, 4}, {1, 5}, {4, 5},
	}
	graph := newMilestone21Graph(t, 6, edges)
	result, err := graph.CohesiveBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	defer result.BlockTree.Close()
	if !reflect.DeepEqual(result.Cohesion, []int{2, 3, 3}) || !reflect.DeepEqual(result.Parents, []int{-1, 0, 0}) {
		t.Fatalf("cohesion/parents = %v/%v", result.Cohesion, result.Parents)
	}
	treeEdges, err := result.BlockTree.Edges()
	if err != nil || !reflect.DeepEqual(treeEdges, []igraph.Edge{{0, 1}, {0, 2}}) {
		t.Fatalf("block tree = %v, %v", treeEdges, err)
	}
	for child := 1; child < len(result.Blocks); child++ {
		if !isVertexSubset(result.Blocks[child], result.Blocks[result.Parents[child]]) {
			t.Fatalf("block %d is not contained in parent", child)
		}
	}
}

func TestMilestone21PercolationAgainstReconstruction(t *testing.T) {
	edges := []igraph.Edge{{0, 1}, {2, 3}, {1, 2}, {3, 4}}
	graph := newMilestone21Graph(t, 5, edges)
	defer graph.Close()
	edgeOrder := []int{2, 0, 3, 1}
	bond, err := graph.BondPercolation(edgeOrder)
	if err != nil {
		t.Fatal(err)
	}
	wantGiant, wantActive := reconstructBond(5, edges, edgeOrder)
	if !reflect.DeepEqual(bond.GiantComponentSizes, wantGiant) || !reflect.DeepEqual(bond.ActiveVertexCounts, wantActive) {
		t.Fatalf("bond = %#v, want %v/%v", bond, wantGiant, wantActive)
	}
	vertexOrder := []int{4, 1, 3, 0, 2}
	site, err := graph.SitePercolation(vertexOrder)
	if err != nil {
		t.Fatal(err)
	}
	wantGiant, wantEdges := reconstructSite(5, edges, vertexOrder)
	if !reflect.DeepEqual(site.GiantComponentSizes, wantGiant) || !reflect.DeepEqual(site.ActiveEdgeCounts, wantEdges) {
		t.Fatalf("site = %#v, want %v/%v", site, wantGiant, wantEdges)
	}
}

func TestMilestone21ConcurrentReadOnlyAnalysis(t *testing.T) {
	graph := newMilestone21Graph(t, 5, []igraph.Edge{{0, 1}, {1, 2}, {2, 0}, {2, 3}, {3, 4}, {4, 2}})
	defer graph.Close()
	selector, err := igraph.VertexIDs(2)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	errors := make(chan error, 32)
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := graph.IsSeparator(selector); err != nil {
				errors <- err
			}
			if _, err := graph.IsBiconnected(); err != nil {
				errors <- err
			}
			if _, err := graph.BondPercolation([]int{0, 1, 2, 3, 4, 5}); err != nil {
				errors <- err
			}
			result, err := graph.CohesiveBlocks()
			if err != nil {
				errors <- err
			} else if err := result.BlockTree.Close(); err != nil {
				errors <- err
			}
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

func newMilestone21Graph(t *testing.T, vertices int, edges []igraph.Edge) *igraph.Graph {
	t.Helper()
	graph, err := igraph.NewGraphFromEdges(vertices, edges, false)
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func isVertexSubset(child, parent []int) bool {
	set := make(map[int]struct{}, len(parent))
	for _, vertex := range parent {
		set[vertex] = struct{}{}
	}
	for _, vertex := range child {
		if _, ok := set[vertex]; !ok {
			return false
		}
	}
	return true
}

func reconstructBond(vertices int, edges []igraph.Edge, order []int) ([]int, []int) {
	activeEdges := make([]igraph.Edge, 0, len(edges))
	giant, active := make([]int, 0, len(order)), make([]int, 0, len(order))
	seen := make([]bool, vertices)
	for _, edgeID := range order {
		edge := edges[edgeID]
		activeEdges = append(activeEdges, edge)
		seen[edge.From], seen[edge.To] = true, true
		giant = append(giant, largestComponent(vertices, activeEdges, seen))
		count := 0
		for _, value := range seen {
			if value {
				count++
			}
		}
		active = append(active, count)
	}
	return giant, active
}

func reconstructSite(vertices int, edges []igraph.Edge, order []int) ([]int, []int) {
	active := make([]bool, vertices)
	giant, counts := make([]int, 0, len(order)), make([]int, 0, len(order))
	for _, vertex := range order {
		active[vertex] = true
		current, count := make([]igraph.Edge, 0), 0
		for _, edge := range edges {
			if active[edge.From] && active[edge.To] {
				current = append(current, edge)
				count++
			}
		}
		giant = append(giant, largestComponent(vertices, current, active))
		counts = append(counts, count)
	}
	return giant, counts
}

func largestComponent(vertices int, edges []igraph.Edge, active []bool) int {
	parent := make([]int, vertices)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(v int) int {
		if parent[v] != v {
			parent[v] = find(parent[v])
		}
		return parent[v]
	}
	for _, edge := range edges {
		a, b := find(edge.From), find(edge.To)
		if a != b {
			parent[b] = a
		}
	}
	sizes := make(map[int]int)
	for vertex, enabled := range active {
		if enabled {
			sizes[find(vertex)]++
		}
	}
	values := make([]int, 0, len(sizes))
	for _, size := range sizes {
		values = append(values, size)
	}
	sort.Ints(values)
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}
