# Lecciones Aprendidas (y Robadas) para Arreglar SMB3 y Contra

Este documento detalla los "trucos", conceptos y correcciones específicas que tuvimos que aplicar (o "robar" del conocimiento colectivo de la comunidad de emulación) para que juegos complejos como **Super Mario Bros 3** y **Contra** funcionaran correctamente en este emulador.

## 1. El Misterio de Super Mario Bros 3 (Mapper 4 / MMC3)

**El Problema:**
Al ejecutar SMB3, la pantalla se quedaba en negro o mostraba basura gráfica, y el juego crasheaba inexplicablemente, a pesar de tener implementado el Mapper 4.

**Lo que aprendimos (robamos):**
El chip MMC3 es complejo, pero el juego depende de algo muy simple que a menudo se olvida: **Work RAM (WRAM)**.

*   **El Concepto:** Muchos cartuchos de MMC3 (TxROM) incluyen 8KB de memoria RAM extra en el cartucho, mapeada en `$6000-$7FFF`.
*   **La Falla:** Nuestro emulador devolvía `0` (Open Bus) para esa región. SMB3 usa esta memoria para guardar estado crítico del nivel y variables. Al no poder escribir/leer, el juego colapsaba.
*   **La Solución:**
    *   Añadir un array de 8KB `WRAM` al `Cartridge`.
    *   Habilitar lectura/escritura en `$6000-$7FFF` en el `Bus`.

> **Lección:** ¡Revisa siempre si el Mapper soporta RAM extra (PRG-RAM/WRAM)! No todo es ROM.

## 2. El Fantasma de Contra (PPU Timing)

**El Problema:**
Contra funcionaba, pero los sprites (el jugador y enemigos) dejaban un rastro o "fantasma" visual al moverse. Parecía un efecto de motion blur mal hecho.

**Lo que aprendimos (robamos):**
El manejo de **CHR-RAM** (memoria de video en RAM) es delicado, pero no tanto como creíamos.

*   **El Error:** En un intento de ser "inteligentes" y evitar *tearing* (parpadeo), implementamos un sistema de **Double Buffering** para la CHR-RAM. Escribíamos en un buffer y leíamos de otro, intercambiándolas solo al final del frame.
*   **La Realidad:** La NES real no tiene doble buffer de video de esa manera. Si la CPU escribe en CHR-RAM, el cambio es **inmediato** y visible por la PPU en ese mismo instante (o scanline).
*   **La Causa del Ghosting:** Nuestro double buffering introducía un **lag de 1 frame** en los gráficos. La posición lógica del personaje se actualizaba (CPU), pero su gráfico (PPU) se mostraba en la posición del frame anterior hasta que se hacía el swap.
*   **La Solución:** Eliminar toda la lógica de buffers. Lectura y escritura directa a `Cartridge.CHR`.

> **Lección:** A veces, la implementación más simple (acceso directo) es la correcta. No intentes solucionar problemas de hardware moderno (tearing) con lógica que el hardware original no tenía.

## 3. Fuentes y Agradecimientos (El "Loot")

Nada de esto se inventó desde cero. Todo el conocimiento técnico necesario para estos fixes salió de los gigantes sobre cuyos hombros nos paramos:

*   **[Nesdev Wiki](https://www.nesdev.org/wiki/Nesdev_Wiki)**: La fuente definitiva de verdad. Sin sus diagramas de timing del PPU y la documentación del registro MMC3, estaríamos perdidos.
    *   Específicamente: [MMC3](https://www.nesdev.org/wiki/MMC3) y [PPU Rendering](https://www.nesdev.org/wiki/PPU_rendering).
*   **[Fogleman/nes](https://github.com/fogleman/nes)**: El estándar de oro para emuladores en Go. Siempre útil para comparar "cómo lo hizo él" cuando nuestra lógica falla.

