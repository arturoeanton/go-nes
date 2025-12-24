package main

// ==========================================
// MAPPER 4 (MMC3) - Implementación Precisa
// ==========================================
// El MMC3 es uno de los mappers más complejos de la NES.
// Características:
// - 8KB PRG bank switching (4 ventanas)
// - 1KB CHR bank switching (8 ventanas)
// - IRQ por contador de scanlines basado en detección A12
// - Mirroring controlado por software
//
// IMPORTANTE: El contador de scanlines se clock por flancos
// ascendentes en la línea A12 del bus de direcciones de la PPU,
// NO simplemente "por scanline". Esto es crítico para juegos
// como Aladdin, Tiny Toon, y TMNT.

// Mapper4 implementa el chip MMC3 con detección precisa de A12.
type Mapper4 struct {
	prgBanks int // Cantidad de bancos PRG de 16KB
	chrBanks int // Cantidad de bancos CHR de 8KB (en términos de 8KB para el cálculo)

	// Registros de control del MMC3
	bankSelect   byte    // $8000: Selección de banco y modos
	registers    [8]byte // R0-R7: Valores de bancos
	prgBankMode  bool    // Bit 6 de $8000: Modo de PRG
	chrInversion bool    // Bit 7 de $8000: Inversión de CHR
	mirrorMode   MirrorMode

	// IRQ Hardware - Contador de Scanlines
	irqLatch   byte // $C000: Valor de recarga
	irqCounter byte // Contador actual
	irqReload  bool // $C001: Flag de recarga pendiente
	irqEnabled bool // $E001: IRQ habilitado
	irqPending bool // IRQ pendiente para la CPU

	// Detección de Flancos A12
	// El hardware real filtra flancos rápidos.
	// Solo cuenta cuando A12 ha estado bajo por ~16 PPU dots (~3 CPU cycles)
	lastA12High    bool
	a12LowCounter  int  // Cuenta ciclos con A12 bajo
	a12FilterReady bool // true cuando el filtro está listo para detectar flanco

	// Cache de mapeo para acceso rápido
	prgMap [4]int // 4 ventanas de 8KB para PRG
	chrMap [8]int // 8 ventanas de 1KB para CHR

	// PRG/CHR sizes para cálculos
	prgSize int
	chrSize int
}

// NewMapper4 crea una nueva instancia del mapper MMC3.
func NewMapper4(prgBanks, chrBanks int) *Mapper4 {
	m := &Mapper4{
		prgBanks:   prgBanks,
		chrBanks:   chrBanks,
		prgSize:    prgBanks * 16384, // Total PRG en bytes
		chrSize:    chrBanks * 8192,  // Total CHR en bytes (o 8192 si CHR-RAM)
		mirrorMode: MirrorVertical,
	}

	// Si no hay CHR banks (CHR-RAM), usamos 8KB
	if m.chrSize == 0 {
		m.chrSize = 8192
	}

	m.updateBanks()
	return m
}

