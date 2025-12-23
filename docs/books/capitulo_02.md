# Capítulo 2: CPU 6502 - Conceptos Básicos

La CPU de la NES es una Ricoh 2A03, que es básicamente un MOS Technology 6502 sin el soporte para decimales (BCD).

## Estructura de la CPU

En `cpu.go`, definimos nuestra estructura `CPU`. Lo más importante son sus **Registros**. Los registros son pequeñas celdas de memoria de acceso ultrarrápido dentro del procesador.

### Registros Principales (8 bits)

*   **A (Acumulador)**: El registro principal para operaciones aritméticas y lógicas. Casi todo pasa por aquí.
*   **X e Y (Índices)**: Usados para bucles y direccionar memoria con offsets (ej: leer el elemento 5 de un array).
*   **SP (Stack Pointer)**: Apunta a la ubicación actual en la Pila (Stack). La pila está hardcodeada en la página 1 de la memoria ($0100-$01FF).
*   **P (Status Register / Flags)**: Contiene 8 bits (banderas) que indican el resultado de la última operación.

### Registro de Programa (16 bits)

*   **PC (Program Counter)**: El más importante. Indica la dirección de memoria de la *siguiente instrucción* a ejecutar.

## Flags del Registro P

Cada bit del registro P tiene un significado:
*   **C (Carry)**: Acarreo en sumas/restas.
*   **Z (Zero)**: Se activa si el resultado de una operación es 0.
*   **I (Interrupt Disable)**: Si está activo, la CPU ignora interrupciones IRQ.
*   **D (Decimal)**: Modo decimal (no usado en NES, pero existe en 6502 genérico).
*   **B (Break)**: Indica si una interrupción fue por software (instrucción BRK).
*   **V (Overflow)**: Desbordamiento en operaciones con signo.
*   **N (Negative)**: Se activa si el bit 7 (signo) del resultado es 1.

## Ciclo Fetch-Decode-Execute

Nuestra función `Run()` en el código hace lo siguiente:
1.  **Fetch**: Lee el byte en la memoria apuntada por `PC`. Ese byte es el "Opcode" (código de operación).
2.  **Decode**: Mira qué instrucción es (ej: 0xA9 es `LDA Immediate`).
3.  **Execute**: Realiza la operación (ej: Cargar un valor en A), actualiza los Flags, e incrementa el PC.
4.  Retorna el número de **Ciclos** que tomó la operación (crucial para sincronizar con la PPU).
