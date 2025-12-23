# Capítulo 13: Mappers Avanzados: MMC3 (Mapper 4)

El **MMC3** es quizás el mapper más popular de la NES, usado en *Super Mario Bros 3*, *Kirby's Adventure* y *Mega Man 3*.

## IRQ Basado en Scanlines

La característica estrella del MMC3 es su contador de IRQ. A diferencia de otros mappers que cuentan ciclos CPU, el MMC3 cuenta **Scanlines de la PPU**.
Esto permite hacer efectos de "Split Screen" (pantalla dividida) en cualquier punto de la pantalla (como la barra de estado estática abajo en SMB3) sin necesidad de hacer trucos complejos de timing CPU (Sprite 0 Hit es menos flexible).

### ¿Cómo funciona?

El MMC3 "espía" el bus de direcciones de la PPU (Address Bus A12). Cuando la PPU renderiza fondo y sprites, la señal A12 oscila (cambia de bajo a alto) en momentos predecibles del scanline.
1.  El MMC3 detecta flancos de subida en A12.
2.  Decrementa un contador interno.
3.  Cuando el contador llega a 0, dispara una **Interrupción IRQ** a la CPU.

En nuestra implementación (`ppu.go`), simulamos esto llamando a `Mapper.Scanline()` aproximadamente en el ciclo 260 de cada línea visible.

## Bank Switching Fino

El MMC3 ofrece una granularidad increíble:
*   **2 bancos de 8KB PRG** conmutables + 2 fijos (o 1 fijo móvil). Permite tener DPCM samples, código de motor y datos de nivel cargados simultáneamente.
*   **6 bancos de CHR**:
    *   2 bancos de 2KB (para sprites grandes o fondo).
    *   4 bancos de 1KB (para animaciones finas de sprites).

Esto permite animar el fondo (agua, bloques ? brillando) y sprites de forma muy eficiente cambiando solo bancos pequeños de 1KB.
