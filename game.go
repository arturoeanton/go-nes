package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Game es la estructura principal que Ebiten usará para correr el bucle del juego.
// Contiene referencias a los tres componentes principales: CPU, Bus y PPU.
type Game struct {
	CPU *CPU
	Bus *Bus
	// cpuCyclesDebt se usa para mantener la sincronización entre CPU y PPU.
	// El PPU va 3 veces más rápido que el CPU. Acumulamos "deuda" de ciclos
	// para que el CPU ejecute instrucciones hasta alcanzar al PPU.
	cpuCyclesDebt int
}

// Update es el corazón del emulador. Ebiten llama a esta función 60 veces por segundo (60Hz).
// Aquí simulamos un frame completo de la NES.
func (g *Game) Update() error {
	// ciclos por frame = 262 scanlines * 341 ciclos por scanline = 89342
	// En hardware real esto es continuo, aquí lo simulamos por bloques de frame.
	const ppuCyclesPerFrame = 89342

	for i := 0; i < ppuCyclesPerFrame; i++ {
		// 1. Reloj Maestro: La PPU dicta el tiempo
		// Tick() avanza un ciclo de PPU (dibujo de píxeles, etc).
		// Retorna true si se activó una NMI (VBlank).
		nmiTriggered := g.Bus.PPU.Tick()

		if nmiTriggered {
			// Si hubo NMI (Non-Maskable Interrupt), avisamos al CPU.
			// Esto ocurre cuando la PPU termina de dibujar el frame (VBlank).
			g.CPU.NMI()

			// Una interrupción toma 7 ciclos de CPU, restamos su "costo".
			g.cpuCyclesDebt -= 7
		}

		// 2. Reloj de CPU (Sincronía 3:1)
		// En la NES (NTSC), por cada 3 ciclos de PPU, pasa 1 ciclo de CPU.
		if i%3 == 0 {
			g.cpuCyclesDebt++
		}

		// 3. Ejecutar CPU mientras tenga ciclos disponibles ("presupuesto")
		// Si la deuda es positiva, significa que la PPU avanzó lo suficiente
		// como para permitirle al CPU ejecutar una o más instrucciones.
		for g.cpuCyclesDebt > 0 {
			// Step() ejecuta una instrucción completa (ej: LDA #$00)
			// y devuelve cuántos ciclos reales tomó (ej: 2).
			used := g.CPU.Step()
			g.cpuCyclesDebt -= used
		}
	}

	// Limpieza de seguridad:
	// Si el CPU se queda muy atrás (ej: bucle infinito o bug), reseteamos la deuda
	// para evitar que en el siguiente frame intente ejecutar millones de ciclos de golpe,
	// lo que congelaría el emulador.
	if g.cpuCyclesDebt < -100 {
		g.cpuCyclesDebt = 0
	}

	return nil
}

// Draw se llama después de Update. Copia el buffer de píxeles generado por la PPU
// a la pantalla de Ebiten para que el usuario lo vea.
func (g *Game) Draw(screen *ebiten.Image) {
	g.Bus.PPU.Draw(screen)
}

// Layout define el tamaño de la pantalla lógica.
// Ebiten escalará esto automáticamente al tamaño de la ventana.
func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}
