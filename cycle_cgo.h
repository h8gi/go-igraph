#ifndef GO_IGRAPH_CYCLE_CGO_H
#define GO_IGRAPH_CYCLE_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_is_acyclic(
    const igraph_t *graph, igraph_bool_t *result);
igraph_error_t go_igraph_is_dag(
    const igraph_t *graph, igraph_bool_t *result);
igraph_error_t go_igraph_topological_sorting(
    const igraph_t *graph, igraph_vector_int_t *result,
    igraph_neimode_t mode);
igraph_error_t go_igraph_find_cycle(
    const igraph_t *graph, igraph_vector_int_t *vertices,
    igraph_vector_int_t *edges, igraph_neimode_t mode);
igraph_error_t go_igraph_girth(
    const igraph_t *graph, igraph_real_t *girth,
    igraph_vector_int_t *vertices);
igraph_error_t go_igraph_simple_cycles(
    const igraph_t *graph, igraph_vector_int_list_t *vertices,
    igraph_vector_int_list_t *edges, igraph_neimode_t mode,
    igraph_int_t min_cycle_length, igraph_int_t max_cycle_length,
    igraph_int_t max_results);

#endif
