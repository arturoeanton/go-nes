package main

import "fmt"

// ==========================================
// MAPPER 69 (Sunsoft FME-7) - Batman Return of the Joker
// ==========================================
// Soporta IRQ basado en ciclos de CPU.
type Mapper69 struct {
	prgBanks int
	chrBanks int

	commandReg byte // Registro de comando seleccionado ($8000)

	// Registros internos FME-7 (0-F)
	chrBanksRegs [8]int // Reg 0-7: Bancos CHR 1KB
	prgBanksRegs [4]int // Reg 8-B: Bancos PRG 8KB

	mirroring         MirrorMode // Reg C: Mirroring
	irqCounter        uint16     // Reg D-E: Contador IRQ 16-bit
	irqEnabled        bool       // Reg D: IRQ Enable
	irqCounterEnabled bool       // Reg D: Counter Enable
	irqActive         bool
}

func NewMapper69(prgBanks, chrBanks int) *Mapper69 {
	m := &Mapper69{
		prgBanks: prgBanks,
		chrBanks: chrBanks,
	}
	// Inicializar valores por defecto (ej: Fix last bank)
	m.prgBanksRegs[3] = m.prgBanks*2 - 1 // Último banco fijo en E000
	return m
}

func (m *Mapper69) Read(addr uint16) int {
	// PRG mapeado en ventanas de 8KB
	window := (int(addr) - 0x6000) / 0x2000 // 0=6000, 1=8000, 2=A000, 3=C000, 4=E000

	if window < 0 {
		return -1
	}

	if window == 0 {
		// Cmd 8: $6000. Check RAM/ROM select (Bit 6).
		// 0=ROM, 1=RAM
		bankVal := m.prgBanksRegs[0]
		if (bankVal & 0x40) != 0 {
			// RAM Selected -> Return -1 to let Bus handle WRAM
			return -1
		}
		// ROM Selected -> Return Physical ROM Index
		return (bankVal&0x3F)*8192 + int(addr&0x1FFF)
	}

	if window >= 1 && window <= 3 {
		return m.prgBanksRegs[window]*8192 + int(addr&0x1FFF)
	}

	// $E000-$FFFF fijo al último banco en implementación standard de FME-7 para la mayoría de juegos
	if window == 4 {
		return (m.prgBanks*2-1)*8192 + int(addr&0x1FFF)
	}

	return -1
}

func (m *Mapper69) Write(addr uint16, data byte) {
	if addr >= 0x8000 && addr <= 0x9FFF {
		// Command Register ($8000)
		m.commandReg = data & 0x0F
		// fmt.Printf("Mapper 69 Select Cmd: %X\n", m.commandReg) // Debug Select
	} else if addr >= 0xA000 && addr <= 0xBFFF {
		// Parameter Register ($A000)
		m.runCommand(data)
		if m.commandReg == 0xD {
			fmt.Printf("Mapper 69 IRQ Ctrl. Data=%02X Val=%04X Act=%v En=%v\n", data, m.irqCounter, m.irqActive, m.irqEnabled)
		}
	}
}

func (m *Mapper69) runCommand(data byte) {
	cmd := m.commandReg
	fmt.Printf("Mapper 69 Cmd: %X Data: %X\n", cmd, data) // Debug
	switch {
	case cmd <= 0x7: // CHR Banks 0-7 ($0000 - $1C00, 1KB chunks)
		m.chrBanksRegs[cmd] = int(data)

	case cmd == 0x8: // PRG Bank 0 ($6000)
		// Bit 0-5: Bank
		// Bit 6: RAM Select (0=ROM, 1=RAM)
		// Bit 7: RAM Enable
		m.prgBanksRegs[0] = int(data) // Guardamos todo el byte para chequear flag de RAM en Read
	case cmd == 0x9: // PRG Bank 1 ($8000)
		m.prgBanksRegs[1] = int(data & 0x3F)
	case cmd == 0xA: // PRG Bank 2 ($A000)
		m.prgBanksRegs[2] = int(data & 0x3F)
	case cmd == 0xB: // PRG Bank 3 ($C000)
		m.prgBanksRegs[3] = int(data & 0x3F)

	case cmd == 0xC: // Mirroring
		switch data & 0x03 {
		case 0:
			m.mirroring = MirrorVertical
		case 1:
			m.mirroring = MirrorHorizontal
		case 2:
			m.mirroring = MirrorSingle0
		case 3:
			m.mirroring = MirrorSingle1
		}

	case cmd == 0xD: // IRQ Control
		m.irqCounterEnabled = (data & 0x80) != 0
		m.irqEnabled = (data & 0x01) != 0
		m.irqActive = false // Acknowledge IRQ
		if data != 0 {
			fmt.Printf("Mapper 69 IRQ Values: Cmd=D Data=%X Enabled=%v CounterEnabled=%v\n", data, m.irqEnabled, m.irqCounterEnabled)
		}

	case cmd == 0xE: // IRQ Counter Low Byte
		m.irqCounter = (m.irqCounter & 0xFF00) | uint16(data)

	case cmd == 0xF: // IRQ Counter High Byte
		m.irqCounter = (m.irqCounter & 0x00FF) | (uint16(data) << 8)
	}
}

func (m *Mapper69) ReadCHR(addr uint16) int {
	chunk := addr / 0x400 // 1KB blocks
	if chunk < 8 {
		return m.chrBanksRegs[chunk]*1024 + int(addr%0x400)
	}
	return int(addr)
}

func (m *Mapper69) Scanline() {} // No usado

func (m *Mapper69) Tick() {
	if m.irqCounterEnabled {
		m.irqCounter--
		if m.irqCounter == 0xFFFF { // Underflow
			m.irqCounter = 0xFFFF
			m.irqActive = true // Flag de IRQ pendiente se activa siempre al desbordar
			fmt.Println("Mapper 69 IRQ PENDING SET!")
		}
	}
}

func (m *Mapper69) IRQState() bool {
	// La interrupción física solo ocurre si hay una pendiente Y están habilitadas
	return m.irqActive && m.irqEnabled
}

func (m *Mapper69) GetMirror() (MirrorMode, bool) {
	return m.mirroring, true
}
