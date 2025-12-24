package main

import "testing"

// ==========================================
// TEST UNITARIOS PARA MAPPER 4 (MMC3)
// ==========================================
// Importante para Super Mario Bros 3.
// Verifica Bank Switching complejo y Scanline IRQ.

func TestMapper4_BankSwitching_PRG(t *testing.T) {
	m := NewMapper4(32, 32) // 512KB PRG, 256KB CHR

	// 1. Configurar Modo Normal (0) en $8000
	// Command 6: PRG Bank en $8000 (R6)
	// Mode = 0 (PRG Swap $8000, $C000 Fixed)
	m.Write(0x8000, 0x06) // Cmd 6, Mode 0

	// Escribir Dato: Banco 5
	m.Write(0x8001, 0x05)

	// Verificar lectura en $8000
	// Debería ser Banco 5 (offset = 5 * 8192)
	off := m.Read(0x8000)
	if off != 5*8192 {
		t.Errorf("PRG Bank 0 falló: esperado %d, obtenido %d", 5*8192, off)
	}

	// 2. Cambiar a Modo Swap (1) en $8000
	// 0x46 = 0100 0110 (Mode=1, Cmd 6) -> R6 va a $C000 ahora!
	m.Write(0x8000, 0x46)

	// La configuración de registros se mantiene (R6 sigue siendo 5),
	// pero al cambiar el modo, la ventana $8000 debe cambiar.
	// En modo 1, $8000 se vuelve Fijo al penúltimo banco (total-2).
	// Total = 64 bancos de 8KB (512KB). Penúltimo = 62.

	off = m.Read(0x8000)
	expected := (32*2 - 2) * 8192
	if off != expected {
		t.Errorf("PRG Swap Mode falló en $8000: esperado %d (Fixed Last-2), obtenido %d", expected, off)
	}

	// Y $C000 ahora debe obedecer a R6 (Banco 5)
	off = m.Read(0xC000)
	if off != 5*8192 {
		t.Errorf("PRG Swap Mode falló en $C000: esperado %d (R6), obtenido %d", 5*8192, off)
	}
}

func TestMapper4_IRQ_Scanline(t *testing.T) {
	m := NewMapper4(16, 16)

	// 1. Configurar IRQ
	// $C000: Latch = 5 scanlines
	m.Write(0xC000, 0x05)
	// $C001: Reload Request
	m.Write(0xC001, 0x00) // Data irrelevante
	// $E001: Enable IRQ
	m.Write(0xE001, 0x00)

	// Estado inicial
	if m.irqActive {
		t.Error("IRQ activa prematuramente")
	}

	// 2. Simular Scanlines
	// Al primer Scanline, el contador se recarga con Latch (5)
	m.Scanline()
	if m.irqCounter != 5 {
		t.Errorf("Scanline Reload falló: Counter %d, esperado 5", m.irqCounter)
	}

	// Siguientes 4 líneas, decerementa 4, 3, 2, 1
	for i := 0; i < 4; i++ {
		m.Scanline()
		if m.irqActive {
			t.Errorf("IRQ disparada muy pronto en línea %d (Counter=%d)", i, m.irqCounter)
		}
	}

	// Siguiente decerementa a 0 -> DISPARA
	m.Scanline()
	if m.irqCounter != 0 {
		t.Errorf("Counter no llegó a 0, valor: %d", m.irqCounter)
	}
	if !m.irqActive {
		t.Error("IRQ no se disparó al llegar a 0")
	}

	// 3. Acknowledge Interrupt
	// Disable ($E000)
	m.Write(0xE000, 0x00)
	if m.irqActive {
		t.Error("Disable IRQ ($E000) no limpió el flag irqActive")
	}
}

func TestMapper4_Zero_Counter_Reload(t *testing.T) {
	m := NewMapper4(16, 16)
	// Caso raro: Latch=0. MMC3 trata 0 como "disparar en cada línea" o comportamiento especial?
	// Depende revisiones. Aquí testeamos nuestra implementación simple: 0 recarga Latch.

	m.Write(0xC000, 10) // Latch 10
	m.irqCounter = 0
	m.irqEnabled = true
	m.irqActive = false

	// Si counter es 0, Scanline() debe recargar Latch
	m.Scanline()
	if m.irqCounter != 10 {
		t.Errorf("Counter 0 no recargó Latch: %d", m.irqCounter)
	}
	// Y no debe disparar IRQ en el ciclo de recarga (normalmente)
	if m.irqActive {
		t.Error("Recarga desde 0 no debería disparar IRQ inmediatamente")
	}
}
