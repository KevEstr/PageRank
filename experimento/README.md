# Sistema de Experimentos PageRank

Este paquete implementa un sistema completo para realizar experimentos controlados comparando las versiones secuencial y concurrente del algoritmo PageRank.

Desarrollado por:
### Kevin Estrada Del Valle
### Maria Cristina Vergara Quinchia

## Diseño del Experimento

### Metodología: RCBD (Randomized Complete Block Design)

El experimento sigue un diseño por bloques completamente aleatorizados donde:

- **Bloques**: Tamaño de grafos (Pequeño, Mediano, Grande)
- **Tratamientos**: Combinaciones de implementación (Secuencial/Concurrente) y nivel de paralelismo
- **Réplicas**: Múltiples ejecuciones de cada tratamiento para validez estadística

### Estructura de Bloques

| Bloque | Nodos | Enlaces/Nodo | Descripción |
|--------|-------|--------------|-------------|
| Pequeño | ~1,000 | 5 | Validación rápida |
| Mediano | ~10,000 | 8 | Caso realista |
| Grande | ~100,000 | 10 | Estrés del sistema |

### Tratamientos

1. **Secuencial**: 1 goroutine (línea base)
2. **Concurrente**: 2, 4, 8, 16 goroutines

### Métricas Capturadas

1. **Tiempo de Ejecución**: Duración total del algoritmo
2. **Speedup**: T_secuencial / T_concurrente
3. **Eficiencia**: Speedup / Número_de_goroutines
4. **Uso de Memoria**: Memoria asignada durante la ejecución

## Componentes

### 1. `tipos.go`
Define las estructuras de datos para configuración y resultados.

### 2. `generador.go`
Genera grafos sintéticos con propiedades controladas:
- Distribución de enlaces realista
- Nodos "hub" que reciben más enlaces
- Reproducibilidad mediante semillas

### 3. `medidor.go`
Captura métricas de rendimiento:
- Tiempo de ejecución preciso
- Uso de memoria

### 4. `ejecutor.go`
Orquesta la ejecución del experimento:
- Aleatoriza el orden de ejecución
- Ejecuta réplicas
- Mantiene condiciones controladas

### 5. `analizador.go`
Procesa resultados y calcula métricas:
- Estadísticas descriptivas
- Speedup y eficiencia
- Verificación de corrección

## Uso

### Experimento Completo

```bash
go run cmd/experimento/main.go -replicas 5 -damping 0.85 -tolerance 0.0001
```

Parámetros:
- `-replicas`: Número de réplicas por tratamiento (default: 3)
- `-damping`: Factor de amortiguación (default: 0.85)
- `-tolerance`: Tolerancia para convergencia (default: 0.0001)
- `-seed`: Semilla para reproducibilidad

### Prueba Simple

Para probar el sistema rápidamente:

```bash
go run cmd/prueba_simple/main.go
```

## Validación de Resultados

El sistema verifica que:
1. Todas las versiones producen el mismo orden de importancia de nodos
2. Los resultados son reproducibles con la misma semilla
3. Las métricas son consistentes entre réplicas

## Interpretación de Resultados

### Speedup
- **> 1**: La versión concurrente es más rápida
- **< 1**: Overhead de concurrencia supera los beneficios
- **Ideal**: Cercano al número de goroutines

### Eficiencia
- **100%**: Uso perfecto de recursos
- **> 50%**: Buena paralelización
- **< 50%**: Overhead significativo o contención

## Ejemplo de Salida

```
=== BLOQUE: Grafo pequeño ===
Generado: 1000 nodos, 5234 enlaces
  [secuencial, 1 goroutines, réplica 1]: 45.2ms
  [concurrente, 4 goroutines, réplica 1]: 15.8ms
  ...

--- Grafo pequeño ---
Implementación  Goroutines   Tiempo Prom.    Speedup    Eficiencia
secuencial      1            45.2ms          -          -
concurrente     2            25.1ms          1.80x      90.0%
concurrente     4            15.8ms          2.86x      71.5%
concurrente     8            12.3ms          3.67x      45.9%
```

## Notas Importantes

1. **Reproducibilidad**: Usa la misma semilla para resultados reproducibles
2. **Condiciones Controladas**: Cierra otras aplicaciones durante el experimento
3. **Warm-up**: Las primeras ejecuciones pueden ser más lentas (JIT, caché)
4. **Memoria**: Grafos grandes requieren RAM significativa
