#ifndef GO_IGRAPH_COMPARISON_CGO_H
#define GO_IGRAPH_COMPARISON_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_compare_communities(
    const igraph_vector_int_t *comm1,
    const igraph_vector_int_t *comm2,
    igraph_real_t *result,
    igraph_community_comparison_t method);

igraph_error_t go_igraph_split_join_distance(
    const igraph_vector_int_t *comm1,
    const igraph_vector_int_t *comm2,
    igraph_int_t *distance12,
    igraph_int_t *distance21);

#endif
