package main

// MirrorMode define cómo se conectan las tablas de nombres (nametables) internamente.
// Esto permite simular mundos más grandes que la pantalla.
type MirrorMode int

const (
	MirrorHorizontal MirrorMode = iota // Scroll vertical (juegos como Ice Climber)
	MirrorVertical                     // Scroll horizontal (juegos como Super Mario Bros)
	MirrorSingle0                      // Solo una pantalla (banco 0)
	MirrorSingle1                      // Solo una pantalla (banco 1)
)

// Cartridge representa el cartucho físico de juego insertado en la consola.
type Cartridge struct {
	PRG      []byte     // Memoria de Programa (Código del juego, conectado a CPU Bus $8000+)
	CHR      []byte     // Memoria de Caracteres (Gráficos/Tiles, conectado a PPU Bus $0000+)
	WRAM     []byte     // Work RAM (8KB) en $6000-$7FFF. Usado por juegos como SMB3.
	Mirror   MirrorMode // Modo de espejo hardcodeado en el cartucho (soldadura)
	Mapper   Mapper     // El chip Mapper (MMC1, MMC3, etc) que controla el acceso a memoria
	IsCHRRAM bool       // Indica si el cartucho usa RAM para gráficos (escribible) en vez de ROM
}
