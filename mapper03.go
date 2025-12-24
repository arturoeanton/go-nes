package main

// ==========================================
// MAPPER 3 (CNROM) - Cybernoid
// ==========================================
// PRG fijo (sin bancos), pero CHR switchable.
type Mapper3 struct {
	prgBanks int
	chrBank  int
}

func NewMapper3(prgBanks int) *Mapper3 {
	return &Mapper3{prgBanks: prgBanks, chrBank: 0}
}
func (m *Mapper3) Read(addr uint16) int {
	if addr >= 0x8000 && addr <= 0xFFFF {
		// PRG completo mapeado (normalmente hasta 32KB)
		return int(addr-0x8000) % (m.prgBanks * 16384)
	}
	return -1
}
func (m *Mapper3) Write(addr uint16, data byte) {
	if addr >= 0x8000 {
		// Escribir cambia el banco de CHR (Gráficos)
		m.chrBank = int(data & 0x03)
	}
}
func (m *Mapper3) ReadCHR(addr uint16) int {
	return m.chrBank*8192 + int(addr)
}
func (m *Mapper3) NotifyA12(high bool)           {}
func (m *Mapper3) Scanline()                     {}
func (m *Mapper3) Tick()                         {}
func (m *Mapper3) IRQState() bool                { return false }
func (m *Mapper3) GetMirror() (MirrorMode, bool) { return 0, false }
