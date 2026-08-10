#include "flow_cgo.h"
#include "igraph_error_cgo.h"

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
    igraph_maxflow_stats_t *stats) {
    GO_IGRAPH_CALL(igraph_maxflow(graph, value, flow, cut, partition, partition2, source, target, capacity, stats));
}

igraph_error_t go_igraph_maxflow_value(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity,
    igraph_maxflow_stats_t *stats) {
    GO_IGRAPH_CALL(igraph_maxflow_value(graph, value, source, target, capacity, stats));
}

igraph_error_t go_igraph_st_mincut(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_int_t *cut,
    igraph_vector_int_t *partition,
    igraph_vector_int_t *partition2,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_st_mincut(graph, value, cut, partition, partition2, source, target, capacity));
}

igraph_error_t go_igraph_st_mincut_value(
    const igraph_t *graph,
    igraph_real_t *res,
    igraph_int_t source,
    igraph_int_t target,
    const igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_st_mincut_value(graph, res, source, target, capacity));
}

igraph_error_t go_igraph_mincut(
    const igraph_t *graph,
    igraph_real_t *value,
    igraph_vector_int_t *partition,
    igraph_vector_int_t *partition2,
    igraph_vector_int_t *cut,
    const igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_mincut(graph, value, partition, partition2, cut, capacity));
}

igraph_error_t go_igraph_mincut_value(
    const igraph_t *graph,
    igraph_real_t *res,
    const igraph_vector_t *capacity) {
    GO_IGRAPH_CALL(igraph_mincut_value(graph, res, capacity));
}
