#ifndef GO_IGRAPH_FLOW_CGO_H
#define GO_IGRAPH_FLOW_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_maxflow(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_t *flow,
    igraph_vector_int_t *cut,
    igraph_vector_int_t *partition,
    igraph_vector_int_t *partition2,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity,
    igraph_maxflow_stats_t *stats);

igraph_error_t go_igraph_maxflow_value(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity,
    igraph_maxflow_stats_t *stats);

igraph_error_t go_igraph_st_mincut(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_int_t *cut,
    igraph_vector_int_t *partition,
    igraph_vector_int_t *partition2,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity);

igraph_error_t go_igraph_st_mincut_value(
    const igraph_t *graph,
    igraph_real_t *res,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity);

igraph_error_t go_igraph_mincut(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_int_t *partition,
    igraph_vector_int_t *partition2,
    igraph_vector_int_t *cut,
    const igraph_vector_t *capacity);

igraph_error_t go_igraph_mincut_value(
    const igraph_t *graph,
    igraph_real_t *res,
    const igraph_vector_t *capacity);

#endif
