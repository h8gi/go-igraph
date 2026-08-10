#include "graphicality_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_is_graphical(
    const igraph_vector_int_t *out_degrees,
    const igraph_vector_int_t *in_degrees,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t *res) {
    GO_IGRAPH_CALL(igraph_is_graphical(out_degrees, in_degrees, allowed_edge_types, res));
}

igraph_error_t go_igraph_is_bigraphical(
    const igraph_vector_int_t *degrees1,
    const igraph_vector_int_t *degrees2,
    igraph_edge_type_sw_t allowed_edge_types,
    igraph_bool_t *res) {
    GO_IGRAPH_CALL(igraph_is_bigraphical(degrees1, degrees2, allowed_edge_types, res));
}
