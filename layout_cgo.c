#include "layout_cgo.h"
#include "igraph_error_cgo.h"

#include <stdlib.h>

igraph_error_t go_igraph_layout_circle(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_vs_t order) {
    GO_IGRAPH_CALL(igraph_layout_circle(graph, res, order));
}

igraph_error_t go_igraph_layout_star(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t center,
    const igraph_vector_int_t *order) {
    GO_IGRAPH_CALL(igraph_layout_star(graph, res, center, order));
}

igraph_error_t go_igraph_layout_grid(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t width) {
    GO_IGRAPH_CALL(igraph_layout_grid(graph, res, width));
}

igraph_error_t go_igraph_layout_random(
    const igraph_t *graph,
    igraph_matrix_t *res) {
    GO_IGRAPH_CALL(igraph_layout_random(graph, res));
}

igraph_error_t go_igraph_layout_reingold_tilford(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_neimode_t mode,
    const igraph_vector_int_t *roots,
    const igraph_vector_int_t *rootlevel) {
    GO_IGRAPH_CALL(igraph_layout_reingold_tilford(graph, res, mode, roots, rootlevel));
}

igraph_error_t go_igraph_layout_reingold_tilford_circular(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_neimode_t mode,
    const igraph_vector_int_t *roots,
    const igraph_vector_int_t *rootlevel) {
    GO_IGRAPH_CALL(igraph_layout_reingold_tilford_circular(graph, res, mode, roots, rootlevel));
}

igraph_error_t go_igraph_layout_bipartite(
    const igraph_t *graph,
    const igraph_vector_bool_t *types,
    igraph_matrix_t *res,
    igraph_real_t hgap,
    igraph_real_t vgap,
    igraph_integer_t maxiter) {
    GO_IGRAPH_CALL(igraph_layout_bipartite(graph, types, res, hgap, vgap, maxiter));
}

igraph_error_t go_igraph_layout_sugiyama(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_matrix_list_t *routing,
    const igraph_vector_int_t *layers,
    igraph_real_t hgap,
    igraph_real_t vgap,
    igraph_integer_t maxiter,
    const igraph_vector_t *weights) {
    GO_IGRAPH_CALL(igraph_layout_sugiyama(
        graph, res, routing, layers, hgap, vgap, maxiter, weights));
}

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
    int dim) {
    if (dim == 2) {
        GO_IGRAPH_CALL(igraph_layout_fruchterman_reingold(
            graph, res, use_seed, niter, start_temp, IGRAPH_LAYOUT_AUTOGRID,
            weights, minx, maxx, miny, maxy));
    }
    GO_IGRAPH_CALL(igraph_layout_fruchterman_reingold_3d(
        graph, res, use_seed, niter, start_temp,
        weights, minx, maxx, miny, maxy, minz, maxz));
}

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
    int dim) {
    if (dim == 2) {
        GO_IGRAPH_CALL(igraph_layout_kamada_kawai(
            graph, res, use_seed, maxiter, epsilon, kkconst,
            weights, minx, maxx, miny, maxy));
    }
    GO_IGRAPH_CALL(igraph_layout_kamada_kawai_3d(
        graph, res, use_seed, maxiter, epsilon, kkconst,
        weights, minx, maxx, miny, maxy, minz, maxz));
}

igraph_error_t go_igraph_layout_mds(
    const igraph_t *graph,
    igraph_matrix_t *res,
    const igraph_matrix_t *dist,
    igraph_integer_t dim) {
    GO_IGRAPH_CALL(igraph_layout_mds(graph, res, dist, dim));
}

igraph_error_t go_igraph_layout_random_3d(
    const igraph_t *graph,
    igraph_matrix_t *res) {
    GO_IGRAPH_CALL(igraph_layout_random_3d(graph, res));
}

igraph_error_t go_igraph_layout_grid_3d(
    const igraph_t *graph,
    igraph_matrix_t *res,
    igraph_integer_t width,
    igraph_integer_t height) {
    GO_IGRAPH_CALL(igraph_layout_grid_3d(graph, res, width, height));
}

igraph_error_t go_igraph_layout_sphere(
    const igraph_t *graph,
    igraph_matrix_t *res) {
    GO_IGRAPH_CALL(igraph_layout_sphere(graph, res));
}

igraph_error_t go_igraph_layout_davidson_harel(
    const igraph_t *graph, igraph_matrix_t *res, igraph_bool_t use_seed,
    igraph_integer_t maxiter, igraph_integer_t fineiter,
    igraph_real_t cool_fact, igraph_real_t weight_node_dist,
    igraph_real_t weight_border, igraph_real_t weight_edge_lengths,
    igraph_real_t weight_edge_crossings, igraph_real_t weight_node_edge_dist) {
    GO_IGRAPH_CALL(igraph_layout_davidson_harel(
        graph, res, use_seed, maxiter, fineiter, cool_fact,
        weight_node_dist, weight_border, weight_edge_lengths,
        weight_edge_crossings, weight_node_edge_dist));
}

