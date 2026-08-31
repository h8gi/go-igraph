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

igraph_error_t go_igraph_layout_davidson_harel(
    const igraph_t *graph, igraph_matrix_t *res, igraph_bool_t use_seed,
    igraph_integer_t maxiter, igraph_integer_t fineiter,
    igraph_real_t cool_fact, igraph_real_t weight_node_dist,
    igraph_real_t weight_border, igraph_real_t weight_edge_lengths,
    igraph_real_t weight_edge_crossings, igraph_real_t weight_node_edge_dist);

igraph_error_t go_igraph_layout_gem(
    const igraph_t *graph, igraph_matrix_t *res, igraph_bool_t use_seed,
    igraph_integer_t maxiter, igraph_real_t temp_max,
    igraph_real_t temp_min, igraph_real_t temp_init);

igraph_error_t go_igraph_layout_graphopt(
    const igraph_t *graph, igraph_matrix_t *res, igraph_integer_t niter,
    igraph_real_t node_charge, igraph_real_t node_mass,
    igraph_real_t spring_length, igraph_real_t spring_constant,
    igraph_real_t max_sa_movement, igraph_bool_t use_seed);

igraph_error_t go_igraph_layout_drl(
    const igraph_t *graph, igraph_matrix_t *res, igraph_bool_t use_seed,
    int preset, const igraph_vector_t *weights, int dim);

igraph_error_t go_igraph_layout_lgl(
    const igraph_t *graph, igraph_matrix_t *res, igraph_integer_t maxiter,
    igraph_real_t maxdelta, igraph_real_t area, igraph_real_t coolexp,
    igraph_real_t repulserad, igraph_real_t cellsize, igraph_integer_t root);

igraph_error_t go_igraph_layout_align(
    const igraph_t *graph, igraph_matrix_t *layout);

igraph_error_t go_igraph_roots_for_tree_layout(
    const igraph_t *graph, igraph_neimode_t mode,
    igraph_vector_int_t *roots, int choice);

igraph_t *go_igraph_layout_graph_array_alloc(igraph_integer_t count);
void go_igraph_layout_graph_array_set(
    igraph_t *graphs, igraph_integer_t index, const igraph_t *graph);
igraph_matrix_t *go_igraph_layout_matrix_array_alloc(igraph_integer_t count);
void go_igraph_layout_matrix_array_set(
    igraph_matrix_t *matrices, igraph_integer_t index,
    const igraph_matrix_t *matrix);
void go_igraph_layout_array_free(void *array);
igraph_error_t go_igraph_layout_merge_dla(
    const igraph_t *graphs, const igraph_matrix_t *coords,
    igraph_integer_t count, igraph_matrix_t *result);

#endif
