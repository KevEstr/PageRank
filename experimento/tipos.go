package experimento

import "time"

// TamanoGrafo representa los diferentes tamaños de grafos para los bloques
type TamanoGrafo string

const (
	// Bloque 1: Pequeño - 20,000 nodos
	// Baseline para ver comportamiento inicial
	Pequeno TamanoGrafo = "pequeno" // 20,000 nodos
	
	// Bloque 2: Mediano - 100,000 nodos (5x más grande)
	// Aquí se debe ver clara ventaja de concurrencia
	Mediano TamanoGrafo = "mediano" // 100,000 nodos
	
	// Bloque 3: Grande - 500,000 nodos (5x más grande)
	// Máxima diferencia observable en speedup
	Grande  TamanoGrafo = "grande"  // 500,000 nodos
)

// TipoImplementacion representa el tipo de implementación a probar
type TipoImplementacion string

const (
	Secuencial  TipoImplementacion = "secuencial"
	Concurrente TipoImplementacion = "concurrente"
)

// ConfigGrafo define la configuración para generar un grafo
type ConfigGrafo struct {
	NumNodos          int     // Número de nodos en el grafo
	EnlacesPorNodo    int     // Promedio de enlaces salientes por nodo
	ProbabilidadHub   float64 // Probabilidad de que un nodo sea un hub (reciba muchos enlaces)
	Seed              int64   // Semilla para reproducibilidad
}

// ConfigExperimento define los parámetros del experimento
type ConfigExperimento struct {
	DampingFactor float64 // Factor de amortiguación (típicamente 0.85)
	Tolerance     float64 // Tolerancia para convergencia
	NumGoroutines int     // Número de goroutines (solo para versión concurrente)
	NumReplicas   int     // Número de réplicas por combinación
}

// ResultadoEjecucion almacena las métricas de una ejecución
type ResultadoEjecucion struct {
	TamanoGrafo     TamanoGrafo
	Implementacion  TipoImplementacion
	NumGoroutines   int
	Replica         int
	TiempoEjecucion time.Duration
	NumNodos        int
	NumEnlaces      int
	MemoriaUsada    uint64 // Bytes
	ResultadosRank  map[int]float64 // NodeID -> Rank
}

// MetricasAgregadas contiene las métricas calculadas después del experimento
type MetricasAgregadas struct {
	TamanoGrafo         TamanoGrafo
	Implementacion      TipoImplementacion
	NumGoroutines       int
	TiempoPromedio      time.Duration
	TiempoMin           time.Duration
	TiempoMax           time.Duration
	DesviacionEstandar  float64
	MemoriaPromedio     uint64
	Speedup             float64 // Solo para versiones concurrentes
	Eficiencia          float64 // Solo para versiones concurrentes
}
