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

/*
 * dim selects the upstream variant for the force-directed layouts: 2 calls
 * the 2D function (minz/maxz must be NULL), any other value calls the 3D one.
 */
igraph_error_t go_igraph_layout_fruchterman_reingold(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_bool_t use_seed,
    igraph_integer_t niter,
    igraph_real_t start_temp,
    const igraph_vector_t *weights,
    const igraph_vector_t *minx,
    const igraph_vector_t *maxx,
    const igraph_vector_t *miny,
    const igraph_vector_t *maxy,
    const igraph_vector_t *minz,
    const igraph_vector_t *maxz,
    int dim);

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
    const igraph_vector_t *maxy,
    const igraph_vector_t *minz,
    const igraph_vector_t *maxz,
    int dim);

igraph_error_t go_igraph_layout_mds(
    const igraph_t *graph,
    igraph_matrix_t *res,
    const igraph_matrix_t *dist,
    igraph_integer_t dim);

igraph_error_t go_igraph_layout_random_3d(
    const igraph_t *graph,
    igraph_matrix_t *res);

igraph_error_t go_igraph_layout_grid_3d(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t width,
    igraph_integer_t height);

igraph_error_t go_igraph_layout_sphere(
    const igraph_t *graph,
    igraph_matrix_t *res);

#endif
