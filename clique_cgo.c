#include "clique_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_complete(const igraph_t *graph,
                                     igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_complete(graph, result));
}

igraph_error_t go_igraph_is_clique(const igraph_t *graph,
                                   igraph_vs_t candidate,
                                   igraph_bool_t directed,
                                   igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_clique(graph, candidate, directed, result));
}

igraph_error_t go_igraph_is_independent_vertex_set(
    const igraph_t *graph, igraph_vs_t candidate, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_independent_vertex_set(graph, candidate, result));
}

igraph_error_t go_igraph_clique_number(const igraph_t *graph,
                                       igraph_int_t *result) {
    GO_IGRAPH_CALL(igraph_clique_number(graph, result));
}

igraph_error_t go_igraph_independence_number(const igraph_t *graph,
                                             igraph_int_t *result) {
    GO_IGRAPH_CALL(igraph_independence_number(graph, result));
}
