package main

// #cgo pkg-config: igraph
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

func booltoint(in bool) C.int {
	if in {
		return C.int(1)
	}
	return C.int(0)
}

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

type Vector struct {
	vector C.igraph_vector_t
}

func (v *Vector) Destroy() {
	C.igraph_vector_destroy(&v.vector)
}

func NewVector(size int) *Vector {
	v := &Vector{}
	runtime.SetFinalizer(v, func(v *Vector) {
		v.Destroy()
	})

	C.igraph_vector_init(&v.vector, C.long(size))

	return v
}

func (v *Vector) Set(pos int, value float64) {
	C.igraph_vector_set(&v.vector, C.long(pos), C.double(value))
}

func NewLattice(dim Vector, nei int, directed bool, mutual bool, circular bool) *Graph {
	g := NewGraph()
	C.igraph_lattice(&g.graph, &dim.vector, C.int(nei),
		booltoint(directed), booltoint(mutual), booltoint(circular))

	return g
}

func (g *Graph) Write(file *os.File) error {
	fstruct := C.fdopen(C.int(file.Fd()), C.CString("w"))
	err := C.igraph_write_graph_edgelist(&g.graph, fstruct)
	if err != 0 {
		return errors.New("Write failed")
	}
	C.fflush(fstruct)
	return nil
}

func shortest() {
	graph := NewGraph()
	dimvector := NewVector(20)
	dimvector.Set(0, 30)
	dimvector.Set(1, 30)
	fmt.Println(IGRAPH_DIRECTED, IGRAPH_UNDIRECTED)
	C.igraph_lattice(&graph.graph, &dimvector.vector, 0, IGRAPH_UNDIRECTED, 0, 1)
	g := NewLattice(*dimvector, 0, false, false, true)

	f, err := os.Create("out.edgelist")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	g.Write(f)

	nf, err := os.Create("hoge.edgelist")
	if err != nil {
		panic(err)
	}
	defer nf.Close()
	graph.Write(nf)
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
	graph.Write(f)
}

func main() {
	hoge()
	fmt.Println("Hello")
	fmt.Println("World")
	shortest()
}
