#include "cycle_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_acyclic(
        const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_acyclic(graph, result));
}

igraph_error_t go_igraph_is_dag(
        const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_dag(graph, result));
}

igraph_error_t go_igraph_topological_sorting(
        const igraph_t *graph, igraph_vector_int_t *result,
        igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_topological_sorting(graph, result, mode));
}

igraph_error_t go_igraph_find_cycle(
        const igraph_t *graph, igraph_vector_int_t *vertices,
        igraph_vector_int_t *edges, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_find_cycle(graph, vertices, edges, mode));
}

igraph_error_t go_igraph_girth(
        const igraph_t *graph, igraph_real_t *girth,
        igraph_vector_int_t *vertices) {
    GO_IGRAPH_CALL(igraph_girth(graph, girth, vertices));
}

igraph_error_t go_igraph_simple_cycles(
        const igraph_t *graph, igraph_vector_int_list_t *vertices,
        igraph_vector_int_list_t *edges, igraph_neimode_t mode,
        igraph_int_t min_cycle_length, igraph_int_t max_cycle_length,
        igraph_int_t max_results) {
    GO_IGRAPH_CALL(igraph_simple_cycles(
        graph, vertices, edges, mode, min_cycle_length, max_cycle_length,
        max_results));
}

igraph_error_t go_igraph_fundamental_cycles(
        const igraph_t *graph, igraph_vector_int_list_t *result,
        igraph_int_t start_vertex, igraph_real_t bfs_cutoff) {
    GO_IGRAPH_CALL(igraph_fundamental_cycles(
        graph, NULL, result, start_vertex, bfs_cutoff));
}

igraph_error_t go_igraph_minimum_cycle_basis(
        const igraph_t *graph, igraph_vector_int_list_t *result,
        igraph_real_t bfs_cutoff, igraph_bool_t complete,
        igraph_bool_t use_cycle_order) {
    GO_IGRAPH_CALL(igraph_minimum_cycle_basis(
        graph, NULL, result, bfs_cutoff, complete, use_cycle_order));
}

igraph_error_t go_igraph_feedback_arc_set(
        const igraph_t *graph, igraph_vector_int_t *result,
        const igraph_vector_t *weights, igraph_fas_algorithm_t algorithm) {
    GO_IGRAPH_CALL(igraph_feedback_arc_set(
        graph, result, weights, algorithm));
}

igraph_error_t go_igraph_feedback_vertex_set(
        const igraph_t *graph, igraph_vector_int_t *result,
        const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_feedback_vertex_set(
        graph, result, weights, IGRAPH_FVS_EXACT_IP));
}
