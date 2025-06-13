package igraph

// #cgo pkg-config: igraph libxml-2.0
// #include <igraph/igraph.h>
import "C"

func booltoint(in bool) C.igraph_bool_t {
	if in {
		return C.igraph_bool_t(C.bool(true))
	}
	return C.igraph_bool_t(C.bool(false))
}
