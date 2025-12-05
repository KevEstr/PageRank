package pagerank

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// Verificar que la versión concurrente produce exactamente los mismos resultados
func TestConcurrentVsSequentialEquality(t *testing.T) {
	tests := []struct {
		name  string
		links [][2]int
	}{
		{
			name: "Simple graph",
			links: [][2]int{
				{0, 1},
				{1, 2},
				{2, 0},
			},
		},
		{
			name: "Wikipedia example",
			links: [][2]int{
				{1, 2}, {2, 1}, {3, 0}, {3, 1}, {4, 3},
				{4, 1}, {4, 5}, {5, 4}, {5, 1}, {6, 1},
				{6, 4}, {7, 1}, {7, 4}, {8, 1}, {8, 4},
				{9, 4}, {10, 4},
			},
		},
		{
			name: "Star graph",
			links: [][2]int{
				{0, 2}, {1, 2}, {2, 2},
			},
		},
		{
			name: "Circular graph",
			links: [][2]int{
				{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 0},
			},
		},
		{
			name: "Converging graph",
			links: [][2]int{
				{0, 1}, {0, 2}, {1, 2}, {2, 2},
			},
		},
		{
			name: "Dangling nodes",
			links: [][2]int{
				{0, 2}, {1, 2},
			},
		},
	}

	const followingProb = 0.85
	const tolerance = 0.0001

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Versión secuencial
			prSeq := New()
			for _, link := range tt.links {
				prSeq.Link(link[0], link[1])
			}

			// Versión concurrente
			prConc := NewConcurrent()
			for _, link := range tt.links {
				prConc.Link(link[0], link[1])
			}

			// Recolectar resultados
			seqResults := make(map[int]float64)
			prSeq.Rank(followingProb, tolerance, func(label int, rank float64) {
				seqResults[label] = rank
			})

			concResults := make(map[int]float64)
			prConc.Rank(followingProb, tolerance, func(label int, rank float64) {
				concResults[label] = rank
			})

			// Comparar resultados
			if len(seqResults) != len(concResults) {
				t.Fatalf("Different number of nodes: seq=%d, conc=%d", len(seqResults), len(concResults))
			}

			for label, seqRank := range seqResults {
				concRank, ok := concResults[label]
				if !ok {
					t.Fatalf("Node %d missing in concurrent results", label)
				}

				diff := math.Abs(seqRank - concRank)
				if diff > 1e-10 {
					t.Errorf("Node %d: sequential=%.15f, concurrent=%.15f, diff=%.15e",
						label, seqRank, concRank, diff)
				}
			}
		})
	}
}

// Test con diferentes números de workers
func TestConcurrentWithDifferentWorkers(t *testing.T) {
	links := [][2]int{
		{1, 2}, {2, 1}, {3, 0}, {3, 1}, {4, 3},
		{4, 1}, {4, 5}, {5, 4}, {5, 1}, {6, 1},
		{6, 4}, {7, 1}, {7, 4}, {8, 1}, {8, 4},
		{9, 4}, {10, 4},
	}

	const followingProb = 0.85
	const tolerance = 0.0001

	// Resultado de referencia (secuencial)
	prSeq := New()
	for _, link := range links {
		prSeq.Link(link[0], link[1])
	}

	referenceResults := make(map[int]float64)
	prSeq.Rank(followingProb, tolerance, func(label int, rank float64) {
		referenceResults[label] = rank
	})

	// Probar con diferentes números de workers
	for _, numWorkers := range []int{1, 2, 4, 8, 16} {
		t.Run(fmt.Sprintf("Workers=%d", numWorkers), func(t *testing.T) {
			prConc := NewConcurrentWithWorkers(numWorkers)
			for _, link := range links {
				prConc.Link(link[0], link[1])
			}

			concResults := make(map[int]float64)
			prConc.Rank(followingProb, tolerance, func(label int, rank float64) {
				concResults[label] = rank
			})

			// Comparar con referencia
			for label, refRank := range referenceResults {
				concRank := concResults[label]
				diff := math.Abs(refRank - concRank)
				if diff > 1e-10 {
					t.Errorf("Workers=%d, Node %d: ref=%.15f, conc=%.15f, diff=%.15e",
						numWorkers, label, refRank, concRank, diff)
				}
			}
		})
	}
}

