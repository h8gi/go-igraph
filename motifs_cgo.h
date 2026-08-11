#ifndef GO_IGRAPH_MOTIFS_CGO_H
#define GO_IGRAPH_MOTIFS_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_dyad_census(const igraph_t *graph,
                                     igraph_real_t *mut,
                                     igraph_real_t *asym,
                                     igraph_real_t *null);

igraph_error_t go_igraph_triad_census(const igraph_t *igraph,
                                      igraph_vector_t *res);

igraph_error_t go_igraph_count_triangles(const igraph_t *graph,
                                         igraph_real_t *res);

igraph_error_t go_igraph_list_triangles(const igraph_t *graph,
                                        igraph_vector_int_t *res);

#endif
