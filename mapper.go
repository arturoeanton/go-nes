package main

// ==========================================
// INTERFAZ MAPPER
// ==========================================
// El Mapper es el chip de soporte dentro del cartucho.
// Permite a la NES ver más memoria de la que puede direccionar (Bank Switching).
// Los juegos simples (NROM) no hacen nada. Los complejos (MMC3) tienen contadores de IRQ,
// registros de desplazamiento, y múltiples ventanas de bancos.
type Mapper interface {
	Read(addr uint16) int          // Traduce dirección CPU ($8000+) a offset ROM físico.
	Write(addr uint16, data byte)  // Escribe registros de control del Mapper.
	ReadCHR(addr uint16) int       // Traduce dirección PPU ($0000-$1FFF) a offset CHR físico.
	Scanline()                     // Hook que la PPU llama cada línea (para IRQ del MMC3).
	IRQState() bool                // Devuelve true si el Mapper pide interrumpir al CPU.
	GetMirror() (MirrorMode, bool) // Devuelve el modo de espejo si el Mapper lo controla dinámicamente.
}

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

func (m *Mapper0) Scanline()      {}
func (m *Mapper0) IRQState() bool { return false }
func (m *Mapper0) GetMirror() (MirrorMode, bool) {
	// Mapper 0 no controla mirroring (se define por soldadura en el cartucho).
	return 0, false
}

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
func (m *Mapper2) Scanline()                     {}
func (m *Mapper2) IRQState() bool                { return false }
func (m *Mapper2) GetMirror() (MirrorMode, bool) { return 0, false }

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
func (m *Mapper3) Scanline()                     {}
func (m *Mapper3) IRQState() bool                { return false }
func (m *Mapper3) GetMirror() (MirrorMode, bool) { return 0, false }

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
func (m *Mapper7) GetMirror() (MirrorMode, bool) {
	return m.mirrorMode, true
}
