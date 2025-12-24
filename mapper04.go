package main

// ==========================================
// MAPPER 4 (MMC3) - Super Mario Bros 3, Kirby
// ==========================================
// El más avanzado de los clásicos.
// Soporta IRQ por Scanline (para efectos de pantalla partida / split screen).
// Granularidad fina de bancos (PRG 8KB, CHR 1KB).
type Mapper4 struct {
	prgBanks int
	chrBanks int // Bancos de 1KB

	// Registros internos MMC3
	targetRegister byte    // Qué registro vamos a actualizar (0-7)
	registers      [8]byte // Los 8 registros de configuración de bancos
	prgBankMode    bool    // Modo de PRG ($8000 vs $C000 fijo)
	chrInversion   bool    // Invierte bancos de CHR altos/bajos
	mirrorMode     MirrorMode

	// Hardware IRQ (Contador de Scanlines)
	irqEnabled bool
	irqLatch   byte
	irqCounter byte
	irqReload  bool
	irqActive  bool // Flag de IRQ pendiente para la CPU

	// Mapeo Rápido (Cache)
	prgMap [4]int // 4 ventanas de 8KB
	chrMap [8]int // 8 ventanas de 1KB
}

func NewMapper4(prgBanks, chrBanks int) *Mapper4 {
	m := &Mapper4{
		prgBanks:   prgBanks,
		chrBanks:   chrBanks,
		mirrorMode: MirrorVertical,
	}
	m.updateBanks()
	return m
}

func (m *Mapper4) Read(addr uint16) int {
	// PRG dividido en 4 ventanas de 8KB
	switch {
	case addr >= 0x8000 && addr <= 0x9FFF:
		return m.prgMap[0] + int(addr-0x8000)
	case addr >= 0xA000 && addr <= 0xBFFF:
		return m.prgMap[1] + int(addr-0xA000)
	case addr >= 0xC000 && addr <= 0xDFFF:
		return m.prgMap[2] + int(addr-0xC000)
	case addr >= 0xE000 && addr <= 0xFFFF:
		return m.prgMap[3] + int(addr-0xE000)
	}
	return -1
}

func (m *Mapper4) Write(addr uint16, data byte) {
	switch {
	case addr >= 0x8000 && addr <= 0x9FFF:
		if addr%2 == 0 {
			// $8000 (Par): Bank Select
			m.targetRegister = data & 0x07 // Qué registro actualizar (0-7)
			m.prgBankMode = (data & 0x40) != 0
			m.chrInversion = (data & 0x80) != 0
			m.updateBanks()
		} else {
			// $8001 (Impar): Bank Data
			// Actualiza el registro seleccionado por $8000
			m.registers[m.targetRegister] = data
			m.updateBanks()
		}
	case addr >= 0xA000 && addr <= 0xBFFF:
		if addr%2 == 0 {
			// $A000: Mirroring
			if data&1 == 0 {
				m.mirrorMode = MirrorVertical
			} else {
				m.mirrorMode = MirrorHorizontal
			}
		} else {
			// $A001: PRG RAM Protect (No implementado en MVP)
		}
	case addr >= 0xC000 && addr <= 0xDFFF:
		if addr%2 == 0 {
			// $C000: IRQ Latch (Valor de recarga)
			m.irqLatch = data
		} else {
			// $C001: IRQ Reload (Fuerza recarga en próximo scanline)
			m.irqReload = true
		}
	case addr >= 0xE000 && addr <= 0xFFFF:
		if addr%2 == 0 {
			// $E000: Disable IRQ y Acknowledge
			m.irqEnabled = false
			m.irqActive = false
		} else {
			// $E001: Enable IRQ
			m.irqEnabled = true
		}
	}
}

func (m *Mapper4) ReadCHR(addr uint16) int {
	chunk := addr / 0x0400 // Bloque de 1KB
	if chunk < 8 {
		return m.chrMap[chunk] + int(addr%0x0400)
	}
	return int(addr)
}

func (m *Mapper4) updateBanks() {
	// Configurar ventanas PRG (8KB) según registros y modo
	total8KB := m.prgBanks * 2
	r6 := int(m.registers[6])
	r7 := int(m.registers[7])

	if !m.prgBankMode {
		// Modo normal: $8000 swappable, $C000 fijo (-2)
		m.prgMap[0] = r6 * 8192
		m.prgMap[1] = r7 * 8192
		m.prgMap[2] = (total8KB - 2) * 8192
		m.prgMap[3] = (total8KB - 1) * 8192
	} else {
		// Modo swap: $C000 swappable, $8000 fijo (-2)
		m.prgMap[0] = (total8KB - 2) * 8192
		m.prgMap[1] = r7 * 8192
		m.prgMap[2] = r6 * 8192
		m.prgMap[3] = (total8KB - 1) * 8192
	}

	// Configurar ventanas CHR (1KB)
	r0 := int(m.registers[0] & 0xFE) // 2KB
	r1 := int(m.registers[1] & 0xFE) // 2KB
	r2 := int(m.registers[2])        // 1KB
	r3 := int(m.registers[3])        // 1KB
	r4 := int(m.registers[4])        // 1KB
	r5 := int(m.registers[5])        // 1KB

	// La inversión CHR ($8000 bit 7) intercambia la región baja (0-1K) con la alta
	if !m.chrInversion {
		m.chrMap[0] = r0 * 1024
		m.chrMap[1] = (r0 + 1) * 1024
		m.chrMap[2] = r1 * 1024
		m.chrMap[3] = (r1 + 1) * 1024
		m.chrMap[4] = r2 * 1024
		m.chrMap[5] = r3 * 1024
		m.chrMap[6] = r4 * 1024
		m.chrMap[7] = r5 * 1024
	} else {
		m.chrMap[0] = r2 * 1024
		m.chrMap[1] = r3 * 1024
		m.chrMap[2] = r4 * 1024
		m.chrMap[3] = r5 * 1024
		m.chrMap[4] = r0 * 1024
		m.chrMap[5] = (r0 + 1) * 1024
		m.chrMap[6] = r1 * 1024
		m.chrMap[7] = (r1 + 1) * 1024
	}
}

// Scanline se llama por la PPU cuando renderiza una línea.
// El MMC3 detecta flancos en A12 para contar líneas.
func (m *Mapper4) Tick() {}

func (m *Mapper4) Scanline() {
	if m.irqCounter == 0 || m.irqReload {
		m.irqCounter = m.irqLatch
		m.irqReload = false
	} else {
		m.irqCounter--
	}
	// Si el contador llega a cero y las IRQ están activadas, ¡Disparar!
	if m.irqCounter == 0 && m.irqEnabled {
		m.irqActive = true
	}
}

func (m *Mapper4) IRQState() bool {
	return m.irqActive
}

func (m *Mapper4) GetMirror() (MirrorMode, bool) {
	return m.mirrorMode, true
}
