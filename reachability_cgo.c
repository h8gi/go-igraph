#include "reachability_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_reachability(
    const igraph_t *graph, igraph_vector_int_t *membership,
    igraph_vector_int_t *sizes, igraph_int_t *component_count,
    igraph_vector_int_list_t *reachable, igraph_neimode_t mode) {
    igraph_error_handler_t *old_error =
        igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning =
        igraph_set_warning_handler(&igraph_warning_handler_ignore);
    igraph_bitset_list_t bitsets;
    igraph_error_t code = igraph_bitset_list_init(&bitsets, 0);
    if (code != IGRAPH_SUCCESS) {
        igraph_set_warning_handler(old_warning);
        igraph_set_error_handler(old_error);
        return code;
    }
    code = igraph_reachability(
        graph, membership, sizes, component_count, &bitsets, mode);
    if (code == IGRAPH_SUCCESS) {
        igraph_int_t count = igraph_bitset_list_size(&bitsets);
        code = igraph_vector_int_list_resize(reachable, count);
        for (igraph_int_t i = 0; code == IGRAPH_SUCCESS && i < count; i++) {
            const igraph_bitset_t *set = igraph_bitset_list_get_ptr(&bitsets, i);
            igraph_vector_int_t *vertices = igraph_vector_int_list_get_ptr(reachable, i);
            for (igraph_int_t vertex = 0; vertex < igraph_vcount(graph); vertex++) {
                if (IGRAPH_BIT_TEST(*set, vertex)) {
                    code = igraph_vector_int_push_back(vertices, vertex);
                    if (code != IGRAPH_SUCCESS) {
                        break;
                    }
                }
            }
        }
    }
    igraph_bitset_list_destroy(&bitsets);
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}

igraph_error_t go_igraph_count_reachable(
    const igraph_t *graph, igraph_vector_int_t *counts,
    igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_count_reachable(graph, counts, mode));
}

igraph_error_t go_igraph_neighborhood_graphs(
    const igraph_t *graph, igraph_graph_list_t *graphs, igraph_vs_t vertices,
    igraph_int_t order, igraph_neimode_t mode, igraph_int_t min_distance) {
    GO_IGRAPH_CALL(igraph_neighborhood_graphs(
        graph, graphs, vertices, order, mode, min_distance));
}

igraph_error_t go_igraph_transitive_closure(
    const igraph_t *graph, igraph_t *closure) {
    GO_IGRAPH_CALL(igraph_transitive_closure(graph, closure));
}
