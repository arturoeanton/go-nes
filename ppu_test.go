package main

import "testing"

// ==========================================
// TEST UNITARIOS DE PPU
// ==========================================
// Pruebas para la Unidad de Procesamiento de Gráficos (PPU) 2C02.

func setupTestPPU() *PPU {
	prg := make([]byte, 32*1024)
	chr := make([]byte, 8*1024)
	cart := &Cartridge{
		PRG:    prg,
		CHR:    chr,
		Mapper: NewMapper0(2),
		Mirror: MirrorHorizontal,
	}
	return NewPPU(cart)
}

// TestPPU_Registers_WriteRead verifica la escritura básica en registros de control.
func TestPPU_Registers_WriteRead(t *testing.T) {
	ppu := setupTestPPU()

	// Prueba $2000 (PPUCTRL)
	// Escribir 0x80 debería activar NMI Enable (Bit 7).
	ppu.Write(0x2000, 0x80)
	if ppu.Ctrl != 0x80 {
		t.Errorf("PPUCTRL Write falló: Esperado 0x80, obtenido 0x%02X", ppu.Ctrl)
	}

	// Prueba $2001 (PPUMASK)
	// Escribir 0x1E (Show BG, Sprites, etc)
	ppu.Write(0x2001, 0x1E)
	if ppu.Mask != 0x1E {
		t.Errorf("PPUMASK Write falló: Esperado 0x1E, obtenido 0x%02X", ppu.Mask)
	}
}

// TestPPU_VRAM_Access prueba la escritura y lectura de VRAM a través de $2006/$2007.
func TestPPU_VRAM_Access(t *testing.T) {
	ppu := setupTestPPU()

	// 1. Setear dirección VRAM $2010
	// Escritura 1: Byte alto (0x20)
	ppu.Write(0x2006, 0x20)
	// Escritura 2: Byte bajo (0x10)
	ppu.Write(0x2006, 0x10)

	// Verificar registro interno V (VRAM Address)
	if ppu.VramAddr != 0x2010 {
		t.Errorf("PPU Address Latch falló: Esperado 0x2010, obtenido 0x%04X", ppu.VramAddr)
	}

	// 2. Escribir dato 0xAB en VRAM
	ppu.Write(0x2007, 0xAB)

	// Verificar que el puntero incrementó (por defecto +1)
	if ppu.VramAddr != 0x2011 {
		t.Errorf("PPU Auto-Increment falló: Esperado 0x2011, obtenido 0x%04X", ppu.VramAddr)
	}

	// 3. Leer dato (requiere setear dirección de nuevo o confiar en el puntero)
	// Volvamos a apuntar a $2010 para leer.
	ppu.Write(0x2006, 0x20)
	ppu.Write(0x2006, 0x10)

	// Lectura dummy (buffer)
	// La PPU tiene un buffer de lectura retrasada para VRAM ($0000-$3EFF).
	dummy := ppu.Read(0x2007)
	_ = dummy

	// Segunda lectura: Dato real
	val := ppu.Read(0x2007)
	if val != 0xAB {
		t.Errorf("PPU VRAM Read falló: Esperado 0xAB, obtenido 0x%02X", val)
	}
}

// TestPPU_Status_Reset verifica que leer $2002 resetee el latch de direcciones y el flag VBlank.
func TestPPU_Status_Reset(t *testing.T) {
	ppu := setupTestPPU()

	// Simular NMI Ocurrido
	ppu.NmiOccurred = true
	ppu.AddrLatch = 1 // Latch sucio (mitad de escritura)

	// Leer Status
	status := ppu.Read(0x2002)

	// Verificar bit 7 (VBlank)
	if status&0x80 == 0 {
		t.Error("PPUSTATUS Read no reportó NmiOccurred (Bit 7)")
	}

	// Verificar efectos secundarios
	if ppu.NmiOccurred {
		t.Error("PPUSTATUS Read no limpió NmiOccurred")
	}
	if ppu.AddrLatch != 0 {
		t.Error("PPUSTATUS Read no reseteó Address Latch")
	}
}
