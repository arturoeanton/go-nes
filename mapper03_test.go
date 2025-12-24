package main

import "testing"

// ==========================================
// TEST UNITARIOS PARA MAPPER 3 (CNROM)
// ==========================================
// Usado en: Cybernoid, juegos simples con mucha grafica.
// Características: PRG Fijo (sin bancos), CHR Switchable.

func TestMapper3_PRG_Fixed(t *testing.T) {
	// 2 Bancos PRG (32KB). Mapea todo directo.
	m := NewMapper3(2)

	// Leer $8000 -> Offset 0
	val := m.Read(0x8000)
	if val != 0 {
		t.Errorf("Read $8000 falló: %d", val)
	}

	// Leer $C000 -> Offset 16384
	val = m.Read(0xC000)
	if val != 16384 {
		t.Errorf("Read $C000 falló: %d", val)
	}
}

func TestMapper3_CHR_Banking(t *testing.T) {
	// 4 Bancos CHR de 8KB (32KB Total).
	// Escribir en ROM ($8000+) selecciona banco CHR completo (8KB).
	m := NewMapper3(2)

	// 1. Estado inicial (Banco 0)
	off := m.ReadCHR(0x0000)
	if off != 0 {
		t.Errorf("CHR Inicial falló: %d", off)
	}

	// 2. Cambiar a Banco 2
	// Escribir valor 2 en $8000
	// bits 0-1 seleccionan el banco.
	m.Write(0x8000, 0x02)

	// Verificar lectura CHR
	// Offset = Banco * 8192 + addr
	off = m.ReadCHR(0x1000)
	expected := 2*8192 + 0x1000
	if off != expected {
		t.Errorf("CHR Bank Switch a 2 falló. Esperado %d, obtenido %d", expected, off)
	}

	// 3. Cambiar a Banco 3 (máximo para 2 bits)
	m.Write(0xFFFF, 0x03)
	off = m.ReadCHR(0x0000)
	if off != 3*8192 {
		t.Errorf("CHR Bank Switch a 3 falló. Esperado %d, obtenido %d", 3*8192, off)
	}
}
