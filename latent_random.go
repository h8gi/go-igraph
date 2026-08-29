package igraph

/*
#include <igraph.h>
#include "random_games_cgo.h"
*/
import "C"

import (
	"fmt"
	"math"
)

// GeometricGraphOptions controls unit-square geometric sampling.
type GeometricGraphOptions struct {
	Seed  *uint64
	Torus bool
}

// LatentGraphOptions controls random-dot-product graph sampling.
type LatentGraphOptions struct {
	Seed     *uint64
	Directed bool
}

// LatentSampleOptions controls latent-vector sampling. Positive restricts
// sphere samples to the positive orthant.
type LatentSampleOptions struct {
	Seed     *uint64
	Positive bool
}

// GeometricGraphResult contains an independently owned graph and a Go-owned
// vertex-ID-aligned matrix with one (x,y) coordinate pair per row.
type GeometricGraphResult struct {
	Graph       *Graph
	Coordinates Matrix
}

func transposeMatrix(input Matrix) (Matrix, error) {
	rows, columns := input.Dims()
	out, err := NewMatrix(columns, rows)
	if err != nil {
		return Matrix{}, err
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < columns; j++ {
			out.values[j*rows+i] = input.values[i*columns+j]
		}
	}
	return out, nil
}

// GeometricRandomGame samples points uniformly in the unit square and joins
// pairs whose Euclidean distance is strictly below radius. Torus uses periodic
// distance. Coordinates remain valid after Graph.Close; Graph must be closed.
//
//igraph:bind igraph_grg_game
func GeometricRandomGame(vertexCount int, radius float64, options GeometricGraphOptions) (GeometricGraphResult, error) {
	if err := validateConstructorSize("vertex count", vertexCount); err != nil {
		return GeometricGraphResult{}, err
	}
	if math.IsNaN(radius) || math.IsInf(radius, 0) || radius < 0 {
		return GeometricGraphResult{}, fmt.Errorf("igraph: geometric radius must be finite and non-negative: %g", radius)
	}
	x, err := newRealVector(nil)
	if err != nil {
		return GeometricGraphResult{}, err
	}
	defer x.close()
	y, err := newRealVector(nil)
	if err != nil {
		return GeometricGraphResult{}, err
	}
	defer y.close()
	var graph C.igraph_t
	err = withRNG(options.Seed, func() error {
		code := C.go_igraph_grg_game(&graph, C.igraph_int_t(vertexCount), C.igraph_real_t(radius), booltoint(options.Torus), &x.value, &y.value)
		if code != C.IGRAPH_SUCCESS {
			return igraphError("igraph_grg_game", int(code))
		}
		return nil
	})
	if err != nil {
		return GeometricGraphResult{}, err
	}
	xs, err := x.slice()
	if err != nil {
		C.igraph_destroy(&graph)
		return GeometricGraphResult{}, err
	}
	ys, err := y.slice()
	if err != nil {
		C.igraph_destroy(&graph)
		return GeometricGraphResult{}, err
	}
	coordinates, err := NewMatrix(vertexCount, 2)
	if err != nil {
		C.igraph_destroy(&graph)
		return GeometricGraphResult{}, err
	}
	for i := 0; i < vertexCount; i++ {
		coordinates.values[i*2] = xs[i]
		coordinates.values[i*2+1] = ys[i]
	}
	return GeometricGraphResult{Graph: adoptInitializedGraph(&graph), Coordinates: coordinates}, nil
}

// DotProductGame samples an edge for each distinct vertex pair with probability
// equal to the dot product of their latent rows. All such dot products are
// checked in [0,1] before C execution; upstream warning/clamping is not exposed.
// positions is borrowed synchronously and Graph must be closed.
//
//igraph:bind igraph_dot_product_game
func DotProductGame(positions Matrix, options LatentGraphOptions) (*Graph, error) {
	vertices, dimensions := positions.Dims()
	for i := 0; i < vertices; i++ {
		for d := 0; d < dimensions; d++ {
			v := positions.values[i*dimensions+d]
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, fmt.Errorf("igraph: latent position[%d,%d] must be finite", i, d)
			}
		}
	}
	for i := 0; i < vertices; i++ {
		for j := i + 1; j < vertices; j++ {
			dot := 0.0
			for d := 0; d < dimensions; d++ {
				a := positions.values[i*dimensions+d]
				b := positions.values[j*dimensions+d]
				dot += a * b
			}
			if math.IsNaN(dot) || math.IsInf(dot, 0) || dot < 0 || dot > 1 {
				return nil, fmt.Errorf("igraph: latent dot product for vertices %d and %d must be in [0,1]: %g", i, j, dot)
			}
		}
	}
	transposed, err := transposeMatrix(positions)
	if err != nil {
		return nil, err
	}
	matrix, err := newCMatrix(transposed)
	if err != nil {
		return nil, err
	}
	defer matrix.close()
	return generateGraph("igraph_dot_product_game", options.Seed, func(graph *C.igraph_t) C.igraph_error_t {
		return C.go_igraph_dot_product_game(graph, &matrix.value, booltoint(options.Directed))
	})
}

