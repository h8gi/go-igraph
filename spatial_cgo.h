#ifndef GO_IGRAPH_SPATIAL_CGO_H
#define GO_IGRAPH_SPATIAL_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_convex_hull_2d(
    const igraph_matrix_t *, igraph_vector_int_t *, igraph_matrix_t *);
igraph_error_t go_igraph_spatial_edge_lengths(
    const igraph_t *, igraph_vector_t *, const igraph_matrix_t *,
    igraph_metric_t);

#endif
