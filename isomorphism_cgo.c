#include "isomorphism_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_isomorphic(
    const igraph_t *left, const igraph_t *right, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_isomorphic(left, right, result));
}

igraph_error_t go_igraph_subisomorphic(
    const igraph_t *pattern, const igraph_t *target, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_subisomorphic(pattern, target, result));
}
