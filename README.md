# GoNES - Emulador de NES Educativo en Go

> [!WARNING]
> **PROYECTO EDUCATIVO**: Este emulador ha sido diseñado exclusivamente con fines de investigación y aprendizaje. **NO es un producto destinado a producción** ni busca competir con emuladores de alto rendimiento o precisión perfecta. Su objetivo es hacer el código legible y comprensible para explicar cómo funcionan las consolas internamente. Use con fines académicos.

¡Bienvenido a **GoNES**! Este proyecto es una implementación educativa de un emulador de **Nintendo Entertainment System (NES)** escrito completamente en **Go** (Golang).

El objetivo principal de este proyecto es desmitificar cómo funcionan las consolas retro por dentro, ofreciendo un código limpio, legible y documentado para que cualquiera pueda aprender sobre emulación, arquitectura de computadoras y el funcionamiento del hardware de la NES.

## 🚀 Lo que se logró

Hemos construido un sistema capaz de ejecutar juegos comerciales icónicos y complejos.

### Características Principales:

*   **CPU 6502 Completa**: Implementación fiel del procesador Ricoh 2A03 (basado en MOS 6502).
    *   Soporte de todos los opcodes oficiales y manejo de instrucciones ilegales (como NOPs).
    *   Manejo de interrupciones (NMI, IRQ, RESET).
    *   Emulación ciclo a ciclo para sincronización precisa.
*   **PPU (Picture Processing Unit)**: Emulación gráfica avanzada.
    *   Rendering basado en Scanlines.
    *   Soporte de Backgrounds (Nametables) y Scroll.
    *   Soporte de Sprites de 8x8 y **8x16** (usado en juegos como *Tecmo World Cup Soccer*).
    *   Mirroring Dinámico (Horizontal, Vertical, Single Screen).
    *   CHR-RAM Double Buffering (solución para glitches visuales).
*   **Sistema de Cartuchos y Mappers**: Soporte para múltiples tipos de hardware de cartucho.
    *   **Mapper 0 (NROM)**: *Super Mario Bros*, *Donkey Kong*.
    *   **Mapper 1 (MMC1)**: *Prince of Persia*, *Metroid*, *Zelda*. (Incluye soporte de carga serial).
    *   **Mapper 2 (UxROM)**: *Contra* (Sin ghosting de sprites), *Castlevania*.
    *   **Mapper 3 (CNROM)**: *Cybernoid*.
    *   **Mapper 4 (MMC3)**: *Super Mario Bros 3* (Soporte de WRAM y Scanline IRQ).
*   **Visualización**: Uso de la librería **Ebiten** para renderizado de buffers de píxeles modernos a 60 FPS.

## 🔇 Nota sobre el Audio

Por razones de simplicidad educativa y para mantener el foco en la lógica central de la computación (CPU/PPU) y el manejo de memoria, **no se ha implementado el APU (Audio Processing Unit)**. El emulador es totalmente silencioso. Esto permite estudiar el código sin la complejidad adicional de la generación de ondas de sonido y sincronización de audio.

## 📚 Documentación Educativa

Este repositorio incluye un "Libro" completo explicando paso a paso cómo se construyó este emulador.
Puedes encontrarlo en el directorio `docs/books/`.

### Índice del Libro:

1.  [Introducción y Arquitectura NES](docs/books/capitulo_01.md)
2.  [CPU 6502: Conceptos Básicos](docs/books/capitulo_02.md)
3.  [CPU 6502: Opcodes y Ciclos](docs/books/capitulo_03.md)
4.  [CPU 6502: Modos de Direccionamiento](docs/books/capitulo_04.md)
5.  [Stack e Interrupciones](docs/books/capitulo_05.md)
6.  [Bus del Sistema](docs/books/capitulo_06.md)
7.  [Cartuchos y Mapper 0](docs/books/capitulo_07.md)
8.  [PPU: Arquitectura](docs/books/capitulo_08.md)
9.  [PPU: Rendering de Fondo](docs/books/capitulo_09.md)
10. [PPU: Sprites](docs/books/capitulo_10.md)
11. [Scrolling y Mirroring](docs/books/capitulo_11.md)
12. [Mappers Avanzados: MMC1](docs/books/capitulo_12.md)
13. [Mappers Avanzados: MMC3](docs/books/capitulo_13.md)
14. [Entrada y Controladores](docs/books/capitulo_14.md)
14. [Entrada y Controladores](docs/books/capitulo_14.md)
15. [Conclusión](docs/books/capitulo_15.md)

### Extras:
*   [Lecciones Aprendidas y Robadas](LECCIONES_APRENDIDAS_ROBADAS.md): Documento sobre los fixes específicos para SMB3 y Contra.


## 🐛 Issues Conocidos

Para ver la lista de limitaciones conocidas o bugs pendientes, consulta [docs/issue.md](docs/issue.md).

## Ejecución

```bash
go run . ruta/al/rom.nes
```
Ejemplos:
```bash
go run . nes/SuperMarioBros.nes
go run . nes/Prince.nes
```

## Referencias y Recursos Adicionales

Este proyecto tiene fines puramente educativos. Para aquellos interesados en profundizar más en la emulación de NES o buscar implementaciones destinadas a un uso general o más avanzado, se recomiendan los siguientes recursos:

*   **Nesdev Wiki**: La enciclopedia técnica definitiva sobre el hardware de la NES. Imprescindible.
    *   [https://www.nesdev.org/wiki/Nesdev_Wiki](https://www.nesdev.org/wiki/Nesdev_Wiki)
*   **fogleman/nes**: Una implementación de referencia en Go. Es un excelente ejemplo de código Go idiomático y limpio.
    *   [https://github.com/fogleman/nes](https://github.com/fogleman/nes)
    
Otros que vi que me parecieron mejores que este :P 
* [https://leeteng.com/blog/content/writing-nes-emulator](https://leeteng.com/blog/content/writing-nes-emulator)
* [https://github.com/nwidger/nintengo](https://github.com/nwidger/nintengo)
* [https://github.com/maxpoletaev/dendy](https://github.com/maxpoletaev/dendy)

## ❤️ Agradecimientos

Un agradecimiento especial a la comunidad de **Nesdev** por mantener viva la documentación técnica que hizo posible corregir los bugs de Super Mario Bros 3 y Contra. Sin esa wiki, este emulador sería solo una pantalla negra.
