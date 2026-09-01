#include "statistical_utilities_cgo.h"
#include "igraph_error_cgo.h"

igraph_error_t go_igraph_running_mean(
    const igraph_vector_t *data,
    igraph_int_t window,
    igraph_vector_t *result) {
    GO_IGRAPH_CALL(igraph_running_mean(data, result, window));
}

igraph_error_t go_igraph_random_sample(
    igraph_int_t low,
    igraph_int_t high,
    igraph_int_t count,
    igraph_bool_t has_seed,
    uint64_t seed,
    igraph_vector_int_t *result) {
    igraph_error_handler_t *old_error = igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning = igraph_set_warning_handler(&igraph_warning_handler_ignore);
    if (has_seed) {
        igraph_error_t seed_code = igraph_rng_seed(igraph_rng_default(), (igraph_uint_t) seed);
        if (seed_code != IGRAPH_SUCCESS) {
            igraph_set_warning_handler(old_warning);
            igraph_set_error_handler(old_error);
            return seed_code;
        }
    }
    igraph_error_t code = igraph_random_sample(result, low, high, count);
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}
