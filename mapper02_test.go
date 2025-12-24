package main

import "testing"

// ==========================================
// TEST UNITARIOS PARA MAPPER 2 (UxROM)
// ==========================================
// Usado en: Contra, Castlevania.
// Características: Banco PRG switchable en $8000, Banco fijo en $C000.

func TestMapper2_PRG_Banking(t *testing.T) {
	// 8 Bancos de 16KB (128KB Total).
	// Banco indices: 0 a 7.
	// Fixed Bank ($C000) debe ser el 7 (Last).
	m := NewMapper2(8)

	// 1. Verificar Banco Fijo en $C000
	offset := m.Read(0xC000)
	expected := 7 * 16384
	if offset != expected {
		t.Errorf("Banco Fijo $C000 incorrecto: Esperado %d (Banco 7), obtenido %d", expected, offset)
	}

	// 2. Verificar Banco Default en $8000 (0)
	offset = m.Read(0x8000)
	if offset != 0 {
		t.Errorf("Banco Inicial $8000 incorrecto: Esperado 0, obtenido %d", offset)
	}

	// 3. Switch Bank a 3
	// Escribir en cualquier dirección >= $8000 cambia el banco de $8000.
	m.Write(0x9000, 0x03)

	offset = m.Read(0x8000)
	expected = 3 * 16384
	if offset != expected {
		t.Errorf("Switch Bank a 3 falló: Esperado %d, obtenido %d", expected, offset)
	}

	// Verificar que Fixed Bank no cambió
	offset = m.Read(0xC000)
	expected = 7 * 16384
	if offset != expected {
		t.Error("Switching $8000 afectó incorrectamente al banco fijo $C000")
	}

	// 4. Switch Bank a 5 usando otra dirección ($FFFF)
	m.Write(0xFFFF, 0x05)

	offset = m.Read(0x8000)
	expected = 5 * 16384
	if offset != expected {
		t.Errorf("Switch Bank a 5 falló: Esperado %d, obtenido %d", expected, offset)
	}
}

func TestMapper2_CHR_Passthrough(t *testing.T) {
	m := NewMapper2(8)
	// Mapper 2 usualmente usa CHR-RAM, mapeo directo.
	val := m.ReadCHR(0x1234)
	if val != 0x1234 {
		t.Errorf("CHR Read no fue directo: %d", val)
	}
}
