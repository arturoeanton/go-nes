package main

import "testing"

// ==========================================
// TEST UNITARIOS PARA MAPPER 69 (Sunsoft FME-7)
// ==========================================
// Crítico para "Batman: Return of the Joker".
// Verifica Bank Switching, Mirroring y sobre todo IRQs.

func TestMapper69_IRQ_Counter(t *testing.T) {
	// Crear Mapper con algunos bancos dummy
	m := NewMapper69(16, 16)

	// 1. Configurar Contador IRQ
	// Seleccionar Comando $D (IRQ Control) -> Disable Enabled, Enable Counter
	// $81 = 1000 0001 (Counter Enabled, IRQ Enabled)
	m.Write(0x8000, 0x0D)
	m.Write(0xA000, 0x81)

	// Seleccionar Comando $E (Low Byte) -> $10
	m.Write(0x8000, 0x0E)
	m.Write(0xA000, 0x10)

	// Seleccionar Comando $F (High Byte) -> $00 (Contador = $0010 = 16)
	m.Write(0x8000, 0x0F)
	m.Write(0xA000, 0x00)

	// Verificar estado inicial (no activo)
	if m.irqPending {
		t.Error("IRQ no debería estar activa al inicio")
	}

	// 2. Hacer Tick 15 veces (Contador baja de 16 a 1)
	for i := 0; i < 15; i++ {
		m.Tick()
		if m.irqPending {
			t.Errorf("IRQ se disparó prematuramente en tick %d", i)
		}
	}

	// 3. Tick 16 (Contador baja a 0 -> Debería disparar?)
	// FME-7 dispara cuando decremental DESDE 0 -> FFFF (Underflow).
	// Osea cuando Counter == 0, siguiente tick -> FFFF y IRQ.

	m.Tick() // 1 -> 0
	if m.irqPending {
		t.Error("IRQ se disparó en contador 0 (debería ser en underflow)")
	}

	m.Tick() // 0 -> FFFF (Underflow!)
	if !m.irqPending {
		t.Error("IRQ NO se disparó tras underflow del contador")
	}

	// 4. Verificar IRQState
	if !m.IRQState() {
		t.Error("IRQState devolvió false con IRQ activa y habilitada")
	}

	// 5. Acknowledge (Deshabilitar IRQ limpia flag)
	// Comando D: $00 (Disable everything)
	m.Write(0x8000, 0x0D)
	m.Write(0xA000, 0x00)

	if m.irqPending {
		t.Error("Escribir comando D no limpió el flag irqPending")
	}
	if m.IRQState() {
		t.Error("IRQState activo tras acknowledge")
	}
}

func TestMapper69_BankSwitching(t *testing.T) {
	m := NewMapper69(16, 16) // 128KB ROM

	// Test PRG Bank 0 ($6000)
	// Comando 8: Seleccionar banco 5
	m.Write(0x8000, 0x08)
	m.Write(0xA000, 0x05)

	// Verificar registro interno
	if m.prgBanksRegs[0] != 0x05 {
		t.Errorf("No se guardó el banco 5 en reg 0. Obtenido %d", m.prgBanksRegs[0])
	}

	// Test RAM Select (Bit 6)
	// $45 = 0100 0101 (RAM Select, Bank 5)
	m.Write(0xA000, 0x45)

	// Read en $6000 debería devolver -1 (RAM)
	offset := m.Read(0x6000)
	if offset != -1 {
		t.Errorf("Esperado -1 para RAM Select, obtenido offset %d", offset)
	}

	// Read en $6000 con ROM Select (Bit 6=0)
	m.Write(0xA000, 0x05)
	offset = m.Read(0x6000)
	// Offset = 5 * 8192 = 40960
	expected := 5 * 8192
	if offset != expected {
		t.Errorf("ROM Offset incorrecto. Esperado %d, obtenido %d", expected, offset)
	}
}
