package main

import (
	"log"
)

// ==========================================
// DEFINICIÓN DE FLAGS (Registro de Estado P)
// ==========================================
// El registro P contiene 8 bits (banderas) que indican el resultado
// de la última operación o el estado del procesador.
const (
	C = 1 << 0 // Carry (Acarreo): Se activa si una suma desborda el byte.
	Z = 1 << 1 // Zero (Cero): Se activa si el resultado de una operación es 0.
	I = 1 << 2 // Interrupt Disable: Si es 1, se ignoran las interrupciones IRQ.
	D = 1 << 3 // Decimal: Modo BCD (No usado en la NES, pero parte del 6502 original).
	B = 1 << 4 // Break: Indica que una interrupción fue generada por software (instrucción BRK).
	U = 1 << 5 // Unused (No usado): Este bit siempre se lee como 1.
	V = 1 << 6 // Overflow (Desbordamiento): Para operaciones con signo (complemento a 2).
	N = 1 << 7 // Negative (Negativo): Se activa si el bit 7 (signo) del resultado es 1.
)

// CPU representa el procesador Ricoh 2A03 (basado en MOS 6502).
type CPU struct {
	Bus *Bus // Referencia al Bus del sistema para leer/escribir memoria.

	// Registros del Procesador
	PC uint16 // Program Counter: Dirección de la siguiente instrucción a ejecutar.
	SP byte   // Stack Pointer: Puntero a la pila (en la página $0100-$01FF).
	A  byte   // Acumulador: Registro principal para matemáticas y lógica.
	X  byte   // Índice X: Registro auxiliar (contadores, offsets).
	Y  byte   // Índice Y: Registro auxiliar.
	P  byte   // Status Flags: Registro de estado (contiene C, Z, I, etc).
}

// NewCPU crea una nueva instancia de la CPU conectada al bus.
func NewCPU(bus *Bus) *CPU {
	return &CPU{Bus: bus}
}

// Reset inicializa la CPU a un estado conocido.
// Se llama al encender la consola o presionar el botón de Reset.
func (cpu *CPU) Reset() {
	cpu.A, cpu.X, cpu.Y = 0, 0, 0
	cpu.P = 0x24 // U(1) e I(1) activados por defecto.
	cpu.SP = 0xFD

	// Leer vector de reset desde 0xFFFC y 0xFFFD
	// El vector es una dirección de 16 bits que dice dónde empieza el programa.
	low := uint16(cpu.Bus.Read(0xFFFC))
	high := uint16(cpu.Bus.Read(0xFFFD))
	cpu.PC = (high << 8) | low
	log.Printf("CPU Reset: PC inicializado en $%04X", cpu.PC)
}

// NMI: Non-Maskable Interrupt (Interrupción No Enmascarable).
// Es la interrupción más importante en la NES. La PPU la dispara
// 60 veces por segundo (VBlank) para que el juego actualice gráficos.
// NMI: Non-Maskable Interrupt (Interrupción No Enmascarable).
// Es la interrupción más importante en la NES. La PPU la dispara
// 60 veces por segundo (VBlank) para que el juego actualice gráficos.
func (cpu *CPU) NMI() {
	log.Println("CPU NMI Triggered")
	cpu.push16(cpu.PC)         // Guardar dónde estábamos
	cpu.push(cpu.P & ^byte(B)) // Guardar estado (sin bit B)
	cpu.P |= I                 // Deshabilitar interrupciones IRQ durante la NMI

	// Saltar a la dirección especificada en el vector NMI ($FFFA)
	low := uint16(cpu.Bus.Read(0xFFFA))
	high := uint16(cpu.Bus.Read(0xFFFB))
	cpu.PC = (high << 8) | low
}