// Read traduce una dirección de CPU ($8000-$FFFF) a offset en PRG-ROM.
func (m *Mapper4) Read(addr uint16) int {
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

// Write maneja escrituras a los registros del MMC3.
func (m *Mapper4) Write(addr uint16, data byte) {
	switch {
	// $8000-$9FFF: Bank Select y Bank Data
	case addr >= 0x8000 && addr <= 0x9FFF:
		if addr&1 == 0 {
			// $8000, $8002, etc. (pares): Bank Select
			m.bankSelect = data
			m.prgBankMode = (data & 0x40) != 0
			m.chrInversion = (data & 0x80) != 0
		} else {
			// $8001, $8003, etc. (impares): Bank Data
			reg := m.bankSelect & 0x07
			m.registers[reg] = data
		}
		m.updateBanks()

	// $A000-$BFFF: Mirroring y PRG RAM Protect
	case addr >= 0xA000 && addr <= 0xBFFF:
		if addr&1 == 0 {
			// $A000: Mirroring
			if data&1 == 0 {
				m.mirrorMode = MirrorVertical
			} else {
				m.mirrorMode = MirrorHorizontal
			}
		}
		// $A001: PRG RAM protect - no implementado

	// $C000-$DFFF: IRQ Latch y Reload
	case addr >= 0xC000 && addr <= 0xDFFF:
		if addr&1 == 0 {
			// $C000: IRQ Latch
			m.irqLatch = data
		} else {
			// $C001: IRQ Reload
			// Escribir aquí resetea el contador y marca recarga pendiente
			m.irqCounter = 0
			m.irqReload = true
		}

	// $E000-$FFFF: IRQ Disable/Enable
	case addr >= 0xE000:
		if addr&1 == 0 {
			// $E000: IRQ Disable y Acknowledge
			m.irqEnabled = false
			m.irqPending = false
		} else {
			// $E001: IRQ Enable
			m.irqEnabled = true
		}
	}
}

// ReadCHR traduce una dirección de PPU ($0000-$1FFF) a offset en CHR.
func (m *Mapper4) ReadCHR(addr uint16) int {
	chunk := addr / 0x0400 // Bloque de 1KB
	if chunk < 8 {
		return m.chrMap[chunk] + int(addr%0x0400)
	}
	return int(addr)
}

// updateBanks recalcula los mapeos de PRG y CHR según los registros actuales.
func (m *Mapper4) updateBanks() {
	// PRG Banking (ventanas de 8KB)
	numPRG8K := m.prgBanks * 2 // Número de bancos de 8KB
	if numPRG8K == 0 {
		numPRG8K = 2 // Mínimo
	}

	r6 := int(m.registers[6]) % numPRG8K
	r7 := int(m.registers[7]) % numPRG8K
	secondLast := (numPRG8K - 2) % numPRG8K
	last := (numPRG8K - 1) % numPRG8K

	if !m.prgBankMode {
		// Modo 0: $8000 = R6, $A000 = R7, $C000 = -2, $E000 = -1
		m.prgMap[0] = r6 * 8192
		m.prgMap[1] = r7 * 8192
		m.prgMap[2] = secondLast * 8192
		m.prgMap[3] = last * 8192
	} else {
		// Modo 1: $8000 = -2, $A000 = R7, $C000 = R6, $E000 = -1
		m.prgMap[0] = secondLast * 8192
		m.prgMap[1] = r7 * 8192
		m.prgMap[2] = r6 * 8192
		m.prgMap[3] = last * 8192
	}

	// CHR Banking (ventanas de 1KB)
	// Configurar ventanas CHR según registro y modo
	r0 := int(m.registers[0] & 0xFE) // 2KB bank
	r1 := int(m.registers[1] & 0xFE) // 2KB bank
	r2 := int(m.registers[2])        // 1KB bank
	r3 := int(m.registers[3])        // 1KB bank
	r4 := int(m.registers[4])        // 1KB bank
	r5 := int(m.registers[5])        // 1KB bank

	// La inversión CHR ($8000 bit 7) intercambia la región baja con la alta
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

// NotifyA12 es llamado por la PPU cada vez que accede a CHR memory.
// Detecta flancos ascendentes en A12 para clockear el contador de IRQ.
// Este es el corazón de la precisión del MMC3.
func (m *Mapper4) NotifyA12(high bool) {
	if high {
		// A12 está alto
		if m.a12FilterReady && !m.lastA12High {
			// ¡Flanco ascendente válido! Clockear el contador.
			m.clockIRQCounter()
		}
		m.a12LowCounter = 0
		m.a12FilterReady = false
	} else {
		// A12 está bajo
		m.a12LowCounter++
		// Filtro: A12 debe estar bajo por al menos 3 ciclos (aprox 16 PPU dots)
		if m.a12LowCounter >= 3 {
			m.a12FilterReady = true
		}
	}
	m.lastA12High = high
}

// clockIRQCounter es llamado en cada flanco ascendente válido de A12.
func (m *Mapper4) clockIRQCounter() {
	// Si el contador es 0 o hay recarga pendiente, recargar
	if m.irqCounter == 0 || m.irqReload {
		m.irqCounter = m.irqLatch
		m.irqReload = false
	} else {
		m.irqCounter--
	}

	// Si el contador llegó a 0 y las IRQ están habilitadas, disparar
	if m.irqCounter == 0 && m.irqEnabled {
		m.irqPending = true
	}
}

// Scanline es la versión legacy - ahora usamos NotifyA12.
// Se mantiene para compatibilidad pero ya no es el método principal.
func (m *Mapper4) Scanline() {
	// Legacy: solo si no se usa NotifyA12, simular un flanco
	// Esto sirve como fallback si la integración con PPU no funciona
	m.clockIRQCounter()
}

// Tick es llamado por ciclo de CPU (para mappers como FME-7).
// MMC3 no usa esto para IRQ.
func (m *Mapper4) Tick() {}

// IRQState retorna true si hay un IRQ pendiente para la CPU.
func (m *Mapper4) IRQState() bool {
	return m.irqPending
}

// GetMirror retorna el modo de espejo controlado por el mapper.
func (m *Mapper4) GetMirror() (MirrorMode, bool) {
	return m.mirrorMode, true
}
