package experimento

import (
	"math/rand"
)

// GeneradorGrafos crea grafos sintéticos para pruebas
type GeneradorGrafos struct {
	rand *rand.Rand
}

// NewGenerador crea un nuevo generador de grafos
func NewGenerador(seed int64) *GeneradorGrafos {
	return &GeneradorGrafos{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Grafo representa un grafo dirigido simple
type Grafo struct {
	NumNodos   int
	Enlaces    [][2]int // [from, to]
	NumEnlaces int
}

// GenerarGrafo crea un grafo sintético basado en la configuración
func (g *GeneradorGrafos) GenerarGrafo(config ConfigGrafo) *Grafo {
	grafo := &Grafo{
		NumNodos: config.NumNodos,
		Enlaces:  make([][2]int, 0),
	}

	// Identificar nodos hub (nodos que recibirán más enlaces)
	hubs := make(map[int]bool)
	numHubs := int(float64(config.NumNodos) * config.ProbabilidadHub)
	for i := 0; i < numHubs; i++ {
		hubs[g.rand.Intn(config.NumNodos)] = true
	}

	// Crear enlaces para cada nodo
	enlacesCreados := make(map[[2]int]bool) // Para evitar duplicados

	for from := 0; from < config.NumNodos; from++ {
		// Número de enlaces salientes (con variación)
		numEnlaces := config.EnlacesPorNodo + g.rand.Intn(config.EnlacesPorNodo/2+1) - config.EnlacesPorNodo/4
		if numEnlaces < 1 {
			numEnlaces = 1
		}

		for i := 0; i < numEnlaces; i++ {
			var to int

			// 70% de probabilidad de enlazar a un hub si existen
			if len(hubs) > 0 && g.rand.Float64() < 0.7 {
				// Seleccionar un hub aleatorio
				hubIndex := g.rand.Intn(len(hubs))
				j := 0
				for hubID := range hubs {
					if j == hubIndex {
						to = hubID
						break
					}
					j++
				}
			} else {
				// Enlazar a un nodo aleatorio
				to = g.rand.Intn(config.NumNodos)
			}

			// Evitar auto-enlaces y duplicados
			if from != to {
				enlace := [2]int{from, to}
				if !enlacesCreados[enlace] {
					grafo.Enlaces = append(grafo.Enlaces, enlace)
					enlacesCreados[enlace] = true
				}
			}
		}
	}

	grafo.NumEnlaces = len(grafo.Enlaces)
	return grafo
}

// ObtenerConfiguracionPorTamano devuelve la configuración apropiada según el tamaño
func ObtenerConfiguracionPorTamano(tamano TamanoGrafo, seed int64) ConfigGrafo {
	configs := map[TamanoGrafo]ConfigGrafo{
		Pequeno: {
			NumNodos:          20000,    // 20K nodos
			EnlacesPorNodo:    8,
			ProbabilidadHub:   0.03,
			Seed:              seed,
		},
		Mediano: {
			NumNodos:          100000,   // 100K nodos (5x)
			EnlacesPorNodo:    10,
			ProbabilidadHub:   0.02,
			Seed:              seed,
		},
		Grande: {
			NumNodos:          500000,   // 500K nodos (5x)
			EnlacesPorNodo:    12,
			ProbabilidadHub:   0.015,
			Seed:              seed,
		},
	}

	return configs[tamano]
}
