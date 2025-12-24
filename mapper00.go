package main

// ==========================================
// MAPPER 0 (NROM) - Super Mario Bros, Donkey Kong
// ==========================================
// El más simple. Sin registros, sin bank switching dinámico.
type Mapper0 struct {
	prgBanks int // Cantidad de bancos de 16KB
}

func NewMapper0(prgBanks int) *Mapper0 {
	return &Mapper0{prgBanks: prgBanks}
}

func (m *Mapper0) Read(addr uint16) int {
	// Si el juego tiene 1 banco (16KB), se repite en $8000 y $C000.
	// Si tiene 2 bancos (32KB), se mapea linealmente.
	if addr >= 0x8000 && addr <= 0xFFFF {
		return int(addr-0x8000) % (m.prgBanks * 16384)
	}
	return -1
}

// Mapper 0 no tiene registros de escritura.
func (m *Mapper0) Write(addr uint16, data byte) {}

// Mapper 0 mapea CHR directament (no hay bancos).
func (m *Mapper0) ReadCHR(addr uint16) int { return int(addr) }

func (m *Mapper0) NotifyA12(high bool) {}
func (m *Mapper0) Scanline()           {}
func (m *Mapper0) Tick()               {}
func (m *Mapper0) IRQState() bool      { return false }
func (m *Mapper0) GetMirror() (MirrorMode, bool) {
	// Mapper 0 no controla mirroring (se define por soldadura en el cartucho).
	return 0, false
}
