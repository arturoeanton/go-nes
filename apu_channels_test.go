package main

import "testing"

// ==========================================
// TEST UNITARIOS DE CANALES APU
// ==========================================
// Verifican Length Counters y Status Register

func TestAPU_Pulse_LengthCounter(t *testing.T) {
	apu := NewAPU()

	// 1. Enable Pulse 1 via $4015
	apu.Write(0x4015, 0x01)
	if !apu.Pulse1.Enabled {
		t.Error("Pulse 1 Enable via $4015 falló")
	}

	// 2. Load Length Counter
	// $4003: LLLL L... -> index 0000 1 (1) -> table[1] = 254 ticks
	// 0000 1000 = 0x08
	apu.Write(0x4003, 0x08)

	if apu.Pulse1.LengthValue != 254 {
		t.Errorf("Pulse 1 Length Load falló. Esperado 254, obtenido %d", apu.Pulse1.LengthValue)
	}

	// 3. Verificar Status
	if (apu.Read(0x4015) & 0x01) == 0 {
		t.Error("Status $4015 bit 0 debería estar activo (Pulse 1 playing)")
	}

	// 4. Tick Half Frame (Clock Length)
	// Necesitamos ~14913 ciclos para llegar al primer clock de Length (Step 2)
	for i := 0; i < 14913; i++ {
		apu.Tick()
	}

	if apu.Pulse1.LengthValue != 253 {
		t.Errorf("Pulse 1 Length Decrement falló. Esperado 253, obtenido %d", apu.Pulse1.LengthValue)
	}

	// 5. Disable via $4015 -> Should clear Length
	apu.Write(0x4015, 0x00)
	if apu.Pulse1.LengthValue != 0 {
		t.Error("Disable via $4015 no limpió Length Counter")
	}
	if (apu.Read(0x4015) & 0x01) != 0 {
		t.Error("Status bit 0 debería estar inactivo tras disable")
	}
}

func TestAPU_Triangle_Linear(t *testing.T) {
	apu := NewAPU()
	apu.Write(0x4015, 0x04) // Enable Triangle

	// Setup Linear Counter
	// $4008: CRRR RRRR -> 1000 0011 (Control=1, Reload=3) -> $83
	apu.Write(0x4008, 0x83)

	// $400B: Load Length (necesario para activar reload flag?)
	// Tri usa reload flag seteado en $400B write.
	apu.Write(0x400B, 0x08) // Length

	// Tick Quarter Frame (Linear Clock)
	// Step 1: 7457 cycles
	for i := 0; i < 7457; i++ {
		apu.Tick()
	}

	// Al tener Control Flag = 1, Linear Counter debería recargar y mantenerse (Halt)
	if apu.Triangle.LinearValue != 3 {
		t.Errorf("Triangle Linear Reload falló. Esperado 3, obtenido %d", apu.Triangle.LinearValue)
	}

	// Disable Control Flag
	// $03 -> Control=0
	apu.Write(0x4008, 0x03)

	// Next Tick -> Should decrement?
	// Depende de lógica de reload flag.
	// Si reload flag estaba activo, reload. Si no, decrement.
	// Requiere ticks precisos. Test simple de "existe el componente".
	if apu.Triangle.LinearCounterReload != 3 {
		t.Errorf("Triangle Register Write falló")
	}
}

func TestAPU_Noise_Length(t *testing.T) {
	apu := NewAPU()
	apu.Write(0x4015, 0x08) // Enable Noise
	apu.Write(0x400F, 0x08) // Length load -> 254

	if apu.Noise.LengthValue != 254 {
		t.Errorf("Noise Length Load faló")
	}
}

func TestAPU_DMC_Status(t *testing.T) {
	apu := NewAPU()
	// Enable DMC
	apu.Write(0x4015, 0x10)

	// No sample loaded -> BytesRemaining=0 -> Status=0
	if (apu.Read(0x4015) & 0x10) != 0 {
		t.Error("DMC Status activo sin sample")
	}

	// Load Sample (Dummy addr/len)
	apu.Write(0x4012, 0x00)
	apu.Write(0x4013, 0x00) // 1 byte

	// Re-enable to load
	apu.Write(0x4015, 0x00)
	apu.Write(0x4015, 0x10)

	// Ahora BytesRemaining deberían ser != 0?
	// NewDMCChannel inicializa a 0. Write 4015 setea bytes remaining si transition 0->1.
	// Length = (0*16)+1 = 1 byte.

	// Verificar struct interno
	if apu.DMC.BytesRemaining != 1 {
		t.Errorf("DMC Sample Load falló. BytesRemaining %d", apu.DMC.BytesRemaining)
	}

	if (apu.Read(0x4015) & 0x10) == 0 {
		t.Error("DMC Status debería estar activo tras carga", apu.DMC.BytesRemaining)
	}
}