func sampledRows(samples, dimensions int, options LatentSampleOptions, operation string, call func(*C.igraph_matrix_t) C.igraph_error_t) (Matrix, error) {
	return sampledRowsWithAdapters(samples, dimensions, options, operation, latentSampleAdapters{
		newMatrix: newCMatrix,
		invoke:    func(output *cMatrix) int { return int(call(&output.value)) },
		convert:   (*cMatrix).matrix,
	})
}

type latentSampleAdapters struct {
	newMatrix func(Matrix) (*cMatrix, error)
	invoke    func(*cMatrix) int
	convert   func(*cMatrix) (Matrix, error)
}

func sampledRowsWithAdapters(samples, dimensions int, options LatentSampleOptions, operation string, adapters latentSampleAdapters) (Matrix, error) {
	if err := validateConstructorSize("sample count", samples); err != nil {
		return Matrix{}, err
	}
	if err := validateConstructorSize("sample dimension", dimensions); err != nil {
		return Matrix{}, err
	}
	if _, err := matrixSize(samples, dimensions); err != nil {
		return Matrix{}, err
	}
	output, err := adapters.newMatrix(Matrix{})
	if err != nil {
		return Matrix{}, err
	}
	defer output.close()
	err = withRNG(options.Seed, func() error {
		code := adapters.invoke(output)
		if code != int(C.IGRAPH_SUCCESS) {
			return igraphError(operation, code)
		}
		return nil
	})
	if err != nil {
		return Matrix{}, err
	}
	columns, err := adapters.convert(output)
	if err != nil {
		return Matrix{}, err
	}
	return transposeMatrix(columns)
}

// SampleDirichlet returns sampleCount rows on the probability simplex. alpha is
// borrowed and supplies one strictly positive finite concentration per column.
//
//igraph:bind igraph_rng_sample_dirichlet
func SampleDirichlet(sampleCount int, alpha []float64, options LatentSampleOptions) (Matrix, error) {
	if len(alpha) < 2 {
		return Matrix{}, fmt.Errorf("igraph: Dirichlet alpha requires at least two values")
	}
	for i, v := range alpha {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
			return Matrix{}, fmt.Errorf("igraph: Dirichlet alpha %d must be finite and positive", i)
		}
	}
	vector, err := newRealVector(alpha)
	if err != nil {
		return Matrix{}, err
	}
	defer vector.close()
	return sampledRows(sampleCount, len(alpha), options, "igraph_rng_sample_dirichlet", func(out *C.igraph_matrix_t) C.igraph_error_t {
		return C.go_igraph_rng_sample_dirichlet(C.igraph_int_t(sampleCount), &vector.value, out)
	})
}

func sampleSphere(sampleCount, dimensions int, radius float64, volume bool, options LatentSampleOptions) (Matrix, error) {
	if dimensions < 2 {
		return Matrix{}, fmt.Errorf("igraph: sphere dimension must be at least two: %d", dimensions)
	}
	if math.IsNaN(radius) || math.IsInf(radius, 0) || radius <= 0 {
		return Matrix{}, fmt.Errorf("igraph: sphere radius must be finite and positive: %g", radius)
	}
	operation := "igraph_rng_sample_sphere_surface"
	if volume {
		operation = "igraph_rng_sample_sphere_volume"
	}
	return sampledRows(sampleCount, dimensions, options, operation, func(out *C.igraph_matrix_t) C.igraph_error_t {
		if volume {
			return C.go_igraph_rng_sample_sphere_volume(C.igraph_int_t(dimensions), C.igraph_int_t(sampleCount), C.igraph_real_t(radius), booltoint(options.Positive), out)
		}
		return C.go_igraph_rng_sample_sphere_surface(C.igraph_int_t(dimensions), C.igraph_int_t(sampleCount), C.igraph_real_t(radius), booltoint(options.Positive), out)
	})
}

// SampleSphereSurface returns sampleCount rows uniformly sampled from a sphere.
//
//igraph:bind igraph_rng_sample_sphere_surface
func SampleSphereSurface(sampleCount, dimensions int, radius float64, options LatentSampleOptions) (Matrix, error) {
	return sampleSphere(sampleCount, dimensions, radius, false, options)
}

// SampleSphereVolume returns sampleCount rows uniformly sampled inside a sphere.
//
//igraph:bind igraph_rng_sample_sphere_volume
func SampleSphereVolume(sampleCount, dimensions int, radius float64, options LatentSampleOptions) (Matrix, error) {
	return sampleSphere(sampleCount, dimensions, radius, true, options)
}
