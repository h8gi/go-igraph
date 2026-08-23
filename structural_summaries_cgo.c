#include "structural_summaries_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_mean_degree(const igraph_t *graph,
                                     igraph_real_t *result,
                                     igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_mean_degree(graph, result, loops));
}

igraph_error_t go_igraph_maxdegree(const igraph_t *graph,
                                   igraph_int_t *result,
                                   igraph_vs_t vertices,
                                   igraph_neimode_t mode,
                                   igraph_loops_t loops) {
    GO_IGRAPH_CALL(igraph_maxdegree(graph, result, vertices, mode, loops));
}

igraph_error_t go_igraph_avg_nearest_neighbor_degree(
        const igraph_t *graph, igraph_vs_t vertices, igraph_neimode_t mode,
        igraph_neimode_t neighbor_degree_mode, igraph_vector_t *by_vertex,
        igraph_vector_t *by_degree, const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_avg_nearest_neighbor_degree(
        graph, vertices, mode, neighbor_degree_mode, by_vertex, by_degree,
        weights));
}

igraph_error_t go_igraph_degree_correlation_vector(
        const igraph_t *graph, const igraph_vector_t *weights,
        igraph_vector_t *result, igraph_neimode_t from_mode,
        igraph_neimode_t to_mode, igraph_bool_t directed_neighbors) {
    GO_IGRAPH_CALL(igraph_degree_correlation_vector(
        graph, weights, result, from_mode, to_mode, directed_neighbors));
}

igraph_error_t go_igraph_reciprocity(const igraph_t *graph,
                                     igraph_real_t *result,
                                     igraph_bool_t ignore_loops,
                                     igraph_reciprocity_t mode) {
    GO_IGRAPH_CALL(igraph_reciprocity(graph, result, ignore_loops, mode));
}

igraph_error_t go_igraph_diversity(const igraph_t *graph,
                                   const igraph_vector_t *weights,
                                   igraph_vector_t *result,
                                   igraph_vs_t vertices) {
    GO_IGRAPH_CALL(igraph_diversity(graph, weights, result, vertices));
}

igraph_error_t go_igraph_rich_club_sequence(
        const igraph_t *graph, const igraph_vector_t *weights,
        igraph_vector_t *result, const igraph_vector_int_t *vertex_order,
        igraph_bool_t normalized, igraph_bool_t loops,
        igraph_bool_t directed) {
    GO_IGRAPH_CALL(igraph_rich_club_sequence(
        graph, weights, result, vertex_order, normalized, loops, directed));
}

igraph_error_t go_igraph_sort_vertex_ids_by_degree(
        const igraph_t *graph, igraph_vector_int_t *result,
        igraph_vs_t vertices, igraph_neimode_t mode, igraph_loops_t loops,
        igraph_order_t order) {
    GO_IGRAPH_CALL(igraph_sort_vertex_ids_by_degree(
        graph, result, vertices, mode, loops, order, false));
}
