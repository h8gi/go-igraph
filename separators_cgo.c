#include "separators_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_separator(
        const igraph_t *graph, igraph_vs_t candidate,
        igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_separator(graph, candidate, result));
}

igraph_error_t go_igraph_is_minimal_separator(
        const igraph_t *graph, igraph_vs_t candidate,
        igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_minimal_separator(graph, candidate, result));
}
