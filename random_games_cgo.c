#include "random_games_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_erdos_renyi_game_gnm(
    igraph_t *graph,
    igraph_int_t n,
    igraph_int_t m,
    igraph_bool_t directed,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t edge_labeled) {
    GO_IGRAPH_CALL(igraph_erdos_renyi_game_gnm(
        graph, n, m, directed, allowed_edge_types, edge_labeled));
}

igraph_error_t go_igraph_erdos_renyi_game_gnp(
    igraph_t *graph,
    igraph_int_t n,
    igraph_real_t p,
    igraph_bool_t directed,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t edge_labeled) {
    GO_IGRAPH_CALL(igraph_erdos_renyi_game_gnp(
        graph, n, p, directed, allowed_edge_types, edge_labeled));
}

igraph_error_t go_igraph_k_regular_game(
    igraph_t *graph,
    igraph_int_t no_of_nodes,
    igraph_int_t k,
    igraph_bool_t directed,
    igraph_bool_t multiple) {
    GO_IGRAPH_CALL(igraph_k_regular_game(graph, no_of_nodes, k, directed, multiple));
}

igraph_error_t go_igraph_tree_game(
    igraph_t *graph,
    igraph_int_t n,
    igraph_bool_t directed,
    igraph_random_tree_t method) {
    GO_IGRAPH_CALL(igraph_tree_game(graph, n, directed, method));
}

igraph_error_t go_igraph_degree_sequence_game(
    igraph_t *graph,
    const igraph_vector_int_t *out_deg,
    const igraph_vector_int_t *in_deg,
    igraph_degseq_t method) {
    GO_IGRAPH_CALL(igraph_degree_sequence_game(graph, out_deg, in_deg, method));
}

igraph_error_t go_igraph_barabasi_game(
    igraph_t *graph,
    igraph_int_t n,
    igraph_real_t power,
    igraph_int_t m,
    const igraph_vector_int_t *outseq,
    igraph_bool_t outpref,
    igraph_real_t A,
    igraph_bool_t directed,
    igraph_barabasi_algorithm_t algo,
    const igraph_t *start_from) {
    GO_IGRAPH_CALL(igraph_barabasi_game(
        graph, n, power, m, outseq, outpref, A, directed, algo, start_from));
}

igraph_error_t go_igraph_watts_strogatz_game(
    igraph_t *graph,
    igraph_int_t dim,
    igraph_int_t size,
    igraph_int_t nei,
    igraph_real_t p,
    igraph_edge_type_sw_t allowed_edge_types) {
    GO_IGRAPH_CALL(igraph_watts_strogatz_game(
        graph, dim, size, nei, p, allowed_edge_types));
}

igraph_error_t go_igraph_sbm_game(
    igraph_t *graph,
    const igraph_matrix_t *pref_matrix,
    const igraph_vector_int_t *block_sizes,
    igraph_bool_t directed,
    igraph_edge_type_sw_t allowed_edge_types) {
    GO_IGRAPH_CALL(igraph_sbm_game(
        graph, pref_matrix, block_sizes, directed, allowed_edge_types));
}

igraph_error_t go_igraph_hsbm_game(igraph_t *graph, igraph_int_t n,
    igraph_int_t m, const igraph_vector_t *rho, const igraph_matrix_t *C,
    igraph_real_t p) {
    GO_IGRAPH_CALL(igraph_hsbm_game(graph, n, m, rho, C, p));
}

static igraph_error_t go_igraph_hsbm_list_game_impl(igraph_t *graph,
    igraph_int_t n, const igraph_vector_int_t *sizes,
    const igraph_vector_int_t *lengths, const igraph_vector_t *rhos,
    const igraph_vector_t *matrices, igraph_real_t p) {
    igraph_int_t count = igraph_vector_int_size(sizes), rho_offset = 0, matrix_offset = 0;
    igraph_vector_list_t rho_list;
    igraph_matrix_list_t matrix_list;
    IGRAPH_CHECK(igraph_vector_list_init(&rho_list, count));
    IGRAPH_FINALLY(igraph_vector_list_destroy, &rho_list);
    IGRAPH_CHECK(igraph_matrix_list_init(&matrix_list, count));
    IGRAPH_FINALLY(igraph_matrix_list_destroy, &matrix_list);
    for (igraph_int_t i = 0; i < count; ++i) {
        igraph_int_t k = VECTOR(*lengths)[i];
        igraph_vector_t *rho = igraph_vector_list_get_ptr(&rho_list, i);
        igraph_matrix_t *matrix = igraph_matrix_list_get_ptr(&matrix_list, i);
        IGRAPH_CHECK(igraph_vector_resize(rho, k));
        IGRAPH_CHECK(igraph_matrix_resize(matrix, k, k));
        for (igraph_int_t j = 0; j < k; ++j) VECTOR(*rho)[j] = VECTOR(*rhos)[rho_offset++];
        for (igraph_int_t row = 0; row < k; ++row) {
            for (igraph_int_t col = 0; col < k; ++col) {
                MATRIX(*matrix, row, col) = VECTOR(*matrices)[matrix_offset++];
            }
        }
    }
    IGRAPH_CHECK(igraph_hsbm_list_game(graph, n, sizes, &rho_list, &matrix_list, p));
    igraph_matrix_list_destroy(&matrix_list);
    igraph_vector_list_destroy(&rho_list);
    IGRAPH_FINALLY_CLEAN(2);
    return IGRAPH_SUCCESS;
}

