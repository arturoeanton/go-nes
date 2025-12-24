package main

import "fmt"

// APU (Audio Processing Unit)
// Implementación mínima para satisfacer dependencias de IRQ (Frame Counter).
type APU struct {
	// Frame Counter ($4017)
	frameCounterMode byte // 0=4-step, 1=5-step
	irqInhibit       bool // Bit 6 de $4017
	frameIrqActive   bool // Flag interno de Frame IRQ

	// Cycles
	cycleCount int
}

func NewAPU() *APU {
	return &APU{}
}

// Read maneja lecturas de registros de la APU (ej: $4015)
func (a *APU) Read(addr uint16) byte {
	switch addr {
	case 0x4015:
		// Status Register
		// Bit 6: Frame Interrupt Flag
		data := byte(0)
		if a.frameIrqActive {
			data |= 0x40
		}

		// Leer $4015 limpia el flag de Frame IRQ
		a.frameIrqActive = false

		fmt.Printf("APU Read $4015: Status=%02X\n", data) // Debug
		return data
	}
	// Otros registros de lectura retornan 0 por ahora
	return 0
}

// Write maneja escrituras a registros de la APU
func (a *APU) Write(addr uint16, data byte) {
	switch addr {
	case 0x4017:
		// Frame Counter Control
		// Bit 7: Mode (0=4-step, 1=5-step)
		// Bit 6: IRQ Inhibit (1=Disable Frame IRQ)
		a.frameCounterMode = (data >> 7) & 1
		a.irqInhibit = (data & 0x40) != 0

		if a.irqInhibit {
			// Si se inhibe, el flag se limpia inmediatamente
			a.frameIrqActive = false
		}

		fmt.Printf("APU Write $4017: Mode=%d Inhibit=%v\n", a.frameCounterMode, a.irqInhibit)

		// Side effect: Escribir a $4017 resetea el contador de frames
		a.cycleCount = 0

	default:
		// Otros registros de sonido ($4000-$4013) ignorados por ahora
	}
}

// Tick avanza el estado de la APU (llamado por CPU o Bus)
func (a *APU) Tick() {
	// Secuenciador de Frames (Simplificado)
	// Mode 0: 4-step sequence
	// Mode 1: 5-step sequence (No sets flag)

	// Solo si es Mode 0
	if a.frameCounterMode == 0 {
		a.cycleCount++
		// Aprox 29830 ciclos por frame NTSC
		// El paso 4 activa el flag de IRQ
		// (Nota: Hardware real tiene timings precisos de 4 pasos, esto es aproximado)
		if a.cycleCount >= 29829 {
			if !a.irqInhibit {
				a.frameIrqActive = true
			}
			// Resetear contador (loop)
			a.cycleCount = 0
		}
	} else {
		// Mode 1: No genera flags de Frame IRQ
		// (Podríamos resetear cycleCount si quisiéramos simular el loop de 5 pasos, pero no afecta flags)
	}
}

// IRQState devuelve true si la APU tiene una interrupción pendiente
func (a *APU) IRQState() bool {
	// La línea IRQ hacia la CPU solo baja si el flag está activo Y no estamos inhibidos.
	// Aunque Tick ya previene setear el flag si está inhibido, esta doble verificación es segura.
	return a.frameIrqActive && !a.irqInhibit
}
