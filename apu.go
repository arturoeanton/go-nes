package main

// APU (Audio Processing Unit)
// Implementación "densa" con canales lógicos y registros de estado.
type APU struct {
	// Canales
	Pulse1   *PulseChannel
	Pulse2   *PulseChannel
	Triangle *TriangleChannel
	Noise    *NoiseChannel
	DMC      *DMCChannel

	// Frame Counter ($4017)
	frameCounterMode byte // 0=4-step, 1=5-step
	irqInhibit       bool // Bit 6 de $4017
	frameIrqActive   bool // Flag interno de Frame IRQ

	// Cycles
	cycleCount int
	audioTimer float64 // Timer acumulativo para sampling de audio preciso (44100Hz)

	// Audio Buffer (Channel for Ebiten)
	AudioBuffer chan float32
}

func NewAPU() *APU {
	return &APU{
		Pulse1:      NewPulseChannel(false),
		Pulse2:      NewPulseChannel(true),
		Triangle:    NewTriangleChannel(),
		Noise:       NewNoiseChannel(),
		DMC:         NewDMCChannel(),
		AudioBuffer: make(chan float32, 4096),
	}
}

// Read maneja lecturas de registros de la APU (ej: $4015)
func (a *APU) Read(addr uint16) byte {
	switch addr {
	case 0x4015:
		// Status Register
		// Bit 0: Pulse 1 Length > 0
		// Bit 1: Pulse 2 Length > 0
		// Bit 2: Triangle Length > 0
		// Bit 3: Noise Length > 0
		// Bit 4: DMC Bytes Remaining > 0
		// Bit 5: Open Bus (unused)
		// Bit 6: Frame Interrupt Flag
		// Bit 7: DMC Interrupt Flag

		data := byte(0)
		if a.Pulse1.Status() {
			data |= 0x01
		}
		if a.Pulse2.Status() {
			data |= 0x02
		}
		if a.Triangle.Status() {
			data |= 0x04
		}
		if a.Noise.Status() {
			data |= 0x08
		}
		if a.DMC.Status() {
			data |= 0x10
		}

		if a.frameIrqActive {
			data |= 0x40
		}
		if a.DMC.IrqPending {
			data |= 0x80
		}

		// Leer $4015 limpia el flag de Frame IRQ (Side Effect)
		a.frameIrqActive = false

		//fmt.Printf("APU Read $4015: Status=%02X\n", data) // Debug
		return data
	}
	return 0
}

// Write maneja escrituras a registros de la APU
func (a *APU) Write(addr uint16, data byte) {
	switch {
	case addr >= 0x4000 && addr <= 0x4003: // Pulse 1
		a.Pulse1.Write(addr-0x4000, data)

	case addr >= 0x4004 && addr <= 0x4007: // Pulse 2
		a.Pulse2.Write(addr-0x4004, data)

	case addr >= 0x4008 && addr <= 0x400B: // Triangle
		a.Triangle.Write(addr-0x4008, data)

	case addr >= 0x400C && addr <= 0x400F: // Noise
		a.Noise.Write(addr-0x400C, data)

	case addr >= 0x4010 && addr <= 0x4013: // DMC
		a.DMC.Write(addr-0x4010, data)

	case addr == 0x4015: // Status Control (Enable/Disable Channels)
		// 000D NT21
		a.Pulse1.SetEnabled((data & 0x01) != 0)
		a.Pulse2.SetEnabled((data & 0x02) != 0)
		a.Triangle.SetEnabled((data & 0x04) != 0)
		a.Noise.SetEnabled((data & 0x08) != 0)
		a.DMC.SetEnabled((data & 0x10) != 0)

		// Limpiar flag DMC IRQ si se escribe en $4015? (Depende documentación,
		// pero normalmente se limpia leyendo, escribir maneja Enables).

	case addr == 0x4017: // Frame Counter
		a.frameCounterMode = (data >> 7) & 1
		a.irqInhibit = (data & 0x40) != 0

		if a.irqInhibit {
			a.frameIrqActive = false
		}

		// Side effect: Reset 4-step/5-step sequence
		a.cycleCount = 0

		// IMPORTANTE: Si es Mode 1, se clockean las unidades inmediatamente (Quarter + Half frame)
		if a.frameCounterMode == 1 {
			a.clockQuarterFrame()
			a.clockHalfFrame()
		}

		// fmt.Printf("APU Write $4017: Mode=%d Inhibit=%v\n", a.frameCounterMode, a.irqInhibit)

	default:
		// Ignorar
	}
}

