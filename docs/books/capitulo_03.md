# Capítulo 3: CPU 6502 - Opcodes y Ciclos

El dialecto del 6502 tiene 56 instrucciones (mnemónicos) oficiales, pero combinadas con distintos modos de direccionamiento, resultan en 151 Opcodes (códigos hexadecimales) únicos oficiales.

## Implementando una Instrucción: LDA (Load Accumulator)

Veamos cómo implementamos `LDA` (Cargar en A). Esta instrucción carga un valor de memoria en el registro A y actualiza los flags Z (Zero) y N (Negative).

```go
func (c *CPU) LDA(mode AddressingMode) {
    addr := c.getOperandAddress(mode) // Obtener dirección de memoria según el modo
    value := c.Bus.Read(addr)         // Leer valor
    c.A = value                       // Guardar en A
    c.updateZN(c.A)                   // Actualizar flags Z y N
}
```

La función `updateZN` es un helper vital:
```go
func (c *CPU) updateZN(value byte) {
    if value == 0 {
        c.SetFlag(FlagZ, true)
    } else {
        c.SetFlag(FlagZ, false)
    }
    if value&0x80 != 0 { // Verificar bit 7
        c.SetFlag(FlagN, true)
    } else {
        c.SetFlag(FlagN, false)
    }
}
```

## El Switch Gigante

En `cpu.go`, tenemos un enorme `switch` que despacha cada opcode:

```go
func (c *CPU) Step() int {
    opcode := c.Bus.Read(c.PC)
    c.PC++

    switch opcode {
    case 0xA9: // LDA Immediate
        c.LDA(ModeImmediate)
        return 2 // Ciclos
    case 0xA5: // LDA ZeroPage
        c.LDA(ModeZeroPage)
        return 3
    // ... y así 150 veces más
    }
}
```

Es tedioso, pero es la forma más clara de escribir un emulador.

## Ciclos de Reloj

Cada instrucción toma un tiempo fijo. `LDA Immediate` toma 2 ciclos. `LDA Absolute` toma 4. Algunos opcodes toman +1 ciclo si cruzan límites de página de memoria (page crossing).
En nuestro emulador, devolvemos el conteo base. Para mayor precisión, el método `getOperandAddress` a veces retorna un booleano `pageCrossed` para sumar ese ciclo extra.
