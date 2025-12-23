# Capítulo 10: PPU - Rendering de Sprites

Los sprites son los objetos móviles (Mario, enemigos, balas). La NES soporta hasta 64 sprites definidos en la **OAM (Object Attribute Memory)**, pero solo puede dibujar 8 por línea de escaneo (Scanline).

## OAM (256 bytes)

Cada sprite se define con 4 bytes:
1.  **Y Position**: Coordenada Y vertical.
2.  **Tile ID**: Índice del gráfico en el Pattern Table.
3.  **Attributes**:
    *   Bits 0-1: Paleta de Sprite (0-3).
    *   Bit 5: Prioridad (0=Delante del fondo, 1=Detrás del fondo).
    *   Bit 6: Flip Horizontal.
    *   Bit 7: Flip Vertical.
4.  **X Position**: Coordenada X horizontal.

## Modos 8x8 vs 8x16

*   **8x8**: El modo estándar. El Tile ID selecciona un tile de la tabla especificada en PPUCTRL.
*   **8x16**: Usado en juegos como *Tecmo World Cup Soccer*. El Tile ID tiene un significado especial:
    *   Si Tile ID es par (ej: 0x46), usa Tiles 0x46 y 0x47 de la Tabla 0 ($0000).
    *   Si Tile ID es impar (ej: 0x47), usa Tiles 0x46 y 0x47 de la Tabla 1 ($1000).
    *   Esto permite duplicar el espacio de patrones accesibles sin cambiar PPUCTRL.

## Pipeline de Sprites

1.  **Evaluación**: Durante cada scanline, la PPU revisa la OAM para encontrar los primeros 8 sprites que caen en la línea actual (`Y <= Scanline < Y + Height`).
2.  **Fetch**: Lee los patrones gráficos de esos sprites.
3.  **Draw**: Cuando el haz de electrones pasa por la coordenada X del sprite, dibuja el píxel si no es transparente (Color 0).

## Limitación de 8 Sprites
Si hay más de 8 sprites en una línea, el bit de **Sprite Overflow** en PPUSTATUS se activa y los sprites sobrantes no se dibujan. Esto causa el famoso "parpadeo" de NES cuando hay muchos enemigos. Los juegos rotan el orden de sprites en OAM para que el parpadeo se distribuya y no desaparezcan sprites permanentemente.
