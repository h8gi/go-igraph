package igraph

// #include <igraph.h>
// static igraph_bool_t go_igraph_bool(int value) { return value != 0; }
import "C"

func booltoint(in bool) C.igraph_bool_t {
	if in {
		return C.go_igraph_bool(1)
	}
	return C.go_igraph_bool(0)
}
