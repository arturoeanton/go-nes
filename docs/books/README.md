# Libro: Construyendo un Emulador de NES en Go

Bienvenido a la documentación educativa de **GoNES**. Esta serie de documentos explica en detalle la teoría y la práctica detrás de cada componente del emulador que has visto en funcionamiento.

## Índice de Contenidos

### Parte I: El Corazón (CPU)
1.  [Introducción y Arquitectura NES](capitulo_01.md) - Visión general del hardware.
2.  [CPU 6502: Conceptos Básicos](capitulo_02.md) - Registros, Flags y el ciclo Fetch-Decode-Execute.
3.  [CPU 6502: Opcodes y Ciclos](capitulo_03.md) - implementando instrucciones y timing preciso.
4.  [CPU 6502: Modos de Direccionamiento](capitulo_04.md) - Cómo la CPU accede a la memoria.
5.  [Stack e Interrupciones](capitulo_05.md) - Manejo de la pila, NMI, IRQ y RESET.

### Parte II: El Sistema (Bus y Memoria)
6.  [Bus del Sistema](capitulo_06.md) - Interconectando CPU, PPU y RAM. Mapa de Memoria.
7.  [Cartuchos y Mapper 0](capitulo_07.md) - Leyendo ROMs (iNES) e implementando el mapper más simple.

### Parte III: Los Gráficos (PPU)
8.  [PPU: Arquitectura](capitulo_08.md) - Registros, VRAM, Paletas y Color.
9.  [PPU: Rendering de Fondo](capitulo_09.md) - Nametables, Attribute Tables y patrones.
10. [PPU: Sprites](capitulo_10.md) - OAM, evaluación de sprites, 8x8 vs 8x16.
11. [Scrolling y Mirroring](capitulo_11.md) - Cómo funciona el desplazamiento y los espejos de memoria.

### Parte IV: Avanzado (Mappers y Periféricos)
12. [Mappers Avanzados: MMC1](capitulo_12.md) - Registros de desplazamiento, bancos CHR/PRG.
13. [Mappers Avanzados: MMC3](capitulo_13.md) - IRQ basado en Scanlines y bancos finos.
14. [Entrada y Controladores](capitulo_14.md) - Polling, strobing y manejo de input.
15. [Conclusión](capitulo_15.md) - Resumen, lo que faltó (APU) y siguientes pasos.
