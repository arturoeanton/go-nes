# Capítulo 14: Entrada y Controladores

La NES usa un registro de desplazamiento serial para leer el estado de los 8 botones del control (A, B, Select, Start, Up, Down, Left, Right).

## Registro $4016

La CPU lee la dirección de memoria `$4016`.
Pero como es un puerto serial de 1 bit, no lees "todo el estado" de una vez.

### El Protocolo de Strobe

1.  **Escritura (Strobe)**: El juego escribe 1 y luego 0 en `$4016`. Esto le dice al controlador "captura (latch) el estado actual de todos los botones".
2.  **Lecturas Secuenciales**:
    *   1ª Lectura de `$4016`: Retorna el estado del botón **A** (en el bit 0).
    *   2ª Lectura de `$4016`: Retorna el estado del botón **B**.
    *   ...
    *   8ª Lectura: Retorna el botón **Right**.
    *   Lecturas siguientes: Retornan 1 (generalmente).

## Implementación

En `input.go` y `bus.go`:
Mantenemos un byte `controllerState` que contiene el estado en vivo (actualizado por Ebiten cada frame).
Cuando la CPU hace "Strobe", copiamos ese estado a una variable `controllerShift`.
Cada lectura desplaza `controllerShift` a la derecha 1 bit y retorna el bit saliente.

```go
func (b *Bus) ReadController(addr uint16) byte {
    if addr == 0x4016 {
        data := (b.controllerShift & 0x80) >> 7
        b.controllerShift <<= 1
        return data
    }
    return 0
}
```
*Nota: El orden de bits y shift en la implementación real puede variar (LSB vs MSB first), pero el concepto es este.*
