# Capítulo 16: APU - Audio Processing Unit

## La Unidad de Procesamiento de Audio de la NES

El APU (Audio Processing Unit) es el chip responsable de generar todo el sonido en la NES. Está integrado dentro del mismo chip que la CPU (el Ricoh 2A03), lo que significa que comparte el mismo reloj maestro y puede ser controlado directamente mediante escrituras a registros mapeados en memoria.

## Arquitectura del APU

El APU de la NES contiene **5 canales de audio** independientes:

| Canal | Tipo | Descripción |
|-------|------|-------------|
| **Pulse 1** | Onda Cuadrada | Canal principal para melodías, con sweep |
| **Pulse 2** | Onda Cuadrada | Segundo canal melódico, idéntico al Pulse 1 |
| **Triangle** | Onda Triangular | Tonos suaves, usado para líneas de bajo |
| **Noise** | Ruido Pseudoaleatorio | Efectos de percusión, explosiones |
| **DMC** | Delta Modulation | Samples digitales (voces, efectos complejos) |

## Registros del APU

Los registros del APU están mapeados en el rango **$4000 - $4017**:

```
$4000-$4003: Pulse 1
$4004-$4007: Pulse 2
$4008-$400B: Triangle
$400C-$400F: Noise
$4010-$4013: DMC
$4015: Status (lectura/escritura)
$4017: Frame Counter
```

## El Frame Sequencer

El corazón del timing del APU es el **Frame Sequencer**, un contador interno que genera "ticks" a frecuencias específicas para controlar los diferentes componentes de cada canal.

### Modos del Frame Sequencer

| Modo | Pasos | Frecuencia | IRQ |
|------|-------|------------|-----|
| **4-Step** | 4 | 240 Hz | Sí |
| **5-Step** | 5 | 192 Hz | No |

El Frame Sequencer controla:
- **Envelope** (240 Hz): Volumen dinámico
- **Sweep** (120 Hz): Cambio de frecuencia automático
- **Length Counter** (120 Hz): Duración de las notas

```
Modo 4-Step (Más común):
  Step 1 (7457 ciclos):  Envelope, Linear Counter
  Step 2 (14913 ciclos): Envelope, Linear Counter, Length Counter, Sweep
  Step 3 (22371 ciclos): Envelope, Linear Counter
  Step 4 (29829 ciclos): Envelope, Linear Counter, Length Counter, Sweep, IRQ
```

## Canal Pulse (Onda Cuadrada)

Los canales Pulse generan ondas cuadradas con **duty cycle** configurable:

| Duty | Forma de Onda | Sonido |
|------|---------------|--------|
| 12.5% | `_-------` | Agudo, metálico |
| 25% | `__------` | Brillante |
| 50% | `____----` | Hueco, como flauta |
| 75% | `______--` | Igual que 25% (invertido) |

### Timer y Frecuencia

La frecuencia de salida se calcula:
```
f = CPU_FREQ / (16 * (timer_period + 1))
```

Para NTSC (CPU = 1.789773 MHz):
```
f = 1789773 / (16 * (timer_period + 1))
```

### Envelope (Envolvente de Volumen)

El envelope permite que las notas decaigan naturalmente:
- **Constant**: Volumen fijo (0-15)
- **Decay**: El volumen decrece de 15 a 0

### Sweep Unit

Permite "deslizar" la frecuencia automáticamente (portamento):
- **Shift**: Cuántos bits desplazar el período
- **Negate**: Dirección (arriba/abajo)
- **Period**: Velocidad del sweep

## Canal Triangle

El canal Triangle genera una onda triangular de 32 pasos, ideal para tonos graves:

```
Secuencia: 15,14,13,12,11,10,9,8,7,6,5,4,3,2,1,0,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15
```

No tiene control de volumen, pero tiene un **Linear Counter** que silencia el canal después de cierto tiempo.

## Canal Noise

El canal Noise genera ruido pseudoaleatorio usando un **Linear Feedback Shift Register (LFSR)**:

```go
// Implementación del LFSR
feedback := (n.shiftReg & 1) ^ ((n.shiftReg >> bit) & 1)
n.shiftReg = (n.shiftReg >> 1) | (feedback << 14)
```

El bit 6 de $400E selecciona entre:
- **Modo 0**: LFSR usa bits 0 y 1 (ruido blanco)
- **Modo 1**: LFSR usa bits 0 y 6 (ruido metálico/tonal)

## Canal DMC (Delta Modulation)

El DMC reproduce samples de audio almacenados en la ROM. Usa **modulación delta** donde cada bit indica si la salida sube (+2) o baja (-2).

> **Nota**: Nuestra implementación actual del DMC es mínima y no reproduce samples correctamente.

## Mezcla de Audio

La NES usa una fórmula de mezcla **no lineal** para combinar los canales:

```
pulse_out = 95.88 / ((8128 / (pulse1 + pulse2)) + 100)
tnd_out = 159.79 / ((1 / (triangle/8227 + noise/12241 + dmc/22638)) + 100)
output = pulse_out + tnd_out
```

Por simplicidad, nuestra implementación usa una **aproximación lineal**:
```go
pulseOut := 0.00752 * (p1 + p2)
tndOut := 0.00851*t + 0.00494*n + 0.00335*d
return pulseOut + tndOut
```

## Implementación en Go-NES

### Estructura del APU

```go
type APU struct {
    Pulse1   *PulseChannel
    Pulse2   *PulseChannel
    Triangle *TriangleChannel
    Noise    *NoiseChannel
    DMC      *DMCChannel
    
    frameCounterMode byte
    irqInhibit       bool
    cycleCount       int
    audioTimer       float64  // Para sampling preciso
    AudioBuffer      chan float32
}
```

### Sampling de Audio

Para generar audio a 44100 Hz desde un reloj de CPU de ~1.79 MHz:

```go
// En APU.Tick():
a.audioTimer += 1.0
if a.audioTimer >= 40.5844 { // 1789773 / 44100
    a.audioTimer -= 40.5844
    select {
    case a.AudioBuffer <- a.Output():
    default: // Buffer lleno
    }
}
```

### Integración con Ebiten

El audio se reproduce mediante un stream que convierte float32 a PCM 16-bit:

```go
type AudioStream struct {
    AudioBuffer chan float32
}

func (s *AudioStream) Read(p []byte) (int, error) {
    // Convierte float32 a int16 stereo
    for n < len(p)-3 {
        sample := <-s.AudioBuffer
        val := int16(sample * 30000)
        // Escribir L y R
        p[n], p[n+1] = byte(val), byte(val>>8)
        p[n+2], p[n+3] = byte(val), byte(val>>8)
        n += 4
    }
    return n, nil
}
```

## Desafíos de la Emulación de Audio

1. **Sincronización**: El audio debe generarse a una tasa constante independiente del framerate del juego.
2. **Latencia**: Buffers grandes = más latencia, buffers pequeños = más cortes.
3. **Precisión de Timer**: Usar divisores enteros causa drift; usamos `float64` para precisión.

## Conclusión

El APU es un chip sorprendentemente sofisticado para su época. Aunque nuestra implementación es básica (sin sweep funcional, DMC incompleto), demuestra los principios fundamentales de la síntesis de audio en sistemas retro.

Para implementaciones más precisas, consultar la [Nesdev Wiki: APU](https://www.nesdev.org/wiki/APU).