// Benchmark comparando versiones
func BenchmarkSequentialSmall(b *testing.B) {
	links := [][2]int{
		{1, 2}, {2, 1}, {3, 0}, {3, 1}, {4, 3},
		{4, 1}, {4, 5}, {5, 4}, {5, 1}, {6, 1},
		{6, 4}, {7, 1}, {7, 4}, {8, 1}, {8, 4},
		{9, 4}, {10, 4},
	}

	for i := 0; i < b.N; i++ {
		pr := New()
		for _, link := range links {
			pr.Link(link[0], link[1])
		}
		pr.Rank(0.85, 0.0001, func(label int, rank float64) {})
	}
}

func BenchmarkConcurrentSmall(b *testing.B) {
	links := [][2]int{
		{1, 2}, {2, 1}, {3, 0}, {3, 1}, {4, 3},
		{4, 1}, {4, 5}, {5, 4}, {5, 1}, {6, 1},
		{6, 4}, {7, 1}, {7, 4}, {8, 1}, {8, 4},
		{9, 4}, {10, 4},
	}

	for i := 0; i < b.N; i++ {
		pr := NewConcurrent()
		for _, link := range links {
			pr.Link(link[0], link[1])
		}
		pr.Rank(0.85, 0.0001, func(label int, rank float64) {})
	}
}

// Benchmark con grafo más grande
func BenchmarkSequentialLarge(b *testing.B) {
	n := 1000
	pr := New()
	
	// Crear grafo aleatorio pero reproducible
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			to := (i*7 + j*13) % n
			pr.Link(i, to)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pr.Rank(0.85, 0.001, func(label int, rank float64) {})
	}
}

func BenchmarkConcurrentLarge(b *testing.B) {
	n := 1000
	pr := NewConcurrent()
	
	// Crear grafo aleatorio pero reproducible
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			to := (i*7 + j*13) % n
			pr.Link(i, to)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pr.Rank(0.85, 0.001, func(label int, rank float64) {})
	}
}

// Test funcional completo
func TestConcurrentFunctionalCorrectness(t *testing.T) {
	prConc := NewConcurrent()
	prConc.Link(0, 1)
	prConc.Link(1, 2)
	prConc.Link(2, 0)

	const tolerance = 0.0001
	results := make(map[int]float64)
	
	prConc.Rank(0.85, tolerance, func(label int, rank float64) {
		results[label] = rank
	})

	// Verificar que todos los nodos tienen rank
	if len(results) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(results))
	}

	// Verificar que la suma es aproximadamente 1.0
	sum := 0.0
	for _, rank := range results {
		sum += rank
		if rank <= 0 {
			t.Errorf("Rank should be positive, got %f", rank)
		}
	}

	if math.Abs(sum-1.0) > tolerance {
		t.Errorf("Sum of ranks should be 1.0, got %f", sum)
	}
}

// Función auxiliar para comparación en tiempo real
func CompareBothVersions() {
	n := 100
	
	// Crear mismo grafo en ambas versiones
	prSeq := New()
	prConc := NewConcurrent()
	
	fmt.Println("Construyendo grafo con", n, "nodos...")
	for i := 0; i < n; i++ {
		for j := 0; j < 5; j++ {
			to := (i*3 + j*7) % n
			prSeq.Link(i, to)
			prConc.Link(i, to)
		}
	}

	// Versión secuencial
	fmt.Println("\n=== Versión Secuencial ===")
	startSeq := time.Now()
	seqResults := make(map[int]float64)
	prSeq.Rank(0.85, 0.0001, func(label int, rank float64) {
		seqResults[label] = rank
	})
	durationSeq := time.Since(startSeq)
	fmt.Printf("Tiempo: %v\n", durationSeq)

	// Versión concurrente
	fmt.Println("\n=== Versión Concurrente ===")
	startConc := time.Now()
	concResults := make(map[int]float64)
	prConc.Rank(0.85, 0.0001, func(label int, rank float64) {
		concResults[label] = rank
	})
	durationConc := time.Since(startConc)
	fmt.Printf("Tiempo: %v\n", durationConc)

	// Comparar resultados
	fmt.Println("\n=== Comparación ===")
	maxDiff := 0.0
	for label, seqRank := range seqResults {
		concRank := concResults[label]
		diff := math.Abs(seqRank - concRank)
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	fmt.Printf("Diferencia máxima: %.15e\n", maxDiff)
	fmt.Printf("Speedup: %.2fx\n", float64(durationSeq)/float64(durationConc))
	
	if maxDiff < 1e-10 {
		fmt.Println("✓ Resultados IDÉNTICOS")
	} else {
		fmt.Println("✗ Resultados diferentes")
	}
}
