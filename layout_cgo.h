#ifndef GO_IGRAPH_LAYOUT_CGO_H
#define GO_IGRAPH_LAYOUT_CGO_H

#include <igraph.h>

igraph_error_t go_igraph_layout_circle(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_vs_t order);

igraph_error_t go_igraph_layout_star(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t center,
    const igraph_vector_int_t *order);

igraph_error_t go_igraph_layout_grid(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t width);

igraph_error_t go_igraph_layout_random(
    const igraph_t *graph,
    igraph_matrix_t *res);

igraph_error_t go_igraph_layout_reingold_tilford(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_neimode_t mode,
    const igraph_vector_int_t *roots,
    const igraph_vector_int_t *rootlevel);

igraph_error_t go_igraph_layout_reingold_tilford_circular(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_neimode_t mode,
    const igraph_vector_int_t *roots,
    const igraph_vector_int_t *rootlevel);

igraph_error_t go_igraph_layout_bipartite(
    const igraph_t *graph,
    const igraph_vector_bool_t *types,
    igraph_matrix_t *res,
    igraph_real_t hgap,
    igraph_real_t vgap,
    igraph_integer_t maxiter);

igraph_error_t go_igraph_layout_sugiyama(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_matrix_list_t *routing,
    const igraph_vector_int_t *layers,
    igraph_real_t hgap,
    igraph_real_t vgap,
    igraph_integer_t maxiter,
    const igraph_vector_t *weights);

igraph_error_t go_igraph_layout_fruchterman_reingold(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_bool_t use_seed,
    igraph_integer_t niter,
    igraph_real_t start_temp,
    igraph_layout_grid_t grid,
    const igraph_vector_t *weights,
    const igraph_vector_t *minx,
    const igraph_vector_t *maxx,
    const igraph_vector_t *miny,
    const igraph_vector_t *maxy);

igraph_error_t go_igraph_layout_kamada_kawai(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_bool_t use_seed,
    igraph_integer_t maxiter,
    igraph_real_t epsilon,
    igraph_real_t kkconst,
    const igraph_vector_t *weights,
    const igraph_vector_t *minx,
    const igraph_vector_t *maxx,
    const igraph_vector_t *miny,
    const igraph_vector_t *maxy);

igraph_error_t go_igraph_layout_mds(
    const igraph_t *graph,
    igraph_matrix_t *res,
    const igraph_matrix_t *dist,
    igraph_integer_t dim);

#endif
