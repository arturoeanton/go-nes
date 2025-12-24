package main

import (
	"testing"
)

// =============================================================================
// MAPPER 4 (MMC3) - SMART A12 EDGE DETECTION TESTS
// =============================================================================
// Tests para desarrollo incremental de A12 edge detection preciso.
//
// OBJETIVO: Hacer que Aladdin y Shinobi funcionen sin romper los juegos existentes.
//
// El hardware MMC3 real detecta flancos ascendentes en A12 del bus PPU con
// un filtro de ~16 PPU dots (~3 CPU cycles) para evitar flancos espurios.
//
// ESTADO ACTUAL:
// - NotifyA12() existe pero no se usa (causaba problemas)
// - Usamos Scanline() hook simplificado
//
// PLAN DE IMPLEMENTACIÓN:
// 1. [ ] Agregar contador de ciclos PPU al mapper
// 2. [ ] Implementar filtrado de ~16 PPU dots en NotifyA12
// 3. [ ] Llamar NotifyA12 solo en transiciones reales (no cada fetch)
// 4. [ ] Verificar que no rompe juegos existentes
// 5. [ ] Verificar que Aladdin/Shinobi funcionan

// =============================================================================
// PHASE 1: Tests básicos de NotifyA12 (actualmente no usada)
// =============================================================================

// TestA12EdgeDetectionBasic verifica que NotifyA12 existe y no crashea
func TestA12EdgeDetectionBasic(t *testing.T) {
	m := NewMapper4(16, 32)

	// NotifyA12 debería existir y no crashear
	m.NotifyA12(false)
	m.NotifyA12(true)
	m.NotifyA12(false)

	// No debería cambiar estado del IRQ sin estar habilitado
	if m.IRQState() {
		t.Error("IRQ should not fire without being enabled")
	}
}

// TestA12EdgeDetectionWithIRQ verifica flancos con IRQ habilitado
func TestA12EdgeDetectionWithIRQ(t *testing.T) {
	m := NewMapper4(16, 32)

	// Configurar IRQ: latch=5, enable
	m.Write(0xC000, 5) // Latch = 5
	m.Write(0xC001, 0) // Reload flag
	m.Write(0xE001, 0) // Enable IRQ

	// NOTA: Esta prueba documenta el comportamiento esperado
	// pero actualmente NotifyA12 no está conectada al IRQ counter
	// porque causaba problemas cuando se llamaba en cada ppuFetch

	// TODO: Cuando se implemente correctamente:
	// - Simular flancos A12 filtrados (uno por scanline aprox)
	// - Verificar que el counter decrementa correctamente
	// - Verificar que dispara IRQ después de N flancos
}

// =============================================================================
// PHASE 2: Tests de filtrado de flancos (16 PPU dots)
// =============================================================================

// TestA12FilterIgnoresRapidTransitions verifica que el filtro funciona
func TestA12FilterIgnoresRapidTransitions(t *testing.T) {
	m := NewMapper4(16, 32)

	// El filtro debería ignorar transiciones rápidas
	// A12 debe estar bajo por ~16 PPU dots antes de que un flanco cuenta

	// El filtro ya está implementado en mapper04.go

	m.Write(0xC000, 5) // Latch = 5
	m.Write(0xC001, 0) // Reload
	m.Write(0xE001, 0) // Enable

	// Simular muchas transiciones rápidas (como en PPU fetch)
	for i := 0; i < 100; i++ {
		m.NotifyA12(true)
		m.NotifyA12(false)
	}

	// El IRQ NO debería disparar porque las transiciones son muy rápidas
	if m.IRQState() {
		t.Error("IRQ should not fire with rapid A12 transitions (filter should prevent)")
	}
}

// TestA12FilterAllowsSlowTransitions verifica flancos válidos
func TestA12FilterAllowsSlowTransitions(t *testing.T) {
	m := NewMapper4(16, 32)

	m.Write(0xC000, 2) // Latch = 2 (dispara después de 2 flancos válidos)
	m.Write(0xC001, 0) // Reload
	m.Write(0xE001, 0) // Enable

	// Simular flancos con suficiente tiempo entre ellos
	// (simulando conteo de ciclos PPU)
	for scanline := 0; scanline < 5; scanline++ {
		// A12 bajo por tiempo suficiente
		for cycle := 0; cycle < 20; cycle++ {
			m.NotifyA12(false)
		}
		// Flanco ascendente
		m.NotifyA12(true)
	}

	// Después de ~3 flancos válidos debería disparar
	if !m.IRQState() {
		t.Error("IRQ should fire after valid A12 transitions")
	}
}

