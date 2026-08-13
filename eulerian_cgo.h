#ifndef GO_IGRAPH_EULERIAN_CGO_H
#define GO_IGRAPH_EULERIAN_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_eulerian(
    const igraph_t *, igraph_bool_t *, igraph_bool_t *);
igraph_error_t go_igraph_eulerian_path(
    const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *);
igraph_error_t go_igraph_eulerian_cycle(
    const igraph_t *, igraph_vector_int_t *, igraph_vector_int_t *);

#endif
