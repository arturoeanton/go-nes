# Capítulo 12: Mappers Avanzados: MMC1 (Mapper 1)

El chip **Memory Management Controller 1 (MMC1)** fue el primer mapper ASIC de Nintendo, permitiendo juegos más grandes y complejos como *The Legend of Zelda* y *Metroid*.

## Características
*   **Bank Switching de PRG**: Hasta 512KB de ROM.
*   **Bank Switching de CHR**: Hasta 128KB de CHR ROM/RAM.
*   **Mirroring Dinámico**: El juego puede cambiar entre H y V en tiempo real.

## Carga Serial

A diferencia del NROM o UxROM donde escribes un byte para cambiar de banco, el MMC1 tiene pocos pines. Para ahorrar espacio, usa un **Registro de Desplazamiento (Shift Register)**.
Para enviar un comando al MMC1, la CPU debe hacer **5 escrituras secuenciales** en memoria ($8000-$FFFF). En cada escritura, solo el bit 0 (el menos significativo) es capturado.

1.  Escritura 1: Bit 0 -> Shift Reg [....1]
2.  Escritura 2: Bit 0 -> Shift Reg [...10]
...
5.  Escritura 5: Bit 0 -> Shift Reg [10101] -> **¡COMANDO EJECUTADO!**

## Registros Internos

Una vez completadas las 5 escrituras, los 5 bits acumulados se copian a uno de los registros internos, dependiendo de la dirección de la última escritura:

*   **$8000-$9FFF (Control)**: Configura Mirroring y modos de bancos (8KB/4KB CHR, 16KB/32KB PRG).
*   **$A000-$BFFF (CHR Bank 0)**: Selecciona el banco de gráficos inferior (4KB).
*   **$C000-$DFFF (CHR Bank 1)**: Selecciona el banco de gráficos superior (4KB).
*   **$E000-$FFFF (PRG Bank)**: Selecciona el banco de código (16KB).

## Implementación
En `mapper.go`, la estructura `Mapper1` incluye un campo `shiftReg` y un contador `writeCnt`. Solo cuando `writeCnt` llega a 5, actualizamos el estado real del mapper. Esto fue crítico para solucionar bugs en juegos que intentan resetear el mapper escribiendo el bit 7 (Reset).
