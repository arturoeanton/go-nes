package main

// ==========================================
// MAPPER 09 (MMC2) - Punch-Out!!
// ==========================================
// Características únicas:
// - PRG: 8KB switchable + 24KB fijos (últimos 3 bancos)
// - CHR: 4KB x 2 con switching automático via tiles $FD/$FE
// - Mirroring: Controlado por software
//
// El MMC2 cambia CHR banks automáticamente cuando la PPU lee
// tiles con índice $FD o $FE, permitiendo efectos mid-frame
// sin intervención de la CPU.

type Mapper9 struct {
	prgBanks int // Número de bancos PRG de 16KB
	chrBanks int // Número de bancos CHR de 8KB

	// PRG Banking
	// $8000-$9FFF: Switchable
	// $A000-$FFFF: Fixed to last 3 banks (24KB)
	prgBank int

	// CHR Banking con latch automático
	// Cada pattern table tiene 2 bancos posibles
	// El latch determina cuál se usa
	chrBank0FD int  // Banco para $0000-$0FFF cuando latch0 = $FD
	chrBank0FE int  // Banco para $0000-$0FFF cuando latch0 = $FE
	chrBank1FD int  // Banco para $1000-$1FFF cuando latch1 = $FD
	chrBank1FE int  // Banco para $1000-$1FFF cuando latch1 = $FE
	latch0     byte // Latch para pattern table 0 ($FD o $FE)
	latch1     byte // Latch para pattern table 1 ($FD o $FE)

	// Mirroring
	mirroring MirrorMode
}

func NewMapper9(prgBanks, chrBanks int) *Mapper9 {
	m := &Mapper9{
		prgBanks:  prgBanks,
		chrBanks:  chrBanks,
		mirroring: MirrorHorizontal,
		// Inicializar latches a $FE (valor por defecto)
		latch0: 0xFE,
		latch1: 0xFE,
	}
	return m
}

func (m *Mapper9) Read(addr uint16) int {
	total8KB := m.prgBanks * 2 // Convertir 16KB a 8KB

	switch {
	case addr >= 0x8000 && addr <= 0x9FFF:
		// Banco switchable
		bank := m.prgBank
		if total8KB > 0 {
			bank = bank % total8KB
		}
		return bank*8192 + int(addr&0x1FFF)

	case addr >= 0xA000 && addr <= 0xBFFF:
		// Penúltimo-2 banco (fijo)
		if total8KB >= 3 {
			return (total8KB-3)*8192 + int(addr&0x1FFF)
		}
		return int(addr & 0x1FFF)

	case addr >= 0xC000 && addr <= 0xDFFF:
		// Penúltimo banco (fijo)
		if total8KB >= 2 {
			return (total8KB-2)*8192 + int(addr&0x1FFF)
		}
		return int(addr & 0x1FFF)

	case addr >= 0xE000 && addr <= 0xFFFF:
		// Último banco (fijo)
		if total8KB >= 1 {
			return (total8KB-1)*8192 + int(addr&0x1FFF)
		}
		return int(addr & 0x1FFF)
	}

	return -1
}

func (m *Mapper9) Write(addr uint16, data byte) {
	switch {
	case addr >= 0xA000 && addr <= 0xAFFF:
		// PRG ROM bank select
		m.prgBank = int(data & 0x0F)

	case addr >= 0xB000 && addr <= 0xBFFF:
		// CHR ROM $FD/0000 bank select
		m.chrBank0FD = int(data & 0x1F)

	case addr >= 0xC000 && addr <= 0xCFFF:
		// CHR ROM $FE/0000 bank select
		m.chrBank0FE = int(data & 0x1F)

	case addr >= 0xD000 && addr <= 0xDFFF:
		// CHR ROM $FD/1000 bank select
		m.chrBank1FD = int(data & 0x1F)

	case addr >= 0xE000 && addr <= 0xEFFF:
		// CHR ROM $FE/1000 bank select
		m.chrBank1FE = int(data & 0x1F)

	case addr >= 0xF000 && addr <= 0xFFFF:
		// Mirroring
		if data&0x01 != 0 {
			m.mirroring = MirrorHorizontal
		} else {
			m.mirroring = MirrorVertical
		}
	}
}

func (m *Mapper9) ReadCHR(addr uint16) int {
	var bank int
	var offset int

	if addr < 0x1000 {
		// Pattern table 0 ($0000-$0FFF)
		if m.latch0 == 0xFD {
			bank = m.chrBank0FD
		} else {
			bank = m.chrBank0FE
		}
		offset = bank*4096 + int(addr&0x0FFF)

		// IMPORTANTE: Actualizar latch DESPUÉS de leer
		// El latch cambia cuando se lee tile $0FD8-$0FDF o $0FE8-$0FEF
		tileAddr := addr & 0x0FF8
		if tileAddr == 0x0FD8 {
			m.latch0 = 0xFD
		} else if tileAddr == 0x0FE8 {
			m.latch0 = 0xFE
		}
	} else {
		// Pattern table 1 ($1000-$1FFF)
		if m.latch1 == 0xFD {
			bank = m.chrBank1FD
		} else {
			bank = m.chrBank1FE
		}
		offset = bank*4096 + int(addr&0x0FFF)

		// Actualizar latch para pattern table 1
		// Usar máscara que preserva bit $1000
		tileAddr := addr & 0x1FF8
		if tileAddr == 0x1FD8 {
			m.latch1 = 0xFD
		} else if tileAddr == 0x1FE8 {
			m.latch1 = 0xFE
		}
	}

	// Bounds check
	chrSize := m.chrBanks * 8192
	if chrSize > 0 {
		offset = offset % chrSize
	}

	return offset
}

func (m *Mapper9) Scanline()           {}
func (m *Mapper9) Tick()               {}
func (m *Mapper9) NotifyA12(high bool) {}
func (m *Mapper9) IRQState() bool      { return false }
func (m *Mapper9) GetMirror() (MirrorMode, bool) {
	return m.mirroring, true
}
