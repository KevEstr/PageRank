package pagerank

import (
	"fmt"
	"math"
	"runtime"
	"sync"
)

// Umbral mínimo de elementos para justificar paralelización
// Por debajo de este valor, el overhead de goroutines supera el beneficio
const parallelizationThreshold = 5000

type pageRankConcurrent struct {
	inLinks               [][]int
	numberOutLinks        []int
	currentAvailableIndex int
	keyToIndex            map[int]int
	indexToKey            map[int]int
	numWorkers            int
}

// workChunk representa un rango de trabajo para un worker
type workChunk struct {
	start int
	end   int
}

// calculateWorkChunks divide el trabajo en chunks balanceados para los workers
func (pr *pageRankConcurrent) calculateWorkChunks(totalSize int) ([]workChunk, int) {
	if totalSize < parallelizationThreshold {
		// No paralelizar si es muy pequeño
		return []workChunk{{start: 0, end: totalSize}}, 1
	}

	numWorkers := pr.numWorkers
	chunkSize := totalSize / numWorkers

	if chunkSize == 0 {
		// Si hay más workers que elementos, ajustar
		chunkSize = 1
		numWorkers = totalSize
	}

	chunks := make([]workChunk, numWorkers)
	for w := 0; w < numWorkers; w++ {
		chunks[w].start = w * chunkSize
		chunks[w].end = chunks[w].start + chunkSize
		if w == numWorkers-1 {
			// El último worker toma el resto
			chunks[w].end = totalSize
		}
	}

	return chunks, numWorkers
}

func NewConcurrent() *pageRankConcurrent {
	pr := new(pageRankConcurrent)
	pr.numWorkers = runtime.NumCPU()
	pr.Clear()
	return pr
}

func NewConcurrentWithWorkers(numWorkers int) *pageRankConcurrent {
	pr := new(pageRankConcurrent)
	pr.numWorkers = numWorkers
	pr.Clear()
	return pr
}

func (pr *pageRankConcurrent) String() string {
	return fmt.Sprintf(
		"PageRank Concurrent Struct:\n"+
			"InLinks: %v\n"+
			"NumberOutLinks: %v\n"+
			"CurrentAvailableIndex: %d\n"+
			"KeyToIndex: %v\n"+
			"IndexToKey: %v\n"+
			"NumWorkers: %d",
		pr.inLinks,
		pr.numberOutLinks,
		pr.currentAvailableIndex,
		pr.keyToIndex,
		pr.indexToKey,
		pr.numWorkers,
	)
}

func (pr *pageRankConcurrent) keyAsArrayIndex(key int) int {
	index, ok := pr.keyToIndex[key]

	if !ok {
		index = pr.currentAvailableIndex
		pr.currentAvailableIndex++
		pr.keyToIndex[key] = index
		pr.indexToKey[index] = key
	}

	return index
}

func (pr *pageRankConcurrent) updateInLinks(fromAsIndex, toAsIndex int) {
	missingSlots := len(pr.keyToIndex) - len(pr.inLinks)

	if missingSlots > 0 {
		pr.inLinks = append(pr.inLinks, make([][]int, missingSlots)...)
	}

	pr.inLinks[toAsIndex] = append(pr.inLinks[toAsIndex], fromAsIndex)
}

func (pr *pageRankConcurrent) updateNumberOutLinks(fromAsIndex int) {
	missingSlots := len(pr.keyToIndex) - len(pr.numberOutLinks)

	if missingSlots > 0 {
		pr.numberOutLinks = append(pr.numberOutLinks, make([]int, missingSlots)...)
	}

	pr.numberOutLinks[fromAsIndex] += 1
}

func (pr *pageRankConcurrent) linkWithIndices(fromAsIndex, toAsIndex int) {
	pr.updateInLinks(fromAsIndex, toAsIndex)
	pr.updateNumberOutLinks(fromAsIndex)
}

func (pr *pageRankConcurrent) Link(from, to int) {
	fromAsIndex := pr.keyAsArrayIndex(from)
	toAsIndex := pr.keyAsArrayIndex(to)

	pr.linkWithIndices(fromAsIndex, toAsIndex)
}

func (pr *pageRankConcurrent) calculateDanglingNodes() []int {
	danglingNodes := make([]int, 0, len(pr.numberOutLinks))

	for i, numberOutLinksForI := range pr.numberOutLinks {
		if numberOutLinksForI == 0 {
			danglingNodes = append(danglingNodes, i)
		}
	}

	return danglingNodes
}

// Calcula el inner product de forma concurrente
func (pr *pageRankConcurrent) calculateInnerProductConcurrent(p []float64, danglingNodes []int) float64 {
	if len(danglingNodes) == 0 {
		return 0.0
	}

	chunks, numWorkers := pr.calculateWorkChunks(len(danglingNodes))

	// Si solo hay un chunk, ejecutar secuencialmente (más eficiente)
	if numWorkers == 1 {
		innerProduct := 0.0
		for _, danglingNode := range danglingNodes {
			innerProduct += p[danglingNode]
		}
		return innerProduct
	}

	// Ejecución paralela con arrays locales (mejor cache locality)
	var wg sync.WaitGroup
	results := make([]float64, numWorkers)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int, chunk workChunk) {
			defer wg.Done()
			localSum := 0.0
			// Optimización: reducir bounds checking
			for i := chunk.start; i < chunk.end; i++ {
				localSum += p[danglingNodes[i]]
			}
			results[workerID] = localSum
		}(w, chunks[w])
	}

	wg.Wait()

	// Reducción final
	innerProduct := 0.0
	for w := 0; w < numWorkers; w++ {
		innerProduct += results[w]
	}

	return innerProduct
}