// Tick avanza el estado de la APU
func (a *APU) Tick() {
	// Frame Sequence:
	// Cycles approx per step: 7457 (NTSC)
	const stepCycles = 7457 // 29829 / 4 approx

	a.cycleCount++

	// Avance de Timers de onda (Waveform generation)
	a.Pulse1.TickTimer()
	a.Pulse2.TickTimer()
	a.Triangle.TickTimer()
	a.Noise.TickTimer()
	// DMC Timer... (Pending)

	// Sampling (Accurate 44100Hz)
	// CPU Freq approx 1.789773 MHz.
	// 1789773 / 44100 = 40.5844...
	a.audioTimer += 1.0
	if a.audioTimer >= 40.5844 {
		a.audioTimer -= 40.5844
		select {
		case a.AudioBuffer <- a.Output():
		default:
			// Buffer lleno, descartar (aunque con 44100 de buffer no debería pasar mucho)
		}
	}

	mode := a.frameCounterMode

	// Check Steps
	// (Implementación simplificada de triggers, idealmente usar rangos)
	// Pero para no perder ticks, usaremos comprobación exacta o >= con reset.
	// Dado el reset en $4017, la cuenta es fiable.

	if mode == 0 { // 4-Step Sequence
		if a.cycleCount == 7457 {
			a.clockQuarterFrame()
		} else if a.cycleCount == 14913 {
			a.clockQuarterFrame()
			a.clockHalfFrame()
		} else if a.cycleCount == 22371 {
			a.clockQuarterFrame()
		} else if a.cycleCount == 29829 { // Paso 4 (last)
			a.clockQuarterFrame()
			a.clockHalfFrame()
			if !a.irqInhibit {
				a.frameIrqActive = true
			}
		} else if a.cycleCount >= 29830 {
			a.cycleCount = 0
		}
	} else { // 5-Step Sequence
		if a.cycleCount == 7457 {
			a.clockQuarterFrame()
		} else if a.cycleCount == 14913 {
			a.clockQuarterFrame()
			a.clockHalfFrame()
		} else if a.cycleCount == 22371 {
			a.clockQuarterFrame()
		} else if a.cycleCount == 29829 {
			// Do nothing in step 4 of mode 1
		} else if a.cycleCount == 37281 { // Paso 5
			a.clockQuarterFrame()
			a.clockHalfFrame()
		} else if a.cycleCount >= 37282 {
			a.cycleCount = 0
		}
	}

	// DMC Tick (si necesario)
	a.DMC.Tick()
}

// clockQuarterFrame: Envuelves, Tri Linear Counter (240 Hz)
func (a *APU) clockQuarterFrame() {
	a.Pulse1.Envelope.Tick()
	a.Pulse2.Envelope.Tick()
	a.Noise.Envelope.Tick()
	a.Triangle.TickLinear()
}

// clockHalfFrame: Length Counters, Sweep (120 Hz)
func (a *APU) clockHalfFrame() {
	a.Pulse1.TickLength()
	a.Pulse2.TickLength()
	a.Triangle.TickLength()
	a.Noise.TickLength()
	// Sweep ticks...
}

// IRQState devuelve true si la APU tiene una interrupción pendiente
func (a *APU) IRQState() bool {
	return (a.frameIrqActive && !a.irqInhibit) || a.DMC.IrqPending
}

// Output genera una muestra de audio mezclada (0.0 - 1.0)
func (a *APU) Output() float32 {
	p1 := float32(a.Pulse1.Output())
	p2 := float32(a.Pulse2.Output())
	t := float32(a.Triangle.Output())
	n := float32(a.Noise.Output())
	d := float32(a.DMC.DirectLoad)

	// Aproximación lineal (más rápida que la no-lineal exacta)
	pulseOut := 0.00752 * (p1 + p2)
	tndOut := 0.00851*t + 0.00494*n + 0.00335*d

	return pulseOut + tndOut
}
