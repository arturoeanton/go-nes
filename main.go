package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// Constantes globales de resolución
const (
	ScreenWidth  = 256
	ScreenHeight = 240
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

		// Chequear IRQ (Interrupciones generadas por el Mapper o APU)
		// Mappers avanzados como MMC3 (Mapper 4) usan esto para efectos de pantalla partida.
		if g.Bus.PPU.Cart.Mapper.IRQState() {
			g.CPU.IRQ()
			// Nota: En una implementación perfecta, verificaríamos si el CPU realmente
			// aceptó la IRQ (flag I deshabilitado) antes de restar ciclos.
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

// main es el punto de entrada del programa.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run . <archivo.nes>")
		return
	}

	// Leer el archivo ROM completo a memoria
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Validar "Magic Number" del formato iNES
	if string(data[:4]) != "NES\x1a" {
		log.Fatal("Archivo inválido: No es una ROM de NES (iNES header)")
	}

	// 1. Parsear Header iNES (los primeros 16 bytes)
	prgBanks := int(data[4]) // Cantidad de bancos de programa (16KB c/u)
	chrBanks := int(data[5]) // Cantidad de bancos de gráficos (8KB c/u)
	flags6 := data[6]        // Flags de control 1 (Mirroring, Mapper bajo)
	flags7 := data[7]        // Flags de control 2 (Mapper alto)

	// Determinar Mirroring (Horizontal o Vertical)
	// Esto define cómo se repite la memoria de video.
	mirror := MirrorHorizontal
	if flags6&0x01 != 0 {
		mirror = MirrorVertical
	}

	// Determinar ID del Mapper
	// El Mapper es el chip dentro del cartucho que maneja bancos de memoria complejos.
	// flags6 tiene los 4 bits bajos, flags7 los 4 bits altos.
	mapperID := (flags7 & 0xF0) | (flags6 >> 4)

	// 2. Calcular dónde empiezan y terminan los datos en el archivo
	prgStart := 16
	prgEnd := prgStart + (prgBanks * 16384)
	chrEnd := prgEnd + (chrBanks * 8192)

	// 3. Crear el objeto Cartucho virtual
	cart := &Cartridge{
		PRG:    data[prgStart:prgEnd], // Código del juego
		Mirror: mirror,                // Configuración de video
	}

	// Inicializar la lógica del Mapper específico según el ID
	switch mapperID {
	case 0:
		// Mapper 0 (NROM): El más simple. Mario Bros, Donkey Kong.
		// Sin bank switching real, solo mapeo directo.
		cart.Mapper = NewMapper0(prgBanks)
	case 1:
		// Mapper 1 (MMC1): Configurable, carga serial.
		// Zelda, Metroid, Prince of Persia (hackeado o real).
		cart.Mapper = NewMapper1(prgBanks, chrBanks)
		log.Printf("Mapper 1 (MMC1) detectado para %s", os.Args[1])
	case 2:
		// Mapper 2 (UxROM): Bank switching simple de PRG.
		// Contra, Castlevania.
		cart.Mapper = NewMapper2(prgBanks)
		log.Printf("Mapper 2 (UxROM) detectado para %s", os.Args[1])
	case 3:
		// Mapper 3 (CNROM): Bank switching de CHR (gráficos).
		// Cybernoid, juegos con muchas animaciones de fondo.
		cart.Mapper = NewMapper3(prgBanks)
		log.Printf("Mapper 3 (CNROM) detectado para %s", os.Args[1])
	case 4:
		// Mapper 4 (MMC3): El más avanzado de la era clásica.
		// Super Mario Bros 3, Kirby. Soporta IRQ por Scanline.
		// Nota: MMC3 usa bancos de CHR de 1KB, por eso multiplicamos por 8.
		cart.Mapper = NewMapper4(prgBanks, chrBanks*8)
		log.Printf("Mapper 4 (MMC3) detectado para %s", os.Args[1])
	default:
		// Fallback de seguridad
		log.Printf("ADVERTENCIA: Mapper %d no soportado plenamente. Usando Mapper 0.", mapperID)
		cart.Mapper = NewMapper0(prgBanks)
	}

	// 4. Manejo de CHR-ROM vs CHR-RAM
	// Si chrBanks > 0, el cartucho trae gráficos fijos (ROM).
	// Si chrBanks == 0, el cartucho usa RAM para gráficos (el juego los dibuja en tiempo real).
	if chrBanks > 0 {
		cart.CHR = data[prgEnd:chrEnd]
		cart.IsCHRRAM = false
	} else {
		log.Println("CHR-RAM detectada (8KB). Activando Double Buffering para evitar glitches.")
		cart.CHR = make([]byte, 8192)
		cart.IsCHRRAM = true
	}

	// 5. Inicializar Componentes del Sistema
	// Conexión: CPU <-> Bus <-> PPU <-> Cartucho
	ppu := NewPPU(cart)
	joy := NewController()
	// El Bus conecta todo.
	bus := &Bus{ROM: cart.PRG, PPU: ppu, Joy1: joy}
	cpu := NewCPU(bus)

	// Resetear la CPU al estado inicial (vector de reset $FFFC)
	cpu.Reset()

	// 6. Arrancar Motor Gráfico (Ebiten)
	game := &Game{CPU: cpu, Bus: bus}
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2) // Escalar x2 para ver mejor
	ebiten.SetWindowTitle("GoNES - Emulador Educativo")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