igraph_error_t go_igraph_layout_gem(
    const igraph_t *graph, igraph_matrix_t *res, igraph_bool_t use_seed,
    igraph_integer_t maxiter, igraph_real_t temp_max,
    igraph_real_t temp_min, igraph_real_t temp_init) {
    GO_IGRAPH_CALL(igraph_layout_gem(
        graph, res, use_seed, maxiter, temp_max, temp_min, temp_init));
}

igraph_error_t go_igraph_layout_graphopt(
    const igraph_t *graph, igraph_matrix_t *res, igraph_integer_t niter,
    igraph_real_t node_charge, igraph_real_t node_mass,
    igraph_real_t spring_length, igraph_real_t spring_constant,
    igraph_real_t max_sa_movement, igraph_bool_t use_seed) {
    GO_IGRAPH_CALL(igraph_layout_graphopt(
        graph, res, niter, node_charge, node_mass, spring_length,
        spring_constant, max_sa_movement, use_seed));
}

igraph_error_t go_igraph_layout_drl(
    const igraph_t *graph, igraph_matrix_t *res, igraph_bool_t use_seed,
    int preset, const igraph_vector_t *weights, int dim) {
    igraph_layout_drl_options_t options;
    igraph_error_handler_t *old_error =
        igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning =
        igraph_set_warning_handler(&igraph_warning_handler_ignore);
    igraph_error_t code = igraph_layout_drl_options_init(
        &options, (igraph_layout_drl_default_t) preset);
    if (code == IGRAPH_SUCCESS) {
        code = dim == 2
            ? igraph_layout_drl(graph, res, use_seed, &options, weights)
            : igraph_layout_drl_3d(graph, res, use_seed, &options, weights);
    }
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}

igraph_error_t go_igraph_layout_lgl(
    const igraph_t *graph, igraph_matrix_t *res, igraph_integer_t maxiter,
    igraph_real_t maxdelta, igraph_real_t area, igraph_real_t coolexp,
    igraph_real_t repulserad, igraph_real_t cellsize, igraph_integer_t root) {
    GO_IGRAPH_CALL(igraph_layout_lgl(
        graph, res, maxiter, maxdelta, area, coolexp,
        repulserad, cellsize, root));
}

igraph_error_t go_igraph_layout_align(
    const igraph_t *graph, igraph_matrix_t *layout) {
    GO_IGRAPH_CALL(igraph_layout_align(graph, layout));
}

igraph_error_t go_igraph_roots_for_tree_layout(
    const igraph_t *graph, igraph_neimode_t mode,
    igraph_vector_int_t *roots, int choice) {
    GO_IGRAPH_CALL(igraph_roots_for_tree_layout(
        graph, mode, roots, (igraph_root_choice_t) choice));
}

igraph_t *go_igraph_layout_graph_array_alloc(igraph_integer_t count) {
    return count == 0 ? NULL : calloc((size_t) count, sizeof(igraph_t));
}

void go_igraph_layout_graph_array_set(
    igraph_t *graphs, igraph_integer_t index, const igraph_t *graph) {
    graphs[index] = *graph;
}

igraph_matrix_t *go_igraph_layout_matrix_array_alloc(igraph_integer_t count) {
    return count == 0 ? NULL : calloc((size_t) count, sizeof(igraph_matrix_t));
}

void go_igraph_layout_matrix_array_set(
    igraph_matrix_t *matrices, igraph_integer_t index,
    const igraph_matrix_t *matrix) {
    matrices[index] = *matrix;
}

void go_igraph_layout_array_free(void *array) {
    free(array);
}

igraph_error_t go_igraph_layout_merge_dla(
    const igraph_t *graphs, const igraph_matrix_t *coords,
    igraph_integer_t count, igraph_matrix_t *result) {
    igraph_error_handler_t *old_error =
        igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning =
        igraph_set_warning_handler(&igraph_warning_handler_ignore);
    igraph_vector_ptr_t pointers;
    igraph_matrix_list_t matrices;
    igraph_error_t code = igraph_vector_ptr_init(&pointers, count);
    if (code != IGRAPH_SUCCESS) {
        goto done;
    }
    code = igraph_matrix_list_init(&matrices, count);
    if (code != IGRAPH_SUCCESS) {
        igraph_vector_ptr_destroy(&pointers);
        goto done;
    }
    for (igraph_integer_t index = 0; index < count; ++index) {
        VECTOR(pointers)[index] = (void *) &graphs[index];
        code = igraph_matrix_update(
            igraph_matrix_list_get_ptr(&matrices, index), &coords[index]);
        if (code != IGRAPH_SUCCESS) {
            break;
        }
    }
    if (code == IGRAPH_SUCCESS) {
        code = igraph_layout_merge_dla(&pointers, &matrices, result);
    }
    igraph_matrix_list_destroy(&matrices);
    igraph_vector_ptr_destroy(&pointers);
done:
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}
