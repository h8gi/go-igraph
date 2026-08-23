#ifndef GO_IGRAPH_STRUCTURAL_SUMMARIES_CGO_H
#define GO_IGRAPH_STRUCTURAL_SUMMARIES_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_mean_degree(
    const igraph_t *, igraph_real_t *, igraph_bool_t);
igraph_error_t go_igraph_maxdegree(
    const igraph_t *, igraph_int_t *, igraph_vs_t, igraph_neimode_t,
    igraph_loops_t);
igraph_error_t go_igraph_avg_nearest_neighbor_degree(
    const igraph_t *, igraph_vs_t, igraph_neimode_t, igraph_neimode_t,
    igraph_vector_t *, igraph_vector_t *, const igraph_vector_t *);
igraph_error_t go_igraph_degree_correlation_vector(
    const igraph_t *, const igraph_vector_t *, igraph_vector_t *,
    igraph_neimode_t, igraph_neimode_t, igraph_bool_t);
igraph_error_t go_igraph_reciprocity(
    const igraph_t *, igraph_real_t *, igraph_bool_t, igraph_reciprocity_t);
igraph_error_t go_igraph_diversity(
    const igraph_t *, const igraph_vector_t *, igraph_vector_t *, igraph_vs_t);
igraph_error_t go_igraph_rich_club_sequence(
    const igraph_t *, const igraph_vector_t *, igraph_vector_t *,
    const igraph_vector_int_t *, igraph_bool_t, igraph_bool_t, igraph_bool_t);
igraph_error_t go_igraph_sort_vertex_ids_by_degree(
    const igraph_t *, igraph_vector_int_t *, igraph_vs_t, igraph_neimode_t,
    igraph_loops_t, igraph_order_t);

#endif