// Calcula el vector v de forma concurrente - AQUÍ ESTÁ EL MAYOR BENEFICIO
func (pr *pageRankConcurrent) step(followingProb, tOverSize float64, p []float64, danglingNodes []int) []float64 {
	innerProduct := pr.calculateInnerProductConcurrent(p, danglingNodes)
	innerProductOverSize := innerProduct / float64(len(p))

	v := make([]float64, len(p))
	chunks, numWorkers := pr.calculateWorkChunks(len(pr.inLinks))

	// Si solo hay un chunk, ejecutar secuencialmente
	if numWorkers == 1 {
		vsum := 0.0
		for i := range pr.inLinks {
			ksum := 0.0
			for _, index := range pr.inLinks[i] {
				ksum += p[index] / float64(pr.numberOutLinks[index])
			}
			v[i] = followingProb*(ksum+innerProductOverSize) + tOverSize
			vsum += v[i]
		}

		// Normalización secuencial
		inverseOfSum := 1.0 / vsum
		for i := range v {
			v[i] *= inverseOfSum
		}
		return v
	}

	// Ejecución paralela: Fase 1 - Calcular v y suma parcial
	// OPTIMIZACIÓN CRÍTICA: Pre-calcular 1.0/numberOutLinks para evitar divisiones repetidas
	inverseOutLinks := make([]float64, len(pr.numberOutLinks))
	for i, outLinks := range pr.numberOutLinks {
		if outLinks > 0 {
			inverseOutLinks[i] = 1.0 / float64(outLinks)
		}
	}

	vsumParts := make([]float64, numWorkers)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int, chunk workChunk) {
			defer wg.Done()
			localVsum := 0.0
			// OPTIMIZACIÓN: Mejor cache locality
			for i := chunk.start; i < chunk.end; i++ {
				ksum := 0.0
				inLinks := pr.inLinks[i]
				// Inner loop optimizado - acceso secuencial
				for j := 0; j < len(inLinks); j++ {
					index := inLinks[j]
					ksum += p[index] * inverseOutLinks[index]
				}
				v[i] = followingProb*(ksum+innerProductOverSize) + tOverSize
				localVsum += v[i]
			}
			vsumParts[workerID] = localVsum
		}(w, chunks[w])
	}

	wg.Wait()

	// Reducción: calcular vsum total
	vsum := 0.0
	for w := 0; w < numWorkers; w++ {
		vsum += vsumParts[w]
	}
	inverseOfSum := 1.0 / vsum

	// Fase 2 - Normalización paralela usando los mismos chunks
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int, chunk workChunk) {
			defer wg.Done()
			for i := chunk.start; i < chunk.end; i++ {
				v[i] *= inverseOfSum
			}
		}(w, chunks[w])
	}

	wg.Wait()

	return v
}

// Calcula el cambio de forma concurrente
func (pr *pageRankConcurrent) calculateChangeConcurrent(p, new_p []float64) float64 {
	chunks, numWorkers := pr.calculateWorkChunks(len(p))

	// Si solo hay un chunk, ejecutar secuencialmente
	if numWorkers == 1 {
		acc := 0.0
		for i := range p {
			acc += math.Abs(p[i] - new_p[i])
		}
		return acc
	}

	// Ejecución paralela
	results := make([]float64, numWorkers)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int, chunk workChunk) {
			defer wg.Done()
			localAcc := 0.0
			// Mejor localidad de caché
			for i := chunk.start; i < chunk.end; i++ {
				diff := p[i] - new_p[i]
				if diff < 0 {
					diff = -diff
				}
				localAcc += diff
			}
			results[workerID] = localAcc
		}(w, chunks[w])
	}

	wg.Wait()

	// Reducción final
	acc := 0.0
	for w := 0; w < numWorkers; w++ {
		acc += results[w]
	}

	return acc
}

func (pr *pageRankConcurrent) Rank(followingProb, tolerance float64, resultFunc func(label int, rank float64)) {
	size := len(pr.keyToIndex)
	inverseOfSize := 1.0 / float64(size)
	tOverSize := (1.0 - followingProb) / float64(size)
	danglingNodes := pr.calculateDanglingNodes()

	p := make([]float64, size)
	for i := range p {
		p[i] = inverseOfSize
	}

	change := 2.0

	for change > tolerance {
		new_p := pr.step(followingProb, tOverSize, p, danglingNodes)
		change = pr.calculateChangeConcurrent(p, new_p)
		p = new_p
	}

	for i, pForI := range p {
		resultFunc(pr.indexToKey[i], pForI)
	}
}

func (pr *pageRankConcurrent) Clear() {
	pr.inLinks = [][]int{}
	pr.numberOutLinks = []int{}
	pr.currentAvailableIndex = 0
	pr.keyToIndex = make(map[int]int)
	pr.indexToKey = make(map[int]int)
}
