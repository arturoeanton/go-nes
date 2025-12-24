package main

import "testing"

// ==========================================
// TEST UNITARIOS DE CPU
// ==========================================
// Este archivo contiene pruebas para verificar el correcto funcionamiento
// de la emulación de la CPU Ricoh 2A03 (MOS 6502).

// Helper para escribir directamente en la ROM (bypass Mapper Write protection)
// Esto es necesario porque Mapper 0 (NROM) no permite escrituras en el rango $8000-$FFFF,
// pero para los tests necesitamos inyectar código.
func writeDirect(cpu *CPU, addr uint16, data byte) {
	if addr >= 0x8000 {
		// Asumiendo Mapper 0 con 32KB
		// $8000 maps to index 0
		idx := int(addr - 0x8000)
		if idx < len(cpu.Bus.Cart.PRG) {
			cpu.Bus.Cart.PRG[idx] = data
		}
	} else {
		// RAM y otros registros escribibles
		cpu.Bus.Write(addr, data)
	}
}

// Helper para crear una CPU lista para pruebas con un Bus básico (Mapper 0)
// Creamos una ROM vacía de 32KB (2 bancos)
func setupTestCPU() *CPU {
	// Crear bancos de memoria vacíos
	prg := make([]byte, 32*1024)
	chr := make([]byte, 8*1024)

	// Crear Cartucho y Mapper 0
	cart := &Cartridge{
		PRG:    prg,
		CHR:    chr,
		Mapper: NewMapper0(2), // 2 Bancos PRG
		Mirror: MirrorHorizontal,
	}

	// Crear PPU y APU con constructores correctos
	ppu := NewPPU(cart)
	apu := NewAPU()

	// Crear Bus
	// Se inicializa directo porque Bus es una estructura simple
	bus := &Bus{
		Cart: cart,
		PPU:  ppu,
		APU:  apu,
		// RAM inicializada a 0 por Go
	}

	// Crear CPU y conectarla
	cpu := NewCPU(bus)

	return cpu
}

// TestCPU_Reset verifica que la señal de Reset inicialice los registros correctamente.
func TestCPU_Reset(t *testing.T) {
	cpu := setupTestCPU()

	// Escribir el Vector de Reset ($FFFC/$FFFD) en la "ROM"
	// Queremos que el PC arranque en $8000
	writeDirect(cpu, 0xFFFC, 0x00)
	writeDirect(cpu, 0xFFFD, 0x80)

	// Ejecutar Reset
	cpu.Reset()

	// 1. Verificar Program Counter (PC)
	if cpu.PC != 0x8000 {
		t.Errorf("Reset falló: PC esperado $8000, obtenido $%04X", cpu.PC)
	}

	// 2. Verificar Stack Pointer (SP)
	// El 6502 inicializa el SP en $FD tras un ciclo de reset.
	if cpu.SP != 0xFD {
		t.Errorf("Reset falló: SP esperado $FD, obtenido $%02X", cpu.SP)
	}

	// 3. Verificar Flags (P)
	// Banderas I (Interrupt Disable) y U (Unused) deben estar activas.
	// 0x24 = 0010 0100 (U=1, I=1)
	if cpu.P != 0x24 {
		t.Errorf("Reset falló: Flags P esperados $24 (I|U), obtenido $%02X", cpu.P)
	}
}

// TestCPU_LDA_Immediate verifica la instrucción LDA en modo inmediato.
// Opcode: $A9, Flags afectados: Z, N.
func TestCPU_LDA_Immediate(t *testing.T) {
	cpu := setupTestCPU()

	// Programa en ROM ($8000):
	// A9 42   ; LDA #$42  (Carga el valor 0x42 en Acumulador)
	writeDirect(cpu, 0x8000, 0xA9)
	writeDirect(cpu, 0x8001, 0x42)

	// Ajustar PC manual
	cpu.PC = 0x8000

	// Ejecutar un paso (Instrucción completa)
	cycles := cpu.Step()

	// 1. Verificar valor en A
	if cpu.A != 0x42 {
		t.Errorf("LDA Imm falló: A esperado $42, obtenido $%02X", cpu.A)
	}

	// 2. Verificar ciclos consumidos (LDA Imm = 2 ciclos)
	if cycles != 2 {
		t.Errorf("LDA Imm Timing falló: Ciclos esperados 2, obtenidos %d", cycles)
	}

	// 3. Verificar Flags
	// 0x42 no es cero (Z=0) y es positivo (N=0).
	if cpu.P&Z != 0 {
		t.Error("LDA Imm falló: Flag Z no debería estar activo para $42")
	}
	if cpu.P&N != 0 {
		t.Error("LDA Imm falló: Flag N no debería estar activo para $42")
	}
}

