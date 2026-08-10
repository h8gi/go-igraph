#include "connectivity_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_edge_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks) {
    GO_IGRAPH_CALL(igraph_edge_connectivity(graph, res, checks));
}

igraph_error_t go_igraph_st_edge_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target) {
    GO_IGRAPH_CALL(igraph_st_edge_connectivity(graph, res, source, target));
}

igraph_error_t go_igraph_vertex_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks) {
    GO_IGRAPH_CALL(igraph_vertex_connectivity(graph, res, checks));
}

igraph_error_t go_igraph_st_vertex_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target,
    igraph_vconn_nei_t neighbors) {
    GO_IGRAPH_CALL(igraph_st_vertex_connectivity(graph, res, source, target, neighbors));
}

igraph_error_t go_igraph_edge_disjoint_paths(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target) {
    GO_IGRAPH_CALL(igraph_edge_disjoint_paths(graph, res, source, target));
}

igraph_error_t go_igraph_vertex_disjoint_paths(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target) {
    GO_IGRAPH_CALL(igraph_vertex_disjoint_paths(graph, res, source, target));
}

igraph_error_t go_igraph_adhesion(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks) {
    GO_IGRAPH_CALL(igraph_adhesion(graph, res, checks));
}

igraph_error_t go_igraph_cohesion(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks) {
    GO_IGRAPH_CALL(igraph_cohesion(graph, res, checks));
}
