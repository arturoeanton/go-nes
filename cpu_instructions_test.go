package main

import "testing"

// ==========================================
// TEST UNITARIOS DE INSTRUCCIONES CPU
// ==========================================
// Pruebas específicas para aritmética, saltos y lógica.

func TestCPU_ADC_Addition(t *testing.T) {
	cpu := setupTestCPU()

	// 1. Suma simple: 10 + 20 = 30
	// CLC (Clear Carry) - $18
	// LDA #10 - $A9 $0A
	// ADC #20 - $69 $14
	runCode(cpu, []byte{
		0x18,       // CLC
		0xA9, 0x0A, // LDA #10
		0x69, 0x14, // ADC #20
	})

	if cpu.A != 30 {
		t.Errorf("ADC Simple falló: 10+20=%d (Esperado 30)", cpu.A)
	}
	if cpu.P&C != 0 {
		t.Error("Carry incorrectamente seteado")
	}

	// 2. Suma con Overflow (Signo): 127 + 1 = 128 (-128 en signed)
	// 127 = 0x7F (+), 1 = 0x01 (+) => 0x80 (-) => Overflow!
	runCode(cpu, []byte{
		0x18,       // CLC
		0xA9, 0x7F, // LDA #127
		0x69, 0x01, // ADC #1
	})

	if cpu.A != 0x80 {
		t.Errorf("ADC Overflow falló valor: %02X", cpu.A)
	}
	if cpu.P&V == 0 {
		t.Error("Overflow Flag (V) no se activó")
	}
	if cpu.P&N == 0 {
		t.Error("Negative Flag (N) no se activó")
	}

	// 3. Suma con Carry generado: 255 + 1 = 0 (C=1)
	runCode(cpu, []byte{
		0x18,       // CLC
		0xA9, 0xFF, // LDA #255
		0x69, 0x01, // ADC #1
	})

	if cpu.A != 0x00 {
		t.Errorf("ADC Carry Gen falló: %02X", cpu.A)
	}
	if cpu.P&Z == 0 {
		t.Error("Zero Flag no se activó")
	}
	if cpu.P&C == 0 {
		t.Error("Carry Flag no se activó")
	}

	// 4. Suma USANDO Carry previo: 10 + 10 + C(1 from prev) = 21
	runCode(cpu, []byte{
		// Carry ya está en 1 del test anterior
		0xA9, 0x0A, // LDA #10
		0x69, 0x0A, // ADC #10
	})
	if cpu.A != 21 {
		t.Errorf("ADC con Carry previo falló: %d", cpu.A)
	}
}

func TestCPU_SBC_Subtraction(t *testing.T) {
	cpu := setupTestCPU()
	// SBC funciona como ADC invertido. Requiere SEC (Set Carry) para restar sin borrow.
	// A - M - (1-C)

	// 1. Resta simple: 20 - 10 = 10
	// SEC (Set Carry = No Borrow) - $38
	runCode(cpu, []byte{
		0x38,       // SEC
		0xA9, 0x14, // LDA #20
		0xE9, 0x0A, // SBC #10
	})
	if cpu.A != 10 {
		t.Errorf("SBC Simple falló: 20-10=%d", cpu.A)
	}
	if cpu.P&C == 0 {
		t.Error("Carry debería mantenerse (No Borrow)")
	}

	// 2. Resta con Borrow: 10 - 20 = -10 (246 o 0xF6)
	runCode(cpu, []byte{
		0x38,       // SEC
		0xA9, 0x0A, // LDA #10
		0xE9, 0x14, // SBC #20
	})
	if cpu.A != 0xF6 {
		t.Errorf("SBC Borrow falló: 10-20=%02X", cpu.A)
	}
	if cpu.P&C != 0 {
		t.Error("Carry debería limpiarse (Borrow ocurrió)")
	}
	if cpu.P&N == 0 {
		t.Error("Negative Flag debería activarse")
	}
}

func TestCPU_Branching(t *testing.T) {
	cpu := setupTestCPU()

	// BNE (Branch Not Equal / Not Zero)
	// Loop que decrementa X desde 5 hasta 0.
	// LDX #5
	// loop: DEX
	// BNE loop ($FD = -3)

	// Opcode DEX = $CA, BNE = $D0
	code := []byte{
		0xA2, 0x05, // LDX #5
		0xCA,       // DEX
		0xD0, 0xFD, // BNE -3 (vuelve a DEX)
	}
	runCodeCycles(cpu, code, 100) // Limite ciclos para no colgar

	if cpu.X != 0 {
		t.Errorf("BNE Loop falló: X=%d (esperado 0)", cpu.X)
	}
	if cpu.P&Z == 0 {
		t.Error("Zero flag debería quedar activo al final")
	}
}

// Helpers locales para facilitar tests de instrucciones
func runCode(cpu *CPU, code []byte) {
	writeCode(cpu, 0x8000, code)
	cpu.PC = 0x8000
	for i := 0; i < len(code); {
		// Asumimos instrucciones simples de 1, 2 o 3 bytes.
		// Difícil saber cuántos steps sin decodificar.
		// Solución chapuza: Ejecutamos N pasos fijos? No.
		// Mejor: un Step por instrucción.
		// Pero no sabemos cuántas instrucciones son code.
		// Ejecutaremos hasta que PC pase el final del código.
		cpu.Step()
		if cpu.PC >= 0x8000+uint16(len(code)) {
			break
		}
		// Safety break
		if cpu.PC < 0x8000 {
			break
		}
	}
}

func runCodeCycles(cpu *CPU, code []byte, maxCycles int) {
	writeCode(cpu, 0x8000, code)
	cpu.PC = 0x8000
	cycles := 0
	for cycles < maxCycles {
		c := cpu.Step()
		cycles += c
		// Si PC sale del rango, paramos (éxito)
		if cpu.PC >= 0x8000+uint16(len(code)) {
			break
		}
	}
}

func writeCode(cpu *CPU, addr uint16, code []byte) {
	for i, b := range code {
		writeDirect(cpu, addr+uint16(i), b)
	}
}
