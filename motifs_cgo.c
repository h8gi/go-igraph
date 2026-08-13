#include "motifs_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_dyad_census(const igraph_t *graph,
                                     igraph_real_t *mutual,
                                     igraph_real_t *asymmetric,
                                     igraph_real_t *null_dyads) {
    GO_IGRAPH_CALL(igraph_dyad_census(graph, mutual, asymmetric, null_dyads));
}

igraph_error_t go_igraph_triad_census(const igraph_t *graph,
                                      igraph_vector_t *result) {
    GO_IGRAPH_CALL(igraph_triad_census(graph, result));
}

igraph_error_t go_igraph_count_adjacent_triangles(const igraph_t *graph,
                                                  igraph_vector_t *result,
                                                  igraph_vs_t vertices) {
    GO_IGRAPH_CALL(igraph_count_adjacent_triangles(graph, result, vertices));
}

igraph_error_t go_igraph_count_triangles(const igraph_t *graph,
                                         igraph_real_t *result) {
    GO_IGRAPH_CALL(igraph_count_triangles(graph, result));
}

igraph_error_t go_igraph_list_triangles(const igraph_t *graph,
                                        igraph_vector_int_t *result) {
    GO_IGRAPH_CALL(igraph_list_triangles(graph, result));
}
