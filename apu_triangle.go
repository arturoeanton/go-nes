package main

// TriangleChannel: Onda triangular (pseudo-analógica).
// Tiene un Linear Counter adicional para control de duración de alta resolución.
type TriangleChannel struct {
	Enabled bool

	// Registros
	LinearCounterReload byte // $4008
	ControlFlag         bool // $4008 Bit 7 (Control Flag / Halt)

	TimerLow    byte // $400A
	TimerPeriod uint16
	LengthReg   byte // $400B

	// Unidades
	LengthValue  byte
	LinearValue  byte
	ReloadLinear bool

	HaltLength bool // Mismo bit que ControlFlag en $4008

	// Timer
	TimerValue uint16

	// Secuenciador (Wave)
	SequenceIndex byte // 0-31
}

func NewTriangleChannel() *TriangleChannel {
	return &TriangleChannel{}
}

func (t *TriangleChannel) Write(addr uint16, data byte) {
	switch addr {
	case 0: // $4008
		// CRRR RRRR
		t.ControlFlag = (data & 0x80) != 0
		t.HaltLength = t.ControlFlag // Bit 7 controla ambos
		t.LinearCounterReload = data & 0x7F

	case 1: // $4009 - Unused

	case 2: // $400A
		// Timer Low
		t.TimerLow = data
		t.TimerPeriod = (t.TimerPeriod & 0xFF00) | uint16(data)

	case 3: // $400B
		// LLLL LTTT
		t.LengthReg = data

		// Timer High
		high := uint16(data & 0x07)
		t.TimerPeriod = (t.TimerPeriod & 0x00FF) | (high << 8)

		// Load Length
		if t.Enabled {
			idx := (data >> 3) & 0x1F
			t.LengthValue = lengthTable[idx]
		}

		// Set Halt flag for Linear Counter
		t.ReloadLinear = true
	}
}

// TickLength clockea el Length Counter
func (t *TriangleChannel) TickLength() {
	if !t.HaltLength && t.LengthValue > 0 {
		t.LengthValue--
	}
}

// TickLinear clockea el Linear Counter (resolución frame / 4 )
func (t *TriangleChannel) TickLinear() {
	if t.ReloadLinear {
		t.LinearValue = t.LinearCounterReload
	} else if t.LinearValue > 0 {
		t.LinearValue--
	}

	if !t.ControlFlag {
		t.ReloadLinear = false
	}
}

func (t *TriangleChannel) Status() bool {
	return t.LengthValue > 0
}

func (t *TriangleChannel) SetEnabled(enabled bool) {
	t.Enabled = enabled
	if !enabled {
		t.LengthValue = 0
	}
}

var triangleTable = [...]byte{
	15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

func (t *TriangleChannel) TickTimer() {
	if t.TimerValue == 0 {
		t.TimerValue = t.TimerPeriod
		// La onda triangular solo avanza si ambos contadores están activos
		if t.LengthValue > 0 && t.LinearValue > 0 {
			t.SequenceIndex = (t.SequenceIndex + 1) & 31
		}
	} else {
		t.TimerValue--
	}
}

func (t *TriangleChannel) Output() byte {
	// Nota: El canal triangular suele silenciarse si se apaga, pero
	// a diferencia de Pulse, output es directo de la secuencia.
	// Si timer period < 2, suele producir ultrasonido y silenciar.
	if t.TimerPeriod < 2 {
		return 0 // Prevent ultrasonic noise
	}
	// El output es simplemente el valor actual del secuenciador
	// No tiene control de volumen (es fijo).
	return triangleTable[t.SequenceIndex]
}
