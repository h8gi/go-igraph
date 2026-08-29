#ifndef GO_IGRAPH_OPERATORS_CGO_H
#define GO_IGRAPH_OPERATORS_CGO_H

#include <igraph.h>

igraph_t *go_igraph_graph_array_alloc(igraph_int_t);
void go_igraph_graph_array_set(igraph_t *, igraph_int_t, const igraph_t *);

igraph_error_t go_igraph_disjoint_union_many(
    igraph_t *, const igraph_t *, igraph_int_t);
igraph_error_t go_igraph_union_many(
    igraph_t *, const igraph_t *, igraph_int_t,
    igraph_vector_int_list_t *);
igraph_error_t go_igraph_intersection_many(
    igraph_t *, const igraph_t *, igraph_int_t,
    igraph_vector_int_list_t *);
igraph_error_t go_igraph_graph_power(
    const igraph_t *, igraph_t *, igraph_int_t, igraph_bool_t);
igraph_error_t go_igraph_connect_neighborhood(
    igraph_t *, igraph_int_t, igraph_neimode_t);
igraph_error_t go_igraph_join(
    igraph_t *, const igraph_t *, const igraph_t *);
igraph_error_t go_igraph_product(
    igraph_t *, const igraph_t *, const igraph_t *, igraph_product_t);
igraph_error_t go_igraph_rooted_product(
    igraph_t *, const igraph_t *, const igraph_t *, igraph_int_t);

igraph_error_t go_igraph_disjoint_union(
    igraph_t *, const igraph_t *, const igraph_t *);
igraph_error_t go_igraph_union(
    igraph_t *, const igraph_t *, const igraph_t *,
    igraph_vector_int_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_intersection(
    igraph_t *, const igraph_t *, const igraph_t *,
    igraph_vector_int_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_difference(
    igraph_t *, const igraph_t *, const igraph_t *);
igraph_error_t go_igraph_compose(
    igraph_t *, const igraph_t *, const igraph_t *,
    igraph_vector_int_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_complementer(
    igraph_t *, const igraph_t *, igraph_bool_t);
igraph_error_t go_igraph_has_multiple(
    const igraph_t *, igraph_bool_t *);

#endif
