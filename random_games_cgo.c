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
