#ifndef GO_IGRAPH_GRAPH_FAMILIES_CGO_H
#define GO_IGRAPH_GRAPH_FAMILIES_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_circulant(igraph_t *, igraph_int_t, const igraph_vector_int_t *, igraph_bool_t);
igraph_error_t go_igraph_wheel(igraph_t *, igraph_int_t, igraph_wheel_mode_t, igraph_int_t);
igraph_error_t go_igraph_generalized_petersen(igraph_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_full_multipartite(igraph_t *, igraph_vector_int_t *, const igraph_vector_int_t *, igraph_bool_t, igraph_neimode_t);
igraph_error_t go_igraph_turan(igraph_t *, igraph_vector_int_t *, igraph_int_t, igraph_int_t);
igraph_error_t go_igraph_full_citation(igraph_t *, igraph_int_t, igraph_bool_t);
igraph_error_t go_igraph_linegraph(const igraph_t *, igraph_t *);

#endif
