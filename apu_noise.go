package main

// NoiseChannel: Generador de ruido pseudo-aleatorio.
type NoiseChannel struct {
	Enabled bool

	// Registros
	EnvelopeReg byte // $400C (Vol/Envelope)
	LoopNoise   bool // $400E Bit 7
	PeriodIndex byte // $400E (0-15)
	LengthReg   byte // $400F

	// Unidades
	Envelope    Envelope
	LengthValue byte
	HaltLength  bool // Bit 5 de $400C

	// Shift Register (LFSR)
	ShiftRegister uint16
	Mode          bool // Bit 7 de $400E (Short mode)

	// Timer
	TimerPeriod uint16
	TimerValue  uint16
}

func NewNoiseChannel() *NoiseChannel {
	return &NoiseChannel{
		ShiftRegister: 1, // Inicializado en 1
	}
}

func (n *NoiseChannel) Write(addr uint16, data byte) {
	switch addr {
	case 0: // $400C
		// --LC VVVV
		n.EnvelopeReg = data
		n.HaltLength = (data & 0x20) != 0

		n.Envelope.loop = n.HaltLength
		n.Envelope.constantVol = (data & 0x10) != 0
		n.Envelope.volCub = data & 0x0F

	case 1: // $400D - Unused

	case 2: // $400E
		// M--- PPPP
		n.LoopNoise = (data & 0x80) != 0 // Mode flag
		n.Mode = n.LoopNoise
		n.PeriodIndex = data & 0x0F
		n.TimerPeriod = noiseTable[n.PeriodIndex]

	case 3: // $400F
		// LLLL L---
		n.LengthReg = data
		if n.Enabled {
			idx := (data >> 3) & 0x1F
			n.LengthValue = lengthTable[idx]
		}

		n.Envelope.startFlag = true
	}
}

var noiseTable = [...]uint16{
	4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068,
}

func (n *NoiseChannel) TickTimer() {
	if n.TimerValue == 0 {
		n.TimerValue = n.TimerPeriod
		shift := byte(1)
		if n.Mode {
			shift = 6
		}
		// Bit 0 xor Bit 1 (or 6)
		feedback := (n.ShiftRegister & 1) ^ ((n.ShiftRegister >> shift) & 1)
		n.ShiftRegister >>= 1
		n.ShiftRegister |= feedback << 14
	} else {
		n.TimerValue--
	}
}

func (n *NoiseChannel) Output() byte {
	if !n.Enabled || n.LengthValue == 0 {
		return 0
	}
	if (n.ShiftRegister & 1) == 0 {
		return n.Envelope.Output()
	}
	return 0
}

func (n *NoiseChannel) TickLength() {
	if !n.HaltLength && n.LengthValue > 0 {
		n.LengthValue--
	}
}

func (n *NoiseChannel) Status() bool {
	return n.LengthValue > 0
}

func (n *NoiseChannel) SetEnabled(enabled bool) {
	n.Enabled = enabled
	if !enabled {
		n.LengthValue = 0
	}
}
