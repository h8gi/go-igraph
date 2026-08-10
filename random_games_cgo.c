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


