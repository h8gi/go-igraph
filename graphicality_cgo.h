#ifndef GO_IGRAPH_GRAPHICALITY_CGO_H
#define GO_IGRAPH_GRAPHICALITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_graphical(
    const igraph_vector_int_t *out_degrees,
    const igraph_vector_int_t *in_degrees,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t *res);

igraph_error_t go_igraph_is_bigraphical(
    const igraph_vector_int_t *degrees1,
    const igraph_vector_int_t *degrees2,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t *res);

#endif
