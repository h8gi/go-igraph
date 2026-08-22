package igraph_test

import (
	"math"
	"os"
	"reflect"
	"testing"

	igraph "github.com/h8gi/go-igraph"
)

func TestMilestone18AttributedAnalysisPipeline(t *testing.T) {
	graph := importMilestone18Graph(t)
	categories, err := graph.VertexStringAttributes("category")
	if err != nil {
		t.Fatal(err)
	}
	values, err := graph.VertexNumericAttributes("value")
	if err != nil {
		t.Fatal(err)
	}
	weights, err := graph.EdgeNumericAttributes("weight")
	if err != nil {
		t.Fatal(err)
	}

	categoryValues := igraph.StringCategories(categories)
	assortativity, err := graph.CategoricalAssortativity(categoryValues, igraph.CategoricalAssortativityOptions{Directed: true, Normalized: false})
	if err != nil {
		t.Fatal(err)
	}
	distribution, err := graph.CategoricalJointDistribution(categoryValues, igraph.CategoryJointDistributionOptions{Directed: true, Normalized: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := categoricalModularity(distribution.Matrix.Rows()); math.Abs(got-assortativity) > 1e-12 {
		t.Fatalf("matrix modularity %v != assortativity %v", got, assortativity)
	}
	if rows, ok := distribution.RowCategories.StringValues(); !ok || !reflect.DeepEqual(rows, []string{"a", "b"}) {
		t.Fatalf("row axis = %v, %v", rows, ok)
	}
	if _, err := graph.NumericAssortativity(values, igraph.NumericAssortativityOptions{Weights: weights, Directed: true, Normalized: true}); err != nil {
		t.Fatal(err)
	}

	selected, _ := igraph.VertexIDs(0, 2)
	matrix, err := graph.NeighborhoodSimilarity(selected, selected, igraph.NeighborhoodSimilarityOptions{Direction: igraph.DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	pairs, err := graph.NeighborhoodSimilarityPairs([]igraph.Edge{{From: 0, To: 2}}, igraph.NeighborhoodSimilarityOptions{Direction: igraph.DirectionOut})
	if err != nil {
		t.Fatal(err)
	}
	cell, err := matrix.At(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 || pairs[0] != cell {
		t.Fatalf("pair %v != matrix cell %v", pairs, cell)
	}

	scan, err := graph.LocalScan(igraph.LocalScanOptions{Direction: igraph.DirectionOut, Weights: weights})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(scan, []float64{5, 0, 5}) {
		t.Fatalf("weighted radius-zero scan = %v", scan)
	}
	subsets, err := graph.LocalScanSubsets([][]int{{0, 1}, {0, 1, 2}}, igraph.SubsetLocalScanOptions{Weights: weights})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(subsets, []float64{2, 10}) {
		t.Fatalf("subset scan = %v", subsets)
	}

	if err := graph.Close(); err != nil {
		t.Fatal(err)
	}
	if got := distribution.Matrix.Rows(); !reflect.DeepEqual(got, [][]float64{{1.0 / 3.0, 2.0 / 3.0}, {0, 0}}) {
		t.Fatalf("matrix after close = %v", got)
	}
}

func TestMilestone18TwoSnapshotPipeline(t *testing.T) {
	neighborhoods, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer neighborhoods.Close()
	comparison, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer comparison.Close()
	if err := comparison.SetEdgeNumericAttributes("weight", []float64{2, 5}); err != nil {
		t.Fatal(err)
	}
	weights, err := comparison.EdgeNumericAttributes("weight")
	if err != nil {
		t.Fatal(err)
	}
	got, err := neighborhoods.CrossGraphLocalScan(comparison, igraph.LocalScanOptions{Radius: 1, Direction: igraph.DirectionAll, Weights: weights})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []float64{2, 7, 0}) {
		t.Fatalf("cross-graph scan = %v", got)
	}
}

func importMilestone18Graph(t *testing.T) *igraph.Graph {
	t.Helper()
	source, err := igraph.NewGraphFromEdges(3, []igraph.Edge{{From: 0, To: 1}, {From: 0, To: 2}, {From: 2, To: 1}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := source.SetVertexStringAttributes("category", []string{"a", "b", "a"}); err != nil {
		t.Fatal(err)
	}
	if err := source.SetVertexNumericAttributes("value", []float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := source.SetEdgeNumericAttributes("weight", []float64{2, 3, 5}); err != nil {
		t.Fatal(err)
	}
	file, err := os.CreateTemp(t.TempDir(), "milestone18-*.graphml")
	if err != nil {
		t.Fatal(err)
	}
	if err := source.WriteGraphML(file, false); err != nil {
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	graph, err := igraph.ReadGraphML(file, 0)
	closeErr := file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	t.Cleanup(func() { _ = graph.Close() })
	return graph
}

func categoricalModularity(matrix [][]float64) float64 {
	rows, columns := make([]float64, len(matrix)), make([]float64, len(matrix))
	diagonal := 0.0
	for i := range matrix {
		for j, value := range matrix[i] {
			rows[i] += value
			columns[j] += value
			if i == j {
				diagonal += value
			}
		}
	}
	expected := 0.0
	for i := range rows {
		expected += rows[i] * columns[i]
	}
	return diagonal - expected
}
