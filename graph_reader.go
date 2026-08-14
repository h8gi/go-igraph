package igraph

// #cgo pkg-config: igraph
// #include <stdio.h>
// #include <igraph.h>
// #include "foreign_cgo.h"
import "C"

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"unsafe"
)

// EdgeListReadOptions controls a plain integer edge-list import. VertexCount
// supplies a minimum vertex count; IDs found in the file may increase it.
// Directed selects the result directedness. The zero value auto-sizes an
// undirected graph.
type EdgeListReadOptions struct {
	VertexCount int
	Directed    bool
}

// ReadEdgeList reads whitespace-separated pairs of non-negative, zero-based
// vertex IDs. The file is borrowed synchronously and its current offset is not
// changed. It must be an open, seekable regular file. Edge lists carry topology
// only and create no attributes. The returned graph is independently owned.
//
//igraph:bind igraph_read_graph_edgelist
func ReadEdgeList(file *os.File, options EdgeListReadOptions) (*Graph, error) {
	if options.VertexCount < 0 {
		return nil, fmt.Errorf("igraph: edge-list vertex count must be non-negative: %d", options.VertexCount)
	}
	return readGraphFile(file, "edge list", func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		return C.go_igraph_read_graph_edgelist(graph, stream, C.igraph_int_t(options.VertexCount), booltoint(options.Directed))
	})
}

// ReadGraphML reads the zero-based graphIndex from a GraphML document. The
// file is borrowed synchronously without changing its current offset and must
// be an open, seekable regular file. Directedness and vertex/edge order follow
// the selected graph. GraphML boolean, numeric, and string graph, vertex, and
// edge attributes are available through the typed APIs. Upstream rejects
// unsupported attribute types and graph structures such as nested graphs and
// hyperedges. The returned graph is independently owned.
//
//igraph:bind igraph_read_graph_graphml
func ReadGraphML(file *os.File, graphIndex int) (*Graph, error) {
	if graphIndex < 0 {
		return nil, fmt.Errorf("igraph: GraphML graph index must be non-negative: %d", graphIndex)
	}
	return readGraphFileWithSnapshot(file, "GraphML", func(file *os.File) (*C.FILE, error) {
		return snapshotGraphMLFile(file, graphIndex)
	}, func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		// The selected graph is the only graph in the snapshot. This also avoids
		// an upstream 1.0.1 parser bug that prevents skipping earlier graphs.
		return C.go_igraph_read_graph_graphml(graph, stream, 0)
	})
}

// ReadGML reads the first graph from a GML document. The file is borrowed
// synchronously without changing its current offset and must be an open,
// seekable regular file. GML controls directedness and vertex identity through
// its directed and node id fields. Simple integer/real values become numeric
// attributes and strings remain strings; compound values are ignored or
// replaced with upstream defaults. The returned graph is independently owned.
//
//igraph:bind igraph_read_graph_gml
func ReadGML(file *os.File) (*Graph, error) {
	return readGraphFile(file, "GML", func(graph *C.igraph_t, stream *C.FILE) C.igraph_error_t {
		return C.go_igraph_read_graph_gml(graph, stream)
	})
}

type graphFileReader func(*C.igraph_t, *C.FILE) C.igraph_error_t

// C-igraph's error/warning handlers and safe-locale state are process-global.
// Serialize the narrow reader call while leaving file snapshotting concurrent.
var graphFileReaderMutex sync.Mutex

//igraph:internal igraph_enter_safelocale
//igraph:internal igraph_exit_safelocale
func readGraphFile(file *os.File, format string, reader graphFileReader) (*Graph, error) {
	return readGraphFileWithSnapshot(file, format, snapshotInputFile, reader)
}

func readGraphFileWithSnapshot(file *os.File, format string, snapshot func(*os.File) (*C.FILE, error), reader graphFileReader) (*Graph, error) {
	if reader == nil {
		return nil, errors.New("igraph: graph reader is nil")
	}
	if snapshot == nil {
		return nil, errors.New("igraph: graph reader snapshotter is nil")
	}
	stream, err := snapshot(file)
	if err != nil {
		return nil, err
	}
	defer C.fclose(stream)
	var value C.igraph_t
	graphFileReaderMutex.Lock()
	defer graphFileReaderMutex.Unlock()
	if code := reader(&value, stream); code != C.IGRAPH_SUCCESS {
		return nil, igraphError("read "+format, int(code))
	}
	return adoptInitializedGraph(&value), nil
}

