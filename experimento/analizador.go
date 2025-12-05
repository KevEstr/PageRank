package experimento

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Analizador procesa los resultados y calcula métricas agregadas
type Analizador struct {
	resultados []ResultadoEjecucion
}

// NuevoAnalizador crea un nuevo analizador con los resultados
func NuevoAnalizador(resultados []ResultadoEjecucion) *Analizador {
	return &Analizador{
		resultados: resultados,
	}
}

// CalcularMetricas calcula todas las métricas agregadas
func (a *Analizador) CalcularMetricas() []MetricasAgregadas {
	// Agrupar resultados por combinación de parámetros
	grupos := make(map[string][]ResultadoEjecucion)
	
	for _, r := range a.resultados {
		clave := fmt.Sprintf("%s_%s_%d", r.TamanoGrafo, r.Implementacion, r.NumGoroutines)
		grupos[clave] = append(grupos[clave], r)
	}
	
	metricas := make([]MetricasAgregadas, 0)
	
	for _, grupo := range grupos {
		if len(grupo) == 0 {
			continue
		}
		
		metrica := a.calcularMetricasGrupo(grupo)
		metricas = append(metricas, metrica)
	}
	
	// Calcular speedup y eficiencia
	a.calcularSpeedupYEficiencia(metricas)
	
	return metricas
}

// calcularMetricasGrupo calcula métricas para un grupo de resultados
func (a *Analizador) calcularMetricasGrupo(grupo []ResultadoEjecucion) MetricasAgregadas {
	if len(grupo) == 0 {
		return MetricasAgregadas{}
	}
	
	tiempos := make([]time.Duration, len(grupo))
	memorias := make([]uint64, 0)
	
	var sumaT time.Duration
	var sumaM uint64
	minT := grupo[0].TiempoEjecucion
	maxT := grupo[0].TiempoEjecucion
	
	for i, r := range grupo {
		tiempos[i] = r.TiempoEjecucion
		sumaT += r.TiempoEjecucion
		sumaM += r.MemoriaUsada
		memorias = append(memorias, r.MemoriaUsada)
		
		if r.TiempoEjecucion < minT {
			minT = r.TiempoEjecucion
		}
		if r.TiempoEjecucion > maxT {
			maxT = r.TiempoEjecucion
		}
	}
	
	promedioT := sumaT / time.Duration(len(grupo))
	promedioM := sumaM / uint64(len(grupo))
	
	// Calcular desviación estándar
	var sumaCuadrados float64
	for _, t := range tiempos {
		diff := float64(t - promedioT)
		sumaCuadrados += diff * diff
	}
	desviacion := math.Sqrt(sumaCuadrados / float64(len(grupo)))
	
	return MetricasAgregadas{
		TamanoGrafo:        grupo[0].TamanoGrafo,
		Implementacion:     grupo[0].Implementacion,
		NumGoroutines:      grupo[0].NumGoroutines,
		TiempoPromedio:     promedioT,
		TiempoMin:          minT,
		TiempoMax:          maxT,
		DesviacionEstandar: desviacion,
		MemoriaPromedio:    promedioM,
	}
}

// calcularSpeedupYEficiencia calcula speedup y eficiencia para versiones concurrentes
func (a *Analizador) calcularSpeedupYEficiencia(metricas []MetricasAgregadas) {
	// Agrupar por tamaño de grafo
	tiemposSecuenciales := make(map[TamanoGrafo]time.Duration)
	
	// Primero, obtener tiempos secuenciales
	for i := range metricas {
		if metricas[i].Implementacion == Secuencial {
			tiemposSecuenciales[metricas[i].TamanoGrafo] = metricas[i].TiempoPromedio
		}
	}
	
	// Calcular speedup y eficiencia para versiones concurrentes
	for i := range metricas {
		if metricas[i].Implementacion == Concurrente {
			tSeq, existe := tiemposSecuenciales[metricas[i].TamanoGrafo]
			if existe && metricas[i].TiempoPromedio > 0 {
				metricas[i].Speedup = float64(tSeq) / float64(metricas[i].TiempoPromedio)
				metricas[i].Eficiencia = metricas[i].Speedup / float64(metricas[i].NumGoroutines)
			}
		}
	}
}