// IRQ: Interrupt Request (Interrupción Enmascarable).
// Puede ser generada por hardware externo (Mappers o audio).
// Solo se ejecuta si el flag I es 0.
func (cpu *CPU) IRQ() {
	if cpu.P&I != 0 {
		return // Interrupciones deshabilitadas, salir.
	}
	cpu.push16(cpu.PC)
	cpu.push(cpu.P & ^byte(B))
	cpu.P |= I

	// Saltar al vector IRQ ($FFFE)
	low := uint16(cpu.Bus.Read(0xFFFE))
	high := uint16(cpu.Bus.Read(0xFFFF))
	cpu.PC = (high << 8) | low
}

// ==========================================
// HELPERS DE MODOS DE DIRECCIONAMIENTO
// Dictan cómo la CPU encuentra los datos (operandos)
// ==========================================

// Immediate: El dato está justo después del opcode.
func (c *CPU) addrImm() uint16 {
	addr := c.PC
	c.PC++ // Avanzamos PC porque hemos "leído" el operando
	return addr
}

// Zero Page: Dirección de 1 byte (0x00 - 0xFF). Ahorra espacio/tiempo.
func (c *CPU) addrZP() uint16 {
	addr := uint16(c.Bus.Read(c.PC))
	c.PC++
	return addr
}

// Zero Page, X: Dirección Zero Page + Registro X. (Direccionamiento indexado).
// Útil para arrays en página cero.
func (c *CPU) addrZPX() uint16 {
	// (Addr + X) con wrap-around en 0xFF (no sale de página cero)
	addr := (uint16(c.Bus.Read(c.PC)) + uint16(c.X)) & 0xFF
	c.PC++
	return addr
}

// Zero Page, Y: Igual que ZPX pero con Y.
func (c *CPU) addrZPY() uint16 {
	addr := (uint16(c.Bus.Read(c.PC)) + uint16(c.Y)) & 0xFF
	c.PC++
	return addr
}

// Absolute: Dirección completa de 2 bytes (0x0000 - 0xFFFF).
func (c *CPU) addrAbs() uint16 {
	lo := uint16(c.Bus.Read(c.PC))
	c.PC++
	hi := uint16(c.Bus.Read(c.PC))
	c.PC++
	return (hi << 8) | lo
}

// Absolute, X: Dirección absoluta + Registro X.
func (c *CPU) addrAbsX() uint16 {
	lo := uint16(c.Bus.Read(c.PC))
	c.PC++
	hi := uint16(c.Bus.Read(c.PC))
	c.PC++
	base := (hi << 8) | lo
	return base + uint16(c.X)
}

// Absolute, Y: Dirección absoluta + Registro Y.
func (c *CPU) addrAbsY() uint16 {
	lo := uint16(c.Bus.Read(c.PC))
	c.PC++
	hi := uint16(c.Bus.Read(c.PC))
	c.PC++
	base := (hi << 8) | lo
	return base + uint16(c.Y)
}

// Indirect, X (Indexed Indirect): Pre-indexado.
// Toma un puntero de la Zero Page (addr+X) y lee la dirección final de ahí.
func (c *CPU) addrIndX() uint16 {
	t := uint16(c.Bus.Read(c.PC)) + uint16(c.X) // Dirección en ZP (offset por X)
	c.PC++
	// Leer el puntero real de 16 bits almacenado en t y t+1
	lo := uint16(c.Bus.Read(t & 0xFF))
	hi := uint16(c.Bus.Read((t + 1) & 0xFF))
	return (hi << 8) | lo
}

// Indirect, Y (Indirect Indexed): Post-indexado.
// Lee un puntero de la Zero Page y suma Y a la dirección resultante.
// Muy usado para recorrer arrays grandes.
func (c *CPU) addrIndY() uint16 {
	t := uint16(c.Bus.Read(c.PC))
	c.PC++
	lo := uint16(c.Bus.Read(t & 0xFF))
	hi := uint16(c.Bus.Read((t + 1) & 0xFF))
	base := (hi << 8) | lo
	return base + uint16(c.Y)
}

// ==========================================
// FUNCIONES AUXILIARES (Stack y Flags)
// ==========================================

// setZN actualiza las flags Zero (Z) y Negative (N) según un valor.
func (c *CPU) setZN(val byte) {
	if val == 0 {
		c.P |= Z // Resultado fue cero
	} else {
		c.P &^= Z
	}
	if val&0x80 != 0 {
		c.P |= N // Bit 7 (signo) encendido
	} else {
		c.P &^= N
	}
}

