package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// Constantes globales de resolución
const (
	ScreenWidth  = 256
	ScreenHeight = 240
)

// main es el punto de entrada del programa.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run . <archivo.nes>")
		return
	}

	// Leer el archivo ROM completo a memoria
	data, err := os.ReadFile(os.Args[1])
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
		WRAM:   make([]byte, 8192),    // 8KB Work RAM
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
	case 7:
		// Mapper 7 (AxROM): Battletoads.
		// Bancos PRG 32KB, Mirroring por software.
		cart.Mapper = NewMapper7(prgBanks)
		log.Printf("Mapper 7 (AxROM) detectado para %s", os.Args[1])
	case 69:
		// Mapper 69 (Sunsoft FME-7): Batman Return of the Joker.
		// IRQ por ciclos de CPU, bancos CHR 1KB.
		cart.Mapper = NewMapper69(prgBanks, chrBanks)
		log.Printf("Mapper 69 (FME-7) detectado para %s", os.Args[1])
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
	apu := NewAPU() // Nuevo: APU mínima
	joy := NewController()
	// El Bus conecta todo.
	bus := &Bus{
		ROM:  cart.PRG,
		PPU:  ppu,
		APU:  apu, // Conectar APU
		Joy1: joy,
		Cart: cart,
	}
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
