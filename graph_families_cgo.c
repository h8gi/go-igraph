#include "graph_families_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_circulant(igraph_t *graph, igraph_int_t n,
                                   const igraph_vector_int_t *shifts,
                                   igraph_bool_t directed) {
  GO_IGRAPH_CALL(igraph_circulant(graph, n, shifts, directed));
}

igraph_error_t go_igraph_wheel(igraph_t *graph, igraph_int_t n,
                               igraph_wheel_mode_t mode,
                               igraph_int_t center) {
  GO_IGRAPH_CALL(igraph_wheel(graph, n, mode, center));
}

igraph_error_t go_igraph_generalized_petersen(igraph_t *graph, igraph_int_t n,
                                              igraph_int_t k) {
  GO_IGRAPH_CALL(igraph_generalized_petersen(graph, n, k));
}

igraph_error_t go_igraph_full_multipartite(
    igraph_t *graph, igraph_vector_int_t *types,
    const igraph_vector_int_t *sizes, igraph_bool_t directed,
    igraph_neimode_t mode) {
  GO_IGRAPH_CALL(
      igraph_full_multipartite(graph, types, sizes, directed, mode));
}

igraph_error_t go_igraph_turan(igraph_t *graph, igraph_vector_int_t *types,
                               igraph_int_t n, igraph_int_t parts) {
  GO_IGRAPH_CALL(igraph_turan(graph, types, n, parts));
}

igraph_error_t go_igraph_full_citation(igraph_t *graph, igraph_int_t n,
                                       igraph_bool_t directed) {
  GO_IGRAPH_CALL(igraph_full_citation(graph, n, directed));
}

igraph_error_t go_igraph_linegraph(const igraph_t *graph,
                                   igraph_t *linegraph) {
  GO_IGRAPH_CALL(igraph_linegraph(graph, linegraph));
}