// Operaciones de Stack (Pila)
// Pila empieza en 0x01FF y crece hacia abajo.
func (c *CPU) push(val byte) { c.Bus.Write(0x0100+uint16(c.SP), val); c.SP-- }
func (c *CPU) pull() byte    { c.SP++; return c.Bus.Read(0x0100 + uint16(c.SP)) }

func (c *CPU) push16(val uint16) {
	c.push(byte(val >> 8))   // Push byte alto
	c.push(byte(val & 0xFF)) // Push byte bajo
}
func (c *CPU) pull16() uint16 {
	lo := uint16(c.pull())
	hi := uint16(c.pull())
	return (hi << 8) | lo
}

// ==========================================
// INSTRUCCIONES LÓGICAS Y MATEMÁTICAS
// ==========================================

// compare: Compara un registro con un valor.
// Básicamente hace (Reg - Val) pero sin guardar el resultado, solo flags.
func (c *CPU) compare(reg, val byte) {
	if reg >= val {
		c.P |= C // Carry se activa si el registro es mayor o igual (sin préstamo)
	} else {
		c.P &^= C
	}
	c.setZN(reg - val)
}

// ADC: Add with Carry (Suma con Acarreo).
// A = A + M + C
// Es compleja porque maneja Flags de Overflow (V) y Carry (C).
func (c *CPU) adc(addr uint16) {
	val := uint16(c.Bus.Read(addr))
	sum := uint16(c.A) + val + uint16(c.P&C)

	// Manejo de Carry
	if sum > 0xFF {
		c.P |= C
	} else {
		c.P &^= C
	}

	// Manejo de Overflow (V): La magia negra del complemento a 2.
	// Se activa si sumamos dos positivos y da negativo, o viceversa.
	// ((A ^ sum) & (val ^ sum) & 0x80)
	if (^uint16(c.A)^val)&(uint16(c.A)^sum)&0x0080 != 0 {
		c.P |= V
	} else {
		c.P &^= V
	}

	c.A = byte(sum)
	c.setZN(c.A)
}

// SBC: Subtract with Carry (Resta con Acarreo).
// A = A - M - (1-C)
// En el 6502, la resta se implementa como suma con el complemento del valor.
// A - B = A + (-B) - 1
func (c *CPU) sbc(addr uint16) {
	val := uint16(c.Bus.Read(addr)) ^ 0x00FF // Complemento a 1
	// Ahora usamos la lógica de ADC
	sum := uint16(c.A) + val + uint16(c.P&C)

	if sum > 0xFF {
		c.P |= C // En resta, Carry limpio significa "sin préstamo" (resultado positivo)
	} else {
		c.P &^= C
	}

	if (^uint16(c.A)^val)&(uint16(c.A)^sum)&0x0080 != 0 {
		c.P |= V
	} else {
		c.P &^= V
	}

	c.A = byte(sum)
	c.setZN(c.A)
}

// Operaciones Bit a Bit
func (c *CPU) and(addr uint16) { c.A &= c.Bus.Read(addr); c.setZN(c.A) }
func (c *CPU) ora(addr uint16) { c.A |= c.Bus.Read(addr); c.setZN(c.A) } // OR inclusivo
func (c *CPU) eor(addr uint16) { c.A ^= c.Bus.Read(addr); c.setZN(c.A) } // XOR exclusivo

// Shifts (Desplazamientos) y Rotaciones
// ASL: Arithmetic Shift Left (<< 1). Bit 7 cae al Carry.
func (c *CPU) asl(val byte) byte {
	if val&0x80 != 0 {
		c.P |= C
	} else {
		c.P &^= C
	}
	res := val << 1
	c.setZN(res)
	return res
}

// LSR: Logical Shift Right (>> 1). Bit 0 cae al Carry.
func (c *CPU) lsr(val byte) byte {
	if val&0x01 != 0 {
		c.P |= C
	} else {
		c.P &^= C
	}
	res := val >> 1
	c.setZN(res)
	return res
}

