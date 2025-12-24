# Lista de ROMs Soportados

Este documento mantiene un registro de los juegos que han sido probados y verificados como funcionales en **GoNES**.

## 🟢 Verificados Recientemente (Funcionan Perfecto)

| Juego | Mapper | Notas |
| :--- | :--- | :--- |
| **Super Mario Bros 3** | Mapper 4 (MMC3) | ✅ Funciona Perfecto. WRAM implementado. |
| **Tiny Toon Adventures** | Mapper 4 (MMC3) | ✅ Funciona. Fix de CHR banking aplicado. |
| **TMNT II: Arcade Game** | Mapper 4 (MMC3) | ✅ Funciona (warnings de RAM normales - usa trampoline). |
| **TMNT: Tournament Fighters** | Mapper 4 (MMC3) | ✅ Funciona Perfecto. |
| **Joe & Mac** | Mapper 4 (MMC3) | ✅ Funciona Perfecto. |
| **Batman - Return of the Joker** | Mapper 69 (FME-7) | ✅ Funciona. IRQ por ciclos de CPU. |
| **Ninja Gaiden** | Mapper 1 (MMC1) | ✅ Funciona. Sprites 8x16 + Mirroring. |
| **Bomberman** | Mapper 0 (NROM) | ✅ Funciona Perfecto. |
| **Adventure Island** | Mapper 3 (CNROM) | ✅ Funciona Perfecto. |
| **Bubble Bobble** | Mapper 1 (MMC1) | ✅ Funciona. Fix de paleta aplicado. |
| **Contra** | Mapper 2 (UxROM) | ✅ Funciona. CHR-RAM timing corregido. |
| **Super Mario Bros** | Mapper 0 (NROM) | ✅ Funciona Perfecto. |
| **Prince of Persia** | Mapper 1 (MMC1) | ✅ Funciona Perfecto. |

## ❗ Con Problemas Conocidos

| Juego | Mapper | Problema |
| :--- | :--- | :--- |
| **Aladdin** | Mapper 4 (MMC3) | Crash (PC en $0000). Requiere A12 edge detection preciso. |
| **Shinobi** | Mapper 4 (MMC3) | Crash (PC oscila entre $0000-$0001). Mismo problema de timing. |

## 🟡 Soportados (Según Implementación)

| Juego | Mapper | Notas |
| :--- | :--- | :--- |
| **Donkey Kong** | Mapper 0 (NROM) | Simple. |
| **The Legend of Zelda** | Mapper 1 (MMC1) | Guardado SRAM no persiste a archivo. |
| **Castlevania** | Mapper 2 (UxROM) | Mismo mapper que Contra. |
| **Cybernoid** | Mapper 3 (CNROM) | CHR bank switching. |

## 🔴 No Soportados

*   Juegos que usen **Mapper 9** (Punch-Out!!) o superiores no implementados.
*   Juegos que requieran **detección precisa de flancos A12** (Aladdin, Shinobi).

## Nota Técnica sobre MMC3

El MMC3 detecta flancos ascendentes en la línea A12 del bus PPU para contar scanlines. Nuestra implementación usa un hook `Scanline()` simplificado que funciona para la mayoría de los juegos, pero algunos títulos (como Aladdin) requieren timing pixel-perfect que no está implementado.

Los warnings de "PC executing from Low Memory" en TMNT II son normales - el juego usa un trampoline en RAM ($0051) para saltos dinámicos.
