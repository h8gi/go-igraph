package igraph

// #cgo pkg-config: igraph
// #include <igraph.h>
// #include "algorithm_cgo.h"
// #include "embedding_cgo.h"
import "C"

import (
	"fmt"
	"math"
)

// SpectralEmbeddingWhich selects which eigenvalues the spectral embedding uses.
type SpectralEmbeddingWhich int

const (
	// EmbeddingLargestMagnitude uses the eigenvalues with the largest magnitude (default).
	EmbeddingLargestMagnitude SpectralEmbeddingWhich = iota
	// EmbeddingLargestAlgebraic uses the largest eigenvalues.
	EmbeddingLargestAlgebraic
	// EmbeddingSmallestAlgebraic uses the smallest eigenvalues.
	EmbeddingSmallestAlgebraic
)

func (w SpectralEmbeddingWhich) cValue() (C.igraph_eigen_which_position_t, error) {
	switch w {
	case EmbeddingLargestMagnitude:
		return C.IGRAPH_EIGEN_LM, nil
	case EmbeddingLargestAlgebraic:
		return C.IGRAPH_EIGEN_LA, nil
	case EmbeddingSmallestAlgebraic:
		return C.IGRAPH_EIGEN_SA, nil
	default:
		return 0, fmt.Errorf("igraph: invalid spectral embedding eigenvalue selection: %d", w)
	}
}

// LaplacianEmbeddingType selects the Laplacian matrix definition used by
// LaplacianSpectralEmbedding. D is the degree matrix, A the adjacency matrix,
// and I the identity matrix. Upstream restricts LaplacianEmbeddingDA,
// LaplacianEmbeddingIDAD, and LaplacianEmbeddingDAD to undirected graphs and
// LaplacianEmbeddingOAP to directed graphs; the mismatched combinations
// return an error.
type LaplacianEmbeddingType int

const (
	// LaplacianEmbeddingDA embeds D - A (default; undirected graphs only).
	LaplacianEmbeddingDA LaplacianEmbeddingType = iota
	// LaplacianEmbeddingIDAD embeds I - D^(-1/2) A D^(-1/2) (undirected graphs only).
	LaplacianEmbeddingIDAD
	// LaplacianEmbeddingDAD embeds D^(-1/2) A D^(-1/2) (undirected graphs only).
	LaplacianEmbeddingDAD
	// LaplacianEmbeddingOAP embeds O^(-1/2) A P^(-1/2), where O and P are the
	// out- and in-degree matrices (directed graphs only).
	LaplacianEmbeddingOAP
)

func (t LaplacianEmbeddingType) cValue() (C.igraph_laplacian_spectral_embedding_type_t, error) {
	switch t {
	case LaplacianEmbeddingDA:
		return C.IGRAPH_EMBEDDING_D_A, nil
	case LaplacianEmbeddingIDAD:
		return C.IGRAPH_EMBEDDING_I_DAD, nil
	case LaplacianEmbeddingDAD:
		return C.IGRAPH_EMBEDDING_DAD, nil
	case LaplacianEmbeddingOAP:
		return C.IGRAPH_EMBEDDING_OAP, nil
	default:
		return 0, fmt.Errorf("igraph: invalid Laplacian embedding type: %d", t)
	}
}

// SpectralEmbeddingOptions controls adjacency and Laplacian spectral embeddings.
// Weights are borrowed for the call, must contain one finite value per edge, and
// nil means unweighted. Solver settings affect only the internal ARPACK run; no
// solver object is exposed or retained.
type SpectralEmbeddingOptions struct {
	// Which selects the eigenvalues used for the embedding.
	Which SpectralEmbeddingWhich
	// Scaled multiplies the eigenvectors by the square root of the singular values when true.
	Scaled bool
	// Weights specifies optional edge weights.
	Weights []float64
	// DegreeCorrection applies only to AdjacencySpectralEmbedding and must be
	// nil for LaplacianSpectralEmbedding. It augments the diagonal of the
	// adjacency matrix with one finite value per vertex; nil uses the
	// upstream-recommended default degree/(V-1).
	DegreeCorrection []float64
	// Type applies only to LaplacianSpectralEmbedding and must be the zero
	// value for AdjacencySpectralEmbedding.
	Type LaplacianEmbeddingType
	// Solver controls ARPACK solver convergence settings.
	Solver SpectralSolverOptions
}

