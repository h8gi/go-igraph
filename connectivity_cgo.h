#ifndef GO_IGRAPH_CONNECTIVITY_CGO_H
#define GO_IGRAPH_CONNECTIVITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_edge_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks);

igraph_error_t go_igraph_st_edge_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target);

igraph_error_t go_igraph_vertex_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks);

igraph_error_t go_igraph_st_vertex_connectivity(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target,
    igraph_vconn_nei_t neighbors);

igraph_error_t go_igraph_edge_disjoint_paths(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target);

igraph_error_t go_igraph_vertex_disjoint_paths(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_int_t source,
    igraph_int_t target);

igraph_error_t go_igraph_adhesion(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks);

igraph_error_t go_igraph_cohesion(
    const igraph_t *graph,
    igraph_int_t *res,
    igraph_bool_t checks);

#endif
