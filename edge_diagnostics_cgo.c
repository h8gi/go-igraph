#include "edge_diagnostics_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_has_loop(const igraph_t *graph,
                                  igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_has_loop(graph, result));
}

igraph_error_t go_igraph_count_loops(const igraph_t *graph,
                                     igraph_int_t *result) {
    GO_IGRAPH_CALL(igraph_count_loops(graph, result));
}

igraph_error_t go_igraph_is_loop(const igraph_t *graph,
                                 igraph_vector_bool_t *result,
                                 igraph_es_t edges) {
    GO_IGRAPH_CALL(igraph_is_loop(graph, result, edges));
}

igraph_error_t go_igraph_count_multiple(const igraph_t *graph,
                                        igraph_vector_int_t *result,
                                        igraph_es_t edges) {
    GO_IGRAPH_CALL(igraph_count_multiple(graph, result, edges));
}

igraph_error_t go_igraph_is_multiple(const igraph_t *graph,
                                     igraph_vector_bool_t *result,
                                     igraph_es_t edges) {
    GO_IGRAPH_CALL(igraph_is_multiple(graph, result, edges));
}

igraph_error_t go_igraph_has_mutual(const igraph_t *graph,
                                    igraph_bool_t *result,
                                    igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_has_mutual(graph, result, loops));
}

igraph_error_t go_igraph_is_mutual(const igraph_t *graph,
                                   igraph_vector_bool_t *result,
                                   igraph_es_t edges,
                                   igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_is_mutual(graph, result, edges, loops));
}
