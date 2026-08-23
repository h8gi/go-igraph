#ifndef GO_IGRAPH_DEGREE_REALIZATION_CGO_H
#define GO_IGRAPH_DEGREE_REALIZATION_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_realize_degree_sequence(
    igraph_t *graph, const igraph_vector_int_t *out_degrees,
    const igraph_vector_int_t *in_degrees,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_realize_degseq_t method);

igraph_error_t go_igraph_realize_bipartite_degree_sequence(
    igraph_t *graph, const igraph_vector_int_t *false_mode_degrees,
    const igraph_vector_int_t *true_mode_degrees,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_realize_degseq_t method);

#endif
