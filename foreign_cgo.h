#ifndef GO_IGRAPH_FOREIGN_CGO_H
#define GO_IGRAPH_FOREIGN_CGO_H

#include <igraph.h>
#include <stdio.h>

igraph_error_t go_igraph_read_graph_edgelist(
    igraph_t *, FILE *, igraph_int_t, igraph_bool_t);
igraph_error_t go_igraph_read_graph_graphml(
    igraph_t *, FILE *, igraph_int_t);
igraph_error_t go_igraph_read_graph_gml(igraph_t *, FILE *);

igraph_error_t go_igraph_write_graph_edgelist(
    const igraph_t *, FILE *);
igraph_error_t go_igraph_write_graph_graphml(
    const igraph_t *, FILE *, igraph_bool_t);
igraph_error_t go_igraph_write_graph_gml(
    const igraph_t *, FILE *, const char *);

#endif