// ROL: Rotate Left. Rota a través del Carry.
//
//	[C] <- [76543210] <- [C]
func (c *CPU) rol(val byte) byte {
	newBit := byte(0)
	if c.P&C != 0 {
		newBit = 1
	}
	if val&0x80 != 0 {
		c.P |= C
	} else {
		c.P &^= C
	}
	res := (val << 1) | newBit
	c.setZN(res)
	return res
}

// ROR: Rotate Right. [0] -> [C] -> [7]
func (c *CPU) ror(val byte) byte {
	newBit := byte(0)
	if c.P&C != 0 {
		newBit = 0x80
	}
	if val&0x01 != 0 {
		c.P |= C
	} else {
		c.P &^= C
	}
	res := (val >> 1) | newBit
	c.setZN(res)
	return res
}

// ==========================================
// STEP: EL BUCLE DE INSTRUCCIONES
// Decodifica y ejecuta UNA instrucción.
// Retorna cuántos ciclos de CPU consumió.
// ==========================================
func (c *CPU) Step() int {
	// Verificar interrupciones IRQ pendientes (nivel bajo)
	// Solo si el flag I (Interrupt Disable) está limpio.
	apuIrq := c.Bus.APU != nil && c.Bus.APU.IRQState()
	if (c.Bus.Cart.Mapper.IRQState() || apuIrq) && (c.P&I == 0) {
		c.IRQ()
		// Una IRQ toma 7 ciclos

		// Tick del Mapper durante los ciclos de IRQ
		for i := 0; i < 7; i++ {
			c.Bus.Cart.Mapper.Tick()
		}
		return 7
	}

	opcode := c.Bus.Read(c.PC)
	cycles := cycleTable[opcode] // Buscar ciclos base en tabla
	c.PC++

	var extraCycles int

	if c.PC < 0x8000 {
		log.Printf("WARNING: PC executing from Low Memory: $%04X (Opcode: %02X)", c.PC, opcode)
	}

	switch opcode {
	// --- INSTRUCCIONES DE CARGA (LOAD) ---
	case 0xA9: // LDA Immediate
		c.A = c.Bus.Read(c.addrImm())
		c.setZN(c.A)
	case 0xA5: // LDA ZeroPage
		c.A = c.Bus.Read(c.addrZP())
		c.setZN(c.A)
	case 0xB5: // LDA ZP, X
		c.A = c.Bus.Read(c.addrZPX())
		c.setZN(c.A)
	case 0xAD: // LDA Absolute
		c.A = c.Bus.Read(c.addrAbs())
		c.setZN(c.A)
	case 0xBD: // LDA Abs, X
		c.A = c.Bus.Read(c.addrAbsX())
		c.setZN(c.A)
	case 0xB9: // LDA Abs, Y
		c.A = c.Bus.Read(c.addrAbsY())
		c.setZN(c.A)
	case 0xA1: // LDA (Indirect, X)
		c.A = c.Bus.Read(c.addrIndX())
		c.setZN(c.A)
	case 0xB1: // LDA (Indirect), Y
		c.A = c.Bus.Read(c.addrIndY())
		c.setZN(c.A)

	case 0xA2: // LDX Immediate
		c.X = c.Bus.Read(c.addrImm())
		c.setZN(c.X)
	case 0xA6: // LDX ZeroPage
		c.X = c.Bus.Read(c.addrZP())
		c.setZN(c.X)
	case 0xB6: // LDX ZP, Y
		c.X = c.Bus.Read(c.addrZPY())
		c.setZN(c.X)
	case 0xAE: // LDX Absolute
		c.X = c.Bus.Read(c.addrAbs())
		c.setZN(c.X)
	case 0xBE: // LDX Abs, Y
		c.X = c.Bus.Read(c.addrAbsY())
		c.setZN(c.X)

	case 0xA0: // LDY Immediate
		c.Y = c.Bus.Read(c.addrImm())
		c.setZN(c.Y)
	case 0xA4: // LDY ZeroPage
		c.Y = c.Bus.Read(c.addrZP())
		c.setZN(c.Y)
	case 0xB4: // LDY ZP, X
		c.Y = c.Bus.Read(c.addrZPX())
		c.setZN(c.Y)
	case 0xAC: // LDY Absolute
		c.Y = c.Bus.Read(c.addrAbs())
		c.setZN(c.Y)
	case 0xBC: // LDY Abs, X
		c.Y = c.Bus.Read(c.addrAbsX())
		c.setZN(c.Y)

	// --- INSTRUCCIONES DE ALMACENAMIENTO (STORE) ---
	case 0x85: // STA ZeroPage
		c.Bus.Write(c.addrZP(), c.A)
	case 0x95: // STA ZP, X
		c.Bus.Write(c.addrZPX(), c.A)
	case 0x8D: // STA Absolute
		c.Bus.Write(c.addrAbs(), c.A)
	case 0x9D: // STA Abs, X
		c.Bus.Write(c.addrAbsX(), c.A)
	case 0x99: // STA Abs, Y
		c.Bus.Write(c.addrAbsY(), c.A)
	case 0x81: // STA (Ind, X)
		c.Bus.Write(c.addrIndX(), c.A)
	case 0x91: // STA (Ind), Y
		c.Bus.Write(c.addrIndY(), c.A)

	case 0x86: // STX ZeroPage
		c.Bus.Write(c.addrZP(), c.X)
	case 0x96: // STX ZP, Y
		c.Bus.Write(c.addrZPY(), c.X)
	case 0x8E: // STX Absolute
		c.Bus.Write(c.addrAbs(), c.X)

	case 0x84: // STY ZeroPage
		c.Bus.Write(c.addrZP(), c.Y)
	case 0x94: // STY ZP, X
		c.Bus.Write(c.addrZPX(), c.Y)
	case 0x8C: // STY Absolute
		c.Bus.Write(c.addrAbs(), c.Y)

	// --- TRANSFERENCIAS DE REGISTRO ---
	case 0xAA: // TAX: Transfer A to X
		c.X = c.A
		c.setZN(c.X)
	case 0xA8: // TAY: Transfer A to Y
		c.Y = c.A
		c.setZN(c.Y)
	case 0x8A: // TXA: Transfer X to A
		c.A = c.X
		c.setZN(c.A)
	case 0x98: // TYA: Transfer Y to A
		c.A = c.Y
		c.setZN(c.A)
	case 0x9A: // TXS: Transfer X to SP (Stack Pointer)
		c.SP = c.X // Nota: NO actualiza flags Z/N
	case 0xBA: // TSX: Transfer SP to X
		c.X = c.SP
		c.setZN(c.X)

	// --- ARITMÉTICA ---
	case 0x69:
		c.adc(c.addrImm())
	case 0x65:
		c.adc(c.addrZP())
	case 0x75:
		c.adc(c.addrZPX())
	case 0x6D:
		c.adc(c.addrAbs())
	case 0x7D:
		c.adc(c.addrAbsX())
	case 0x79:
		c.adc(c.addrAbsY())
	case 0x61:
		c.adc(c.addrIndX())
	case 0x71:
		c.adc(c.addrIndY())

	case 0xE9:
		c.sbc(c.addrImm())
	case 0xE5:
		c.sbc(c.addrZP())
	case 0xF5:
		c.sbc(c.addrZPX())
	case 0xED:
		c.sbc(c.addrAbs())
	case 0xFD:
		c.sbc(c.addrAbsX())
	case 0xF9:
		c.sbc(c.addrAbsY())
	case 0xE1:
		c.sbc(c.addrIndX())
	case 0xF1:
		c.sbc(c.addrIndY())

	// --- INCREMENTOS / DECREMENTOS ---
	case 0xE8: // INX (X++)
		c.X++
		c.setZN(c.X)
	case 0xC8: // INY (Y++)
		c.Y++
		c.setZN(c.Y)
	case 0xCA: // DEX (X--)
		c.X--
		c.setZN(c.X)
	case 0x88: // DEY (Y--)
		c.Y--
		c.setZN(c.Y)

	case 0xE6: // INC Memoria (Zero Page)
		addr := c.addrZP()
		val := c.Bus.Read(addr) + 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xF6: // INC Memoria (ZP, X)
		addr := c.addrZPX()
		val := c.Bus.Read(addr) + 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xEE: // INC Memoria (Abs)
		addr := c.addrAbs()
		val := c.Bus.Read(addr) + 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xFE: // INC Memoria (Abs, X)
		addr := c.addrAbsX()
		val := c.Bus.Read(addr) + 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xC6: // DEC Memoria (ZP)
		addr := c.addrZP()
		val := c.Bus.Read(addr) - 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xD6: // DEC Memoria (ZP, X)
		addr := c.addrZPX()
		val := c.Bus.Read(addr) - 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xCE: // DEC Memoria (Abs)
		addr := c.addrAbs()
		val := c.Bus.Read(addr) - 1
		c.Bus.Write(addr, val)
		c.setZN(val)
	case 0xDE: // DEC Memoria (Abs, X)
		addr := c.addrAbsX()
		val := c.Bus.Read(addr) - 1
		c.Bus.Write(addr, val)
		c.setZN(val)

	// --- LÓGICA (AND, EOR, ORA) ---
	case 0x29:
		c.and(c.addrImm())
	case 0x25:
		c.and(c.addrZP())
	case 0x35:
		c.and(c.addrZPX())
	case 0x2D:
		c.and(c.addrAbs())
	case 0x3D:
		c.and(c.addrAbsX())
	case 0x39:
		c.and(c.addrAbsY())
	case 0x21:
		c.and(c.addrIndX())
	case 0x31:
		c.and(c.addrIndY())

	case 0x49:
		c.eor(c.addrImm())
	case 0x45:
		c.eor(c.addrZP())
	case 0x55:
		c.eor(c.addrZPX())
	case 0x4D:
		c.eor(c.addrAbs())
	case 0x5D:
		c.eor(c.addrAbsX())
	case 0x59:
		c.eor(c.addrAbsY())
	case 0x41:
		c.eor(c.addrIndX())
	case 0x51:
		c.eor(c.addrIndY())

	case 0x09:
		c.ora(c.addrImm())
	case 0x05:
		c.ora(c.addrZP())
	case 0x15:
		c.ora(c.addrZPX())
	case 0x0D:
		c.ora(c.addrAbs())
	case 0x1D:
		c.ora(c.addrAbsX())
	case 0x19:
		c.ora(c.addrAbsY())
	case 0x01:
		c.ora(c.addrIndX())
	case 0x11:
		c.ora(c.addrIndY())

	// --- COMPARACIONES ---
	case 0xC9:
		c.compare(c.A, c.Bus.Read(c.addrImm()))
	case 0xC5:
		c.compare(c.A, c.Bus.Read(c.addrZP()))
	case 0xD5:
		c.compare(c.A, c.Bus.Read(c.addrZPX()))
	case 0xCD:
		c.compare(c.A, c.Bus.Read(c.addrAbs()))
	case 0xDD:
		c.compare(c.A, c.Bus.Read(c.addrAbsX()))
	case 0xD9:
		c.compare(c.A, c.Bus.Read(c.addrAbsY()))
	case 0xC1:
		c.compare(c.A, c.Bus.Read(c.addrIndX()))
	case 0xD1:
		c.compare(c.A, c.Bus.Read(c.addrIndY()))

	case 0xE0: // CPX (Compare X)
		c.compare(c.X, c.Bus.Read(c.addrImm()))
	case 0xE4:
		c.compare(c.X, c.Bus.Read(c.addrZP()))
	case 0xEC:
		c.compare(c.X, c.Bus.Read(c.addrAbs()))

	case 0xC0: // CPY (Compare Y)
		c.compare(c.Y, c.Bus.Read(c.addrImm()))
	case 0xC4:
		c.compare(c.Y, c.Bus.Read(c.addrZP()))
	case 0xCC:
		c.compare(c.Y, c.Bus.Read(c.addrAbs()))

	// --- BIT TEST ---
	case 0x24: // BIT ZP
		val := c.Bus.Read(c.addrZP())
		if (c.A & val) == 0 {
			c.P |= Z
		} else {
			c.P &^= Z
		}
		c.P = (c.P & 0x3F) | (val & 0xC0) // Copia bits 6 y 7 a flags V y N
	case 0x2C: // BIT Abs
		val := c.Bus.Read(c.addrAbs())
		if (c.A & val) == 0 {
			c.P |= Z
		} else {
			c.P &^= Z
		}
		c.P = (c.P & 0x3F) | (val & 0xC0)

	// --- SHIFTS & ROTATES ---
	case 0x0A: // ASL A
		c.A = c.asl(c.A)
	case 0x06: // ASL ZP
		addr := c.addrZP()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.asl(val))
	case 0x16: // ASL ZP, X
		addr := c.addrZPX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.asl(val))
	case 0x0E: // ASL Abs
		addr := c.addrAbs()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.asl(val))
	case 0x1E: // ASL Abs, X
		addr := c.addrAbsX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.asl(val))

	case 0x4A: // LSR A
		c.A = c.lsr(c.A)
	case 0x46:
		addr := c.addrZP()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.lsr(val))
	case 0x56:
		addr := c.addrZPX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.lsr(val))
	case 0x4E:
		addr := c.addrAbs()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.lsr(val))
	case 0x5E:
		addr := c.addrAbsX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.lsr(val))

	case 0x2A: // ROL A
		c.A = c.rol(c.A)
	case 0x26:
		addr := c.addrZP()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.rol(val))
	case 0x36:
		addr := c.addrZPX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.rol(val))
	case 0x2E:
		addr := c.addrAbs()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.rol(val))
	case 0x3E:
		addr := c.addrAbsX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.rol(val))

	case 0x6A: // ROR A
		c.A = c.ror(c.A)
	case 0x66:
		addr := c.addrZP()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.ror(val))
	case 0x76:
		addr := c.addrZPX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.ror(val))
	case 0x6E:
		addr := c.addrAbs()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.ror(val))
	case 0x7E:
		addr := c.addrAbsX()
		val := c.Bus.Read(addr)
		c.Bus.Write(addr, c.ror(val))

	// --- STACK Y SISTEMA ---
	case 0x48: // PHA (Push A)
		c.push(c.A)
	case 0x08: // PHP (Push P)
		c.push(c.P | B) // B flag se activa al pushear por PHP/BRK
	case 0x68: // PLA (Pull A)
		c.A = c.pull()
		c.setZN(c.A)
	case 0x28: // PLP (Pull P)
		c.P = c.pull() | U // Flag U siempre 1

	case 0x20: // JSR (Jump to Subroutine)
		addr := c.addrAbs()
		c.push16(c.PC - 1) // Guardamos la dirección de retorno
		c.PC = addr
	case 0x60: // RTS (Return from Subroutine)
		c.PC = c.pull16() + 1

	case 0x00: // BRK (Break - Interrupción por Software)
		c.PC++
		c.push16(c.PC)
		c.push(c.P | B)
		c.P |= I
		lo, hi := uint16(c.Bus.Read(0xFFFE)), uint16(c.Bus.Read(0xFFFF))
		c.PC = (hi << 8) | lo

	case 0x40: // RTI (Return from Interrupt)
		c.P = c.pull() | U
		c.PC = c.pull16()

	case 0x4C: // JMP Absolute
		c.PC = c.addrAbs()
	case 0x6C: // JMP Indirect
		addr := c.addrAbs()
		lo := uint16(c.Bus.Read(addr))
		// Bug hardware del 6502: Si la dirección cae en un borde de página (0xXXFF),
		// no cruza la página correctamente para leer el byte alto, sino que hace
		// wrap-around al inicio de la MISMA página (0xXX00).
		hiAddr := addr + 1
		if addr&0x00FF == 0x00FF {
			hiAddr = addr & 0xFF00
		}
		hi := uint16(c.Bus.Read(hiAddr))
		c.PC = (hi << 8) | lo

	// --- BRANCHES (SALTOS CONDICIONALES) ---
	// Calculan ciclos extra si la condición se cumple y si cambia de página.
	case 0x90: // BCC (Branch if Carry Clear)
		extraCycles = c.branch((c.P & C) == 0)
	case 0xB0: // BCS (Branch if Carry Set)
		extraCycles = c.branch((c.P & C) != 0)
	case 0xF0: // BEQ (Branch if Equal / Zero Set)
		extraCycles = c.branch((c.P & Z) != 0)
	case 0xD0: // BNE (Branch if Not Equal / Zero Clear)
		extraCycles = c.branch((c.P & Z) == 0)
	case 0x30: // BMI (Branch if Minus / Negative Set)
		extraCycles = c.branch((c.P & N) != 0)
	case 0x10: // BPL (Branch if Plus / Negative Clear)
		extraCycles = c.branch((c.P & N) == 0)
	case 0x50: // BVC (Branch if Overflow Clear)
		extraCycles = c.branch((c.P & V) == 0)
	case 0x70: // BVS (Branch if Overflow Set)
		extraCycles = c.branch((c.P & V) != 0)

	// --- CLC/SEC ... (FLAGS) ---
	case 0x18: // CLC
		c.P &^= C
	case 0x38: // SEC
		c.P |= C
	case 0x58: // CLI
		c.P &^= I
	case 0x78: // SEI
		c.P |= I
	case 0xB8: // CLV
		c.P &^= V
	case 0xD8: // CLD
		c.P &^= D
	case 0xF8: // SED
		c.P |= D

	case 0xEA: // NOP (No Operation)
		// No hacer nada, solo consumir ciclos.

	default:
		// Instrucciones ilegales: Las tratamos como NOP para seguridad,
		// aunque en hardware real hacen cosas raras.
	}

	// Tick del Mapper para los ciclos extra (branches, page crossing, etc)
	for i := 0; i < extraCycles; i++ {
		c.Bus.Cart.Mapper.Tick()
	}

	// Tick del Mapper y APU (post-instrucción para capturar IRQs generados por la instrucción)
	// Usamos cycles + extraCycles para simular el tiempo total de la instrucción.
	totalCycles := cycles + extraCycles
	for i := 0; i < totalCycles; i++ {
		c.Bus.Cart.Mapper.Tick()
		if c.Bus.APU != nil {
			c.Bus.APU.Tick()
		}
	}

	return totalCycles
}

