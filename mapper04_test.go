package main

import (
	"testing"
)

// =============================================================================
// MAPPER 4 (MMC3) TEST SUITE
// =============================================================================
// Tests para la implementación actual del MMC3 que funciona con:
// - Super Mario Bros 3
// - Tiny Toon Adventures
// - TMNT II: Arcade Game
// - TMNT: Tournament Fighters
// - Joe & Mac
//
// Nota: Aladdin y Shinobi requieren A12 edge detection (ver mapper04_smartA12_test.go)

// TestMapper4Creation verifica que el mapper se crea correctamente
func TestMapper4Creation(t *testing.T) {
	m := NewMapper4(16, 32) // 256KB PRG, 256KB CHR

	if m == nil {
		t.Fatal("NewMapper4 returned nil")
	}

	if m.prgBanks != 16 {
		t.Errorf("Expected prgBanks=16, got %d", m.prgBanks)
	}

	if m.chrBanks != 32 {
		t.Errorf("Expected chrBanks=32, got %d", m.chrBanks)
	}

	// Verificar estado inicial
	if m.irqEnabled {
		t.Error("IRQ should be disabled initially")
	}

	if m.irqPending {
		t.Error("IRQ should not be pending initially")
	}
}

// TestMapper4PRGBankingMode0 verifica el modo de PRG banking normal
func TestMapper4PRGBankingMode0(t *testing.T) {
	m := NewMapper4(16, 32) // 16 bancos de 16KB = 32 bancos de 8KB

	// Modo 0: $8000=R6, $A000=R7, $C000=-2, $E000=-1
	m.Write(0x8000, 0x06) // Seleccionar registro 6, modo PRG=0
	m.Write(0x8001, 0x05) // R6 = banco 5
	m.Write(0x8000, 0x07) // Seleccionar registro 7
	m.Write(0x8001, 0x10) // R7 = banco 16

	// $8000-$9FFF debería ser banco 5 (offset 5*8192 = 40960)
	offset := m.Read(0x8000)
	if offset != 5*8192 {
		t.Errorf("$8000 expected offset %d, got %d", 5*8192, offset)
	}

	// $A000-$BFFF debería ser banco 16
	offset = m.Read(0xA000)
	if offset != 16*8192 {
		t.Errorf("$A000 expected offset %d, got %d", 16*8192, offset)
	}

	// $C000-$DFFF debería ser penúltimo banco (30)
	offset = m.Read(0xC000)
	expected := 30 * 8192
	if offset != expected {
		t.Errorf("$C000 expected offset %d, got %d", expected, offset)
	}

	// $E000-$FFFF debería ser último banco (31)
	offset = m.Read(0xE000)
	expected = 31 * 8192
	if offset != expected {
		t.Errorf("$E000 expected offset %d, got %d", expected, offset)
	}
}

// TestMapper4PRGBankingMode1 verifica el modo de PRG banking swapped
func TestMapper4PRGBankingMode1(t *testing.T) {
	m := NewMapper4(16, 32)

	// Modo 1: $8000=-2, $A000=R7, $C000=R6, $E000=-1
	m.Write(0x8000, 0x46) // Seleccionar registro 6, modo PRG=1 (bit 6 set)
	m.Write(0x8001, 0x05) // R6 = banco 5
	m.Write(0x8000, 0x47) // Seleccionar registro 7, mantener modo
	m.Write(0x8001, 0x10) // R7 = banco 16

	// $8000 debería ser penúltimo banco (30) en modo 1
	offset := m.Read(0x8000)
	expected := 30 * 8192
	if offset != expected {
		t.Errorf("Mode1 $8000 expected offset %d, got %d", expected, offset)
	}

	// $C000 debería ser R6 (banco 5) en modo 1
	offset = m.Read(0xC000)
	if offset != 5*8192 {
		t.Errorf("Mode1 $C000 expected offset %d, got %d", 5*8192, offset)
	}
}

// TestMapper4CHRBankingNoInversion verifica CHR banking sin inversión
func TestMapper4CHRBankingNoInversion(t *testing.T) {
	m := NewMapper4(16, 32)

	// Sin inversión (bit 7 = 0)
	// R0, R1 = 2KB banks en $0000-$0FFF
	// R2-R5 = 1KB banks en $1000-$1FFF
	m.Write(0x8000, 0x00) // Seleccionar R0
	m.Write(0x8001, 0x00) // R0 = 0 (banco 2KB)
	m.Write(0x8000, 0x01) // Seleccionar R1
	m.Write(0x8001, 0x04) // R1 = 4 (banco 2KB alineado, bit 0 ignorado)
	m.Write(0x8000, 0x02) // Seleccionar R2
	m.Write(0x8001, 0x10) // R2 = 16 (banco 1KB)
	m.Write(0x8000, 0x03) // Seleccionar R3
	m.Write(0x8001, 0x11) // R3 = 17
	m.Write(0x8000, 0x04) // Seleccionar R4
	m.Write(0x8001, 0x12) // R4 = 18
	m.Write(0x8000, 0x05) // Seleccionar R5
	m.Write(0x8001, 0x13) // R5 = 19

	// Verificar $0000 = R0 * 1024
	offset := m.ReadCHR(0x0000)
	if offset != 0*1024 {
		t.Errorf("$0000 expected offset %d, got %d", 0*1024, offset)
	}

	// Verificar $0400 = (R0+1) * 1024
	offset = m.ReadCHR(0x0400)
	if offset != 1*1024 {
		t.Errorf("$0400 expected offset %d, got %d", 1*1024, offset)
	}

	// Verificar $1000 = R2 * 1024
	offset = m.ReadCHR(0x1000)
	if offset != 16*1024 {
		t.Errorf("$1000 expected offset %d, got %d", 16*1024, offset)
	}
}

