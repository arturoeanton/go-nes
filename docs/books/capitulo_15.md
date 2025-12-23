# Capítulo 15: Conclusión

¡Felicidades! Hemos recorrido todo el camino desde los opcodes básicos de la CPU hasta los complejos sistemas de interrupciones del MMC3.

## Lo que hemos logrado

*   Un emulador funcional capaz de correr juegos legendarios.
*   Comprensión profunda de cómo el hardware de 8 bits gestiona recursos limitados.
*   Implementación de sistemas en tiempo real (PPU pipeline).

## Lo que falta (Retos para el lector)

Este proyecto educativo dejó fuera deliberadamente la **APU (Audio)**. Implementarla sería el siguiente gran paso:
1.  Generadores de ondas cuadradas (Square 1 & 2) con duty cycles y sweep.
2.  Generador de onda triangular (Triangle) para bajos.
3.  Generador de ruido (Noise) para percusión y explosiones.
4.  Canal DMC (Delta Modulation Channel) para samples de voz (como el "Go!" de los perros en Duck Hunt).

Otros desafíos:
*   Soporte para **Save States** (serializar toda la struct NES a disco).
*   Soporte para **Light Gun** (Zapper).
*   Cycle-accuracy perfecta en PPU (pasar tests como blargg's ppu tests).

¡Gracias por leer y Happy Hacking!
