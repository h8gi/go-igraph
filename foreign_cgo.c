#include "foreign_cgo.h"

static igraph_error_t go_igraph_read_with_locale(
        igraph_error_t (*reader)(igraph_t *, FILE *, void *),
        igraph_t *graph, FILE *stream, void *context) {
    igraph_error_handler_t *old_error =
        igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning =
        igraph_set_warning_handler(&igraph_warning_handler_ignore);
    igraph_safelocale_t locale;
    igraph_error_t code = igraph_enter_safelocale(&locale);
    if (code == IGRAPH_SUCCESS) {
        code = reader(graph, stream, context);
        igraph_exit_safelocale(&locale);
    }
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}

typedef struct {
    igraph_int_t vertices;
    igraph_bool_t directed;
} go_igraph_edgelist_context_t;

static igraph_error_t go_igraph_read_edgelist_adapter(
        igraph_t *graph, FILE *stream, void *context) {
    const go_igraph_edgelist_context_t *options = context;
    return igraph_read_graph_edgelist(
        graph, stream, options->vertices, options->directed);
}

igraph_error_t go_igraph_read_graph_edgelist(
        igraph_t *graph, FILE *stream, igraph_int_t vertices,
        igraph_bool_t directed) {
    go_igraph_edgelist_context_t context = { vertices, directed };
    return go_igraph_read_with_locale(
        go_igraph_read_edgelist_adapter, graph, stream, &context);
}

static igraph_error_t go_igraph_read_graphml_adapter(
        igraph_t *graph, FILE *stream, void *context) {
    const igraph_int_t *index = context;
    return igraph_read_graph_graphml(graph, stream, *index);
}

igraph_error_t go_igraph_read_graph_graphml(
        igraph_t *graph, FILE *stream, igraph_int_t index) {
    return go_igraph_read_with_locale(
        go_igraph_read_graphml_adapter, graph, stream, &index);
}

static igraph_error_t go_igraph_read_gml_adapter(
        igraph_t *graph, FILE *stream, void *context) {
    (void) context;
    return igraph_read_graph_gml(graph, stream);
}

igraph_error_t go_igraph_read_graph_gml(igraph_t *graph, FILE *stream) {
    return go_igraph_read_with_locale(
        go_igraph_read_gml_adapter, graph, stream, NULL);
}

static igraph_error_t go_igraph_write_with_locale(
        igraph_error_t (*writer)(const igraph_t *, FILE *, void *),
        const igraph_t *graph, FILE *stream, void *context) {
    igraph_error_handler_t *old_error =
        igraph_set_error_handler(&igraph_error_handler_ignore);
    igraph_warning_handler_t *old_warning =
        igraph_set_warning_handler(&igraph_warning_handler_ignore);
    igraph_safelocale_t locale;
    igraph_error_t code = igraph_enter_safelocale(&locale);
    if (code == IGRAPH_SUCCESS) {
        code = writer(graph, stream, context);
        igraph_exit_safelocale(&locale);
    }
    igraph_set_warning_handler(old_warning);
    igraph_set_error_handler(old_error);
    return code;
}

static igraph_error_t go_igraph_write_edgelist_adapter(
        const igraph_t *graph, FILE *stream, void *context) {
    (void) context;
    return igraph_write_graph_edgelist(graph, stream);
}

igraph_error_t go_igraph_write_graph_edgelist(
        const igraph_t *graph, FILE *stream) {
    return go_igraph_write_with_locale(
        go_igraph_write_edgelist_adapter, graph, stream, NULL);
}

typedef struct {
    igraph_bool_t prefixattr;
} go_igraph_graphml_write_context_t;

static igraph_error_t go_igraph_write_graphml_adapter(
        const igraph_t *graph, FILE *stream, void *context) {
    const go_igraph_graphml_write_context_t *options = context;
    return igraph_write_graph_graphml(graph, stream, options->prefixattr);
}

igraph_error_t go_igraph_write_graph_graphml(
        const igraph_t *graph, FILE *stream, igraph_bool_t prefixattr) {
    go_igraph_graphml_write_context_t context = { prefixattr };
    return go_igraph_write_with_locale(
        go_igraph_write_graphml_adapter, graph, stream, &context);
}

typedef struct {
    const char *creator;
} go_igraph_gml_write_context_t;

static igraph_error_t go_igraph_write_gml_adapter(
        const igraph_t *graph, FILE *stream, void *context) {
    const go_igraph_gml_write_context_t *options = context;
    return igraph_write_graph_gml(
        graph, stream, 0, NULL, options->creator);
}

igraph_error_t go_igraph_write_graph_gml(
        const igraph_t *graph, FILE *stream, const char *creator) {
    go_igraph_gml_write_context_t context = { creator };
    return go_igraph_write_with_locale(
        go_igraph_write_gml_adapter, graph, stream, &context);
}