// TestMapper4CHRBankingWithInversion verifica CHR banking con inversión
func TestMapper4CHRBankingWithInversion(t *testing.T) {
	m := NewMapper4(16, 32)

	// Con inversión (bit 7 = 1)
	// R2-R5 = 1KB banks en $0000-$0FFF
	// R0, R1 = 2KB banks en $1000-$1FFF
	m.Write(0x8000, 0x80) // Seleccionar R0, CHR inversion ON
	m.Write(0x8001, 0x00) // R0 = 0
	m.Write(0x8000, 0x82) // Seleccionar R2
	m.Write(0x8001, 0x10) // R2 = 16

	// Con inversión, $0000 debería ser R2
	offset := m.ReadCHR(0x0000)
	if offset != 16*1024 {
		t.Errorf("Inverted $0000 expected offset %d, got %d", 16*1024, offset)
	}

	// Con inversión, $1000 debería ser R0
	offset = m.ReadCHR(0x1000)
	if offset != 0*1024 {
		t.Errorf("Inverted $1000 expected offset %d, got %d", 0*1024, offset)
	}
}

// TestMapper4Mirroring verifica el control de mirroring
func TestMapper4Mirroring(t *testing.T) {
	m := NewMapper4(16, 32)

	// Por defecto debería ser vertical
	mode, controlled := m.GetMirror()
	if !controlled {
		t.Error("MMC3 should control mirroring")
	}
	if mode != MirrorVertical {
		t.Error("Default mirroring should be vertical")
	}

	// Cambiar a horizontal
	m.Write(0xA000, 0x01)
	mode, _ = m.GetMirror()
	if mode != MirrorHorizontal {
		t.Error("Mirroring should be horizontal after $A000 write with bit 0 set")
	}

	// Cambiar a vertical
	m.Write(0xA000, 0x00)
	mode, _ = m.GetMirror()
	if mode != MirrorVertical {
		t.Error("Mirroring should be vertical after $A000 write with bit 0 clear")
	}
}

// TestMapper4IRQBasic verifica el comportamiento básico del IRQ
func TestMapper4IRQBasic(t *testing.T) {
	m := NewMapper4(16, 32)

	// Configurar IRQ latch
	m.Write(0xC000, 10) // Latch = 10
	m.Write(0xC001, 0)  // Reload flag

	// Habilitar IRQ
	m.Write(0xE001, 0)

	// Verificar que no hay IRQ pendiente aún
	if m.IRQState() {
		t.Error("IRQ should not be pending before any scanlines")
	}

	// Simular scanlines
	for i := 0; i < 10; i++ {
		m.Scanline()
	}

	// Después de 10 scanlines, el contador debería llegar a 0 y disparar
	// Nota: El primer Scanline recarga, entonces necesitamos 11 total
	m.Scanline() // Este debería disparar

	if !m.IRQState() {
		t.Error("IRQ should be pending after counter reaches 0")
	}
}

// TestMapper4IRQDisable verifica que se puede deshabilitar el IRQ
func TestMapper4IRQDisable(t *testing.T) {
	m := NewMapper4(16, 32)

	// Configurar y habilitar IRQ
	m.Write(0xC000, 5) // Latch = 5
	m.Write(0xC001, 0) // Reload
	m.Write(0xE001, 0) // Enable

	// Simular suficientes scanlines para disparar
	for i := 0; i < 7; i++ {
		m.Scanline()
	}

	if !m.IRQState() {
		t.Error("IRQ should be pending")
	}

	// Deshabilitar y acknowledge
	m.Write(0xE000, 0)

	if m.IRQState() {
		t.Error("IRQ should be cleared after $E000 write")
	}
}

// TestMapper4ReadCHRBounds verifica que ReadCHR maneja correctamente los límites
func TestMapper4ReadCHRBounds(t *testing.T) {
	m := NewMapper4(16, 32)

	// Todas las direcciones de 0x0000 a 0x1FFF deberían funcionar
	for addr := uint16(0); addr < 0x2000; addr += 0x100 {
		offset := m.ReadCHR(addr)
		if offset < 0 {
			t.Errorf("ReadCHR($%04X) returned negative offset %d", addr, offset)
		}
	}
}

// TestMapper4PRGReadAddresses verifica todas las ventanas de PRG
func TestMapper4PRGReadAddresses(t *testing.T) {
	m := NewMapper4(8, 16) // 128KB PRG

	testCases := []struct {
		addr     uint16
		expected bool // true if should return valid offset
	}{
		{0x7FFF, false}, // Below PRG range
		{0x8000, true},  // Start of PRG
		{0x9FFF, true},  // End of first window
		{0xA000, true},  // Start of second window
		{0xBFFF, true},  // End of second window
		{0xC000, true},  // Start of third window
		{0xDFFF, true},  // End of third window
		{0xE000, true},  // Start of fourth window
		{0xFFFF, true},  // End of PRG
	}

	for _, tc := range testCases {
		offset := m.Read(tc.addr)
		if tc.expected && offset < 0 {
			t.Errorf("Read($%04X) expected valid offset, got %d", tc.addr, offset)
		}
		if !tc.expected && offset >= 0 {
			t.Errorf("Read($%04X) expected -1, got %d", tc.addr, offset)
		}
	}
}
