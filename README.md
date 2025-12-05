# PageRank: Implementación Secuencial y Concurrente

Implementación del algoritmo PageRank en Go con versiones secuencial y concurrente, incluyendo sistema de experimentación para análisis de rendimiento.

![PageRank](http://upload.wikimedia.org/wikipedia/commons/thumb/f/fb/PageRanks-Example.svg/596px-PageRanks-Example.svg.png)

## Autores

- Kevin Estrada Del Valle
- María Cristina Vergara Quinchia

## Descripción

Este proyecto implementa el algoritmo PageRank para calcular la importancia de nodos en grafos dirigidos. Incluye:

- **Implementación secuencial**: Versión tradicional de un solo hilo
- **Implementación concurrente**: Versión concurrente con múltiples goroutines
- **Sistema de experimentación**: Para comparar rendimiento mediante diseño de bloques aleatorizados

## Estructura del Proyecto

```
pagerank/
├── pagerank.go                 # Implementación secuencial
├── pagerank_concurrent.go      # Implementación concurrente
├── pagerank_concurrent_test.go
├── cmd/experimento/main.go    # Programa principal de experimentación
├── experimento/               # Sistema de análisis experimental
│   ├── tipos.go
│   ├── generador.go
│   ├── ejecutor.go
│   ├── medidor.go
│   └── analizador.go
└── resultados_experimento.csv # Resultados de ejecuciones
```

## Uso

### Instalación

```bash
git clone https://github.com/KevEstr/PageRank.git
cd PageRank
go mod download
```

### Uso Básico

```go
package main

import "github.com/dcadenas/pagerank"

func main() {
    // Versión secuencial
    graph := pagerank.New()
    graph.Link(1, 2)
    graph.Link(2, 3)
    graph.Link(3, 1)
    
    graph.Rank(0.85, 0.0001, func(id int, rank float64) {
        println("Nodo", id, "tiene rank", rank)
    })
    
    // Versión concurrente con 4 workers
    graphConcurrent := pagerank.NewConcurrentWithWorkers(4)
    graphConcurrent.Link(1, 2)
    // ... resto del código igual
}
```

### Ejecutar Experimentos

```bash
# Experimento completo con configuración por defecto
go run cmd/experimento/main.go

# Configuración personalizada
go run cmd/experimento/main.go -replicas 5 -damping 0.85 -tolerance 0.0001
```

## Diseño Experimental

El sistema utiliza un diseño de bloques completamente aleatorizados (RCBD) que evalúa:

- **Bloques**: Tres tamaños de grafos (20K, 100K, 500K nodos)
- **Tratamientos**: Secuencial y concurrente con 2, 4, 8 y 16 workers
- **Métricas**: Tiempo de ejecución, speedup, eficiencia y uso de memoria

Los resultados se exportan a CSV para análisis estadístico.

## Tests

```bash
go test -v
```

## Requisitos

- Go 1.20 o superior