igraph_error_t go_igraph_hsbm_list_game(igraph_t *graph, igraph_int_t n,
    const igraph_vector_int_t *sizes, const igraph_vector_int_t *lengths,
    const igraph_vector_t *rhos, const igraph_vector_t *matrices, igraph_real_t p) {
    GO_IGRAPH_CALL(go_igraph_hsbm_list_game_impl(graph, n, sizes, lengths, rhos, matrices, p));
}

igraph_error_t go_igraph_preference_game(igraph_t *graph, igraph_int_t nodes,
    igraph_int_t types, const igraph_vector_t *dist, igraph_bool_t fixed,
    const igraph_matrix_t *pref, igraph_vector_int_t *node_types,
    igraph_bool_t directed, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_preference_game(graph, nodes, types, dist, fixed,
        pref, node_types, directed, loops));
}

igraph_error_t go_igraph_asymmetric_preference_game(igraph_t *graph,
    igraph_int_t nodes, igraph_int_t out_types, igraph_int_t in_types,
    const igraph_matrix_t *dist, const igraph_matrix_t *pref,
    igraph_vector_int_t *out, igraph_vector_int_t *in, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_asymmetric_preference_game(graph, nodes, out_types,
        in_types, dist, pref, out, in, loops));
}

igraph_error_t go_igraph_simple_interconnected_islands_game(igraph_t *graph,
    igraph_int_t count, igraph_int_t size, igraph_real_t pin, igraph_int_t inter) {
    GO_IGRAPH_CALL(igraph_simple_interconnected_islands_game(graph, count, size, pin, inter));
}

igraph_error_t go_igraph_chung_lu_game(
    igraph_t *graph,
    const igraph_vector_t *expected_out_degrees,
    const igraph_vector_t *expected_in_degrees,
    igraph_bool_t loops,
    igraph_chung_lu_t variant) {
    GO_IGRAPH_CALL(igraph_chung_lu_game(
        graph, expected_out_degrees, expected_in_degrees, loops, variant));
}

igraph_error_t go_igraph_static_fitness_game(
    igraph_t *graph,
    igraph_int_t edge_count,
    const igraph_vector_t *out_fitness,
    const igraph_vector_t *in_fitness,
    igraph_edge_type_sw_t allowed_edge_types) {
    GO_IGRAPH_CALL(igraph_static_fitness_game(
        graph, edge_count, out_fitness, in_fitness, allowed_edge_types));
}

igraph_error_t go_igraph_static_power_law_game(
    igraph_t *graph,
    igraph_int_t vertex_count,
    igraph_int_t edge_count,
    igraph_real_t out_exponent,
    igraph_real_t in_exponent,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t finite_size_correction) {
    GO_IGRAPH_CALL(igraph_static_power_law_game(
        graph, vertex_count, edge_count, out_exponent, in_exponent,
        allowed_edge_types, finite_size_correction));
}

igraph_error_t go_igraph_rewire(
    igraph_t *graph,
    igraph_int_t n,
    igraph_edge_type_sw_t allowed_edge_types) {
    GO_IGRAPH_CALL(igraph_rewire(graph, n, allowed_edge_types, NULL));
}

igraph_error_t go_igraph_rewire_edges(
    igraph_t *graph,
    igraph_real_t prob,
    igraph_edge_type_sw_t allowed_edge_types) {
    GO_IGRAPH_CALL(igraph_rewire_edges(graph, prob, allowed_edge_types));
}

igraph_error_t go_igraph_random_walk(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_vector_int_t *vertices,
    igraph_vector_int_t *edges,
    igraph_int_t start,
    igraph_neimode_t mode,
    igraph_int_t steps,
    igraph_random_walk_stuck_t stuck) {
    GO_IGRAPH_CALL(igraph_random_walk(
        graph, weights, vertices, edges, start, mode, steps, stuck));
}

igraph_error_t go_igraph_random_spanning_tree(
    const igraph_t *graph,
    igraph_vector_int_t *res,
    igraph_int_t vid) {
    GO_IGRAPH_CALL(igraph_random_spanning_tree(graph, res, vid));
}
