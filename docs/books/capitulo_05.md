# Capítulo 5: Stack e Interrupciones

## El Stack (Pila)

La pila es una estructura LIFO (Last In, First Out) en el rango de memoria `$0100-$01FF`.
El **SP (Stack Pointer)** empieza en `$FD` (o `$FF`) y decrece cuando "pusheamos" datos.

*   `PHA`: Push A (Guarda A en el Stack).
*   `PLA`: Pull A (Recupera A del Stack).
*   `JSR` (Jump to Subroutine): Guarda el PC actual en el Stack antes de saltar.
*   `RTS` (Return from Subroutine): Recupera el PC del Stack para volver.

Implementamos helpers `push(byte)` y `pop()` en `cpu.go` para manejar esto.

## Interrupciones

Las interrupciones detienen el flujo normal del programa para atender eventos urgentes.

### Tipos de Interrupciones

1.  **NMI (Non-Maskable Interrupt)**:
    *   **Causa**: Generada por la **PPU** al inicio del VBlank (cuando termina de dibujar un frame).
    *   **Importancia**: Vital. Es el "latido del corazón" gráfico. Los juegos usan NMI para actualizar gráficos y música una vez por frame (60Hz).
    *   **Vector**: El CPU salta a la dirección almacenada en `$FFFA-$FFFB`.
    *   **No Enmascarable**: El flag `I` no puede bloquearla.

2.  **IRQ (Interrupt Request)**:
    *   **Causa**: Hardware externo (Mappers complejos como MMC3) o APU.
    *   **Vector**: Salta a la dirección en `$FFFE-$FFFF`.
    *   **Enmascarable**: Si el flag `I` está en 1, se ignora.

3.  **RESET**:
    *   **Causa**: Encendido de la consola o botón Reset.
    *   **Comportamiento**: Inicializa registros, silencia APU y carga el PC desde `$FFFC-$FFFD`.

### Implementación

Cuando ocurre una NMI (detectada en el bucle principal o activada por la PPU):
1.  Pusheamos el PC (High y Low) al Stack.
2.  Pusheamos el Registro P (Status) al Stack.
3.  Ponemos el flag `I` a 1 (para evitar IRQs durante la NMI).
4.  Leemos la dirección de destino desde `$FFFA`.
5.  Asignamos esa dirección al PC.

El juego ejecuta su rutina y termina con `RTI` (Return from Interrupt), que deshace todo esto.
