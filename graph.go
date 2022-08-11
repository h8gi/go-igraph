package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <stdio.h>
// #include <igraph.h>
import "C"
import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

const (
	IGRAPH_DIRECTED   = C.IGRAPH_DIRECTED
	IGRAPH_UNDIRECTED = C.IGRAPH_UNDIRECTED
)

type Graph struct {
	graph C.igraph_t
}

func (g *Graph) Destroy() {
	C.igraph_destroy(&g.graph)
}

func NewGraph() *Graph {
	g := &Graph{}
	runtime.SetFinalizer(g, func(g *Graph) {
		g.Destroy()
	})
	return g
}

func NewLattice(dim Vector, nei int, directed bool, mutual bool, circular bool) *Graph {
	g := NewGraph()
	C.igraph_lattice(&g.graph, &dim.vector, C.int(nei),
		booltoint(directed), booltoint(mutual), booltoint(circular))

	return g
}

func (g *Graph) WriteEdgeList(file *os.File) error {
	fstruct := C.fdopen(C.int(file.Fd()), C.CString("w"))
	if err := C.igraph_write_graph_edgelist(&g.graph, fstruct); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil
}

func (g *Graph) WriteGraphML(file *os.File, prefixattr bool) error {
	fstruct := C.fdopen(C.int(file.Fd()), C.CString("w"))
	if err := C.igraph_write_graph_graphml(&g.graph, fstruct, booltoint(prefixattr)); err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil

}

func shortest() {
	var avg_path_len C.igraph_real_t
	graph := NewGraph()
	dimvector := NewVector(2)
	dimvector.Set(0, 30)
	dimvector.Set(1, 30)
	C.igraph_lattice(&graph.graph, &dimvector.vector, 0, IGRAPH_UNDIRECTED, 0, 1)

	C.igraph_average_path_length(&graph.graph, &avg_path_len, nil, C.IGRAPH_UNDIRECTED, 1)

	fmt.Printf("Average path length (lattice): %g\n", avg_path_len)

	nf, err := os.Create("hoge.graphqml")
	if err != nil {
		panic(err)
	}
	defer nf.Close()
	graph.WriteGraphML(nf, true)
}

func hoge() {
	var diameter C.igraph_real_t
	graph := NewGraph()

	rand.Seed(time.Now().Unix())

	C.igraph_rng_seed(C.igraph_rng_default(), C.ulong(rand.Uint64()))

	C.igraph_erdos_renyi_game(&graph.graph, C.IGRAPH_ERDOS_RENYI_GNM, 1000, 3000,
		C.IGRAPH_UNDIRECTED, C.IGRAPH_NO_LOOPS)

	var zero = C.int(0)
	var path C.igraph_vector_t
	C.igraph_vector_init(&path, 0)
	C.igraph_diameter(&graph.graph, &diameter,
		&zero, &zero, &path, C.IGRAPH_UNDIRECTED, 1)
	fmt.Printf("Diameter of a random graph with average degree %v: %f\n",
		2.0*C.igraph_ecount(&graph.graph)/C.igraph_vcount(&graph.graph), diameter)

	f, err := os.Create("foa.edgelist")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	graph.WriteEdgeList(f)
}