// branch ejecuta la lógica de salto relativo condicional.
// Retorna ciclos extra consumidos si se tomó el salto.
func (c *CPU) branch(cond bool) int {
	// Leer offset relativo (puede ser negativo, -128 a +127)
	rel := int8(c.Bus.Read(c.PC))
	c.PC++
	if cond {
		oldPC := c.PC
		c.PC = uint16(int16(c.PC) + int16(rel))
		extra := 1 // Salto costó 1 ciclo extra
		// Si cambiamos de página (bytes altos diferentes), cuesta otro ciclo más.
		if (oldPC & 0xFF00) != (c.PC & 0xFF00) {
			extra++
		}
		return extra
	}
	return 0
}

// cycleTable contiene los ciclos de reloj BASE de cada opcode del 6502.
var cycleTable = [256]int{
	7, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6, // 0x00
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7, // 0x10
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6, // 0x20
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7, // 0x30
	6, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6, // 0x40
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7, // 0x50
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6, // 0x60
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7, // 0x70
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4, // 0x80
	2, 6, 2, 6, 4, 4, 4, 4, 2, 5, 2, 5, 4, 5, 5, 5, // 0x90
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4, // 0xA0
	2, 6, 2, 6, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4, // 0xB0
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6, // 0xC0
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7, // 0xD0
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6, // 0xE0
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7, // 0xF0
}
