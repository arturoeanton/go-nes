# Issues y Limitaciones Conocidas

Este documento lista las limitaciones actuales y problemas conocidos de **GoNES**.

## 1. Falta de Audio (APU)
La limitación más obvia es la ausencia total de sonido. La **APU (Audio Processing Unit)** no ha sido implementada.
- **Impacto**: Silencio total. Algunos juegos que sincronizan lógica con el frame counter de la APU podrían comportarse de forma extraña (aunque la mayoría usa NMI del PPU).

## 2. Persistencia de Partidas (SRAM)
Aunque el emulador soporta Cartuchos con batería (como *The Legend of Zelda* o *Kirby*), actualmente **no se guarda el archivo .sav** en el disco.
- **Impacto**: Si cierras el emulador, pierdes tu progreso en juegos con guardado.

## 3. Opcodes Ilegales/No Documentados
El CPU 6502 del NES tiene instrucciones "no oficiales" que resultan de combinaciones de bits.
- **Estado Actual**: La mayoría se tratan como `NOP` (No Operation) con una advertencia en el log.
- **Impacto**: Juegos que dependen críticamente de efectos secundarios de opcodes ilegales podrían glitchear o crashear. Sin embargo, se ha verificado que juegos como *Prince of Persia* funcionan correctamente con este enfoque.

## 4. Precisión del PPU
El PPU renderiza píxeles ciclo a ciclo, pero algunas sutilezas del timing analógico o efectos de "mid-scanline" muy precisos (como cambios de scroll en el *mismo* scanline para efectos de agua) podrían no ser 100% perfectos.
- **Estado**: Funcional para la gran mayoría de juegos (incluyendo scroll split de *SMB3*).
- **Correcciones Recientes**: Se eliminó el "Double Buffering" de CHR-RAM que causaba ghosting en juegos como *Contra*. Ahora la actualización es inmediata.

## 5. Entrada
Solo se soporta el **Controlador 1** mapeado al teclado.
- No hay soporte para Controlador 2.
- No hay soporte para Zapper (Pistola).
