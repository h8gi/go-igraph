# Network Flow and Connectivity

This document details the Network Flow, Cut, and Connectivity APIs in `go-igraph` delivered in Milestone 7.

## Key Design Principles

1. **Explicit Memory Ownership**:
   - All input slices (`capacities`, `flows`) are borrowed only for the duration of the API call.
   - All returned data structures (`MaxFlowResult`, `MinCutResult`, `STCut`, `GomoryHuTreeResult`, etc.) contain Go-owned slices (`[]float64`, `[]int`) and independently owned `*Graph` instances.
   - Derived graphs and trees (`ResidualGraph`, `GomoryHuTree`, `DominatorTree`, etc.) must be closed by the caller via `graph.Close()`. They remain valid even after the original source graph is closed.

2. **Validation**:
   - Source and target vertex IDs are validated to be distinct and within `[0, NumVertices())`.
   - Optional capacity and flow slices, if provided, must match the graph's edge count (`NumEdges()`) and contain non-negative numbers (no `NaN` or negative values).

## Supported Operations

### Flow & Cuts

- `MaxFlow(source, target int, capacities []float64) (*MaxFlowResult, error)`
- `MaxFlowValue(source, target int, capacities []float64) (float64, error)`
- `STMinCut(source, target int, capacities []float64) (*STMinCutResult, error)`
- `STMinCutValue(source, target int, capacities []float64) (float64, error)`
- `MinCut(capacities []float64) (*MinCutResult, error)`
- `MinCutValue(capacities []float64) (float64, error)`

### Connectivity & Disjoint Paths

- `EdgeConnectivity(checks bool) (int, error)`
- `STEdgeConnectivity(source, target int) (int, error)`
- `VertexConnectivity(checks bool) (int, error)`
- `STVertexConnectivity(source, target int, neighbors VertexConnectivityNeighbors) (int, error)`
- `EdgeDisjointPaths(source, target int) (int, error)`
- `VertexDisjointPaths(source, target int) (int, error)`
- `Adhesion(checks bool) (int, error)`
- `Cohesion(checks bool) (int, error)`

### Cut Enumeration

- `AllSTCuts(source, target int) ([]STCut, error)`
- `AllSTMincuts(source, target int, capacities []float64) (float64, []STCut, error)`

### Derived Flow Graphs & Trees

- `ResidualGraph(capacities []float64, flows []float64) (*ResidualGraphResult, error)`
- `ReverseResidualGraph(capacities []float64, flows []float64) (*Graph, error)`
- `GomoryHuTree(capacities []float64) (*GomoryHuTreeResult, error)`
- `DominatorTree(root int, mode DirectionMode) (*DominatorTreeResult, error)`
- `EvenTarjanReduction() (*TarjanReductionResult, error)`
