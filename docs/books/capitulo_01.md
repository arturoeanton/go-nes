# Capítulo 1: Introducción y Arquitectura de la NES

## La Nintendo Entertainment System (NES)

La NES, lanzada en Japón como Famicom en 1983, es una consola de 8 bits que definió una era. Para emularla, necesitamos simular sus componentes internos clave trabajando en sincronía.

## Arquitectura General

El sistema se compone de tres partes fundamentales conectadas por un Bus de Datos:

1.  **CPU (Central Processing Unit)**: El cerebro. Una variante del MOS 6502 fabricada por Ricoh (2A03 para NTSC, 2A07 para PAL). Es responsable de la lógica del juego, la IA y de decirle al resto de componentes qué hacer.
2.  **PPU (Picture Processing Unit)**: El chip gráfico. Mucho más complejo que la CPU en ciertos aspectos, genera la señal de video compuesta. Maneja fondos (tilemaps) y objetos móviles (sprites).
3.  **APU (Audio Processing Unit)**: Integrada en el chip de la CPU (normalmente), genera ondas de sonido cuadradas, triangulares y ruido. *Nota: En este proyecto educativo, por simplicidad, no emularemos la APU.*

Adicionalmente, el sistema tiene memoria RAM (2KB internos para la CPU, 2KB para la PPU) y la circuitería del cartucho (ROMs y Mappers).

## ¿Cómo funciona un Emulador?

Un emulador es un bucle infinito ("Game Loop") que:
1.  Ejecuta una instrucción de la **CPU**.
2.  Ejecuta 3 ciclos de la **PPU** (la PPU es 3 veces más rápida que la CPU en NTSC).
3.  Comprueba si hay **Interrupciones** (input, timers).
4.  Renderiza el frame de video resultante en tu pantalla moderna (usando Ebiten en nuestro caso).

En los siguientes capítulos, implementaremos cada uno de estos componentes paso a paso.
