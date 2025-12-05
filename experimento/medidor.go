package experimento

import (
	"runtime"
	"time"
)

// Medidor captura métricas de rendimiento durante la ejecución
type Medidor struct {
	inicio       time.Time
	memoriaInicio uint64
}

func NuevoMedidor() *Medidor {
	// Forzar garbage collection para medición limpia
	runtime.GC()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return &Medidor{
		inicio:       time.Now(),
		memoriaInicio: m.Alloc,
	}
}

// Detener finaliza la medición y retorna las métricas
func (m *Medidor) Detener() (duracion time.Duration, memoriaUsada uint64) {
	duracion = time.Since(m.inicio)
	
	runtime.GC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	
	memoriaUsada = mem.Alloc - m.memoriaInicio
	if mem.Alloc < m.memoriaInicio {
		memoriaUsada = mem.Alloc // En caso de que GC haya liberado memoria
	}
	
	return duracion, memoriaUsada
}

// MedirConRepeticiones ejecuta múltiples veces para mejorar precisión en operaciones rápidas
// Útil para grafos pequeños donde una ejecución es < 1ms
func MedirConRepeticiones(fn func(), repeticiones int) time.Duration {
	runtime.GC()
	
	inicio := time.Now()
	for i := 0; i < repeticiones; i++ {
		fn()
	}
	tiempoTotal := time.Since(inicio)
	
	// Retornar tiempo promedio por ejecución
	return tiempoTotal / time.Duration(repeticiones)
}