func snapshotGraphMLFile(file *os.File, graphIndex int) (*C.FILE, error) {
	contents, err := readInputFile(file)
	if err != nil {
		return nil, err
	}
	selected, err := selectGraphML(contents, graphIndex)
	if err != nil {
		return nil, err
	}
	return snapshotBytes(selected)
}

// selectGraphML retains document-level declarations and exactly one top-level
// graph. encoding/xml preserves values and namespace URIs while normalizing
// insignificant lexical details.
func selectGraphML(contents []byte, graphIndex int) ([]byte, error) {
	const graphMLNamespace = "http://graphml.graphdrawing.org/xmlns"
	decoder := xml.NewDecoder(bytes.NewReader(contents))
	var output bytes.Buffer
	encoder := xml.NewEncoder(&output)
	depth, graphCount := 0, 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("igraph: parse GraphML XML: %w", err)
		}
		if start, ok := token.(xml.StartElement); ok && depth == 1 && start.Name.Local == "graph" {
			if graphCount != graphIndex {
				if err := decoder.Skip(); err != nil {
					return nil, fmt.Errorf("igraph: skip GraphML graph: %w", err)
				}
				graphCount++
				continue
			}
			graphCount++
		}
		// encoding/xml otherwise adds redundant namespace declarations when it
		// re-encodes names that already live under GraphML's default namespace.
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Space == graphMLNamespace {
				value.Name.Space = ""
			}
			for i := range value.Attr {
				if value.Attr[i].Name.Space == graphMLNamespace {
					value.Attr[i].Name.Space = ""
				}
			}
			token = value
		case xml.EndElement:
			if value.Name.Space == graphMLNamespace {
				value.Name.Space = ""
			}
			token = value
		}
		if err := encoder.EncodeToken(token); err != nil {
			return nil, fmt.Errorf("igraph: encode GraphML snapshot: %w", err)
		}
		switch token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	if graphCount <= graphIndex {
		return nil, fmt.Errorf("igraph: GraphML graph index %d is out of range", graphIndex)
	}
	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("igraph: encode GraphML snapshot: %w", err)
	}
	return output.Bytes(), nil
}

func snapshotInputFile(file *os.File) (*C.FILE, error) {
	contents, err := readInputFile(file)
	if err != nil {
		return nil, err
	}
	return snapshotBytes(contents)
}

func readInputFile(file *os.File) ([]byte, error) {
	if file == nil {
		return nil, errors.New("igraph: input file is nil")
	}
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("igraph: inspect input file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("igraph: input file must be a seekable regular file")
	}
	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("igraph: inspect input file offset: %w", err)
	}
	var contents bytes.Buffer
	buffer := make([]byte, 64*1024)
	for position := offset; ; {
		count, readErr := file.ReadAt(buffer, position)
		if count > 0 {
			_, _ = contents.Write(buffer[:count])
			position += int64(count)
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("igraph: snapshot input file: %w", readErr)
		}
	}
	return contents.Bytes(), nil
}

func snapshotBytes(contents []byte) (*C.FILE, error) {
	stream := C.tmpfile()
	if stream == nil {
		return nil, errors.New("igraph: create input snapshot stream")
	}
	failed := true
	defer func() {
		if failed {
			C.fclose(stream)
		}
	}()
	if len(contents) > 0 {
		written := C.fwrite(unsafe.Pointer(&contents[0]), 1, C.size_t(len(contents)), stream)
		if written != C.size_t(len(contents)) {
			return nil, errors.New("igraph: write input snapshot")
		}
	}
	if C.fseek(stream, 0, C.SEEK_SET) != 0 {
		return nil, errors.New("igraph: rewind input snapshot")
	}
	failed = false
	return stream, nil
}
