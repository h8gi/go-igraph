#include "comparison_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_compare_communities(
    const igraph_vector_int_t *comm1,
    const igraph_vector_int_t *comm2,
    igraph_real_t *result,
    igraph_community_comparison_t method) {
    GO_IGRAPH_CALL(igraph_compare_communities(comm1, comm2, result, method));
}

igraph_error_t go_igraph_split_join_distance(
    const igraph_vector_int_t *comm1,
    const igraph_vector_int_t *comm2,
    igraph_int_t *distance12,
    igraph_int_t *distance21) {
    GO_IGRAPH_CALL(igraph_split_join_distance(comm1, comm2, distance12, distance21));
}
