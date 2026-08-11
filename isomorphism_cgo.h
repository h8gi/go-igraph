#ifndef GO_IGRAPH_ISOMORPHISM_CGO_H
#define GO_IGRAPH_ISOMORPHISM_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_isomorphic(
    const igraph_t *left, const igraph_t *right, igraph_bool_t *result);

igraph_error_t go_igraph_subisomorphic(
    const igraph_t *pattern, const igraph_t *target, igraph_bool_t *result);

#endif
