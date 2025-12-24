package main

// ==========================================
// MAPPER 69 (Sunsoft FME-7 / 5A / 5B)
// ==========================================
// Usado por: Batman Return of the Joker, Gimmick!, Hebereke, Gremlins 2
//
// Características:
// - PRG ROM: hasta 512KB en ventanas de 8KB
// - PRG RAM: 8KB en $6000-$7FFF (opcional)
// - CHR: 256KB en ventanas de 1KB x 8
// - IRQ: Contador de 16-bit decrementado por ciclo CPU
// - Mirroring: H, V, o Single Screen

type Mapper69 struct {
	prgBanks int // Número de bancos PRG de 16KB
	chrBanks int // Número de bancos CHR de 8KB (o 0 si CHR-RAM)
	chrSize  int // Tamaño total de CHR en bytes

	// Registro de comando seleccionado ($8000)
	commandReg byte

	// CHR Banking: 8 registros para 8 ventanas de 1KB
	chrBanksRegs [8]int

	// PRG Banking: 4 registros para $6000, $8000, $A000, $C000
	// $E000-$FFFF siempre fijo al último banco
	prgBanksRegs [4]int
	prgRAMEnable bool // Bit 7 de comando $8
	prgRAMSelect bool // Bit 6 de comando $8 (1=RAM, 0=ROM en $6000)

	// Mirroring
	mirroring MirrorMode

	// IRQ: contador de 16-bit decrementado por ciclo CPU
	irqCounter        uint16
	irqEnabled        bool // Bit 0 de comando $D
	irqCounterEnabled bool // Bit 7 de comando $D
	irqPending        bool // Flag de IRQ pendiente
}

func NewMapper69(prgBanks, chrBanks int) *Mapper69 {
	m := &Mapper69{
		prgBanks:  prgBanks,
		chrBanks:  chrBanks,
		chrSize:   chrBanks * 8192, // chrBanks está en unidades de 8KB
		mirroring: MirrorVertical,
	}

	// Inicializar CHR banks a identidad (0-7)
	for i := 0; i < 8; i++ {
		m.chrBanksRegs[i] = i
	}

	// Inicializar PRG banks
	total8KB := prgBanks * 2 // Convertir 16KB banks a 8KB banks
	if total8KB >= 4 {
		m.prgBanksRegs[0] = 0            // $6000 = banco 0 (o RAM)
		m.prgBanksRegs[1] = 0            // $8000 = banco 0
		m.prgBanksRegs[2] = 1            // $A000 = banco 1
		m.prgBanksRegs[3] = total8KB - 2 // $C000 = penúltimo banco
	}
	// $E000-$FFFF está fijo al último banco (manejado en Read)

	return m
}

func (m *Mapper69) Read(addr uint16) int {
	total8KB := m.prgBanks * 2

	switch {
	case addr >= 0x6000 && addr <= 0x7FFF:
		// $6000-$7FFF: PRG RAM o ROM según configuración
		if m.prgRAMSelect {
			// RAM seleccionada - devolver -1 para que Bus maneje WRAM
			return -1
		}
		// ROM seleccionada
		bank := m.prgBanksRegs[0] & 0x3F
		if total8KB > 0 {
			bank = bank % total8KB
		}
		return bank*8192 + int(addr&0x1FFF)

	case addr >= 0x8000 && addr <= 0x9FFF:
		bank := m.prgBanksRegs[1] & 0x3F
		if total8KB > 0 {
			bank = bank % total8KB
		}
		return bank*8192 + int(addr&0x1FFF)

	case addr >= 0xA000 && addr <= 0xBFFF:
		bank := m.prgBanksRegs[2] & 0x3F
		if total8KB > 0 {
			bank = bank % total8KB
		}
		return bank*8192 + int(addr&0x1FFF)

	case addr >= 0xC000 && addr <= 0xDFFF:
		bank := m.prgBanksRegs[3] & 0x3F
		if total8KB > 0 {
			bank = bank % total8KB
		}
		return bank*8192 + int(addr&0x1FFF)

	case addr >= 0xE000 && addr <= 0xFFFF:
		// Último banco fijo
		if total8KB > 0 {
			return (total8KB-1)*8192 + int(addr&0x1FFF)
		}
		return int(addr & 0x1FFF)
	}

	return -1
}

func (m *Mapper69) Write(addr uint16, data byte) {
	switch {
	case addr >= 0x8000 && addr <= 0x9FFF:
		// Command Register - selecciona qué registro modificar
		m.commandReg = data & 0x0F

	case addr >= 0xA000 && addr <= 0xBFFF:
		// Parameter Register - escribe al registro seleccionado
		m.runCommand(data)
	}
}

func (m *Mapper69) runCommand(data byte) {
	cmd := m.commandReg

	switch {
	case cmd <= 0x07:
		// CHR Banks 0-7 (ventanas de 1KB)
		m.chrBanksRegs[cmd] = int(data)

	case cmd == 0x08:
		// PRG Bank 0 ($6000) con flags de RAM
		m.prgRAMEnable = (data & 0x80) != 0
		m.prgRAMSelect = (data & 0x40) != 0
		m.prgBanksRegs[0] = int(data & 0x3F)

	case cmd == 0x09:
		// PRG Bank 1 ($8000)
		m.prgBanksRegs[1] = int(data & 0x3F)

	case cmd == 0x0A:
		// PRG Bank 2 ($A000)
		m.prgBanksRegs[2] = int(data & 0x3F)

	case cmd == 0x0B:
		// PRG Bank 3 ($C000)
		m.prgBanksRegs[3] = int(data & 0x3F)

	case cmd == 0x0C:
		// Mirroring
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

	case cmd == 0x0D:
		// IRQ Control
		// Bit 7: Counter Enable
		// Bit 0: IRQ Enable
		// Escribir a este registro hace acknowledge del IRQ
		m.irqCounterEnabled = (data & 0x80) != 0
		m.irqEnabled = (data & 0x01) != 0
		m.irqPending = false // Acknowledge

	case cmd == 0x0E:
		// IRQ Counter Low Byte
		m.irqCounter = (m.irqCounter & 0xFF00) | uint16(data)

	case cmd == 0x0F:
		// IRQ Counter High Byte
		m.irqCounter = (m.irqCounter & 0x00FF) | (uint16(data) << 8)
	}
}

func (m *Mapper69) ReadCHR(addr uint16) int {
	chunk := addr / 0x0400 // Dividir en bloques de 1KB
	if chunk < 8 {
		bank := m.chrBanksRegs[chunk]
		offset := bank*1024 + int(addr&0x03FF)

		// Bounds checking para CHR-ROM
		if m.chrSize > 0 {
			offset = offset % m.chrSize
		}
		return offset
	}
	return int(addr)
}

func (m *Mapper69) Scanline() {} // FME-7 usa IRQ por ciclos CPU, no scanlines

func (m *Mapper69) Tick() {
	// El contador se decrementa una vez por ciclo CPU si está habilitado
	if m.irqCounterEnabled {
		// Primero decrementar
		oldCounter := m.irqCounter
		m.irqCounter--

		// IRQ se dispara en underflow (0 → 0xFFFF)
		if oldCounter == 0x0000 && m.irqCounter == 0xFFFF {
			m.irqPending = true
		}
	}
}

func (m *Mapper69) NotifyA12(high bool) {} // FME-7 no usa A12 detection

func (m *Mapper69) IRQState() bool {
	// IRQ físico solo si hay pendiente Y está habilitado
	return m.irqPending && m.irqEnabled
}

func (m *Mapper69) GetMirror() (MirrorMode, bool) {
	return m.mirroring, true
}