// SpectralEmbeddingResult contains Go-owned embedding coordinates using the
// vertex-to-row convention: row i of X holds the dim-dimensional embedding of
// vertex i. For directed graphs Y holds the second (right singular vector)
// half of the embedding; for undirected graphs Y is an empty matrix.
// SingularValues holds the dim values selected by Which: singular values for
// directed graphs, and the corresponding eigenvalues (possibly negative) for
// undirected graphs. All values remain valid after the graph is closed.
type SpectralEmbeddingResult struct {
	X              Matrix
	Y              Matrix
	SingularValues []float64
}

// AdjacencySpectralEmbedding computes the adjacency spectral embedding of the
// graph into dim dimensions, following Sussman et al. (2012). Option slices
// are borrowed only for the call and returned values are Go-owned.
//
//igraph:bind igraph_adjacency_spectral_embedding
func (g *Graph) AdjacencySpectralEmbedding(dim int, options SpectralEmbeddingOptions) (*SpectralEmbeddingResult, error) {
	if options.Type != LaplacianEmbeddingDA {
		return nil, fmt.Errorf("igraph: Type applies only to LaplacianSpectralEmbedding")
	}
	return g.spectralEmbedding(dim, options, true)
}

// LaplacianSpectralEmbedding computes the Laplacian spectral embedding of the
// graph into dim dimensions; Type selects the Laplacian definition. Option
// slices are borrowed only for the call and returned values are Go-owned.
//
//igraph:bind igraph_laplacian_spectral_embedding
func (g *Graph) LaplacianSpectralEmbedding(dim int, options SpectralEmbeddingOptions) (*SpectralEmbeddingResult, error) {
	if options.DegreeCorrection != nil {
		return nil, fmt.Errorf("igraph: DegreeCorrection applies only to AdjacencySpectralEmbedding")
	}
	return g.spectralEmbedding(dim, options, false)
}

func (g *Graph) spectralEmbedding(dim int, options SpectralEmbeddingOptions, adjacency bool) (*SpectralEmbeddingResult, error) {
	if g == nil {
		return nil, ErrClosed
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return nil, ErrClosed
	}

	numVertices := int(C.igraph_vcount(&g.graph))
	numEdges := int(C.igraph_ecount(&g.graph))

	if dim < 1 {
		return nil, fmt.Errorf("igraph: embedding dimension must be positive: %d", dim)
	}
	if dim > numVertices {
		return nil, fmt.Errorf("igraph: embedding dimension %d exceeds vertex count %d", dim, numVertices)
	}

	which, err := options.Which.cValue()
	if err != nil {
		return nil, err
	}
	embeddingType, err := options.Type.cValue()
	if err != nil {
		return nil, err
	}
	maxIterations, tolerance, err := validateSpectralSolver(options.Solver)
	if err != nil {
		return nil, err
	}

	weights, err := newOptionalEdgeWeights(options.Weights, numEdges)
	if err != nil {
		return nil, err
	}
	if weights != nil {
		defer weights.close()
	}

	var correction *realVector
	if adjacency {
		correction, err = newDegreeCorrection(g, options.DegreeCorrection, numVertices)
		if err != nil {
			return nil, err
		}
		defer correction.close()
	}

	xMat, err := newCMatrix(Matrix{})
	if err != nil {
		return nil, err
	}
	defer xMat.close()
	yMat, err := newCMatrix(Matrix{})
	if err != nil {
		return nil, err
	}
	defer yMat.close()
	singular, err := newRealVectorSize(0)
	if err != nil {
		return nil, err
	}
	defer singular.close()

	// The Y half of the embedding is only meaningful for directed graphs,
	// so it is only requested from upstream when the graph is directed.
	var yPtr *C.igraph_matrix_t
	if bool(C.igraph_is_directed(&g.graph)) {
		yPtr = &yMat.value
	}

	cDim, err := intToIgraphInt(dim, "embedding dimension")
	if err != nil {
		return nil, err
	}

	var code C.igraph_error_t
	operation := "calculate Laplacian spectral embedding"
	if adjacency {
		operation = "calculate adjacency spectral embedding"
		code = C.go_igraph_adjacency_spectral_embedding(
			&g.graph, cDim, edgeWeightPointer(weights), which,
			booltoint(options.Scaled), &xMat.value, yPtr, &singular.value,
			&correction.value, C.int(maxIterations), C.igraph_real_t(tolerance),
		)
	} else {
		code = C.go_igraph_laplacian_spectral_embedding(
			&g.graph, cDim, edgeWeightPointer(weights), which, embeddingType,
			booltoint(options.Scaled), &xMat.value, yPtr, &singular.value,
			C.int(maxIterations), C.igraph_real_t(tolerance),
		)
	}
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError(operation, int(code))
	}

	x, err := xMat.matrix()
	if err != nil {
		return nil, err
	}
	y := Matrix{}
	if yPtr != nil {
		if y, err = yMat.matrix(); err != nil {
			return nil, err
		}
	}
	values, err := singular.slice()
	if err != nil {
		return nil, err
	}
	return &SpectralEmbeddingResult{X: x, Y: y, SingularValues: values}, nil
}

