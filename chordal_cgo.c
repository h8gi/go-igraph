#include "chordal_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_maximum_cardinality_search(const igraph_t *graph, igraph_vector_int_t *alpha, igraph_vector_int_t *alpham1) {
    GO_IGRAPH_CALL(igraph_maximum_cardinality_search(graph, alpha, alpham1));
}

igraph_error_t go_igraph_is_chordal(const igraph_t *graph, const igraph_vector_int_t *alpha, const igraph_vector_int_t *alpham1, igraph_bool_t *chordal, igraph_vector_int_t *fill_in, igraph_t *newgraph) {
    GO_IGRAPH_CALL(igraph_is_chordal(graph, alpha, alpham1, chordal, fill_in, newgraph));
}

igraph_error_t go_igraph_is_perfect(const igraph_t *graph, igraph_bool_t *perfect) {
    GO_IGRAPH_CALL(igraph_is_perfect(graph, perfect));
}

igraph_error_t go_igraph_is_simple_for_perfect(const igraph_t *graph, igraph_bool_t *simple) {
    GO_IGRAPH_CALL(igraph_is_simple(graph, simple, false));
}
