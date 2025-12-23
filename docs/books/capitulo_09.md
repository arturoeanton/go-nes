# Capítulo 9: PPU - Rendering de Fondo

El fondo de la NES se compone de una rejilla de **Tiles** de 8x8 píxeles. La pantalla es de 32x30 tiles (256x240 px).

## Estructura de Datos

1.  **Pattern Tables (CHR)**: La "fuente" de gráficos. Contiene los bitmaps de los tiles. Cada tile ocupa 16 bytes (2 planos de color para 2 bits por píxel = 4 colores posibles).
2.  **Nametables**: Un mapa de 32x30 bytes (1KB) que dice "qué tile va en qué posición de la pantalla".
3.  **Attribute Tables**: Los últimos 64 bytes de la Nametable. Definen la **Paleta de Color** para bloques de 16x16 píxeles (4 tiles).

## El Pipeline de Renderizado

La PPU dibuja la pantalla línea por línea (Scanline). En cada ciclo de reloj, "fetchea" (lee) datos para dibujar los píxeles.

Para dibujar un píxel de fondo en (X, Y):
1.  Determinar en qué **Nametable** estamos (según Scroll y PPUCTRL).
2.  Calcular qué **Tile ID** corresponde a (X, Y).
3.  Leer los 2 bytes del Pattern Table correspondientes a ese Tile ID y la fila (Y % 8).
4.  Combinar los bits para obtener un valor de 2 bits (0-3).
5.  Leer el byte de **Atributo** correspondiente para saber qué Paleta usar (0-3).
6.  Combinar Paleta + Color del Tile para obtener el índice final (0x00 - 0x3F).
7.  Leer el color RGB real desde la Paleta del Sistema.

## Optimizaciones en GoNES

Para simplicidad y rendimiento, a veces podemos pre-renderizar o cachear ciertos datos, pero para máxima compatibilidad (como el cambio de scroll a mitad de pantalla en SMB3), debemos emular este proceso casi pixel a pixel o al menos con granularidad fina.
