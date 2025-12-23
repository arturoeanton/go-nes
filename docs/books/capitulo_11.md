# Capítulo 11: Scrolling y Mirroring

Una pantalla de 256x240 no es suficiente para un mundo como el de Mario. La NES usa técnicas de Scroll y Mirroring para simular mundos gigantes.

## Nametables Lógicas vs Físicas

La PPU puede direccionar 4 Nametables Lógicas ($2000, $2400, $2800, $2C00). Imagine esto como un lienzo de 2x2 pantallas.
Sin embargo, la NES solo tiene 2KB de VRAM interna, suficiente para solo **2 Nametables Físicas**.
Para llenar los huecos, se usan "Espejos" (Mirroring).

### Tipos de Mirroring

1.  **Horizontal**:
    *   Las pantallas están dispuestas verticalmente A sobre B.
    *   NT $2000 y NT $2400 son iguales (NT física 0).
    *   NT $2800 y NT $2C00 son iguales (NT física 1).
    *   Ideal para scrolling vertical (juegos de naves como *Ice Climber*).

2.  **Vertical**:
    *   Las pantallas están dispuestas horizontalmente A al lado de B.
    *   NT $2000 y NT $2800 son iguales (NT física 0).
    *   NT $2400 y NT $2C00 son iguales (NT física 1).
    *   Ideal para scrolling horizontal (como *Super Mario Bros*).

3.  **Single Screen (One Screen)**:
    *   Todas las nametables apuntan a la misma área física de memoria. Usado en juegos sin scroll o con cargas de pantalla completa.

4.  **4-Screen**: El cartucho trae VRAM extra para tener 4 nametables reales únicas (ej: *Gauntlet*).

## Scroll Fino y Grueso
*   **Scroll Grueso (Nametable Select)**: PPUCTRL Bits 0-1 eligen cuál de las 4 NT lógicas es la base.
*   **Scroll Fino (PPUSCROLL)**: Desplaza la vista píxel a píxel dentro de esa NT.

Cuando Mario camina a la derecha:
1.  El Scroll X aumenta.
2.  Cuando pasa de 255, el Scroll X vuelve a 0 y se cambia el bit de Nametable base en PPUCTRL.
3.  La PPU dibuja la nueva parte del mundo que "entra" por la derecha.
