#include "power_law_cgo.h"
#include <stdint.h>
#ifndef _WIN32
#include <dlfcn.h>
#endif

typedef int (*go_igraph_omp_get_max_threads_t)(void);
typedef void (*go_igraph_omp_set_num_threads_t)(int);

igraph_error_t go_igraph_power_law_fit(
    const igraph_vector_t *data,
    igraph_real_t xmin,
    igraph_bool_t force_continuous,
    igraph_bool_t *continuous,
    igraph_real_t *alpha,
    igraph_real_t *fitted_xmin,
    igraph_real_t *log_likelihood,
    igraph_real_t *ks_statistic) {
    igraph_plfit_result_t result;
    igraph_error_handler_t *old_error = igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning = igraph_set_warning_handler(&igraph_warning_handler_ignore);
    igraph_error_t code = igraph_power_law_fit(data, &result, xmin, force_continuous);
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    if (code == IGRAPH_SUCCESS) {
        *continuous = result.continuous;
        *alpha = result.alpha;
        *fitted_xmin = result.xmin;
        *log_likelihood = result.L;
        *ks_statistic = result.D;
    }
    return code;
}

igraph_error_t go_igraph_power_law_p_value(
    const igraph_vector_t *data,
    igraph_bool_t continuous,
    igraph_real_t alpha,
    igraph_real_t xmin,
    igraph_real_t log_likelihood,
    igraph_real_t ks_statistic,
    igraph_real_t precision,
    igraph_bool_t has_seed,
    uint64_t seed,
    igraph_real_t *p_value) {
    igraph_plfit_result_t model = {
        .continuous = continuous,
        .alpha = alpha,
        .xmin = xmin,
        .L = log_likelihood,
        .D = ks_statistic,
        .data = data,
    };
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
#ifndef _WIN32
    go_igraph_omp_get_max_threads_t get_max_threads =
        (go_igraph_omp_get_max_threads_t) dlsym(RTLD_DEFAULT, "omp_get_max_threads");
    go_igraph_omp_set_num_threads_t set_num_threads =
        (go_igraph_omp_set_num_threads_t) dlsym(RTLD_DEFAULT, "omp_set_num_threads");
    int previous_threads = 0;
    if (get_max_threads && set_num_threads) {
        previous_threads = get_max_threads();
        set_num_threads(1);
    }
#endif
    igraph_error_t code = igraph_plfit_result_calculate_p_value(&model, p_value, precision);
#ifndef _WIN32
    if (previous_threads > 0 && set_num_threads) {
        set_num_threads(previous_threads);
    }
#endif
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}
