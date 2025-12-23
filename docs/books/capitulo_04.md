# Capítulo 4: CPU 6502 - Modos de Direccionamiento

El poder del 6502 reside en sus modos de direccionamiento. Una misma instrucción como `LDA` puede comportarse muy diferente según *dónde* busque el dato.

## Modos Comunes

1.  **Implicit (Implícito)**: El operando está en la propia instrucción (ej: `TAX` transfiere A a X). No busca en memoria externa.
2.  **Immediate (Inmediato)**: El dato está justo después del opcode.
    *   Ej: `LDA #$10` -> Carga el valor 0x10 en A.
3.  **Zero Page (Página Cero)**: Usa solo 1 byte para la dirección, asumiendo que el byte alto es $00. Accede a las direcciones $0000-$00FF. Es más rápido y ocupa menos memoria.
    *   Ej: `LDA $10` -> Carga el valor en la dirección $0010.
4.  **Absolute (Absoluto)**: Usa 2 bytes para la dirección completa ($0000-$FFFF).
    *   Ej: `LDA $1234` -> Carga el valor en la dirección $1234.

## Modos Indexados (Arrays)

5.  **Absolute, X**: Toma una dirección base y le suma el registro X.
    *   Ej: `LDA $3000, X`. Si X=5, lee de $3005. Ideal para recorrer arrays.
6.  **Absolute, Y**: Igual, pero suma Y.

## Modos Indirectos (Punteros)

Estos son los más complejos y confusos del 6502.

7.  **Indirect, X (Indexed Indirect)**:
    *   Toma un byte (Zero Page), le suma X, y usa el resultado como *dirección de un puntero* de 2 bytes en la página cero. Ese puntero nos dice la dirección final real.
    *   Rara vez usado.
8.  **Indirect, Y (Indirect Indexed)**:
    *   Toma un byte (Zero Page), lee el *puntero* de 2 bytes ahí almacenado, y a esa dirección leída le suma Y.
    *   ¡Muy usado! Permite acceder a estructuras de datos grandes apuntadas por punteros en página cero.

En `cpu.go`, la función `getOperandAddress(mode)` encapsula toda esta lógica compleja para que las instrucciones (`LDA`, `STA`, etc.) no tengan que preocuparse por ello.
