package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dcadenas/pagerank/experimento"
)

func main() {
	// Parámetros configurables por línea de comandos
	numReplicas := flag.Int("replicas", 3, "Número de réplicas por tratamiento")
	dampingFactor := flag.Float64("damping", 0.85, "Factor de amortiguación")
	tolerance := flag.Float64("tolerance", 0.0001, "Tolerancia para convergencia")
	seed := flag.Int64("seed", time.Now().UnixNano(), "Semilla para generación de grafos")
	
	flag.Parse()

	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     EXPERIMENTO: PageRank Secuencial vs Concurrente                  ║")
	fmt.Println("║     Diseño de Bloques Totalmente Aleatorizados                       ║")
	fmt.Println("║     TAMAÑOS GRANDES PARA RESULTADOS IMPRESIONANTES                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Configuración del Experimento:\n")
	fmt.Printf("  - Réplicas por tratamiento: %d\n", *numReplicas)
	fmt.Printf("  - Damping factor: %.2f\n", *dampingFactor)
	fmt.Printf("  - Tolerancia: %.4f\n", *tolerance)
	fmt.Printf("  - Semilla: %d\n", *seed)
	fmt.Println()
	fmt.Println("Diseño Experimental - Progresión 5x:")
	fmt.Println("  BLOQUE 1 (Pequeño):   20,000 nodos   (baseline)")
	fmt.Println("  BLOQUE 2 (Mediano):  100,000 nodos   (5x más grande)")
	fmt.Println("  BLOQUE 3 (Grande):   500,000 nodos   (5x más grande)")
	fmt.Println()
	fmt.Println("Tratamientos por bloque:")
	fmt.Println("  - Secuencial (1 worker)")
	fmt.Println("  - Concurrente (2 workers)")
	fmt.Println("  - Concurrente (4 workers)")
	fmt.Println("  - Concurrente (8 workers)")
	fmt.Println("  - Concurrente (16 workers)")
	fmt.Printf("\nTotal de ejecuciones: 3 bloques x 5 tratamientos x %d replicas = %d\n", *numReplicas, 3*5*(*numReplicas))
	fmt.Println()
	fmt.Println("Tiempo estimado: 10-15 minutos (3 replicas)")
	fmt.Println()

	// Configurar experimento
	config := experimento.ConfigExperimento{
		DampingFactor: *dampingFactor,
		Tolerance:     *tolerance,
		NumReplicas:   *numReplicas,
	}

	// Crear ejecutor
	ejecutor := experimento.NuevoEjecutor(config, *seed)

	// Ejecutar experimento completo
	fmt.Println("Iniciando experimento con aleatorización completa...")
	fmt.Println("Esto puede tomar varios minutos dependiendo del hardware...")
	fmt.Println()

	inicio := time.Now()
	resultados := ejecutor.EjecutarExperimentoCompleto()
	duracionTotal := time.Since(inicio)

	fmt.Printf("\n✓ Experimento completado en %v\n", duracionTotal)
	fmt.Printf("  Total de ejecuciones: %d\n", len(resultados))

	// Analizar resultados
	analizador := experimento.NuevoAnalizador(resultados)
	
	// Verificar corrección (que todas las versiones den mismo resultado)
	fmt.Println()
	analizador.VerificarCorreccion()

	// Calcular y mostrar métricas
	metricas := analizador.CalcularMetricas()
	analizador.ImprimirResultados(metricas)

	// Guardar resultados a CSV para análisis posterior
	fmt.Println("\n📊 Guardando resultados en CSV...")
	analizador.GuardarCSV("resultados_experimento.csv", resultados)
	
	fmt.Println("\n✅ Experimento completado exitosamente!")
	fmt.Println("   Archivo generado: resultados_experimento.csv")
}
