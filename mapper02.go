package main

// ==========================================
// MAPPER 2 (UxROM) - Contra, Castlevania
// ==========================================
// Banco fijo en $C000 (último), banco switchable en $8000.
// Sin bank switching de CHR (usa CHR-RAM habitualmente).
type Mapper2 struct {
	prgBanks   int
	activeBank int
}

func NewMapper2(prgBanks int) *Mapper2 {
	return &Mapper2{prgBanks: prgBanks, activeBank: 0}
}

func (m *Mapper2) Read(addr uint16) int {
	if addr >= 0x8000 && addr <= 0xBFFF {
		// Banco Switchable
		return m.activeBank*16384 + int(addr-0x8000)
	}
	if addr >= 0xC000 && addr <= 0xFFFF {
		// Banco Fijo (Último)
		return (m.prgBanks-1)*16384 + int(addr-0xC000)
	}
	return -1
}

func (m *Mapper2) Write(addr uint16, data byte) {
	if addr >= 0x8000 {
		// Escribir en la ROM cambia el banco activo para $8000
		m.activeBank = int(data) % m.prgBanks
	}
}

func (m *Mapper2) ReadCHR(addr uint16) int       { return int(addr) }
func (m *Mapper2) NotifyA12(high bool)           {}
func (m *Mapper2) Scanline()                     {}
func (m *Mapper2) Tick()                         {}
func (m *Mapper2) IRQState() bool                { return false }
func (m *Mapper2) GetMirror() (MirrorMode, bool) { return 0, false }
