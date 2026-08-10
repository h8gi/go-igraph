#ifndef GO_IGRAPH_RANDOM_GAMES_CGO_H
#define GO_IGRAPH_RANDOM_GAMES_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_erdos_renyi_game_gnm(
    igraph_t *graph,
    igraph_int_t n,
    igraph_int_t m,
    igraph_bool_t directed,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t edge_labeled);

igraph_error_t go_igraph_erdos_renyi_game_gnp(
    igraph_t *graph,
    igraph_int_t n,
    igraph_real_t p,
    igraph_bool_t directed,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t edge_labeled);

igraph_error_t go_igraph_k_regular_game(
    igraph_t *graph,
    igraph_int_t no_of_nodes,
    igraph_int_t k,
    igraph_bool_t directed,
    igraph_bool_t multiple);

igraph_error_t go_igraph_tree_game(
    igraph_t *graph,
    igraph_int_t n,
    igraph_bool_t directed,
    igraph_random_tree_t method);

igraph_error_t go_igraph_degree_sequence_game(
    igraph_t *graph,
    const igraph_vector_int_t *out_deg,
    const igraph_vector_int_t *in_deg,
    igraph_degseq_t method);

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
    const igraph_t *start_from);

igraph_error_t go_igraph_watts_strogatz_game(
    igraph_t *graph,
    igraph_int_t dim,
    igraph_int_t size,
    igraph_int_t nei,
    igraph_real_t p,
    igraph_edge_type_sw_t allowed_edge_types);

igraph_error_t go_igraph_sbm_game(
    igraph_t *graph,
    const igraph_matrix_t *pref_matrix,
    const igraph_vector_int_t *block_sizes,
    igraph_bool_t directed,
    igraph_edge_type_sw_t allowed_edge_types);

#endif

