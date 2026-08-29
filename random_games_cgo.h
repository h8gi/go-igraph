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

igraph_error_t go_igraph_hsbm_game(igraph_t *, igraph_int_t, igraph_int_t,
    const igraph_vector_t *, const igraph_matrix_t *, igraph_real_t);
igraph_error_t go_igraph_hsbm_list_game(igraph_t *, igraph_int_t,
    const igraph_vector_int_t *, const igraph_vector_int_t *,
    const igraph_vector_t *, const igraph_vector_t *, igraph_real_t);
igraph_error_t go_igraph_preference_game(igraph_t *, igraph_int_t, igraph_int_t,
    const igraph_vector_t *, igraph_bool_t, const igraph_matrix_t *,
    igraph_vector_int_t *, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_asymmetric_preference_game(igraph_t *, igraph_int_t,
    igraph_int_t, igraph_int_t, const igraph_matrix_t *, const igraph_matrix_t *,
    igraph_vector_int_t *, igraph_vector_int_t *, igraph_bool_t);
igraph_error_t go_igraph_simple_interconnected_islands_game(igraph_t *,
    igraph_int_t, igraph_int_t, igraph_real_t, igraph_int_t);
igraph_error_t go_igraph_growing_random_game(igraph_t *, igraph_int_t,
    igraph_int_t, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_forest_fire_game(igraph_t *, igraph_int_t,
    igraph_real_t, igraph_real_t, igraph_int_t, igraph_bool_t);
igraph_error_t go_igraph_barabasi_aging_game(igraph_t *, igraph_int_t,
    igraph_int_t, const igraph_vector_int_t *, igraph_bool_t, igraph_real_t,
    igraph_real_t, igraph_int_t, igraph_real_t, igraph_real_t, igraph_real_t,
    igraph_real_t, igraph_bool_t);
igraph_error_t go_igraph_recent_degree_game(igraph_t *, igraph_int_t,
    igraph_real_t, igraph_int_t, igraph_int_t, const igraph_vector_int_t *,
    igraph_bool_t, igraph_real_t, igraph_bool_t);
igraph_error_t go_igraph_recent_degree_aging_game(igraph_t *, igraph_int_t,
    igraph_int_t, const igraph_vector_int_t *, igraph_bool_t, igraph_real_t,
    igraph_real_t, igraph_int_t, igraph_int_t, igraph_real_t, igraph_bool_t);
igraph_error_t go_igraph_callaway_traits_game(igraph_t *, igraph_int_t,
    igraph_int_t, igraph_int_t, const igraph_vector_t *, const igraph_matrix_t *,
    igraph_bool_t, igraph_vector_int_t *);
igraph_error_t go_igraph_establishment_game(igraph_t *, igraph_int_t,
    igraph_int_t, igraph_int_t, const igraph_vector_t *, const igraph_matrix_t *,
    igraph_bool_t, igraph_vector_int_t *);
igraph_error_t go_igraph_lastcit_game(igraph_t *, igraph_int_t, igraph_int_t,
    igraph_int_t, const igraph_vector_t *, igraph_bool_t);
igraph_error_t go_igraph_cited_type_game(igraph_t *, igraph_int_t,
    const igraph_vector_int_t *, const igraph_vector_t *, igraph_int_t,
    igraph_bool_t);
igraph_error_t go_igraph_citing_cited_type_game(igraph_t *, igraph_int_t,
    const igraph_vector_int_t *, const igraph_matrix_t *, igraph_int_t,
    igraph_bool_t);
igraph_error_t go_igraph_correlated_game(igraph_t *, const igraph_t *,
    igraph_real_t, igraph_real_t, const igraph_vector_int_t *);
igraph_error_t go_igraph_correlated_pair_game(igraph_t *, igraph_t *,
    igraph_int_t, igraph_real_t, igraph_real_t, igraph_bool_t,
    const igraph_vector_int_t *, igraph_bool_t *, igraph_bool_t *);
igraph_error_t go_igraph_grg_game(igraph_t *, igraph_int_t, igraph_real_t,
    igraph_bool_t, igraph_vector_t *, igraph_vector_t *);
igraph_error_t go_igraph_dot_product_game(igraph_t *, const igraph_matrix_t *,
    igraph_bool_t);
igraph_error_t go_igraph_rng_sample_dirichlet(igraph_int_t,
    const igraph_vector_t *, igraph_matrix_t *);
igraph_error_t go_igraph_rng_sample_sphere_surface(igraph_int_t, igraph_int_t,
    igraph_real_t, igraph_bool_t, igraph_matrix_t *);
igraph_error_t go_igraph_rng_sample_sphere_volume(igraph_int_t, igraph_int_t,
    igraph_real_t, igraph_bool_t, igraph_matrix_t *);

igraph_error_t go_igraph_chung_lu_game(
    igraph_t *graph,
    const igraph_vector_t *expected_out_degrees,
    const igraph_vector_t *expected_in_degrees,
    igraph_bool_t loops,
    igraph_chung_lu_t variant);

igraph_error_t go_igraph_static_fitness_game(
    igraph_t *graph,
    igraph_int_t edge_count,
    const igraph_vector_t *out_fitness,
    const igraph_vector_t *in_fitness,
    igraph_edge_type_sw_t allowed_edge_types);

igraph_error_t go_igraph_static_power_law_game(
    igraph_t *graph,
    igraph_int_t vertex_count,
    igraph_int_t edge_count,
    igraph_real_t out_exponent,
    igraph_real_t in_exponent,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t finite_size_correction);

igraph_error_t go_igraph_rewire(
    igraph_t *graph,
    igraph_int_t n,
    igraph_edge_type_sw_t allowed_edge_types);

igraph_error_t go_igraph_rewire_edges(
    igraph_t *graph,
    igraph_real_t prob,
    igraph_edge_type_sw_t allowed_edge_types);

igraph_error_t go_igraph_iea_game(
    igraph_t *graph,
    igraph_int_t vertex_count,
    igraph_int_t edge_count,
    igraph_bool_t directed,
    igraph_bool_t loops);

igraph_error_t go_igraph_rewire_directed_edges(
    igraph_t *graph,
    igraph_real_t prob,
    igraph_bool_t loops,
    igraph_neimode_t mode);

igraph_error_t go_igraph_random_walk(
    const igraph_t *graph,
    const igraph_vector_t *weights,
    igraph_vector_int_t *vertices,
    igraph_vector_int_t *edges,
    igraph_int_t start,
    igraph_neimode_t mode,
    igraph_int_t steps,
    igraph_random_walk_stuck_t stuck);

igraph_error_t go_igraph_random_spanning_tree(
    const igraph_t *graph,
    igraph_vector_int_t *res,
    igraph_int_t vid);

#endif
