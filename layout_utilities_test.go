package igraph_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/h8gi/go-igraph"
)

func TestAlignLayout(t *testing.T) {
	graph, err := igraph.NewRing(4, false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer graph.Close()
	input, _ := igraph.NewMatrixFromRows([][]float64{{10, 10}, {12, 11}, {13, 13}, {11, 12}})
	wantInput := input.Rows()
	aligned, err := graph.AlignLayout(input)
	if err != nil {
		t.Fatalf("AlignLayout failed: %v", err)
	}
	assertAdvancedLayout(t, "aligned", aligned, 4)
	if !reflect.DeepEqual(input.Rows(), wantInput) {
		t.Fatal("AlignLayout mutated its input")
	}
	rows := aligned.Rows()
	for column := 0; column < 2; column++ {
		sum := 0.0
		for row := range rows {
			sum += rows[row][column]
		}
		if math.Abs(sum) > 1e-10 {
			t.Fatalf("aligned column %d is not centered: sum %g", column, sum)
		}
	}
	graph.Close()
	if _, err := graph.AlignLayout(input); err != igraph.ErrClosed {
		t.Fatalf("closed AlignLayout error = %v", err)
	}
}

func TestAlignLayoutDimensionsEmptyAndInvalid(t *testing.T) {
	graph, _ := igraph.NewGraphFromEdges(2, []igraph.Edge{{From: 0, To: 1}}, false)
	defer graph.Close()
	threeD, _ := igraph.NewMatrixFromRows([][]float64{{1, 2, 3}, {4, 5, 6}})
	aligned, err := graph.AlignLayout(threeD)
	if err != nil {
		t.Fatalf("3D AlignLayout failed: %v", err)
	}
	rows, columns := aligned.Dims()
	if rows != 2 || columns != 3 {
		t.Fatalf("3D aligned dimensions = (%d, %d)", rows, columns)
	}
	badRows, _ := igraph.NewMatrix(1, 2)
	zeroDim, _ := igraph.NewMatrix(2, 0)
	badFinite, _ := igraph.NewMatrixFromRows([][]float64{{0, 1}, {math.Inf(1), 2}})
	for index, matrix := range []igraph.Matrix{badRows, zeroDim, badFinite} {
		if _, err := graph.AlignLayout(matrix); err == nil {
			t.Errorf("invalid alignment %d succeeded", index)
		}
	}
	empty, _ := igraph.NewGraph()
	defer empty.Close()
	emptyCoordinates, _ := igraph.NewMatrix(0, 3)
	result, err := empty.AlignLayout(emptyCoordinates)
	if err != nil {
		t.Fatalf("empty AlignLayout failed: %v", err)
	}
	rows, columns = result.Dims()
	if rows != 0 || columns != 3 {
		t.Fatalf("empty aligned dimensions = (%d, %d)", rows, columns)
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.AlignLayout(emptyCoordinates); err != igraph.ErrClosed {
		t.Fatalf("nil AlignLayout error = %v", err)
	}
}

func TestTreeLayoutRoots(t *testing.T) {
	path, err := igraph.NewGraphFromEdges(5, []igraph.Edge{{From: 0, To: 1}, {From: 1, To: 2}, {From: 2, To: 3}, {From: 3, To: 4}}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer path.Close()
	degree, err := path.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByDegree)
	if err != nil {
		t.Fatalf("degree roots failed: %v", err)
	}
	if !reflect.DeepEqual(degree, []int{1}) {
		t.Fatalf("degree roots = %v, want [1]", degree)
	}
	eccentricity, err := path.TreeLayoutRoots(igraph.DegOut, igraph.TreeRootByEccentricity)
	if err != nil {
		t.Fatalf("eccentricity roots failed: %v", err)
	}
	if !reflect.DeepEqual(eccentricity, []int{2}) {
		t.Fatalf("eccentricity roots = %v, want [2]", eccentricity)
	}
	coords, err := path.LayoutReingoldTilford(igraph.DegAll, eccentricity)
	if err != nil {
		t.Fatalf("tree layout with selected roots failed: %v", err)
	}
	assertAdvancedLayout(t, "rooted tree", coords, 5)
}

func TestTreeLayoutRootsDirectedDisconnectedEmptyAndInvalid(t *testing.T) {
	directed, _ := igraph.NewGraphFromEdges(4, []igraph.Edge{{From: 0, To: 1}, {From: 2, To: 1}}, true)
	defer directed.Close()
	roots, err := directed.TreeLayoutRoots(igraph.DegOut, igraph.TreeRootByDegree)
	if err != nil {
		t.Fatalf("directed roots failed: %v", err)
	}
	if !reflect.DeepEqual(roots, []int{3, 2, 0}) {
		t.Fatalf("directed roots = %v, want [3 2 0]", roots)
	}
	if _, err := directed.TreeLayoutRoots(igraph.DegMode(99), igraph.TreeRootByDegree); err == nil {
		t.Fatal("invalid direction succeeded")
	}
	if _, err := directed.TreeLayoutRoots(igraph.DegOut, igraph.TreeRootChoice(99)); err == nil {
		t.Fatal("invalid root choice succeeded")
	}

	empty, _ := igraph.NewGraph()
	roots, err = empty.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByDegree)
	if err != nil || roots == nil || len(roots) != 0 {
		t.Fatalf("empty roots = %v, %v", roots, err)
	}
	empty.Close()
	if _, err := empty.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByDegree); err != igraph.ErrClosed {
		t.Fatalf("closed roots error = %v", err)
	}
	var nilGraph *igraph.Graph
	if _, err := nilGraph.TreeLayoutRoots(igraph.DegAll, igraph.TreeRootByDegree); err != igraph.ErrClosed {
		t.Fatalf("nil roots error = %v", err)
	}
}
