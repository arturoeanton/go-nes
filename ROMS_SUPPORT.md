# Lista de ROMs Soportados

Este documento mantiene un registro de los juegos que han sido probados y verificados como funcionales en **GoNES**.

## 🟢 Verificados Recientemente (Funcionan Perfecto)

Estos juegos han sido objeto de sesiones intensivas de debugging y funcionan correctamente con las últimas mejoras (WRAM, PPU fixes).

| Juego | Mapper | Notas |
| :--- | :--- | :--- |
| **Ninja Gaiden** | Mapper 1 (MMC1) | Funciona Perfecto. Requiere soporte de Sprites 8x16 (implementado) y corrección de Mirroring inicial (fixed). |
| **Batman - Return of the Joker** | Mapper 69 (FME-7) | **Funciona Perfecto**. Nueva implementación de Mapper 69 con IRQ por ciclos de CPU. |
| **Bomberman** | Mapper 0 (NROM) | Funciona Perfecto. |
| **Adventure Island** | Mapper 3 (CNROM) | Funciona Perfecto. |
| **Bubble Bobble** | Mapper 1 (MMC1) | Funciona Perfecto. Fix aplicado a crash por índice de paleta. |
| **Super Mario Bros 3** | Mapper 4 (MMC3) | Requiere soporte de WRAM (arreglado). IRQ de Scanline funciona bien. |
| **Contra** | Mapper 2 (UxROM) | Requiere timing preciso de CHR-RAM (sin double buffering) para evitar ghosting. |
| **Super Mario Bros** | Mapper 0 (NROM) | Verificado. Funciona perfecto. |
| **Prince of Persia** | Mapper 1 (MMC1) | Verificado. Funciona perfecto. |
| **Tecmo World Cup Soccer** | Mapper 0 (NROM) | Verificado. Sprites 8x16 funcionan bien. |

## ❗ Con Problemas Conocidos

| Juego | Mapper | Problema |
| :--- | :--- | :--- |
| **Bubble Bobble** | Mapper 1 (MMC1) | **Crash al inicio**. Error de índice fuera de rango en paleta. Requiere debug. |
| **Batman - Return of the Joker** | Mapper 69 (FME-7) | **No funciona**. Mapper 69 no implementado. |


## 🟡 Soportados (Según Implementación)

Estos juegos utilizan Mappers que están implementados y deberían funcionar, basados en pruebas anteriores.

| Juego | Mapper | Notas |
| :--- | :--- | :--- |
| **Donkey Kong** | Mapper 0 (NROM) | Simple. |
| **The Legend of Zelda** | Mapper 1 (MMC1) | Usa guardado de batería (SRAM) - *El guardado en archivo aún no está implementado*. |
| **Castlevania** | Mapper 2 (UxROM) | Mismo mapper que Contra. |
| **Cybernoid** | Mapper 3 (CNROM) | Usa intercambio de bancos solo para gráficos (CHR). |

## 🔴 No Soportados / Con Problemas

*   Juegos que usen **Mapper 9** (Punch-Out!!) o superiores no implementados.
*   Juegos que dependan críticamente del **Audio (APU)** para la lógica del juego (se congelan si no hay IRQ de APU).