// newDegreeCorrection validates a caller-provided adjacency degree correction
// or computes the upstream-recommended default degree/(V-1). The caller must
// hold g.mu. The returned vector is never nil on success.
func newDegreeCorrection(g *Graph, values []float64, numVertices int) (*realVector, error) {
	if values != nil {
		if len(values) != numVertices {
			return nil, fmt.Errorf("igraph: DegreeCorrection length %d does not match vertex count %d", len(values), numVertices)
		}
		for index, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("igraph: DegreeCorrection at index %d must be finite: %v", index, value)
			}
		}
		return newRealVector(values)
	}

	selector, err := newCVertexSelector(AllVertices())
	if err != nil {
		return nil, err
	}
	defer selector.close()
	degrees, err := newIntVector(nil)
	if err != nil {
		return nil, err
	}
	defer degrees.close()
	code := C.go_igraph_degree(
		&g.graph, &degrees.value, selector.value,
		C.igraph_neimode_t(C.IGRAPH_ALL), C.igraph_loops_t(C.IGRAPH_LOOPS),
	)
	if code != C.IGRAPH_SUCCESS {
		return nil, igraphError("calculate default degree correction", int(code))
	}
	degreeValues, err := degrees.slice()
	if err != nil {
		return nil, err
	}
	denominator := float64(numVertices - 1)
	if denominator <= 0 {
		denominator = 1
	}
	correction := make([]float64, len(degreeValues))
	for index, degree := range degreeValues {
		correction[index] = float64(degree) / denominator
	}
	return newRealVector(correction)
}

// DimSelect selects an embedding dimensionality from a decreasing singular
// value profile using the profile-likelihood method of Zhu and Ghodsi (2006);
// it returns the position of the elbow. The input slice is borrowed, must not
// be empty, and must contain only finite values.
//
//igraph:bind igraph_dim_select
func DimSelect(values []float64) (int, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("igraph: singular value slice must not be empty")
	}
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, fmt.Errorf("igraph: singular value at index %d must be finite: %v", index, value)
		}
	}
	vector, err := newRealVector(values)
	if err != nil {
		return 0, err
	}
	defer vector.close()
	var dim C.igraph_integer_t
	if code := C.go_igraph_dim_select(&vector.value, &dim); code != C.IGRAPH_SUCCESS {
		return 0, igraphError("select embedding dimension", int(code))
	}
	return int(dim), nil
}
