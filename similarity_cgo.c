#include "similarity_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_similarity_jaccard(
        const igraph_t *graph, igraph_matrix_t *result,
        igraph_vs_t rows, igraph_vs_t columns,
        igraph_neimode_t mode, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_similarity_jaccard(
        graph, result, rows, columns, mode, loops));
}

igraph_error_t go_igraph_similarity_dice(
        const igraph_t *graph, igraph_matrix_t *result,
        igraph_vs_t rows, igraph_vs_t columns,
        igraph_neimode_t mode, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_similarity_dice(
        graph, result, rows, columns, mode, loops));
}

igraph_error_t go_igraph_similarity_inverse_log_weighted(
        const igraph_t *graph, igraph_matrix_t *result,
        igraph_vs_t vertices, igraph_neimode_t mode) {
    GO_IGRAPH_CALL(igraph_similarity_inverse_log_weighted(
        graph, result, vertices, mode));
}

igraph_error_t go_igraph_cocitation(
        const igraph_t *graph, igraph_matrix_t *result,
        igraph_vs_t vertices) {
    GO_IGRAPH_CALL(igraph_cocitation(graph, result, vertices));
}

igraph_error_t go_igraph_bibcoupling(
        const igraph_t *graph, igraph_matrix_t *result,
        igraph_vs_t vertices) {
    GO_IGRAPH_CALL(igraph_bibcoupling(graph, result, vertices));
}

igraph_error_t go_igraph_similarity_jaccard_pairs(
        const igraph_t *graph, igraph_vector_t *result,
        const igraph_vector_int_t *pairs,
        igraph_neimode_t mode, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_similarity_jaccard_pairs(
        graph, result, pairs, mode, loops));
}

igraph_error_t go_igraph_similarity_dice_pairs(
        const igraph_t *graph, igraph_vector_t *result,
        const igraph_vector_int_t *pairs,
        igraph_neimode_t mode, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_similarity_dice_pairs(
        graph, result, pairs, mode, loops));
}

igraph_error_t go_igraph_similarity_jaccard_es(
        const igraph_t *graph, igraph_vector_t *result,
        igraph_es_t edges, igraph_neimode_t mode, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_similarity_jaccard_es(
        graph, result, edges, mode, loops));
}

igraph_error_t go_igraph_similarity_dice_es(
        const igraph_t *graph, igraph_vector_t *result,
        igraph_es_t edges, igraph_neimode_t mode, igraph_bool_t loops) {
    GO_IGRAPH_CALL(igraph_similarity_dice_es(
        graph, result, edges, mode, loops));
}