// =============================================================================
// PHASE 3: Tests de integración PPU-Mapper
// =============================================================================

// TestA12CalledOnCHRAccess documenta cuándo debería llamarse NotifyA12
func TestA12CalledOnCHRAccess(t *testing.T) {
	// DOCUMENTACIÓN del comportamiento esperado:
	//
	// En el hardware real, A12 cambia cuando:
	// - PPU accede a pattern table 0 ($0000-$0FFF) -> A12 = 0
	// - PPU accede a pattern table 1 ($1000-$1FFF) -> A12 = 1
	//
	// El MMC3 detecta cuando A12 va de 0 a 1 (rising edge)
	// pero FILTRADO para ignorar cambios muy rápidos.
	//
	// Durante un scanline típico:
	// - Cycles 0-255: Background tiles (A12 depende de PPUCTL)
	// - Cycles 257-320: Sprite tiles (A12 = usualmente 1)
	//
	// El flanco típico ocurre entre cycle 257-260 cuando
	// la PPU cambia de BG fetch a sprite fetch.

	t.Log("This test documents expected A12 behavior - no actual assertions")
}

// =============================================================================
// PHASE 4: Tests de regresión (no romper juegos existentes)
// =============================================================================

// TestRegressionSMB3Style verifica comportamiento tipo SMB3
func TestRegressionSMB3Style(t *testing.T) {
	m := NewMapper4(16, 32)

	// SMB3 usa IRQ para scroll split-screen
	// Típicamente: latch = N, habilita IRQ, espera N scanlines

	m.Write(0xC000, 10) // Disparar después de 10 scanlines
	m.Write(0xC001, 0)  // Reload
	m.Write(0xE001, 0)  // Enable

	// Simular usando Scanline() (el método actual que funciona)
	for i := 0; i < 11; i++ {
		m.Scanline()
	}

	// IRQ debería disparar
	if !m.IRQState() {
		t.Error("Regression: SMB3-style IRQ should still work with Scanline()")
	}
}

// TestRegressionTinyToonStyle verifica comportamiento tipo Tiny Toon
func TestRegressionTinyToonStyle(t *testing.T) {
	m := NewMapper4(16, 32)

	// Tiny Toon usa IRQ para HUD en la parte inferior
	// Similar a SMB3 pero con diferente timing

	m.Write(0xC000, 200) // Cerca del final de la pantalla
	m.Write(0xC001, 0)
	m.Write(0xE001, 0)

	// Simular frame completo
	for i := 0; i < 240; i++ {
		m.Scanline()
	}

	if !m.IRQState() {
		t.Error("Regression: Tiny Toon-style IRQ should work")
	}
}

// =============================================================================
// PHASE 5: Tests específicos para Aladdin/Shinobi
// =============================================================================

// TestAladdinIRQPattern documenta el patrón de IRQ que usa Aladdin
func TestAladdinIRQPattern(t *testing.T) {
	// Test para documentar el patrón esperado de Aladdin

	// NOTA: Este test se habilitará cuando implementemos A12 correctamente
	//
	// Aladdin parece requerir:
	// 1. Detección precisa de flancos A12 durante la inicialización
	// 2. El juego probablemente deshabilita rendering durante inicio
	//    pero espera que el IRQ cuente basado en accesos CHR
	// 3. Con nuestro Scanline() hook, el IRQ no cuenta si rendering está off
}

// TestShinobiIRQPattern documenta el patrón de IRQ que usa Shinobi
func TestShinobiIRQPattern(t *testing.T) {
	// Test para documentar el patrón esperado de Shinobi

	// NOTA: Similar a Aladdin
	// Shinobi crashea con PC oscilando entre $0000-$0001
	// indicando que algo en la inicialización falla
}
