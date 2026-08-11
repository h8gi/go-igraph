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

igraph_error_t go_igraph_cliques(const igraph_t *graph,
                                 igraph_vector_int_list_t *result,
                                 igraph_int_t min_size,
                                 igraph_int_t max_size,
                                 igraph_int_t max_results) {
    GO_IGRAPH_CALL(igraph_cliques(
        graph, result, min_size, max_size, max_results));
}

igraph_error_t go_igraph_clique_size_hist(const igraph_t *graph,
                                          igraph_vector_t *result,
                                          igraph_int_t min_size,
                                          igraph_int_t max_size) {
    GO_IGRAPH_CALL(igraph_clique_size_hist(
        graph, result, min_size, max_size));
}

igraph_error_t go_igraph_maximal_cliques(
    const igraph_t *graph, igraph_vector_int_list_t *result,
    igraph_int_t min_size, igraph_int_t max_size,
    igraph_int_t max_results) {
    GO_IGRAPH_CALL(igraph_maximal_cliques(
        graph, result, min_size, max_size, max_results));
}

igraph_error_t go_igraph_maximal_cliques_count(
    const igraph_t *graph, igraph_int_t *result,
    igraph_int_t min_size, igraph_int_t max_size) {
    GO_IGRAPH_CALL(igraph_maximal_cliques_count(
        graph, result, min_size, max_size));
}

igraph_error_t go_igraph_maximal_cliques_hist(
    const igraph_t *graph, igraph_vector_t *result,
    igraph_int_t min_size, igraph_int_t max_size) {
    GO_IGRAPH_CALL(igraph_maximal_cliques_hist(
        graph, result, min_size, max_size));
}

igraph_error_t go_igraph_maximal_cliques_subset(
    const igraph_t *graph, const igraph_vector_int_t *initial_vertices,
    igraph_vector_int_list_t *result, igraph_int_t min_size,
    igraph_int_t max_size, igraph_int_t max_results) {
    igraph_int_t ignored_count;
    GO_IGRAPH_CALL(igraph_maximal_cliques_subset(
        graph, initial_vertices, result, &ignored_count, NULL,
        min_size, max_size, max_results));
}

igraph_error_t go_igraph_weighted_cliques(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_vector_int_list_t *result, igraph_bool_t maximal,
    igraph_real_t min_weight, igraph_real_t max_weight,
    igraph_int_t max_results) {
    GO_IGRAPH_CALL(igraph_weighted_cliques(
        graph, weights, result, maximal, min_weight, max_weight, max_results));
}

igraph_error_t go_igraph_weighted_clique_number(
    const igraph_t *graph, const igraph_vector_t *weights,
    igraph_real_t *result) {
    GO_IGRAPH_CALL(igraph_weighted_clique_number(graph, weights, result));
}
