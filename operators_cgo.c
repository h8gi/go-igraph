#include "operators_cgo.h"

/* Keep handler installation and restoration in the same cgo call. */
#define GO_IGRAPH_OPERATOR_CALL(expression)                               \
    do {                                                                  \
        igraph_error_handler_t *old_error =                               \
            igraph_set_error_handler(&igraph_error_handler_ignore);       \
        igraph_warning_handler_t *old_warning =                           \
            igraph_set_warning_handler(&igraph_warning_handler_ignore);   \
        igraph_error_t code = (expression);                               \
        igraph_set_warning_handler(old_warning);                          \
        igraph_set_error_handler(old_error);                              \
        return code;                                                      \
    } while (0)

igraph_error_t go_igraph_disjoint_union(
    igraph_t *result, const igraph_t *left, const igraph_t *right) {
    GO_IGRAPH_OPERATOR_CALL(igraph_disjoint_union(result, left, right));
}

igraph_error_t go_igraph_union(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_vector_int_t *left_map, igraph_vector_int_t *right_map) {
    GO_IGRAPH_OPERATOR_CALL(igraph_union(
        result, left, right, left_map, right_map));
}

igraph_error_t go_igraph_intersection(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_vector_int_t *left_map, igraph_vector_int_t *right_map) {
    GO_IGRAPH_OPERATOR_CALL(igraph_intersection(
        result, left, right, left_map, right_map));
}

igraph_error_t go_igraph_difference(
    igraph_t *result, const igraph_t *left, const igraph_t *right) {
    GO_IGRAPH_OPERATOR_CALL(igraph_difference(result, left, right));
}

igraph_error_t go_igraph_compose(
    igraph_t *result, const igraph_t *left, const igraph_t *right,
    igraph_vector_int_t *left_map, igraph_vector_int_t *right_map) {
    GO_IGRAPH_OPERATOR_CALL(igraph_compose(
        result, left, right, left_map, right_map));
}

igraph_error_t go_igraph_complementer(
    igraph_t *result, const igraph_t *graph, igraph_bool_t loops) {
    GO_IGRAPH_OPERATOR_CALL(igraph_complementer(result, graph, loops));
}

igraph_error_t go_igraph_has_multiple(
    const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_OPERATOR_CALL(igraph_has_multiple(graph, result));
}
