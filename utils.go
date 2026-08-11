package igraph

// #include <igraph.h>
// static igraph_bool_t go_igraph_bool(int value) { return value != 0; }
import "C"
import "fmt"

func booltoint(in bool) C.igraph_bool_t {
	if in {
		return C.go_igraph_bool(1)
	}
	return C.go_igraph_bool(0)
}

//igraph:internal igraph_strerror
//igraph:internal igraph_set_error_handler
//igraph:internal igraph_set_warning_handler
func igraphError(operation string, code int) error {
	description := C.GoString(C.igraph_strerror(C.igraph_error_t(code)))
	return fmt.Errorf("igraph: %s: %s (code %d)", operation, description, code)
}
