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

func (g *Graph) writeGraphFile(file *os.File, format string, preflight func(*C.igraph_t) error, writer graphFileWriter) (err error) {
	if g == nil {
		return ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrClosed
	}

	if preflight != nil {
		if err := preflight(&g.graph); err != nil {
			return err
		}
	}

	fstruct, err := openFileStream(file)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := C.fclose(fstruct); closeErr != 0 && err == nil {
			err = fmt.Errorf("igraph: failed to close %s output stream", format)
		}
	}()

	graphFileIOMutex.Lock()
	code := writer(&g.graph, fstruct)
	if code == C.IGRAPH_SUCCESS && C.fflush(fstruct) != 0 {
		code = C.IGRAPH_EFILE
	}
	graphFileIOMutex.Unlock()

	if code != C.IGRAPH_SUCCESS {
		return igraphError("write "+format, int(code))
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
	return g.writeGraphFile(file, "edge list", nil, func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
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
	return g.writeGraphFile(file, "GraphML", nil, func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
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
// Attribute names across graph, vertex, and edge scopes must start with an ASCII
// letter ([a-zA-Z]) and consist solely of ASCII alphanumeric characters ([a-zA-Z0-9]).
// Attribute names that start with a digit, contain non-alphanumeric characters, or
// match reserved GML keywords ('Creator', 'directed', 'id', 'source', 'target',
// 'node', 'edge', 'graph', 'label') are rejected to prevent upstream key mangling
// and silent attribute omission.
//
//igraph:bind igraph_write_graph_gml
func (g *Graph) WriteGML(file *os.File, options GMLWriteOptions) error {
	if options.Creator != "" {
		if err := validateGMLCreator(options.Creator); err != nil {
			return err
		}
	}
	preflight := func(graph *C.igraph_t) error {
		for _, scope := range []AttributeScope{AttributeGraph, AttributeVertex, AttributeEdge} {
			meta, err := attributeMetadataLocked(graph, scope)
			if err != nil {
				return err
			}
			for _, m := range meta {
				if err := validateGMLAttributeName(m.Name); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return g.writeGraphFile(file, "GML", preflight, func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		var cCreator *C.char
		if options.Creator != "" {
			cCreator = C.CString(options.Creator)
			defer C.free(unsafe.Pointer(cCreator))
		}
		return C.go_igraph_write_graph_gml(graph, stream, cCreator)
	})
}

func validateGMLCreator(creator string) error {
	if err := validateAttributeString("GML creator", creator); err != nil {
		return err
	}
	for i := 0; i < len(creator); i++ {
		c := creator[i]
		if c == '"' || c == '\n' || c == '\r' {
			return fmt.Errorf("igraph: GML creator contains invalid character %q", c)
		}
	}
	return nil
}

var reservedGMLKeywords = map[string]struct{}{
	"creator":  {},
	"directed": {},
	"id":       {},
	"source":   {},
	"target":   {},
	"node":     {},
	"edge":     {},
	"graph":    {},
	"label":    {},
}

func validateGMLAttributeName(name string) error {
	if name == "" {
		return errors.New("igraph: GML attribute name must not be empty")
	}
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return fmt.Errorf("igraph: GML attribute name %q must start with an ASCII letter", name)
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return fmt.Errorf("igraph: GML attribute name %q contains non-alphanumeric character %q", name, c)
		}
	}
	// Case-insensitive check for reserved keywords
	lower := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			lower[i] = c + ('a' - 'A')
		} else {
			lower[i] = c
		}
	}
	if _, reserved := reservedGMLKeywords[string(lower)]; reserved {
		return fmt.Errorf("igraph: GML attribute name %q is a reserved GML keyword", name)
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