// TestCPU_Flags_ZeroNegative verifica que los flags Z y N se actualicen correctamente.
func TestCPU_Flags_ZeroNegative(t *testing.T) {
	cpu := setupTestCPU()

	// Caso 1: Cero
	// A9 00   ; LDA #$00
	writeDirect(cpu, 0x8000, 0xA9)
	writeDirect(cpu, 0x8001, 0x00)
	cpu.PC = 0x8000
	cpu.Step()

	if cpu.A != 0x00 {
		t.Errorf("Caso Z: A vale $%02X", cpu.A)
	}
	if cpu.P&Z == 0 {
		t.Error("Caso Z: Flag Z debería estar activo")
	}
	if cpu.P&N != 0 {
		t.Error("Caso Z: Flag N debería estar inactivo")
	}

	// Caso 2: Negativo
	// A9 80   ; LDA #$80 (0x80 tiene bit 7 activo)
	writeDirect(cpu, 0x8002, 0xA9)
	writeDirect(cpu, 0x8003, 0x80)

	// Verificar PC
	if cpu.PC != 0x8002 {
		t.Errorf("PC desalineado: esperado $8002, obtenido $%04X", cpu.PC)
	}
	cpu.Step()

	if cpu.A != 0x80 {
		t.Errorf("Caso N: A vale $%02X", cpu.A)
	}
	if cpu.P&Z != 0 {
		t.Error("Caso N: Flag Z debería estar inactivo")
	}
	if cpu.P&N == 0 {
		t.Error("Caso N: Flag N debería estar activo")
	}
}

// TestCPU_StackOperations verifica PHP, PHA, PLP, PLA.
func TestCPU_StackOperations(t *testing.T) {
	cpu := setupTestCPU()

	// Reset para poner SP en $FD
	writeDirect(cpu, 0xFFFC, 0x00)
	writeDirect(cpu, 0xFFFD, 0x80)
	cpu.Reset()

	// Programa:
	// A9 55   ; LDA #$55
	// 48      ; PHA (Push A)
	// A9 AA   ; LDA #$AA (Sobrescribir A)
	// 68      ; PLA (Pull A) -> A debería volver a ser $55

	base := uint16(0x8000)
	code := []byte{
		0xA9, 0x55, // LDA #$55
		0x48,       // PHA
		0xA9, 0xAA, // LDA #$AA
		0x68, // PLA
	}
	for i, b := range code {
		writeDirect(cpu, base+uint16(i), b)
	}

	cpu.Step() // LDA #$55
	cpu.Step() // PHA

	// Verificar que el Stack Pointer bajó
	if cpu.SP != 0xFC {
		t.Errorf("PHA no decrementó SP correctamente. SP=$%02X", cpu.SP)
	}
	// Verificar que el valor está en la RAM ($0100 + SP + 1) -> $01FD
	val := cpu.Bus.Read(0x01FD)
	if val != 0x55 {
		t.Errorf("PHA no guardó el valor correcto en Stack. Mem[$01FD]=$%02X", val)
	}

	cpu.Step() // LDA #$AA
	if cpu.A != 0xAA {
		t.Error("Fallo intermedio en LDA #$AA")
	}

	cpu.Step() // PLA
	// Verificar recuperación
	if cpu.A != 0x55 {
		t.Errorf("PLA falló: recuperó $%02X, esperaba $55", cpu.A)
	}
	// Verificar SP restaurado
	if cpu.SP != 0xFD {
		t.Errorf("PLA no incrementó SP correctamente. SP=$%02X", cpu.SP)
	}
}
