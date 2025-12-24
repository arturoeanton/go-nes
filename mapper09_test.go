package main

import "testing"

// ==========================================
// TEST UNITARIOS PARA MAPPER 09 (MMC2)
// ==========================================
// Crítico para "Punch-Out!!"
// Verifica PRG banking, CHR latch automático, y mirroring.

func TestMapper9Creation(t *testing.T) {
	m := NewMapper9(8, 16) // 128KB PRG, 128KB CHR

	if m == nil {
		t.Fatal("NewMapper9 returned nil")
	}

	if m.prgBanks != 8 {
		t.Errorf("Expected prgBanks=8, got %d", m.prgBanks)
	}

	// Latch inicial debe ser $FE
	if m.latch0 != 0xFE {
		t.Errorf("Expected latch0=0xFE, got %02X", m.latch0)
	}
	if m.latch1 != 0xFE {
		t.Errorf("Expected latch1=0xFE, got %02X", m.latch1)
	}
}

func TestMapper9PRGBanking(t *testing.T) {
	m := NewMapper9(8, 16) // 128KB PRG = 16 bancos de 8KB

	// Banco por defecto = 0
	offset := m.Read(0x8000)
	if offset != 0 {
		t.Errorf("$8000 default offset expected 0, got %d", offset)
	}

	// Cambiar banco a 5
	m.Write(0xA000, 0x05)
	offset = m.Read(0x8000)
	expected := 5 * 8192
	if offset != expected {
		t.Errorf("$8000 expected offset %d, got %d", expected, offset)
	}

	// Verificar bancos fijos
	// $A000-$BFFF = banco 13 (16-3)
	offset = m.Read(0xA000)
	expected = 13 * 8192
	if offset != expected {
		t.Errorf("$A000 fixed expected offset %d, got %d", expected, offset)
	}

	// $C000-$DFFF = banco 14 (16-2)
	offset = m.Read(0xC000)
	expected = 14 * 8192
	if offset != expected {
		t.Errorf("$C000 fixed expected offset %d, got %d", expected, offset)
	}

	// $E000-$FFFF = banco 15 (16-1)
	offset = m.Read(0xE000)
	expected = 15 * 8192
	if offset != expected {
		t.Errorf("$E000 fixed expected offset %d, got %d", expected, offset)
	}
}

func TestMapper9CHRLatchAutomatic(t *testing.T) {
	m := NewMapper9(8, 16)

	// Configurar bancos CHR
	m.Write(0xB000, 0x05) // chrBank0FD = 5
	m.Write(0xC000, 0x0A) // chrBank0FE = 10
	m.Write(0xD000, 0x03) // chrBank1FD = 3
	m.Write(0xE000, 0x07) // chrBank1FE = 7

	// Latch inicial es $FE, así que debería usar chrBank0FE (10)
	offset := m.ReadCHR(0x0000)
	expected := 10 * 4096
	if offset != expected {
		t.Errorf("Initial CHR0 expected offset %d, got %d", expected, offset)
	}

	// Leer tile en rango $0FD8-$0FDF (tile $FD) cambia latch a FD
	_ = m.ReadCHR(0x0FD8)
	if m.latch0 != 0xFD {
		t.Errorf("Latch0 should be FD after reading $0FD8, got %02X", m.latch0)
	}

	// Ahora debería usar chrBank0FD (5)
	offset = m.ReadCHR(0x0000)
	expected = 5 * 4096
	if offset != expected {
		t.Errorf("After FD latch, CHR0 expected offset %d, got %d", expected, offset)
	}

	// Leer tile $FE (rango $0FE8-$0FEF) vuelve latch a FE
	_ = m.ReadCHR(0x0FE8)
	if m.latch0 != 0xFE {
		t.Errorf("Latch0 should be FE after reading $0FE8, got %02X", m.latch0)
	}
}

func TestMapper9CHRLatchPatternTable1(t *testing.T) {
	m := NewMapper9(8, 16)

	m.Write(0xD000, 0x02) // chrBank1FD = 2
	m.Write(0xE000, 0x08) // chrBank1FE = 8

	// Latch1 inicial es FE, usar banco 8
	offset := m.ReadCHR(0x1000)
	expected := 8 * 4096
	if offset != expected {
		t.Errorf("Initial CHR1 expected offset %d, got %d", expected, offset)
	}

	// Leer tile $FD en pattern table 1 ($1FD8-$1FDF)
	_ = m.ReadCHR(0x1FD8)
	if m.latch1 != 0xFD {
		t.Errorf("Latch1 should be FD after reading $1FD8, got %02X", m.latch1)
	}

	// Ahora usa banco 2
	offset = m.ReadCHR(0x1000)
	expected = 2 * 4096
	if offset != expected {
		t.Errorf("After FD latch, CHR1 expected offset %d, got %d", expected, offset)
	}
}

func TestMapper9Mirroring(t *testing.T) {
	m := NewMapper9(8, 16)

	// Default es horizontal
	mode, controlled := m.GetMirror()
	if !controlled {
		t.Error("MMC2 should control mirroring")
	}
	if mode != MirrorHorizontal {
		t.Error("Default mirroring should be horizontal")
	}

	// Cambiar a vertical (bit 0 = 0)
	m.Write(0xF000, 0x00)
	mode, _ = m.GetMirror()
	if mode != MirrorVertical {
		t.Error("Mirroring should be vertical after $F000 write with bit 0 = 0")
	}

	// Cambiar a horizontal (bit 0 = 1)
	m.Write(0xF000, 0x01)
	mode, _ = m.GetMirror()
	if mode != MirrorHorizontal {
		t.Error("Mirroring should be horizontal after $F000 write with bit 0 = 1")
	}
}

func TestMapper9NoIRQ(t *testing.T) {
	m := NewMapper9(8, 16)

	// MMC2 no tiene IRQ
	if m.IRQState() {
		t.Error("MMC2 should not generate IRQ")
	}
}

func TestMapper9CHRBoundsCheck(t *testing.T) {
	m := NewMapper9(8, 8) // 64KB CHR

	// Configurar banco alto que excede CHR size
	m.Write(0xC000, 0x1F) // Bank 31, pero solo hay 16 bancos de 4KB
	m.latch0 = 0xFE

	// Debería hacer wrap-around
	offset := m.ReadCHR(0x0000)
	chrSize := 8 * 8192 // 64KB
	if offset >= chrSize {
		t.Errorf("CHR offset %d exceeds size %d", offset, chrSize)
	}
}