// ImprimirResultados imprime un resumen de las métricas
func (a *Analizador) ImprimirResultados(metricas []MetricasAgregadas) {
	// Ordenar por tamaño y luego por número de goroutines
	sort.Slice(metricas, func(i, j int) bool {
		if metricas[i].TamanoGrafo != metricas[j].TamanoGrafo {
			return metricas[i].TamanoGrafo < metricas[j].TamanoGrafo
		}
		return metricas[i].NumGoroutines < metricas[j].NumGoroutines
	})
	
	fmt.Println("\n" + strings.Repeat("=", 110))
	fmt.Println("RESULTADOS DEL EXPERIMENTO - DISEÑO DE BLOQUES TOTALMENTE ALEATORIZADOS")
	fmt.Println(strings.Repeat("=", 110))
	
	// Verificar si hay versiones concurrentes
	hayConcurrente := false
	for _, m := range metricas {
		if m.Implementacion == Concurrente {
			hayConcurrente = true
			break
		}
	}
	
	if !hayConcurrente {
		fmt.Println("\n⚠️  ADVERTENCIA: Solo se ejecutó la versión secuencial.")
		fmt.Println("    Las métricas de Speedup y Eficiencia requieren la implementación concurrente.")
		fmt.Println()
	}
	
	tamanoActual := TamanoGrafo("")
	
	for _, m := range metricas {
		if m.TamanoGrafo != tamanoActual {
			tamanoActual = m.TamanoGrafo
			fmt.Printf("\n━━━ BLOQUE: Grafo %s ━━━\n", strings.ToUpper(string(tamanoActual)))
			
			if hayConcurrente {
				fmt.Printf("%-15s %-12s %-18s %-18s %-12s %-15s\n", 
					"Implementación", "Workers", "Tiempo Promedio", "Desv. Est. (ms)", "Speedup", "Eficiencia")
			} else {
				fmt.Printf("%-15s %-18s %-18s %-18s\n", 
					"Implementación", "Tiempo Promedio", "Desv. Est. (ms)", "Memoria Prom.")
			}
			fmt.Println(strings.Repeat("-", 110))
		}
		
		if hayConcurrente {
			speedupStr := "-"
			eficienciaStr := "-"
			
			if m.Implementacion == Concurrente {
				speedupStr = fmt.Sprintf("%.3fx", m.Speedup)
				eficienciaStr = fmt.Sprintf("%.1f%%", m.Eficiencia*100)
			}
			
			fmt.Printf("%-15s %-12d %-18v %-18.2f %-12s %-15s\n",
				m.Implementacion,
				m.NumGoroutines,
				m.TiempoPromedio,
				m.DesviacionEstandar/float64(time.Millisecond),
				speedupStr,
				eficienciaStr)
		} else {
			// Solo mostrar métricas relevantes para versión secuencial
			fmt.Printf("%-15s %-18v %-18.2f %-18.2f MB\n",
				m.Implementacion,
				m.TiempoPromedio,
				m.DesviacionEstandar/float64(time.Millisecond),
				float64(m.MemoriaPromedio)/(1024*1024))
		}
	}
	
	fmt.Println(strings.Repeat("=", 110))
	
	if !hayConcurrente {
		fmt.Println("\n💡 Próximo paso: Implementar la versión concurrente de PageRank")
		fmt.Println("   para poder calcular Speedup y Eficiencia.")
	} else {
		fmt.Println("\n📊 INTERPRETACIÓN DE RESULTADOS:")
		fmt.Println("   - Speedup > 1.0:  La versión concurrente es MÁS RÁPIDA")
		fmt.Println("   - Speedup = 1.0:  Ambas versiones tienen el MISMO rendimiento")
		fmt.Println("   - Speedup < 1.0:  La versión secuencial es MÁS RÁPIDA (overhead de goroutines)")
		fmt.Println("   - Eficiencia:     Qué tan bien se aprovechan los workers (ideal = 100%)")
	}
}

// VerificarCorreccion verifica que las versiones produzcan el mismo orden de nodos
func (a *Analizador) VerificarCorreccion() bool {
	// Agrupar por tamaño de grafo
	gruposPorTamano := make(map[TamanoGrafo][]ResultadoEjecucion)
	
	for _, r := range a.resultados {
		gruposPorTamano[r.TamanoGrafo] = append(gruposPorTamano[r.TamanoGrafo], r)
	}
	
	todoCorrecto := true
	
	for tamano, grupo := range gruposPorTamano {
		if len(grupo) < 2 {
			continue
		}
		
		// Obtener top 10 nodos del primer resultado como referencia
		referencia := obtenerTopNodos(grupo[0].ResultadosRank, 10)
		
		// Comparar con todos los demás
		for i := 1; i < len(grupo); i++ {
			actual := obtenerTopNodos(grupo[i].ResultadosRank, 10)
			
			if !compararOrden(referencia, actual) {
				fmt.Printf("⚠️  ADVERTENCIA: Orden diferente en grafo %s entre ejecuciones\n", tamano)
				todoCorrecto = false
			}
		}
	}
	
	if todoCorrecto {
		fmt.Println("✓ Verificación de corrección: Todas las versiones producen el mismo orden de nodos")
	}
	
	return todoCorrecto
}

// obtenerTopNodos obtiene los N nodos con mayor rank
func obtenerTopNodos(ranks map[int]float64, n int) []int {
	type nodoRank struct {
		nodo int
		rank float64
	}
	
	lista := make([]nodoRank, 0, len(ranks))
	for nodo, rank := range ranks {
		lista = append(lista, nodoRank{nodo, rank})
	}
	
	sort.Slice(lista, func(i, j int) bool {
		return lista[i].rank > lista[j].rank
	})
	
	resultado := make([]int, 0, n)
	for i := 0; i < n && i < len(lista); i++ {
		resultado = append(resultado, lista[i].nodo)
	}
	
	return resultado
}

// compararOrden compara si dos listas de nodos tienen el mismo orden
func compararOrden(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	
	return true
}

// GuardarCSV guarda los resultados en formato CSV para análisis posterior
func (a *Analizador) GuardarCSV(nombreArchivo string, resultados []ResultadoEjecucion) error {
	file, err := os.Create(nombreArchivo)
	if err != nil {
		return fmt.Errorf("error creando archivo CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Escribir encabezados
	headers := []string{
		"bloque",
		"tratamiento",
		"num_goroutines",
		"replica",
		"tiempo_ms",
		"num_nodos",
		"num_enlaces",
		"memoria_mb",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error escribiendo encabezados: %v", err)
	}

	// Escribir datos
	for _, r := range resultados {
		record := []string{
			string(r.TamanoGrafo),
			string(r.Implementacion),
			strconv.Itoa(r.NumGoroutines),
			strconv.Itoa(r.Replica),
			fmt.Sprintf("%.2f", float64(r.TiempoEjecucion.Microseconds())/1000.0),
			strconv.Itoa(r.NumNodos),
			strconv.Itoa(r.NumEnlaces),
			fmt.Sprintf("%.2f", float64(r.MemoriaUsada)/(1024*1024)),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error escribiendo registro: %v", err)
		}
	}

	return nil
}
