package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "igraph_error_cgo.h"
// static igraph_error_t go_hrg_sample_many(const igraph_hrg_t *hrg, igraph_graph_list_t *samples, igraph_int_t count) { GO_IGRAPH_CALL(igraph_hrg_sample_many(hrg, samples, count)); }
import "C"

import "fmt"

type hrgSampleAdapters struct {
	model      func(HRGModel) (*cHRG, error)
	closeModel func(*cHRG)
	newList    func() (*graphList, error)
	closeList  func(*graphList)
	run        func(*cHRG, *graphList, int) error
	take       func(*graphList) ([]*Graph, error)
}

func defaultHRGSampleAdapters() hrgSampleAdapters {
	return hrgSampleAdapters{
		model: newCHRG, closeModel: (*cHRG).close,
		newList: newGraphList, closeList: (*graphList).close,
		run: func(h *cHRG, list *graphList, count int) error {
			if code := C.go_hrg_sample_many(&h.value, &list.value, C.igraph_int_t(count)); code != C.IGRAPH_SUCCESS {
				return igraphError("sample graphs from HRG model", int(code))
			}
			return nil
		},
		take: (*graphList).takeGraphs,
	}
}

// Sample generates count independently owned undirected simple graphs from m.
// Vertex IDs preserve the model's leaf ordering. Each result has LeafCount
// vertices, no loops, and no parallel edges. Probability zero never creates
// the corresponding cross-subtree edges; probability one always creates them.
//
// Count must be positive; count == 1 is the one-sample convenience contract.
// Seed optionally resets the package C/igraph RNG, and equal non-nil seeds
// reproduce result edge lists exactly. The receiver is borrowed only during
// this call and reconstructed as temporary C state. The returned non-nil slice
// and every graph in it are independently owned; callers must close each graph.
// Pinned igraph 1.0.1 cannot sample its otherwise valid one-leaf model, so at
// least two leaves are required.
//
//igraph:bind igraph_hrg_sample_many
func (m HRGModel) Sample(count int, seed *uint64) ([]*Graph, error) {
	return m.sample(count, seed, nil)
}

func (m HRGModel) sample(count int, seed *uint64, adapters *hrgSampleAdapters) ([]*Graph, error) {
	if count <= 0 {
		return nil, fmt.Errorf("igraph: HRG sample count must be positive: %d", count)
	}
	if _, err := intToIgraphInt(count, "HRG sample count"); err != nil {
		return nil, err
	}
	validated, err := NewHRGModel(m.left, m.right, m.probabilities, m.edges)
	if err != nil {
		return nil, err
	}
	if validated.LeafCount() < 2 {
		return nil, fmt.Errorf("igraph: HRG sampling requires at least two leaves")
	}
	op := defaultHRGSampleAdapters()
	if adapters != nil {
		op = *adapters
	}
	h, err := op.model(validated)
	if err != nil {
		return nil, err
	}
	defer op.closeModel(h)
	list, err := op.newList()
	if err != nil {
		return nil, err
	}
	defer op.closeList(list)
	var graphs []*Graph
	err = withRNG(seed, func() error {
		if err := op.run(h, list, count); err != nil {
			return err
		}
		graphs, err = op.take(list)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(graphs) != count {
		for _, graph := range graphs {
			_ = graph.Close()
		}
		return nil, fmt.Errorf("igraph: HRG sampling returned %d graphs, want %d", len(graphs), count)
	}
	return graphs, nil
}
