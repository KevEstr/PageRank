package experimento

import (
	"fmt"
	"runtime"
	"time"

	"github.com/dcadenas/pagerank"
)

// Ejecutor orquesta la ejecución de experimentos
type Ejecutor struct {
	config    ConfigExperimento
	generador *GeneradorGrafos
}

// NuevoEjecutor crea un nuevo ejecutor de experimentos
func NuevoEjecutor(config ConfigExperimento, seed int64) *Ejecutor {
	return &Ejecutor{
		config:    config,
		generador: NewGenerador(seed),
	}
}

// EjecutarExperimentoCompleto ejecuta todos los bloques y tratamientos
func (e *Ejecutor) EjecutarExperimentoCompleto() []ResultadoEjecucion {
	// Orden aleatorizado generado por R
	ordenR := []struct {
		bloque  TamanoGrafo
		workers int
		replica int
	}{
		{Grande, 8, 1}, {Grande, 2, 2}, {Grande, 16, 3}, {Grande, 16, 1}, {Grande, 4, 2},
		{Grande, 2, 3}, {Grande, 4, 1}, {Grande, 1, 2}, {Grande, 2, 1}, {Grande, 8, 3},
		{Grande, 8, 2}, {Grande, 4, 3}, {Grande, 16, 2}, {Grande, 1, 3}, {Grande, 1, 1},
		{Mediano, 4, 1}, {Mediano, 2, 2}, {Mediano, 2, 3}, {Mediano, 8, 2}, {Mediano, 8, 1},
		{Mediano, 16, 3}, {Mediano, 1, 2}, {Mediano, 16, 1}, {Mediano, 8, 3}, {Mediano, 2, 1},
		{Mediano, 1, 3}, {Mediano, 1, 1}, {Mediano, 16, 2}, {Mediano, 4, 2}, {Mediano, 4, 3},
		{Pequeno, 8, 2}, {Pequeno, 16, 1}, {Pequeno, 8, 3}, {Pequeno, 1, 2}, {Pequeno, 4, 2},
		{Pequeno, 1, 1}, {Pequeno, 16, 2}, {Pequeno, 4, 3}, {Pequeno, 4, 1}, {Pequeno, 2, 1},
		{Pequeno, 2, 2}, {Pequeno, 8, 1}, {Pequeno, 2, 3}, {Pequeno, 16, 3}, {Pequeno, 1, 3},
	}

	// Pre-generar grafos
	grafos := make(map[TamanoGrafo]*Grafo)
	for _, tamano := range []TamanoGrafo{Pequeno, Mediano, Grande} {
		configGrafo := ObtenerConfiguracionPorTamano(tamano, time.Now().UnixNano())
		grafo := e.generador.GenerarGrafo(configGrafo)
		grafos[tamano] = grafo
		fmt.Printf("Generado: Grafo %s - %d nodos, %d enlaces\n", tamano, grafo.NumNodos, grafo.NumEnlaces)
	}

	resultados := make([]ResultadoEjecucion, 0, len(ordenR))
	fmt.Printf("\nEjecutando %d corridas en orden aleatorizado de R...\n\n", len(ordenR))

	// Ejecutar en el orden de R
	for i, corrida := range ordenR {
		grafo := grafos[corrida.bloque]
		
		var impl TipoImplementacion
		if corrida.workers == 1 {
			impl = Secuencial
		} else {
			impl = Concurrente
		}
		
		trat := Tratamiento{
			Implementacion: impl,
			NumGoroutines:  corrida.workers,
		}
		
		resultado := e.ejecutarTratamiento(grafo, trat, corrida.replica)
		resultados = append(resultados, resultado)
		
		fmt.Printf("[%2d/45] %8s | %2d workers | Réplica %d | %v\n",
			i+1, corrida.bloque, corrida.workers, corrida.replica, resultado.TiempoEjecucion)
	}
	
	return resultados
}

// Tratamiento representa una combinación de implementación y configuración
type Tratamiento struct {
	Implementacion TipoImplementacion
	NumGoroutines  int
}

// crearTratamientos genera todas las combinaciones de tratamientos
func (e *Ejecutor) crearTratamientos() []Tratamiento {
	tratamientos := []Tratamiento{
		{Implementacion: Secuencial, NumGoroutines: 1},
	}
	
	// Versiones concurrentes con diferentes números de workers
	// Para grafos grandes, probamos hasta 16 workers
	nivelesGoroutines := []int{2, 4, 8, 16}
	for _, n := range nivelesGoroutines {
		tratamientos = append(tratamientos, Tratamiento{
			Implementacion: Concurrente,
			NumGoroutines:  n,
		})
	}
	
	return tratamientos
}

