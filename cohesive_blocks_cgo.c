#include "cohesive_blocks_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_cohesive_blocks(
        const igraph_t *graph, igraph_vector_int_list_t *blocks,
        igraph_vector_int_t *cohesion, igraph_vector_int_t *parent,
        igraph_t *block_tree) {
    GO_IGRAPH_CALL(igraph_cohesive_blocks(
        graph, blocks, cohesion, parent, block_tree));
}

igraph_error_t go_igraph_is_simple_for_cohesive(
        const igraph_t *graph, igraph_bool_t *result) {
    GO_IGRAPH_CALL(igraph_is_simple(graph, result, IGRAPH_DIRECTED));
}
