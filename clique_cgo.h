#ifndef GO_IGRAPH_CLIQUE_CGO_H
#define GO_IGRAPH_CLIQUE_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_complete(const igraph_t *, igraph_bool_t *);
igraph_error_t go_igraph_is_clique(const igraph_t *, igraph_vs_t,
                                   igraph_bool_t, igraph_bool_t *);
igraph_error_t go_igraph_is_independent_vertex_set(
    const igraph_t *, igraph_vs_t, igraph_bool_t *);
igraph_error_t go_igraph_clique_number(const igraph_t *, igraph_int_t *);
igraph_error_t go_igraph_independence_number(const igraph_t *, igraph_int_t *);
igraph_error_t go_igraph_cliques(const igraph_t *, igraph_vector_int_list_t *,
                                 igraph_int_t, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_clique_size_hist(const igraph_t *, igraph_vector_t *,
                                          igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_maximal_cliques(
    const igraph_t *, igraph_vector_int_list_t *, igraph_int_t, igraph_int_t,
    igraph_int_t);
igraph_error_t go_igraph_maximal_cliques_count(
    const igraph_t *, igraph_int_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_maximal_cliques_hist(
    const igraph_t *, igraph_vector_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_maximal_cliques_subset(
    const igraph_t *, const igraph_vector_int_t *, igraph_vector_int_list_t *,
    igraph_int_t, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_weighted_cliques(
    const igraph_t *, const igraph_vector_t *, igraph_vector_int_list_t *,
    igraph_bool_t, igraph_real_t, igraph_real_t, igraph_int_t);
igraph_error_t go_igraph_weighted_clique_number(
    const igraph_t *, const igraph_vector_t *, igraph_real_t *);

#endif
