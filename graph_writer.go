package igraph

// #cgo pkg-config: igraph
// #include <stdio.h>
// #include <unistd.h>
// #include <igraph.h>
// #include "foreign_cgo.h"
import "C"

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

type graphFileWriter func(*C.igraph_t, *C.FILE) C.igraph_error_t

func (g *Graph) writeGraphFile(file *os.File, format string, writer graphFileWriter) error {
	if g == nil {
		return ErrClosed
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.closed {
		return ErrClosed
	}
	if file == nil {
		return errors.New("igraph: output file is nil")
	}

	fstruct, err := openFileStream(file)
	if err != nil {
		return err
	}
	defer C.fclose(fstruct)

	graphFileIOMutex.Lock()
	code := writer(&g.graph, fstruct)
	graphFileIOMutex.Unlock()

	if code != C.IGRAPH_SUCCESS {
		return igraphError("write "+format, int(code))
	}
	if C.fflush(fstruct) != 0 {
		return fmt.Errorf("igraph: failed to flush %s", format)
	}
	return nil
}

// WriteEdgeList writes whitespace-separated pairs of vertex IDs representing
// each edge, one per line. The file is borrowed synchronously and never closed.
// Edge lists carry topology only; vertex count for isolated vertices and all
// graph, vertex, and edge attributes are omitted. Edge output order is sorted,
// so original edge order is not preserved. For undirected graphs, endpoints
// within an edge are additionally normalized (source <= target).
//
//igraph:bind igraph_write_graph_edgelist
func (g *Graph) WriteEdgeList(file *os.File) error {
	return g.writeGraphFile(file, "edge list", func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		return C.go_igraph_write_graph_edgelist(graph, stream)
	})
}

// WriteGraphML serializes the graph to GraphML format. The file is borrowed
// synchronously and never closed. Directedness, vertex count, edge topology,
// and all typed graph, vertex, and edge attributes (numeric, boolean, string)
// are serialized. When prefixattr is true, attribute IDs are prefixed with
// 'g_', 'v_', or 'e_' to prevent attribute ID collisions across scopes.
//
//igraph:bind igraph_write_graph_graphml
func (g *Graph) WriteGraphML(file *os.File, prefixattr bool) error {
	return g.writeGraphFile(file, "GraphML", func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		return C.go_igraph_write_graph_graphml(graph, stream, booltoint(prefixattr))
	})
}

// GMLWriteOptions controls GML serialization. Creator overrides the upstream
// Creator header line when non-empty. The zero value uses upstream defaults.
type GMLWriteOptions struct {
	Creator string
}

// WriteGML serializes the graph to GML format. The file is borrowed
// synchronously and never closed. Directedness, vertex count, edge topology,
// and numeric and string graph, vertex, and edge attributes are preserved.
// Boolean attributes are deterministically converted to numeric values (0 and 1).
//
// Attribute names across graph, vertex, and edge scopes must consist solely
// of ASCII alphanumeric characters ([a-zA-Z0-9]); attribute names containing
// underscores or non-alphanumeric characters are rejected to prevent upstream
// key mangling and attribute collisions. Reserved fields include 'Creator',
// 'directed', 'id', 'source', and 'target'.
//
//igraph:bind igraph_write_graph_gml
func (g *Graph) WriteGML(file *os.File, options GMLWriteOptions) error {
	if options.Creator != "" {
		if err := validateAttributeString("GML creator", options.Creator); err != nil {
			return err
		}
	}
	if g == nil {
		return ErrClosed
	}
	g.mu.RLock()
	if g.closed {
		g.mu.RUnlock()
		return ErrClosed
	}
	var meta []AttributeMetadata
	for _, scope := range []AttributeScope{AttributeGraph, AttributeVertex, AttributeEdge} {
		m, err := attributeMetadataLocked(&g.graph, scope)
		if err != nil {
			g.mu.RUnlock()
			return err
		}
		meta = append(meta, m...)
	}
	g.mu.RUnlock()
	for _, m := range meta {
		if err := validateGMLAttributeName(m.Name); err != nil {
			return err
		}
	}
	return g.writeGraphFile(file, "GML", func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		var cCreator *C.char
		if options.Creator != "" {
			cCreator = C.CString(options.Creator)
			defer C.free(unsafe.Pointer(cCreator))
		}
		return C.go_igraph_write_graph_gml(graph, stream, cCreator)
	})
}

func validateGMLAttributeName(name string) error {
	if name == "" {
		return errors.New("igraph: GML attribute name must not be empty")
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return fmt.Errorf("igraph: GML attribute name %q contains non-alphanumeric character %q", name, c)
		}
	}
	return nil
}

func openFileStream(file *os.File) (*C.FILE, error) {
	if file == nil {
		return nil, errors.New("igraph: output file is nil")
	}

	fd := C.dup(C.int(file.Fd()))
	if fd < 0 {
		return nil, errors.New("igraph: failed to duplicate output file descriptor")
	}

	mode := C.CString("w")
	defer C.free(unsafe.Pointer(mode))
	fstruct := C.fdopen(fd, mode)
	if fstruct == nil {
		C.close(fd)
		return nil, errors.New("igraph: failed to open output stream")
	}
	return fstruct, nil
}
