#ifndef GO_IGRAPH_SIMILARITY_CGO_H
#define GO_IGRAPH_SIMILARITY_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_similarity_jaccard(
    const igraph_t *, igraph_matrix_t *, igraph_vs_t, igraph_vs_t,
    igraph_neimode_t, igraph_bool_t);
igraph_error_t go_igraph_similarity_dice(
    const igraph_t *, igraph_matrix_t *, igraph_vs_t, igraph_vs_t,
    igraph_neimode_t, igraph_bool_t);
igraph_error_t go_igraph_similarity_inverse_log_weighted(
    const igraph_t *, igraph_matrix_t *, igraph_vs_t, igraph_neimode_t);
igraph_error_t go_igraph_cocitation(
    const igraph_t *, igraph_matrix_t *, igraph_vs_t);
igraph_error_t go_igraph_bibcoupling(
    const igraph_t *, igraph_matrix_t *, igraph_vs_t);

#endif
