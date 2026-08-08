package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
//
// static igraph_bool_t go_igraph_vit_end(const igraph_vit_t *iterator) {
//   return IGRAPH_VIT_END(*iterator);
// }
// static void go_igraph_vit_next(igraph_vit_t *iterator) {
//   IGRAPH_VIT_NEXT(*iterator);
// }
// static igraph_int_t go_igraph_vit_get(const igraph_vit_t *iterator) {
//   return IGRAPH_VIT_GET(*iterator);
// }
// static igraph_int_t go_igraph_vit_size(const igraph_vit_t *iterator) {
//   return IGRAPH_VIT_SIZE(*iterator);
// }
//
// static igraph_bool_t go_igraph_eit_end(const igraph_eit_t *iterator) {
//   return IGRAPH_EIT_END(*iterator);
// }
// static void go_igraph_eit_next(igraph_eit_t *iterator) {
//   IGRAPH_EIT_NEXT(*iterator);
// }
// static igraph_int_t go_igraph_eit_get(const igraph_eit_t *iterator) {
//   return IGRAPH_EIT_GET(*iterator);
// }
// static igraph_int_t go_igraph_eit_size(const igraph_eit_t *iterator) {
//   return IGRAPH_EIT_SIZE(*iterator);
// }
import "C"

// SelectedVertexIDs returns the selected vertex IDs as independent Go-owned
// storage. Selection errors are reported before return. The graph mutex and all
// C iterator resources are held only while this method materializes the slice,
// so callers may stop iterating early or retain the result after graph closure.
func (g *Graph) SelectedVertexIDs(selector VertexSelector) ([]int, error) {
	return g.vertexIDs(selector)
}

// SelectedEdgeIDs returns the selected edge IDs as independent Go-owned
// storage. Its eager error, ownership, and locking behavior matches
// SelectedVertexIDs.
func (g *Graph) SelectedEdgeIDs(selector EdgeSelector) ([]int, error) {
	return g.edgeIDs(selector)
}

type cVertexIterator struct {
	value    C.igraph_vit_t
	selector *cVertexSelector
}

//igraph:internal igraph_vit_create
func newCVertexIterator(graph *C.igraph_t, selector VertexSelector) (*cVertexIterator, error) {
	cSelector, err := newCVertexSelector(selector)
	if err != nil {
		return nil, err
	}
	result := &cVertexIterator{selector: cSelector}
	if code := C.igraph_vit_create(graph, cSelector.value, &result.value); code != C.IGRAPH_SUCCESS {
		cSelector.close()
		return nil, igraphError("initialize vertex iterator", int(code))
	}
	return result, nil
}

func (iterator *cVertexIterator) IDs() ([]int, error) {
	size, err := igraphIntToInt(C.go_igraph_vit_size(&iterator.value), "vertex iterator size")
	if err != nil {
		return nil, err
	}
	result := make([]int, 0, size)
	for C.go_igraph_vit_end(&iterator.value) == booltoint(false) {
		id, err := igraphIntToInt(C.go_igraph_vit_get(&iterator.value), "vertex iterator value")
		if err != nil {
			return nil, err
		}
		result = append(result, id)
		C.go_igraph_vit_next(&iterator.value)
	}
	return result, nil
}

//igraph:internal igraph_vit_destroy
func (iterator *cVertexIterator) close() {
	C.igraph_vit_destroy(&iterator.value)
	iterator.selector.close()
}

func materializeVertexIDs(graph *C.igraph_t, selector VertexSelector) ([]int, error) {
	iterator, err := newCVertexIterator(graph, selector)
	if err != nil {
		return nil, err
	}
	defer iterator.close()
	return iterator.IDs()
}

type cEdgeIterator struct {
	value    C.igraph_eit_t
	selector *cEdgeSelector
}

//igraph:internal igraph_eit_create
func newCEdgeIterator(graph *C.igraph_t, selector EdgeSelector) (*cEdgeIterator, error) {
	cSelector, err := newCEdgeSelector(selector)
	if err != nil {
		return nil, err
	}
	result := &cEdgeIterator{selector: cSelector}
	if code := C.igraph_eit_create(graph, cSelector.value, &result.value); code != C.IGRAPH_SUCCESS {
		cSelector.close()
		return nil, igraphError("initialize edge iterator", int(code))
	}
	return result, nil
}

func (iterator *cEdgeIterator) IDs() ([]int, error) {
	size, err := igraphIntToInt(C.go_igraph_eit_size(&iterator.value), "edge iterator size")
	if err != nil {
		return nil, err
	}
	result := make([]int, 0, size)
	for C.go_igraph_eit_end(&iterator.value) == booltoint(false) {
		id, err := igraphIntToInt(C.go_igraph_eit_get(&iterator.value), "edge iterator value")
		if err != nil {
			return nil, err
		}
		result = append(result, id)
		C.go_igraph_eit_next(&iterator.value)
	}
	return result, nil
}

//igraph:internal igraph_eit_destroy
func (iterator *cEdgeIterator) close() {
	C.igraph_eit_destroy(&iterator.value)
	iterator.selector.close()
}

func materializeEdgeIDs(graph *C.igraph_t, selector EdgeSelector) ([]int, error) {
	iterator, err := newCEdgeIterator(graph, selector)
	if err != nil {
		return nil, err
	}
	defer iterator.close()
	return iterator.IDs()
}
