package main

// PulseChannel representa los dos canales de onda cuadrada (Pulse 1 y 2).
type PulseChannel struct {
	Enabled  bool
	IsPulse2 bool // Para comportamiento específico de Sweep (Negate)

	// Registros
	GlobalVol byte // $4000
	SweepReg  byte // $4001
	TimerLow  byte // $4002
	LengthReg byte // $4003

	// Unidades Lógicas
	Envelope Envelope

	// Length Counter
	LengthValue byte
	HaltLength  bool // Bit 5 de $4000 (Loop/Halt)

	// Timer
	TimerPeriod uint16
	TimerValue  uint16

	// Duty
	DutyMode     byte // 0-3 (12.5%, 25%, 50%, 75%)
	DutySequence byte // índice actual en la secuencia de onda
}

func NewPulseChannel(isPulse2 bool) *PulseChannel {
	return &PulseChannel{
		IsPulse2: isPulse2,
	}
}

func (p *PulseChannel) Write(addr uint16, data byte) {
	switch addr {
	case 0: // $4000 / $4004
		// Ctrl: DDLC VVVV
		p.GlobalVol = data
		p.DutyMode = (data >> 6) & 0x03
		p.HaltLength = (data & 0x20) != 0 // Length Counter Halt / Loop Envelope

		// Envelope Config
		p.Envelope.loop = p.HaltLength
		p.Envelope.constantVol = (data & 0x10) != 0
		p.Envelope.volCub = data & 0x0F

	case 1: // $4001 / $4005
		// Sweep: EPPP NSSS
		p.SweepReg = data
		// TODO: Implementar lógica de Sweep si es necesario para FME-7 IRQ timing (no suele serlo)

	case 2: // $4002 / $4006
		// Timer Low: TTTT TTTT
		p.TimerLow = data
		p.TimerPeriod = (p.TimerPeriod & 0xFF00) | uint16(data)

	case 3: // $4003 / $4007
		// Length: LLLL LTTT
		p.LengthReg = data

		// Timer High
		high := uint16(data & 0x07)
		p.TimerPeriod = (p.TimerPeriod & 0x00FF) | (high << 8)

		// Cargar Length Counter (Lookup table)
		if p.Enabled {
			idx := (data >> 3) & 0x1F
			p.LengthValue = lengthTable[idx]
		}

		// Restart Envelope
		p.Envelope.startFlag = true

		// Reset Duty (phase)
		p.DutySequence = 0
	}
}

// TickLength clockea el contador de duración (llamado frames aprox 60hz/120hz)
func (p *PulseChannel) TickLength() {
	if !p.HaltLength && p.LengthValue > 0 {
		p.LengthValue--
	}
}

// CheckStatus devuelve true si el canal está "activo" (Length > 0)
// Usado para lectura de $4015
func (p *PulseChannel) Status() bool {
	return p.LengthValue > 0
}

// Enable/Disable maneja escrituras a $4015
func (p *PulseChannel) SetEnabled(enabled bool) {
	p.Enabled = enabled
	if !enabled {
		p.LengthValue = 0
	}
}

var dutyTable = [4][8]byte{
	{0, 1, 0, 0, 0, 0, 0, 0}, // 12.5%
	{0, 1, 1, 0, 0, 0, 0, 0}, // 25%
	{0, 1, 1, 1, 1, 0, 0, 0}, // 50%
	{1, 0, 0, 1, 1, 1, 1, 1}, // 25% negated (75%)
}

func (p *PulseChannel) TickTimer() {
	if p.TimerValue == 0 {
		p.TimerValue = p.TimerPeriod
		p.DutySequence = (p.DutySequence + 1) & 7
	} else {
		p.TimerValue--
	}
}

func (p *PulseChannel) Output() byte {
	if !p.Enabled || p.LengthValue == 0 || p.TimerPeriod < 8 {
		return 0
	}
	if dutyTable[p.DutyMode][p.DutySequence] != 0 {
		return p.Envelope.Output()
	}
	return 0
}
