package igraph

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestSimplifyCombinesAttributesAtomically(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}, {0, 1}, {1, 1}}, true)
	defer graph.Close()
	if err := graph.SetGraphStringAttribute("name", "source"); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetVertexBooleanAttributes("active", []bool{true, false}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("weight", []float64{2, 3, 5}); err != nil {
		t.Fatal(err)
	}

	if _, err := graph.SimplifyInPlace(SimplifyOptions{RemoveParallel: true}); err == nil {
		t.Fatal("SimplifyInPlace without an edge policy succeeded")
	}
	if got, _ := graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{2, 3, 5}) {
		t.Fatalf("failed simplification changed attributes: %v", got)
	}

	result, err := graph.SimplifyInPlace(SimplifyOptions{
		RemoveParallel: true,
		RemoveLoops:    true,
		EdgeAttributes: &AttributeCombinationPolicy{Default: AttributeCombineSum},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{5}) {
		t.Fatalf("weights = %v, want [5]", got)
	}
	if got, _ := graph.GraphStringAttribute("name"); got != "source" {
		t.Fatalf("graph name = %q", got)
	}
	if got, _ := graph.VertexBooleanAttributes("active"); !reflect.DeepEqual(got, []bool{true, false}) {
		t.Fatalf("vertex attributes = %v", got)
	}
	if !reflect.DeepEqual(result.Mapping.Edges.OldToNew, []int{0, 0, RemovedID}) {
		t.Fatalf("edge mapping = %v", result.Mapping.Edges.OldToNew)
	}
}