// ejecutarTratamiento ejecuta un tratamiento específico y mide su rendimiento
func (e *Ejecutor) ejecutarTratamiento(grafo *Grafo, trat Tratamiento, replica int) ResultadoEjecucion {
	var pr pagerank.Interface
	
	// Crear instancia según el tipo de implementación
	if trat.Implementacion == Concurrente {
		pr = pagerank.NewConcurrentWithWorkers(trat.NumGoroutines)
	} else {
		pr = pagerank.New()
	}
	
	// Cargar el grafo
	for _, enlace := range grafo.Enlaces {
		pr.Link(enlace[0], enlace[1])
	}
	
	// WARM-UP: Ejecutar una vez sin medir (solo en la primera réplica)
	// Esto calienta la CPU, carga caches, etc.
	if replica == 1 {
		pr.Rank(e.config.DampingFactor, e.config.Tolerance, func(nodeID int, rank float64) {
			// Descartamos estos resultados
		})
		
		// Recrear para la medición real
		if trat.Implementacion == Concurrente {
			pr = pagerank.NewConcurrentWithWorkers(trat.NumGoroutines)
		} else {
			pr = pagerank.New()
		}
		for _, enlace := range grafo.Enlaces {
			pr.Link(enlace[0], enlace[1])
		}
	}
	
	// Almacenar resultados
	resultadosRank := make(map[int]float64)
	
	var duracion time.Duration
	var memoria uint64
	
	// DECISIÓN: Para grafos pequeños, usar repeticiones para mejor precisión
	esGrafoPequeno := grafo.NumNodos <= 50000
	
	if esGrafoPequeno {
		// Grafos pequeños: ejecutar múltiples veces (mejor precisión)
		repeticiones := 10
		if grafo.NumNodos <= 20000 {
			repeticiones = 50 // Más repeticiones para grafos muy pequeños
		}
		
		duracion = MedirConRepeticiones(func() {
			// Recrear resultados en cada iteración
			resultadosRank = make(map[int]float64)
			pr.Rank(e.config.DampingFactor, e.config.Tolerance, func(nodeID int, rank float64) {
				resultadosRank[nodeID] = rank
			})
		}, repeticiones)
		
		// Memoria: medir una vez (no cambia significativamente)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memoriaInicio := m.Alloc
		
		pr.Rank(e.config.DampingFactor, e.config.Tolerance, func(nodeID int, rank float64) {
			resultadosRank[nodeID] = rank
		})
		
		runtime.ReadMemStats(&m)
		memoria = m.Alloc - memoriaInicio
		
	} else {
		// Grafos medianos/grandes: medición normal (suficiente precisión)
		medidor := NuevoMedidor()
		
		pr.Rank(e.config.DampingFactor, e.config.Tolerance, func(nodeID int, rank float64) {
			resultadosRank[nodeID] = rank
		})
		
		duracion, memoria = medidor.Detener()
	}
	
	return ResultadoEjecucion{
		TamanoGrafo:     obtenerTamanoPorNumNodos(grafo.NumNodos),
		Implementacion:  trat.Implementacion,
		NumGoroutines:   trat.NumGoroutines,
		Replica:         replica,
		TiempoEjecucion: duracion,
		NumNodos:        grafo.NumNodos,
		NumEnlaces:      grafo.NumEnlaces,
		MemoriaUsada:    memoria,
		ResultadosRank:  resultadosRank,
	}
}

// obtenerTamanoPorNumNodos determina el tamaño del grafo por su número de nodos
func obtenerTamanoPorNumNodos(numNodos int) TamanoGrafo {
	if numNodos <= 20000 {
		return Pequeno
	} else if numNodos <= 100000 {
		return Mediano
	}
	return Grande
}


// CrearPageRankDesdeGrafo crea una instancia de PageRank y carga el grafo
func CrearPageRankDesdeGrafo(grafo *Grafo) pagerank.Interface {
	pr := pagerank.New()
	for _, enlace := range grafo.Enlaces {
		pr.Link(enlace[0], enlace[1])
	}
	return pr
}
