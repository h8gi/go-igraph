#include "degree_realization_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_realize_degree_sequence(
    igraph_t *graph, const igraph_vector_int_t *out_degrees,
    const igraph_vector_int_t *in_degrees,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_realize_degseq_t method) {
  GO_IGRAPH_CALL(igraph_realize_degree_sequence(
      graph, out_degrees, in_degrees, allowed_edge_types, method));
}

igraph_error_t go_igraph_realize_bipartite_degree_sequence(
    igraph_t *graph, const igraph_vector_int_t *false_mode_degrees,
    const igraph_vector_int_t *true_mode_degrees,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_realize_degseq_t method) {
  GO_IGRAPH_CALL(igraph_realize_bipartite_degree_sequence(
      graph, false_mode_degrees, true_mode_degrees,
      allowed_edge_types, method));
}
