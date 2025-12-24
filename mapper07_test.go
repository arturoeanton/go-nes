package main

import "testing"

// ==========================================
// TEST UNITARIOS PARA MAPPER 7 (AxROM)
// ==========================================
// Usado en: Battletoads.
// Características: Single Screen Mirroring switchable, 32KB PRG Switching.

func TestMapper7_PRG_Switching_32KB(t *testing.T) {
	// 8 Bancos de 32KB (256KB Total).
	m := NewMapper7(16)

	// 1. Inicial (Banco 0)
	// $8000 map to 0
	if m.Read(0x8000) != 0 {
		t.Error("Banco inicial incorrecto")
	}

	// 2. Cambiar a Banco 5
	// Escribir XXX1 0101 (Bit 4 mirror, bits 0-2 bank) -> $15
	m.Write(0x8000, 0x15) // Mirror High, Bank 5

	// Verificar mapeo en $8000 (Inicio de banco)
	off := m.Read(0x8000)
	if off != 5*32768 {
		t.Errorf("PRG Switch a 5 falló en $8000: %d", off)
	}

	// Verificar mapeo en $C000 (Mitad del banco)
	// $C000 - $8000 = $4000 (16384)
	off = m.Read(0xC000)
	expected := 5*32768 + 16384
	if off != expected {
		t.Errorf("PRG Switch a 5 falló en $C000: %d", off)
	}
}

func TestMapper7_Mirroring_Control(t *testing.T) {
	m := NewMapper7(16)

	// 1. Inicial (Single Screen Low - 0)
	mode, ok := m.GetMirror()
	if !ok || mode != MirrorSingle0 {
		t.Errorf("Mirror inicial incorrecto: %d", mode)
	}

	// 2. Cambiar a Single Screen High (1) -> Bit 4 = 1
	// Escribir 0x10 (Bank 0, Mirror 1)
	m.Write(0x8000, 0x10)

	mode, ok = m.GetMirror()
	if !ok || mode != MirrorSingle1 {
		t.Errorf("Switch a MirrorSingle1 falló: %d", mode)
	}

	// 3. Volver a Single Screen Low -> Bit 4 = 0
	// Escribir 0x00 (Bank 0, Mirror 0)
	m.Write(0x8000, 0x00)

	mode, _ = m.GetMirror()
	if mode != MirrorSingle0 {
		t.Errorf("Switch back a MirrorSingle0 falló: %d", mode)
	}
}
