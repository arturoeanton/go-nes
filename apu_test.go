package main

import "testing"

// ==========================================
// TEST UNITARIOS DE APU
// ==========================================
// Verifican la lógica básica de la unidad de audio (principalmente Frame Counter e IRQ).

func TestAPU_FrameCounter_Write(t *testing.T) {
	apu := NewAPU()

	// Escribir modo 5-step, Inhibit IRQ
	// $C0 = 1100 0000 (Mode=1, Inhibit=1)
	apu.Write(0x4017, 0xC0)

	// Verificar estado interno
	// (Necesitamos exportar campos o usar métodos públicos).
	// Como son campos privados en main package, los tests tienen acceso (mismo paquete).

	if apu.frameCounterMode != 1 {
		t.Errorf("APU Mode Incorrecto: Esperado 1, obtenido %d", apu.frameCounterMode)
	}
	if !apu.irqInhibit {
		t.Error("APU IRQ Inhibit falló: debería ser true")
	}
}

func TestAPU_FrameIRQ_Logic(t *testing.T) {
	apu := NewAPU()

	// Configurar Modo 0 (4-step), IRQ Enabled
	apu.Write(0x4017, 0x00) // Mode=0, Inhibit=0

	// Simular Ticks hasta el punto de IRQ
	// Paso 4 es ~29830 ciclos
	for i := 0; i < 29830; i++ {
		apu.Tick()
	}

	// Verificar si el Flag Interno está activo
	if !apu.frameIrqActive {
		t.Error("APU Frame IRQ no se activó tras 29830 ciclos en Modo 0")
	}

	// Verificar si la línea IRQ está activa
	if !apu.IRQState() {
		t.Error("APU IRQState devuelve false aunque flag está activo y no inhibido")
	}

	// Leer Status ($4015) debería limpiar el flag
	status := apu.Read(0x4015)

	// Verificar que el bit 6 estaba activo en la lectura
	if status&0x40 == 0 {
		t.Error("Lectura $4015 no reportó Frame IRQ (Bit 6)")
	}

	// Verificar que el flag se limpió
	if apu.frameIrqActive {
		t.Error("Lectura $4015 no limpió el flag Frame IRQ")
	}
}

func TestAPU_Inhibit_Prevents_IRQ(t *testing.T) {
	apu := NewAPU()

	// Configurar Modo 0 (4-step), IRQ Inhibited
	// $40 = 0100 0000 (Mode=0, Inhibit=1)
	apu.Write(0x4017, 0x40)

	// Avanzar tiempo
	for i := 0; i < 30000; i++ {
		apu.Tick()
	}

	// El flag interno NO debería activarse si inhibit es checked en Tick
	// (Dependiendo de la implementación, hardware real activa flag interno pero no salida IRQ? No, inhibit limpia flag).

	if apu.frameIrqActive {
		t.Error("APU activó Frame IRQ a pesar de Inhibit=true")
	}

	if apu.IRQState() {
		t.Error("APU IRQState activado con Inhibit=true")
	}
}
