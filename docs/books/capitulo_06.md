# Capítulo 6: El Bus del Sistema

El BUS es la autopista por donde viajan los datos. En nuestra implementación en Go, el Bus es el objeto que conecta todo.

## Mapa de Memoria (CPU Map)

La CPU del 6502 puede direccionar 64KB (`0x0000 - 0xFFFF`). Pero la NES no tiene 64KB de RAM. El mapa es el siguiente:

*   **$0000 - $07FF**: RAM interna (2KB).
*   **$0800 - $1FFF**: Espejos de la RAM interna (Mirrors). Escribir en $0800 es lo mismo que en $0000.
*   **$2000 - $2007**: Registros de la PPU. Son puertos de comunicación con el chip gráfico.
*   **$2008 - $3FFF**: Espejos de los registros PPU.
*   **$4000 - $4017**: Registros de APU (Audio), I/O (Joysticks) y OAM DMA.
*   **$4018 - $401F**: Funcionalidad de prueba (normalmente deshabilitada).
*   **$4020 - $FFFF**: Espacio del Cartucho (Cartridge Space).
    *   **$6000 - $7FFF**: PRG RAM (SRAM, a veces con batería para guardar partidas).
    *   **$8000 - $FFFF**: PRG ROM (El código del juego).

## Mapa de Memoria de la PPU

La PPU tiene su propio bus de datos separado, puede direccionar 16KB (`0x0000 - 0x3FFF`):

*   **$0000 - $1FFF**: Pattern Tables (CHR ROM/RAM). Aquí viven los gráficos (tiles) de sprites y fondo.
*   **$2000 - $2FFF**: Nametables (VRAM interna). Aquí se define qué tiles se dibujan en pantalla.
*   **$3000 - $3EFF**: Espejos de $2000-$2EFF.
*   **$3F00 - $3FFF**: Paletas de color (RAM de paleta).

## Implementación en bus.go

En `bus.go`, definimos una estructura que contiene referencias a la CPU, PPU, RAM y Cartucho. Su método `Read(addr)` actúa como un guardia de tráfico:

```go
func (b *Bus) Read(addr uint16) byte {
    if addr <= 0x1FFF {
        return b.RAM[addr & 0x07FF] // Manejo de espejos
    } else if addr <= 0x3FFF {
        return b.PPU.Read(addr) // Delegar a PPU
    } else if addr >= 0x4016 && addr <= 0x4017 {
        return b.ReadController(addr) // Joysticks
    } else if addr >= 0x8000 {
        return b.Cart.Read(addr) // Delegar a Cartucho
    }
    return 0
}
```

Es crucial entender este "Routing" de direcciones.
