#include "scan_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_local_scan_0(
        const igraph_t *graph, igraph_vector_t *result,
        const igraph_vector_t *weights, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_scan_0(graph, result, weights, mode));
}

igraph_error_t go_igraph_local_scan_1_ecount(
        const igraph_t *graph, igraph_vector_t *result,
        const igraph_vector_t *weights, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_scan_1_ecount(graph, result, weights, mode));
}

igraph_error_t go_igraph_local_scan_k_ecount(
        const igraph_t *graph, igraph_int_t radius, igraph_vector_t *result,
        const igraph_vector_t *weights, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_scan_k_ecount(
        graph, radius, result, weights, mode));
}

igraph_error_t go_igraph_local_scan_subset_ecount(
        const igraph_t *graph, igraph_vector_t *result,
        const igraph_vector_t *weights,
        const igraph_vector_int_list_t *subsets) {
    GO_IGRAPH_CALL(igraph_local_scan_subset_ecount(
        graph, result, weights, subsets));
}

igraph_error_t go_igraph_scan_list_append_copy(
        igraph_vector_int_list_t *list,
        const igraph_vector_int_t *vector) {
    GO_IGRAPH_CALL(igraph_vector_int_list_push_back_copy(list, vector));
}

igraph_error_t go_igraph_local_scan_0_them(
        const igraph_t *neighborhood_graph, const igraph_t *comparison_graph,
        igraph_vector_t *result, const igraph_vector_t *weights,
        igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_scan_0_them(
        neighborhood_graph, comparison_graph, result, weights, mode));
}

igraph_error_t go_igraph_local_scan_1_ecount_them(
        const igraph_t *neighborhood_graph, const igraph_t *comparison_graph,
        igraph_vector_t *result, const igraph_vector_t *weights,
        igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_scan_1_ecount_them(
        neighborhood_graph, comparison_graph, result, weights, mode));
}

igraph_error_t go_igraph_local_scan_k_ecount_them(
        const igraph_t *neighborhood_graph, const igraph_t *comparison_graph,
        igraph_int_t radius, igraph_vector_t *result,
        const igraph_vector_t *weights, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_local_scan_k_ecount_them(
        neighborhood_graph, comparison_graph, radius, result, weights, mode));
}