func TestUndirectedConversionCombinesEdgeAttributes(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}, {1, 0}}, true)
	defer graph.Close()
	if err := graph.SetEdgeStringAttributes("label", []string{"out", "back"}); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.ConvertToUndirectedInPlace(UndirectedConversionCollapse, nil); err == nil {
		t.Fatal("collapse without an edge policy succeeded")
	}
	if directed, _ := graph.IsDirected(); !directed {
		t.Fatal("failed collapse changed the graph")
	}
	_, err := graph.ConvertToUndirectedInPlace(UndirectedConversionCollapse, &AttributeCombinationPolicy{
		Default:   AttributeCombineDrop,
		Overrides: map[string]AttributeCombination{"label": AttributeCombineConcat},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := graph.EdgeStringAttributes("label"); !reflect.DeepEqual(got, []string{"outback"}) {
		t.Fatalf("labels = %v", got)
	}
}

func TestLosslessDirectednessConversionsPreserveAttributes(t *testing.T) {
	toUndirected := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	defer toUndirected.Close()
	if err := toUndirected.SetEdgeStringAttributes("label", []string{"edge"}); err != nil {
		t.Fatal(err)
	}
	if _, err := toUndirected.ConvertToUndirectedInPlace(UndirectedConversionEach, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := toUndirected.EdgeStringAttributes("label"); !reflect.DeepEqual(got, []string{"edge"}) {
		t.Fatalf("undirected labels = %v", got)
	}

	toDirected := testGraphFromEdges(t, 2, []Edge{{0, 1}}, false)
	defer toDirected.Close()
	if err := toDirected.SetEdgeNumericAttributes("weight", []float64{2}); err != nil {
		t.Fatal(err)
	}
	if _, err := toDirected.ConvertToDirectedInPlace(DirectedConversionMutual); err != nil {
		t.Fatal(err)
	}
	if got, _ := toDirected.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{2, 2}) {
		t.Fatalf("directed weights = %v", got)
	}
}

func TestBinaryOperatorsCombineAndOwnAttributes(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	defer left.Close()
	defer right.Close()
	for _, item := range []struct {
		graph *Graph
		base  float64
	}{{left, 1}, {right, 10}} {
		if err := item.graph.SetGraphNumericAttribute("score", item.base); err != nil {
			t.Fatal(err)
		}
		if err := item.graph.SetVertexNumericAttributes("score", []float64{item.base, item.base + 1}); err != nil {
			t.Fatal(err)
		}
		if err := item.graph.SetEdgeNumericAttributes("score", []float64{item.base}); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := left.Union(right, nil); err == nil || result.Graph != nil {
		t.Fatalf("union without policy = %#v, %v", result, err)
	}
	policy := &GraphOperatorAttributePolicy{
		Graph:    AttributeCombinationPolicy{Default: AttributeCombineSum},
		Vertices: AttributeCombinationPolicy{Default: AttributeCombineSum},
		Edges:    AttributeCombinationPolicy{Default: AttributeCombineSum},
	}
	result, err := left.Union(right, policy)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	if got, _ := result.Graph.GraphNumericAttribute("score"); got != 11 {
		t.Fatalf("graph score = %v", got)
	}
	if got, _ := result.Graph.VertexNumericAttributes("score"); !reflect.DeepEqual(got, []float64{11, 13}) {
		t.Fatalf("vertex scores = %v", got)
	}
	if got, _ := result.Graph.EdgeNumericAttributes("score"); !reflect.DeepEqual(got, []float64{11}) {
		t.Fatalf("edge scores = %v", got)
	}
	intersection, err := left.Intersection(right, policy)
	if err != nil {
		t.Fatal(err)
	}
	defer intersection.Graph.Close()
	if got, _ := intersection.Graph.EdgeNumericAttributes("score"); !reflect.DeepEqual(got, []float64{11}) {
		t.Fatalf("intersection edge scores = %v", got)
	}
	if err := left.SetGraphNumericAttribute("score", 99); err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Graph.GraphNumericAttribute("score"); got != 11 {
		t.Fatalf("result aliases source attributes: %v", got)
	}
	if err := left.Close(); err != nil {
		t.Fatal(err)
	}
	if err := right.Close(); err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Graph.EdgeNumericAttributes("score"); !reflect.DeepEqual(got, []float64{11}) {
		t.Fatalf("result after operand closure = %v", got)
	}
}

func TestDisjointUnionPreservesNonConflictingElementValues(t *testing.T) {
	left := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 2, []Edge{{1, 0}}, true)
	defer left.Close()
	defer right.Close()
	if err := left.SetVertexStringAttributes("label", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := right.SetVertexStringAttributes("label", []string{"c", "d"}); err != nil {
		t.Fatal(err)
	}
	if err := left.SetGraphStringAttribute("name", "left"); err != nil {
		t.Fatal(err)
	}
	if err := right.SetGraphStringAttribute("name", "right"); err != nil {
		t.Fatal(err)
	}
	if result, err := left.DisjointUnion(right, nil); err == nil || result.Graph != nil {
		t.Fatalf("disjoint union graph conflict = %#v, %v", result, err)
	}
	result, err := left.DisjointUnion(right, &GraphOperatorAttributePolicy{
		Graph: AttributeCombinationPolicy{Default: AttributeCombineConcat},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	if got, _ := result.Graph.VertexStringAttributes("label"); !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("labels = %v", got)
	}
	if got, _ := result.Graph.GraphStringAttribute("name"); got != "leftright" {
		t.Fatalf("name = %q", got)
	}
}

func TestCompositionCombinesContributingEdgeAttributes(t *testing.T) {
	left := testGraphFromEdges(t, 3, []Edge{{0, 1}}, true)
	right := testGraphFromEdges(t, 3, []Edge{{1, 2}}, true)
	defer left.Close()
	defer right.Close()
	if err := left.SetEdgeNumericAttributes("weight", []float64{2}); err != nil {
		t.Fatal(err)
	}
	if err := right.SetEdgeNumericAttributes("weight", []float64{3}); err != nil {
		t.Fatal(err)
	}
	if err := left.SetGraphStringAttribute("name", "left"); err != nil {
		t.Fatal(err)
	}
	if err := right.SetGraphStringAttribute("name", "right"); err != nil {
		t.Fatal(err)
	}
	if err := left.SetVertexBooleanAttributes("active", []bool{true, false, true}); err != nil {
		t.Fatal(err)
	}
	if err := right.SetVertexBooleanAttributes("active", []bool{false, true, false}); err != nil {
		t.Fatal(err)
	}
	if err := left.SetEdgeBooleanAttributes("active", []bool{false}); err != nil {
		t.Fatal(err)
	}
	if err := right.SetEdgeBooleanAttributes("active", []bool{true}); err != nil {
		t.Fatal(err)
	}
	result, err := left.Compose(right, &GraphOperatorAttributePolicy{
		Graph:    AttributeCombinationPolicy{Default: AttributeCombineConcat},
		Vertices: AttributeCombinationPolicy{Default: AttributeCombineLast},
		Edges: AttributeCombinationPolicy{
			Default:   AttributeCombineProduct,
			Overrides: map[string]AttributeCombination{"active": AttributeCombineLast},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Graph.Close()
	if got, _ := result.Graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{6}) {
		t.Fatalf("weights = %v", got)
	}
	if got, _ := result.Graph.GraphStringAttribute("name"); got != "leftright" {
		t.Fatalf("name = %q", got)
	}
	if got, _ := result.Graph.VertexBooleanAttributes("active"); !reflect.DeepEqual(got, []bool{false, true, false}) {
		t.Fatalf("active vertices = %v", got)
	}
	if got, _ := result.Graph.EdgeBooleanAttributes("active"); !reflect.DeepEqual(got, []bool{true}) {
		t.Fatalf("active edges = %v", got)
	}
}

func TestAttributeCombinationValidationAndPartialCleanup(t *testing.T) {
	metadata := []AttributeMetadata{{Name: "weight", Scope: AttributeEdge, Type: AttributeNumeric}}
	policy := &AttributeCombinationPolicy{
		Default:   AttributeCombineFirst,
		Overrides: map[string]AttributeCombination{"weight": AttributeCombineSum},
	}
	initError := errors.New("init")
	if result, err := newAttributeCombinationWithHooks(policy, metadata, attributeCombinationHooks{init: func() error { return initError }}); !errors.Is(err, initError) || result != nil {
		t.Fatalf("init failure = %#v, %v", result, err)
	}
	addError := errors.New("add")
	destroyed := 0
	if result, err := newAttributeCombinationWithHooks(policy, metadata, attributeCombinationHooks{
		init:    func() error { return nil },
		add:     func(string, AttributeCombination) error { return addError },
		destroy: func() { destroyed++ },
	}); !errors.Is(err, addError) || result != nil || destroyed != 1 {
		t.Fatalf("partial add failure = %#v, %v, destroyed %d", result, err, destroyed)
	}

	if err := validateCombinationPolicy(&AttributeCombinationPolicy{Default: AttributeCombineConcat}, metadata); err == nil {
		t.Fatal("string concatenation accepted for numeric attributes")
	}
	if err := validateCombinationType(AttributeCombination(255), AttributeNumeric); err == nil {
		t.Fatal("invalid combination type accepted")
	}
	if err := validateCombinationPolicy(&AttributeCombinationPolicy{
		Overrides: map[string]AttributeCombination{"bad\x00name": AttributeCombineFirst},
	}, metadata); err == nil {
		t.Fatal("invalid override name accepted")
	}
	if err := validateCombinationPolicy(&AttributeCombinationPolicy{
		Overrides: map[string]AttributeCombination{"weight": AttributeCombineConcat},
	}, metadata); err == nil {
		t.Fatal("invalid override type accepted")
	}
	if err := validateCombinationPolicy(&AttributeCombinationPolicy{Overrides: map[string]AttributeCombination{"missing": AttributeCombineFirst}}, metadata); !errors.Is(err, ErrAttributeNotFound) {
		t.Fatalf("missing override error = %v", err)
	}
}

func TestLosslessDerivedGraphsPreserveIndependentAttributes(t *testing.T) {
	source := testGraphFromEdges(t, 4, []Edge{{0, 1}, {2, 3}}, false)
	if err := source.SetGraphStringAttribute("name", "source"); err != nil {
		t.Fatal(err)
	}
	if err := source.SetVertexStringAttributes("label", []string{"a", "b", "c", "d"}); err != nil {
		t.Fatal(err)
	}
	if err := source.SetEdgeNumericAttributes("weight", []float64{1, 2}); err != nil {
		t.Fatal(err)
	}

	clone, err := source.Clone()
	if err != nil {
		t.Fatal(err)
	}
	defer clone.Close()
	selector, _ := VertexIDs(2, 3)
	subgraph, err := source.InducedSubgraph(selector)
	if err != nil {
		t.Fatal(err)
	}
	defer subgraph.Graph.Close()
	edgeSelector, _ := EdgeIDs(1)
	edgeSubgraph, err := source.EdgeSubgraph(edgeSelector, false)
	if err != nil {
		t.Fatal(err)
	}
	defer edgeSubgraph.Graph.Close()
	components, err := source.Decompose(DecomposeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, component := range components {
			_ = component.Close()
		}
	}()

	if err := source.SetGraphStringAttribute("name", "changed"); err != nil {
		t.Fatal(err)
	}
	if got, _ := clone.GraphStringAttribute("name"); got != "source" {
		t.Fatalf("clone graph attribute = %q", got)
	}
	if got, _ := subgraph.Graph.VertexStringAttributes("label"); !reflect.DeepEqual(got, []string{"c", "d"}) {
		t.Fatalf("subgraph labels = %v", got)
	}
	if got, _ := subgraph.Graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{2}) {
		t.Fatalf("subgraph weights = %v", got)
	}
	if got, _ := edgeSubgraph.Graph.VertexStringAttributes("label"); !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("edge-subgraph labels = %v", got)
	}
	if got, _ := edgeSubgraph.Graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{2}) {
		t.Fatalf("edge-subgraph weights = %v", got)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	for _, component := range components {
		if got, _ := component.GraphStringAttribute("name"); got != "source" {
			t.Fatalf("component graph attribute = %q", got)
		}
		if metadata, err := component.VertexAttributes(); err != nil || len(metadata) != 1 {
			t.Fatalf("component vertex metadata = %v, %v", metadata, err)
		}
		if metadata, err := component.EdgeAttributes(); err != nil || len(metadata) != 1 {
			t.Fatalf("component edge metadata = %v, %v", metadata, err)
		}
	}
	if len(components) == 2 {
		if err := components[0].SetGraphStringAttribute("name", "first"); err != nil {
			t.Fatal(err)
		}
		if got, _ := components[1].GraphStringAttribute("name"); got != "source" {
			t.Fatalf("component attributes alias a sibling: %q", got)
		}
	}
}

func TestDeletionPreservesAlignedAttributes(t *testing.T) {
	graph := testGraphFromEdges(t, 3, []Edge{{0, 1}, {1, 2}}, true)
	defer graph.Close()
	if err := graph.SetVertexStringAttributes("label", []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetEdgeNumericAttributes("weight", []float64{1, 2}); err != nil {
		t.Fatal(err)
	}
	edges, _ := EdgeIDs(0)
	if _, err := graph.DeleteEdges(edges); err != nil {
		t.Fatal(err)
	}
	if got, _ := graph.VertexStringAttributes("label"); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("vertex labels = %v", got)
	}
	if got, _ := graph.EdgeNumericAttributes("weight"); !reflect.DeepEqual(got, []float64{2}) {
		t.Fatalf("edge weights = %v", got)
	}
}

func TestAttributeCombinationRules(t *testing.T) {
	var nilOperatorPolicy *GraphOperatorAttributePolicy
	if nilOperatorPolicy.scope(AttributeGraph) != nil {
		t.Fatal("nil operator policy returned a scope")
	}
	operatorPolicy := &GraphOperatorAttributePolicy{}
	if operatorPolicy.scope(AttributeScope(255)) != nil {
		t.Fatal("invalid operator policy scope accepted")
	}
	if (graphAttributeSnapshot{}).scope(AttributeScope(255)) != nil {
		t.Fatal("invalid snapshot scope accepted")
	}
	if err := validateCombinationType(AttributeCombineSum, AttributeBoolean); err == nil {
		t.Fatal("numeric combination accepted for Boolean attributes")
	}
	rules := []AttributeCombination{
		AttributeCombineDrop, AttributeCombineFirst, AttributeCombineLast,
		AttributeCombineSum, AttributeCombineProduct, AttributeCombineMinimum,
		AttributeCombineMaximum, AttributeCombineMean, AttributeCombineConcat,
	}
	for _, rule := range rules {
		if _, err := rule.cValue(); err != nil {
			t.Errorf("rule %d: %v", rule, err)
		}
	}
	if _, err := AttributeCombination(255).cValue(); err == nil {
		t.Fatal("invalid combination accepted")
	}

	numbers := []float64{4, 1, 3, 2}
	wants := map[AttributeCombination]float64{
		AttributeCombineFirst:   4,
		AttributeCombineLast:    2,
		AttributeCombineSum:     10,
		AttributeCombineProduct: 24,
		AttributeCombineMinimum: 1,
		AttributeCombineMaximum: 4,
		AttributeCombineMean:    2.5,
	}
	for rule, want := range wants {
		if got := combineNumbers(numbers, rule); got != want {
			t.Errorf("combineNumbers(%d) = %v, want %v", rule, got, want)
		}
	}
	if got := combineNumbers(numbers, AttributeCombineDrop); got != numbers[0] {
		t.Fatalf("fallback number = %v", got)
	}

	strings, err := combineAttributeValues(AttributeString, [][]attributeScalar{
		{{stringV: "a"}, {stringV: "b"}}, {},
	}, AttributeCombineConcat)
	if err != nil || !reflect.DeepEqual(strings.strings, []string{"ab", ""}) {
		t.Fatalf("combined strings = %#v, %v", strings, err)
	}
	booleans, err := combineAttributeValues(AttributeBoolean, [][]attributeScalar{
		{{boolean: false}, {boolean: true}}, {},
	}, AttributeCombineLast)
	if err != nil || !reflect.DeepEqual(booleans.booleans, []bool{true, false}) {
		t.Fatalf("combined booleans = %#v, %v", booleans, err)
	}
	firstStrings, _ := combineAttributeValues(AttributeString, [][]attributeScalar{{{stringV: "a"}, {stringV: "b"}}}, AttributeCombineFirst)
	lastStrings, _ := combineAttributeValues(AttributeString, [][]attributeScalar{{{stringV: "a"}, {stringV: "b"}}}, AttributeCombineLast)
	firstBooleans, _ := combineAttributeValues(AttributeBoolean, [][]attributeScalar{{{boolean: true}, {boolean: false}}}, AttributeCombineFirst)
	if firstStrings.strings[0] != "a" || lastStrings.strings[0] != "b" || !firstBooleans.booleans[0] {
		t.Fatalf("first/last combinations = %v, %v, %v", firstStrings.strings, lastStrings.strings, firstBooleans.booleans)
	}
	emptyNumeric, err := combineAttributeValues(AttributeNumeric, [][]attributeScalar{{}}, AttributeCombineFirst)
	if err != nil || len(emptyNumeric.numeric) != 1 || !math.IsNaN(emptyNumeric.numeric[0]) {
		t.Fatalf("empty numeric = %#v, %v", emptyNumeric, err)
	}
	if _, err := combineAttributeValues(AttributeType(255), nil, AttributeCombineFirst); err == nil {
		t.Fatal("unsupported type accepted")
	}
}

func TestRawAttributeRestorationSupportsEveryTypeAndScope(t *testing.T) {
	graph := testGraphFromEdges(t, 2, []Edge{{0, 1}}, true)
	defer graph.Close()
	tests := []struct {
		scope AttributeScope
		name  string
		value attributeValues
	}{
		{AttributeGraph, "gn", attributeValues{typeOf: AttributeNumeric, numeric: []float64{1}}},
		{AttributeGraph, "gs", attributeValues{typeOf: AttributeString, strings: []string{"x"}}},
		{AttributeGraph, "gb", attributeValues{typeOf: AttributeBoolean, booleans: []bool{true}}},
		{AttributeVertex, "vn", attributeValues{typeOf: AttributeNumeric, numeric: []float64{1, 2}}},
		{AttributeVertex, "vs", attributeValues{typeOf: AttributeString, strings: []string{"a", "b"}}},
		{AttributeVertex, "vb", attributeValues{typeOf: AttributeBoolean, booleans: []bool{true, false}}},
		{AttributeEdge, "en", attributeValues{typeOf: AttributeNumeric, numeric: []float64{3}}},
		{AttributeEdge, "es", attributeValues{typeOf: AttributeString, strings: []string{"c"}}},
		{AttributeEdge, "eb", attributeValues{typeOf: AttributeBoolean, booleans: []bool{true}}},
	}
	for _, test := range tests {
		if err := setRawAttributeValues(graph, test.scope, test.name, test.value); err != nil {
			t.Fatalf("set %s: %v", test.name, err)
		}
	}
	if got, _ := graph.GraphNumericAttribute("gn"); got != 1 {
		t.Errorf("gn = %v", got)
	}
	if got, _ := graph.GraphStringAttribute("gs"); got != "x" {
		t.Errorf("gs = %q", got)
	}
	if got, _ := graph.GraphBooleanAttribute("gb"); !got {
		t.Errorf("gb = %v", got)
	}
	if got, _ := graph.VertexBooleanAttributes("vb"); !reflect.DeepEqual(got, []bool{true, false}) {
		t.Errorf("vb = %v", got)
	}
	if got, _ := graph.EdgeStringAttributes("es"); !reflect.DeepEqual(got, []string{"c"}) {
		t.Errorf("es = %v", got)
	}
}

func TestOperatorAttributePolicyValidation(t *testing.T) {
	graph := testGraphFromEdges(t, 1, nil, true)
	defer graph.Close()
	numeric := map[string]attributeValues{
		"value": {typeOf: AttributeNumeric, numeric: []float64{1}},
	}
	strings := map[string]attributeValues{
		"value": {typeOf: AttributeString, strings: []string{"one"}},
	}
	sources := []mappedAttributeSource{{numeric, []int{0}}, {strings, []int{0}}}
	if err := restoreMappedAttributes(graph, AttributeVertex, 1, sources, nil); err == nil {
		t.Fatal("type conflict without a policy succeeded")
	}
	if err := restoreMappedAttributes(graph, AttributeVertex, 1, sources, &AttributeCombinationPolicy{Default: AttributeCombineFirst}); !errors.Is(err, ErrAttributeTypeMismatch) {
		t.Fatalf("type conflict error = %v", err)
	}
	if err := restoreMappedAttributes(graph, AttributeVertex, 1, sources, &AttributeCombinationPolicy{Default: AttributeCombineDrop}); err != nil {
		t.Fatalf("drop type conflict: %v", err)
	}
	if err := restoreMappedAttributes(graph, AttributeEdge, 0, nil, &AttributeCombinationPolicy{
		Overrides: map[string]AttributeCombination{"missing": AttributeCombineFirst},
	}); !errors.Is(err, ErrAttributeNotFound) {
		t.Fatalf("unknown override error = %v", err)
	}
	if err := restoreMappedAttributes(graph, AttributeEdge, 0, nil, &AttributeCombinationPolicy{
		Overrides: map[string]AttributeCombination{"bad\x00name": AttributeCombineFirst},
	}); err == nil {
		t.Fatal("invalid operator override name accepted")
	}
	if err := restoreMappedAttributes(graph, AttributeEdge, 0, nil, &AttributeCombinationPolicy{Default: AttributeCombination(255)}); err == nil {
		t.Fatal("invalid default accepted")
	}

	unique := []mappedAttributeSource{{numeric, []int{0}}}
	if err := restoreMappedAttributes(graph, AttributeVertex, 1, unique, nil); err != nil {
		t.Fatalf("unique value restore: %v", err)
	}
	if got, _ := graph.VertexNumericAttributes("value"); !reflect.DeepEqual(got, []float64{1}) {
		t.Fatalf("unique value = %v", got)
	}
}

func TestAttributeCombinationDefaultAddFailureCleanup(t *testing.T) {
	metadata := []AttributeMetadata{{Name: "weight", Scope: AttributeEdge, Type: AttributeNumeric}}
	policy := &AttributeCombinationPolicy{
		Default:   AttributeCombineFirst,
		Overrides: map[string]AttributeCombination{"weight": AttributeCombineSum},
	}
	forced := errors.New("default add")
	adds := 0
	destroyed := 0
	result, err := newAttributeCombinationWithHooks(policy, metadata, attributeCombinationHooks{
		init: func() error { return nil },
		add: func(string, AttributeCombination) error {
			adds++
			if adds == 2 {
				return forced
			}
			return nil
		},
		destroy: func() { destroyed++ },
	})
	if !errors.Is(err, forced) || result != nil || adds != 2 || destroyed != 1 {
		t.Fatalf("default failure = %#v, %v, adds %d, destroyed %d", result, err, adds, destroyed)
	}
	var nilCombination *attributeCombination
	nilCombination.close()
}
