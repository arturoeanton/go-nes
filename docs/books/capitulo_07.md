# Capítulo 7: Cartuchos y Mapper 0 (NROM)

Los cartuchos de NES son complejos. No son solo memoria "muerta"; contienen hardware que puede reconfigurar la consola.

## Formato iNES (.nes)

Las ROMs que descargamos suelen estar en formato **iNES 1.0**. Tienen un encabezado de 16 bytes que nos dice:
*   Cuántos bancos de PRG ROM (código) tiene (unidades de 16KB).
*   Cuántos bancos de CHR ROM (gráficos) tiene (unidades de 8KB).
*   Qué **Mapper ID** usa.
*   Si tiene Battery Backed RAM, qué tipo de mirroring usa, etc.

El emulador lee este encabezado para saber cómo comportarse.

## Mapper 0: NROM

El Mapper 0 es el más simple. No tiene capacidad de "paginación" (bank switching).
Se usa en juegos tempranos como *Super Mario Bros*, *Donkey Kong*, y *Excitebike*.

### Mapeo de PRG ROM
*   Si el juego tiene 16KB de PRG: Se mapea en `$8000-$BFFF` y se repite en `$C000-$FFFF` (Mirror).
*   Si el juego tiene 32KB de PRG: Se mapea directamente en `$8000-$FFFF`.

### Mapeo de CHR ROM
*   Los 8KB de gráficos se cargan directamente en el espacio PPU `$0000-$1FFF`.

### Implementación

```go
type Mapper0 struct {
    prgBanks int
}

func (m *Mapper0) Read(addr uint16) int {
    // Si tenemos 1 banco (16KB), usamos módulo para repetir el espejo en $C000+
    if addr >= 0x8000 && addr <= 0xFFFF {
        return int(addr-0x8000) % (m.prgBanks * 16384)
    }
    return -1
}
```

Así de simple es NROM. Otros mappers (como el 1 o el 4) son mucho más complejos, como veremos más adelante.
