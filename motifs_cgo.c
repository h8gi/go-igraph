#include "motifs_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_dyad_census(const igraph_t *graph,
                                     igraph_real_t *mut,
                                     igraph_real_t *asym,
                                     igraph_real_t *null) {
    GO_IGRAPH_CALL(igraph_dyad_census(graph, mut, asym, null));
}

igraph_error_t go_igraph_triad_census(const igraph_t *igraph,
                                      igraph_vector_t *res) {
    GO_IGRAPH_CALL(igraph_triad_census(igraph, res));
}

igraph_error_t go_igraph_count_triangles(const igraph_t *graph,
                                         igraph_real_t *res) {
    GO_IGRAPH_CALL(igraph_count_triangles(graph, res));
}

igraph_error_t go_igraph_list_triangles(const igraph_t *graph,
                                        igraph_vector_int_t *res) {
    GO_IGRAPH_CALL(igraph_list_triangles(graph, res));
}
