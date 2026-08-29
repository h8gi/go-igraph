#ifndef GO_IGRAPH_COHESIVE_BLOCKS_CGO_H
#define GO_IGRAPH_COHESIVE_BLOCKS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_cohesive_blocks(
    const igraph_t *, igraph_vector_int_list_t *, igraph_vector_int_t *,
    igraph_vector_int_t *, igraph_t *);
igraph_error_t go_igraph_is_simple_for_cohesive(
    const igraph_t *, igraph_bool_t *);

#endif
