#include "motifs_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_dyad_census(const igraph_t *graph,
                                     igraph_real_t *mutual,
                                     igraph_real_t *asymmetric,
                                     igraph_real_t *null_dyads) {
    GO_IGRAPH_CALL(igraph_dyad_census(graph, mutual, asymmetric, null_dyads));
}

igraph_error_t go_igraph_triad_census(const igraph_t *graph,
                                      igraph_vector_t *result) {
    GO_IGRAPH_CALL(igraph_triad_census(graph, result));
}

igraph_error_t go_igraph_count_adjacent_triangles(const igraph_t *graph,
                                                  igraph_vector_t *result,
                                                  igraph_vs_t vertices) {
    GO_IGRAPH_CALL(igraph_count_adjacent_triangles(graph, result, vertices));
}

igraph_error_t go_igraph_count_triangles(const igraph_t *graph,
                                         igraph_real_t *result) {
    GO_IGRAPH_CALL(igraph_count_triangles(graph, result));
}

igraph_error_t go_igraph_list_triangles(const igraph_t *graph,
                                        igraph_vector_int_t *result) {
    GO_IGRAPH_CALL(igraph_list_triangles(graph, result));
}

igraph_error_t go_igraph_motifs_randesu(const igraph_t *graph,
                                        igraph_vector_t *histogram,
                                        igraph_int_t size,
                                        const igraph_vector_t *cut_probability) {
    GO_IGRAPH_CALL(igraph_motifs_randesu(graph, histogram, size,
                                         cut_probability));
}

igraph_error_t go_igraph_motifs_randesu_estimate(
    const igraph_t *graph,
    igraph_real_t *estimate,
    igraph_int_t size,
    const igraph_vector_t *cut_probability,
    igraph_int_t sample_size,
    const igraph_vector_int_t *sample) {
    GO_IGRAPH_CALL(igraph_motifs_randesu_estimate(
        graph, estimate, size, cut_probability, sample_size, sample));
}

igraph_error_t go_igraph_motifs_randesu_no(
    const igraph_t *graph,
    igraph_real_t *count,
    igraph_int_t size,
    const igraph_vector_t *cut_probability) {
    GO_IGRAPH_CALL(igraph_motifs_randesu_no(graph, count, size,
                                            cut_probability));
}
