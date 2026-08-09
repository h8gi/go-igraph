#ifndef GO_IGRAPH_OPERATORS_CGO_H
#define GO_IGRAPH_OPERATORS_CGO_H

#include <igraph.h>

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
