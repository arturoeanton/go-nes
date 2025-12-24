package main

// ==========================================
// MAPPER 7 (AxROM) - Battletoads
// ==========================================
// Banco PRG de 32KB switchable.
// Mirroring Single Screen switchable por software.
type Mapper7 struct {
	prgBanks   int
	prgBank    int
	mirrorMode MirrorMode
}

func NewMapper7(prgBanks int) *Mapper7 {
	return &Mapper7{
		prgBanks:   prgBanks,
		prgBank:    0,
		mirrorMode: MirrorSingle0,
	}
}

func (m *Mapper7) Read(addr uint16) int {
	if addr >= 0x8000 && addr <= 0xFFFF {
		// 32KB Bank Switching
		// Offset del banco actual + offset dentro del banco
		return m.prgBank*32768 + int(addr-0x8000)
	}
	return -1
}

func (m *Mapper7) Write(addr uint16, data byte) {
	if addr >= 0x8000 {
		// AxROM usa los bits bajos para seleccionar el banco PRG (32KB)
		// y el bit 4 para seleccionar el Mirroring (1-Screen).

		// PRG Select (Bits 0-2 para 8 bancos max de 32KB = 256KB)
		// Usamos modulo por seguridad si el juego es mas chico.
		num32KB := m.prgBanks / 2
		if num32KB > 0 {
			m.prgBank = int(data&0x07) % num32KB
		}

		// Mirroring (Bit 4)
		// 0: Single Screen Low, 1: Single Screen High
		if data&0x10 == 0 {
			m.mirrorMode = MirrorSingle0
		} else {
			m.mirrorMode = MirrorSingle1
		}
	}
}

func (m *Mapper7) ReadCHR(addr uint16) int {
	// Mapper 7 usa CHR-RAM usualmente (Battletoads), mapeo directo.
	return int(addr)
}

func (m *Mapper7) Scanline()      {}
func (m *Mapper7) IRQState() bool { return false }
func (m *Mapper7) Tick()          {} // Implementación vacía para interfaz
func (m *Mapper7) GetMirror() (MirrorMode, bool) {
	return m.mirrorMode, true
}
