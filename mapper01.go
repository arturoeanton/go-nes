package main

// ==========================================
// MAPPER 1 (MMC1) - Zelda, Metroid, Prince of Persia (MMC1 real)
// ==========================================
// El primer ASIC de Nintendo. Usa "Serial Loading" para ahorrar pines.
// Permite cambiar bancos de PRG, CHR y el Mirroring.
type Mapper1 struct {
	prgBanks int
	chrBanks int

	shiftReg byte // Registro de desplazamiento para recibir datos bit a bit.
	writeCnt int  // Contador de escrituras (se ejecuta al 5º bit).

	control  byte // REG 0 ($8000): Configuración general
	chrBank0 byte // REG 1 ($A000): Banco CHR bajo
	chrBank1 byte // REG 2 ($C000): Banco CHR alto
	prgBank  byte // REG 3 ($E000): Banco PRG

	// Cache de offsets calculados para velocidad
	prgOffsets [2]int // Offset para ventana $8000 y $C000
	chrOffsets [2]int // Offset para ventana $0000 y $1000
}

func NewMapper1(prgBanks, chrBanks int) *Mapper1 {
	m := &Mapper1{
		prgBanks: prgBanks,
		chrBanks: chrBanks,
		shiftReg: 0x10, // Inicializado con bit marcador
		control:  0x0E, // Por defecto: PRG Fix Last ($C000 fijo), Mirror Vertical (0x0E = 0000 1110)
	}
	m.updateOffsets()
	return m
}

func (m *Mapper1) Read(addr uint16) int {
	if addr >= 0x8000 && addr <= 0xFFFF {
		// Determinar si estamos en banco bajo (0) o alto (1)
		bank := (addr / 0x4000) & 1
		offset := int(addr % 0x4000)
		return m.prgOffsets[bank] + offset
	}
	return -1
}

// Write en MMC1 es especial: Escritura Serial.
// Escribir cualquier valor con bit 7 alto (0x80) resetea el mapper.
// Si no, se toma el bit 0 del dato y se mete al shift register.
// Después de 5 escrituras, se actualiza el registro interno correspondiente a la dirección.
func (m *Mapper1) Write(addr uint16, data byte) {
	if addr < 0x8000 {
		return
	}

	// Reset
	if data&0x80 != 0 {
		m.shiftReg = 0x10
		m.writeCnt = 0
		m.control |= 0x0C // Volver a modo seguro (Fix Last Bank)
		m.updateOffsets()
		return
	}

	// Carga Serial
	m.shiftReg >>= 1
	if data&1 != 0 {
		m.shiftReg |= 0x10 // Poner bit en posición 4
	} else {
		m.shiftReg &= 0x0F
	}
	m.writeCnt++

	// Al quinto bit, aplicar cambios
	if m.writeCnt == 5 {
		val := m.shiftReg
		// Resetear para próxima carga
		m.shiftReg = 0x10
		m.writeCnt = 0

		// La dirección decide qué registro interno actualizar
		switch {
		case addr >= 0x8000 && addr <= 0x9FFF: // Register 0: Control
			m.control = val
		case addr >= 0xA000 && addr <= 0xBFFF: // Register 1: CHR Bank 0
			m.chrBank0 = val
		case addr >= 0xC000 && addr <= 0xDFFF: // Register 2: CHR Bank 1
			m.chrBank1 = val
		case addr >= 0xE000 && addr <= 0xFFFF: // Register 3: PRG Bank
			m.prgBank = val
		}
		m.updateOffsets() // Recalcular punteros de memoria
	}
}

// updateOffsets traduce los valores de los registros a offsets lineales físicos
func (m *Mapper1) updateOffsets() {
	// --- PRG ROM ---
	// Modos (Control bits 2,3):
	// 0, 1: Modo 32KB (banco contiguo)
	// 2:    Fija Primer Banco en $8000, $C000 switchable
	// 3:    Fija Último Banco en $C000, $8000 switchable (Lo más común)
	prgMode := (m.control >> 2) & 0x03
	prgBank := int(m.prgBank & 0x0F)
	total16KBPB := m.prgBanks

	switch prgMode {
	case 0, 1: // 32KB
		idx := prgBank & 0xFE // Ignorar bit bajo para alinear a 32K
		m.prgOffsets[0] = (idx % total16KBPB) * 16384
		m.prgOffsets[1] = ((idx + 1) % total16KBPB) * 16384
	case 2: // Fix First
		m.prgOffsets[0] = 0 // Primer banco físico
		m.prgOffsets[1] = (prgBank % total16KBPB) * 16384
	case 3: // Fix Last
		m.prgOffsets[0] = (prgBank % total16KBPB) * 16384
		m.prgOffsets[1] = (total16KBPB - 1) * 16384 // Último banco físico
	}

	// --- CHR ROM/RAM ---
	// Modo (Control bit 4):
	// 0: 8KB (Un solo banco gigante)
	// 1: 4KB (Dos bancos separados $0000 y $1000)
	chrMode := (m.control >> 4) & 1
	total4KBCHRB := m.chrBanks * 2
	if total4KBCHRB == 0 {
		total4KBCHRB = 2 // Si es RAM (0 bancos ROM), asumimos 8KB RAM (2x4KB)
	}

	if chrMode == 0 { // 8KB Mode
		idx := int(m.chrBank0 & 0xFE)
		m.chrOffsets[0] = (idx % total4KBCHRB) * 4096
		m.chrOffsets[1] = ((idx + 1) % total4KBCHRB) * 4096
	} else { // 4KB Mode
		m.chrOffsets[0] = (int(m.chrBank0) % total4KBCHRB) * 4096
		m.chrOffsets[1] = (int(m.chrBank1) % total4KBCHRB) * 4096
	}
}

func (m *Mapper1) ReadCHR(addr uint16) int {
	bank := (addr / 0x1000) & 1 // 0 o 1
	offset := int(addr % 0x1000)
	return m.chrOffsets[bank] + offset
}

func (m *Mapper1) Scanline()      {}
func (m *Mapper1) Tick()          {}
func (m *Mapper1) IRQState() bool { return false }

// GetMirror devuelve el mirroring dinámico configurado en el registro de Control.
func (m *Mapper1) GetMirror() (MirrorMode, bool) {
	// Bits 0-1 de control:
	// 0: One Screen Lower
	// 1: One Screen Upper
	// 2: Vertical
	// 3: Horizontal
	switch m.control & 0x03 {
	case 0:
		return MirrorSingle0, true
	case 1:
		return MirrorSingle1, true
	case 2:
		return MirrorVertical, true
	case 3:
		return MirrorHorizontal, true
	}
	return MirrorHorizontal, false
}
