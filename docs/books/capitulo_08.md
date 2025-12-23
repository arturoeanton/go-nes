# Capítulo 8: PPU - Arquitectura y Registros

La **Picture Processing Unit (PPU)** es el corazón gráfico de la NES. Genera una señal de video compuesta de 256x240 píxeles a 60 cuadros por segundo (NTSC).

## Registros Mapeados en Memoria

La CPU se comunica con la PPU a través de 8 registros mapeados en `$2000-$2007`.

### $2000 PPUCTRL (Control)
Configura el comportamiento base:
*   Habilita/Deshabilita NMI.
*   Selecciona qué Nametable base usar (scroll grueso).
*   Selecciona tamaño de Sprites (**8x8** o **8x16**).
*   Selecciona qué Pattern Table usar para Fondo ($0000/$1000) y Sprites.

### $2001 PPUMASK (Mask)
Controla el renderizado:
*   Muestra/Oculta Fondo.
*   Muestra/Oculta Sprites.
*   Enfatiza colores (tintes).

### $2002 PPUSTATUS (Status)
Solo lectura. Informa a la CPU sobre:
*   **VBlank Flag**: El frame terminó de dibujarse.
*   **Sprite 0 Hit**: Colisión entre Sprite #0 y el fondo (crucial para sincronización).
*   **Sprite Overflow**: Más de 8 sprites en una línea.

### $2005 PPUSCROLL y $2006 PPUADDR
Controlan la dirección de video y el scroll.
Lo curioso es que **comparten un registro interno temporal** (conocido como `loopy_t` o `TempAddr`).
Escribir en ellos actualiza partes de este registro.

### $2007 PPUDATA
Puerto de lectura/escritura de VRAM. Cuando la CPU quiere cambiar un tile, escribe la dirección en $2006 y el dato en $2007.

## El "Latch" de Escritura
Los registros de 16 bits (Scroll y Address) se escriben en dos pasos (byte alto, byte bajo) a través de un puerto de 8 bits. Un "latch" interno (`w`) lleva la cuenta de si es la primera o segunda escritura. Leer `PPUSTATUS` resetea este latch.
