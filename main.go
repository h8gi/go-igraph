package main

// #cgo pkg-config: igraph
// #include <stdio.h>
// #include <igraph.h>
import "C"
import (
	"fmt"
	"math/rand"
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

// I don't know 'float64 -> double' is reasonable
func (v *Vector) Set(pos int, value float64) {
	C.igraph_vector_set(&v.vector, C.long(pos), C.double(value))
}

func shortest() {
	graph := NewGraph()
	dimvector := NewVector(20)
	dimvector.Set(0, 30)
	dimvector.Set(1, 30)
	fmt.Println(IGRAPH_DIRECTED, IGRAPH_UNDIRECTED)
	C.igraph_lattice(&graph.graph, &dimvector.vector, 0, IGRAPH_UNDIRECTED, 0, 1)
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

}

func main() {
	hoge()
	hoge()
	fmt.Println("Hello")
	runtime.GC()
	time.Sleep(time.Second * 2)
	fmt.Println("World")
	shortest()
}
