#ifndef GO_IGRAPH_MOTIFS_CGO_H
#define GO_IGRAPH_MOTIFS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_dyad_census(const igraph_t *graph,
                                     igraph_real_t *mutual,
                                     igraph_real_t *asymmetric,
                                     igraph_real_t *null_dyads);
igraph_error_t go_igraph_triad_census(const igraph_t *graph,
                                      igraph_vector_t *result);
igraph_error_t go_igraph_count_adjacent_triangles(const igraph_t *graph,
                                                  igraph_vector_t *result,
                                                  igraph_vs_t vertices);
igraph_error_t go_igraph_count_triangles(const igraph_t *graph,
                                         igraph_real_t *result);
igraph_error_t go_igraph_list_triangles(const igraph_t *graph,
                                        igraph_vector_int_t *result);
igraph_error_t go_igraph_motifs_randesu(const igraph_t *graph,
                                        igraph_vector_t *histogram,
                                        igraph_int_t size,
                                        const igraph_vector_t *cut_probability);
igraph_error_t go_igraph_motifs_randesu_estimate(
    const igraph_t *graph,
    igraph_real_t *estimate,
    igraph_int_t size,
    const igraph_vector_t *cut_probability,
    igraph_int_t sample_size,
    const igraph_vector_int_t *sample);
igraph_error_t go_igraph_motifs_randesu_no(
    const igraph_t *graph,
    igraph_real_t *count,
    igraph_int_t size,
    const igraph_vector_t *cut_probability);

#endif
