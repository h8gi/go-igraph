#ifndef GO_IGRAPH_LAYOUT_CGO_H
#define GO_IGRAPH_LAYOUT_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_layout_circle(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_vs_t order);

igraph_error_t go_igraph_layout_star(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t center,
    const igraph_vector_int_t *order);

igraph_error_t go_igraph_layout_grid(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t width);

igraph_error_t go_igraph_layout_random(
    const igraph_t *graph,
    igraph_matrix_t *res);

#endif
