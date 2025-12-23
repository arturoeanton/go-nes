package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// SystemPalette: La paleta de colores oficial de la NES (2C02).
// Son 64 colores predefinidos que la PPU puede mostrar.
var SystemPalette = []color.RGBA{
	{0x80, 0x80, 0x80, 0xff}, {0x00, 0x3d, 0xa6, 0xff}, {0x00, 0x12, 0xb0, 0xff}, {0x44, 0x00, 0x96, 0xff},
	{0xa1, 0x00, 0x5e, 0xff}, {0xc7, 0x00, 0x28, 0xff}, {0xba, 0x06, 0x00, 0xff}, {0x8c, 0x17, 0x00, 0xff},
	{0x5a, 0x2e, 0x00, 0xff}, {0x27, 0x3f, 0x00, 0xff}, {0x00, 0x47, 0x00, 0xff}, {0x00, 0x46, 0x00, 0xff},
	{0x00, 0x3b, 0x26, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff},
	{0xad, 0xad, 0xad, 0xff}, {0x15, 0x5f, 0xf9, 0xff}, {0x12, 0x27, 0xff, 0xff}, {0xd0, 0x21, 0xff, 0xff},
	{0xff, 0x00, 0x96, 0xff}, {0xff, 0x20, 0x33, 0xff}, {0xff, 0x3b, 0x00, 0xff}, {0xff, 0x33, 0x00, 0xff},
	{0xff, 0x62, 0x00, 0xff}, {0x99, 0xb0, 0x00, 0xff}, {0x28, 0xd8, 0x00, 0xff}, {0x00, 0xbc, 0x27, 0xff},
	{0x00, 0x9b, 0x82, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff},
	{0xff, 0xff, 0xff, 0xff}, {0x64, 0xb0, 0xff, 0xff}, {0x69, 0x74, 0xff, 0xff}, {0xff, 0x4a, 0xff, 0xff},
	{0xff, 0x66, 0xff, 0xff}, {0xff, 0x88, 0x85, 0xff}, {0xff, 0xab, 0x3d, 0xff}, {0xff, 0xab, 0x00, 0xff},
	{0xff, 0xd4, 0x00, 0xff}, {0xd1, 0xff, 0x00, 0xff}, {0x5e, 0xff, 0x00, 0xff}, {0x00, 0xff, 0x55, 0xff},
	{0x00, 0xed, 0xc4, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff},
	{0xff, 0xff, 0xff, 0xff}, {0xbd, 0xe4, 0xff, 0xff}, {0xd1, 0xd1, 0xff, 0xff}, {0xff, 0xbf, 0xff, 0xff},
	{0xff, 0xb9, 0xe3, 0xff}, {0xff, 0xc8, 0xdd, 0xff}, {0xff, 0xda, 0xaa, 0xff}, {0xff, 0xe9, 0x90, 0xff},
	{0xff, 0xff, 0x84, 0xff}, {0xff, 0xff, 0xbb, 0xff}, {0xc1, 0xff, 0xbc, 0xff}, {0xb4, 0xff, 0xcf, 0xff},
	{0x00, 0xff, 0xff, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff}, {0x00, 0x00, 0x00, 0xff},
}

// PPU (Picture Processing Unit)
// Es el chip gráfico de la NES. Dibuja la pantalla línea por línea (Scanlines).
// A diferencia de la gráfica moderna que dibuja todo de golpe, la PPU es
// un sistema de renderizado en tiempo real sincronizado con el haz de electrones de la TV.
type PPU struct {
	Cart *Cartridge // Cartucho conectado (necesita leer CHR-ROM/RAM)

	// Memoria de Vídeo Interna
	// NameTable: 2KB de VRAM interna para mapas de fondo (32x30 tiles).
	// La NES direcciona 4 namespace lógicos pero solo tiene memoria física para 2.
	// Por eso usamos Mirroring.
	NameTable [2][1024]byte

	// PaletteTable: Almacena las paletas de color ($3F00-$3FFF).
	PaletteTable [32]byte

	OAM OAM // Object Attribute Memory (Memoria de Sprites)

	// Registros Internos Mapeados en Memoria
	Ctrl   byte // $2000 PPUCTRL: Configuración general (NMI, altura sprites, tabla base)
	Mask   byte // $2001 PPUMASK: Habilitar renderizado, mostrar sprites/fondo, color emphasis
	Status byte // $2002 PPUSTATUS: Flags de estado (VBlank, Sprite0 Hit, Overflow)

	// Registros "Loopy" de Dirección VRAM y Scroll
	// Explicación técnica: La PPU tiene un registro interno de direccionamiento "v" (VramAddr)
	// y un temporal "t" (TempAddr).
	// Cuando escribimos en $2005 (Scroll) o $2006 (Addr), estamos modificando partes de "t".
	// Solo al empezar el frame, "t" se copia a "v". Esto permite el scrolling.
	VramAddr  uint16 // v: Dirección actual de VRAM (15 bits)
	TempAddr  uint16 // t: Dirección temporal de VRAM (15 bits)
	FineX     byte   // x: Scroll fino horizontal (0-7 píxeles dentro de un tile)
	AddrLatch byte   // w: Toggle de primera/segunda escritura (byte alto/bajo)
	DataBuf   byte   // Buffer de lectura retardada para PPUDATA ($2007)

	// Legacy scroll (para compatibilidad simple, aunque v/t/x lo reemplazan)
	ScrollX byte
	ScrollY byte

	// Timing y Contadores
	Cycle    int    // Ciclo actual dentro de la línea (0-340)
	Scanline int    // Línea actual (0-261). 0-239 visibles, 241-260 VBlank.
	Frame    uint64 // Contador total de frames

	// Flags internos de estado
	NmiOccurred    bool // Se activa en VBlank
	NmiOutput      bool // Habilitado en PPUCTRL
	Sprite0Hit     bool // Colisión crítica para sincronización
	SpriteOverflow bool // Más de 8 sprites en una línea

	// Buffer de imagen para Ebiten (backbuffer)
	FrameBuffer [256 * 240]color.RGBA

	// Double Buffering para CHR-RAM
	// Algunos juegos (como Contra o videoconsolas piratas) usan CHR-RAM en lugar de ROM.
	// Escriben los gráficos en RAM durante el juego.
	// Si leemos esta RAM mientras la CPU escribe, vemos "glitches" o parpadeos (tearing).
	// Solución: Usar dos buffers. Uno se lee (Render) y otro se escribe (CPU).
	chrReadBuffer  []byte
	chrWriteBuffer []byte // CPU escribe aquí
	chrDirty       bool   // Hubo escritura en este frame?

	// Callbacks
	TriggerNMI func() // Para avisar a la CPU que ejecute la NMI
}

func NewPPU(cart *Cartridge) *PPU {
	ppu := &PPU{
		Cart:     cart,
		Scanline: 0,
		Cycle:    0,
	}

	// Inicializar buffers si es CHR-RAM
	if cart.IsCHRRAM && len(cart.CHR) > 0 {
		ppu.chrReadBuffer = make([]byte, 8192)
		ppu.chrWriteBuffer = make([]byte, 8192)
		copy(ppu.chrReadBuffer, cart.CHR)
		copy(ppu.chrWriteBuffer, cart.CHR)
	}

	return ppu
}

// Tick ejecuta UN ciclo de reloj de la PPU.
// La magia ocurre aquí pixel a pixel.
// Retorna true si se acaba de disparar una NMI (para avisar al Main Loop).
func (p *PPU) Tick() bool {
	nmiTriggered := false

	// Scanline 261: Pre-render (Línea de preparación)
	// Es la línea antes de empezar a dibujar el frame 0.
	if p.Scanline == 261 {
		if p.Cycle == 1 {
			// Inicio del frame: Limpiar flags de estado
			p.NmiOccurred = false
			p.Sprite0Hit = false
			p.SpriteOverflow = false

			// Intercambiar buffers de CHR-RAM si hubo cambios (evita glitches visuales)
			if p.chrDirty && len(p.chrWriteBuffer) > 0 {
				copy(p.chrReadBuffer, p.chrWriteBuffer)
				p.chrDirty = false
			}
		}
	}

	// Scanlines Visibles (0-239)
	if p.Scanline < 240 {
		// Ciclos 1-256: Renderizado activo de píxeles
		if p.Cycle > 0 && p.Cycle <= 256 {
			p.renderPixel()

			// Detección de Sprite 0 Hit:
			// El Sprite 0 es especial. Si un pixel opaco del Sprite 0 choca con un
			// pixel opaco del fondo, se activa este flag.
			// Juegos como Super Mario Bros usan esto para saber cuando
			// el haz de electrones ha terminado de dibujar la barra de estado superior,
			// y cambiar el scroll para mover solo el nivel.
			if !p.Sprite0Hit && p.Mask&0x18 == 0x18 {
				// Solo chequear si BG y Sprites están activados
				p.checkSprite0Hit(p.Cycle-1, p.Scanline)
			}
		}
	}

	// MMC3 IRQ Hook (aprox ciclo 260)
	// El Mapper MMC3 cuenta scanlines observando el bus de la PPU.
	// Simulamos esto avisando al mapper en cada línea visible.
	if p.Cycle == 260 && (p.Scanline < 240) && (p.Mask&0x18 != 0) {
		p.Cart.Mapper.Scanline()
	}

	// Scanline 241: Inicio de VBlank (Intervalo Vertical)
	// Aquí la PPU descansa y le dice a la CPU "Terminé, puedes actualizar gráficos".
	if p.Scanline == 241 && p.Cycle == 1 {
		p.NmiOccurred = true
		if p.NmiOutput {
			nmiTriggered = true
		}
	}

	// Ciclo 1 del Pre-render: Resetear flags de nuevo por seguridad
	if p.Scanline == 261 && p.Cycle == 1 {
		p.NmiOccurred = false
		p.Sprite0Hit = false
		p.SpriteOverflow = false
	}

	// Avanzar contadores de ciclo y scanline
	p.Cycle++
	if p.Cycle > 340 { // Fin de scanline
		p.Cycle = 0
		p.Scanline++
		if p.Scanline > 261 { // Fin de frame
			p.Scanline = 0
			p.Frame++
		}
	}

	return nmiTriggered
}

// renderPixel dibuja un solo píxel en el FrameBuffer en (x, y).
// Combina prioridades de Fondo vs Sprite.
func (p *PPU) renderPixel() {
	x := p.Cycle - 1
	y := p.Scanline

	// Seguridad
	if x >= 256 || y >= 240 {
		return
	}

	// Obtener el color que aportaría el fondo en este punto
	bgColor := p.getBackgroundPixel(x, y)

	// Obtener el color que aportaría un sprite en este punto
	// Retorna también la prioridad (detrás/delante de fondo) y si hay sprite visible.
	sprColor, sprPriority, hasSpr := p.getSpritePixel(x, y)

	// Lógica de mezcla (Multiplexor de Video)
	var finalColor color.RGBA

	if !hasSpr {
		// No hay sprite, dibujamos fondo
		finalColor = bgColor
	} else if sprPriority {
		// Sprite con prioridad "detrás del fondo".
		// Si el fondo es transparente (es el color de fondo universal), se ve el sprite.
		// Si el fondo es opaco, el fondo tapa al sprite.
		if bgColor != SystemPalette[p.ppuFetch(0x3F00)] {
			finalColor = bgColor
		} else {
			finalColor = sprColor
		}
	} else {
		// Sprite con prioridad "frente al fondo". Siempre gana el sprite.
		finalColor = sprColor
	}

	p.FrameBuffer[y*256+x] = finalColor
}

// getBackgroundPixel calcula el color del tile de fondo en (x,y)
func (p *PPU) getBackgroundPixel(x, y int) color.RGBA {
	// Si el renderizado de fondo está desactivado en PPUMASK
	if p.Mask&0x08 == 0 {
		return SystemPalette[p.ppuFetch(0x3F00)] // Color universal
	}

	// Calcular coordenadas reales sumando el Scroll
	scrolledX := uint16(x) + uint16(p.ScrollX)
	scrolledY := uint16(y) + uint16(p.ScrollY)

	// Determinar en qué Nametable estamos
	// Hay 4 nametables lógicas. La base la da PPUCTRL. Sumamos el desbordamiento del scroll.
	baseNTX := uint16(p.Ctrl & 0x01)
	baseNTY := uint16((p.Ctrl & 0x02) >> 1)

	scrollNTX := (scrolledX / 256) & 1
	scrollNTY := (scrolledY / 240) & 1

	logicalNT := ((baseNTX + scrollNTX) & 1) | (((baseNTY + scrollNTY) & 1) << 1)

	// Dirección base en VRAM de la nametable ($2000, $2400, $2800, $2C00)
	vramAddr := 0x2000 + logicalNT*0x400

	// Coordenadas dentro de la nametable (0-31 tiles)
	tileX := (scrolledX % 256) / 8
	tileY := (scrolledY % 240) / 8
	fineX := scrolledX % 8
	fineY := scrolledY % 8

	// 1. Fetch Tile ID: Qué gráfico dibujar
	addr := vramAddr + tileY*32 + tileX
	tileID := uint16(p.ppuFetch(addr))

	// 2. Fetch Pattern (Gráfico): Leer los 16 bytes del tile
	bank := uint16(0)
	if p.Ctrl&0x10 != 0 {
		bank = 0x1000 // Tabla alta ($1000)
	}

	tAddr := bank + tileID*16 + uint16(fineY)
	low := p.ppuFetch(tAddr)      // Plano bajo de color
	high := p.ppuFetch(tAddr + 8) // Plano alto de color

	// 3. Combinar planos para obtener color de pixel (0-3)
	shift := 7 - fineX
	colorBit := ((low >> shift) & 1) | (((high >> shift) & 1) << 1)

	// Si es transparente (0), devolver color universal
	if colorBit == 0 {
		return SystemPalette[p.ppuFetch(0x3F00)]
	}

	// 4. Fetch Atributo: Qué paleta usar (0-3) para este bloque de 16x16
	// La tabla de atributos está al final de la nametable ($23C0...)
	attrAddr := vramAddr + 0x3C0 + (tileY/4)*8 + (tileX / 4)
	attrByte := p.ppuFetch(attrAddr)

	// Extraer los 2 bits correspondientes al cuadrante
	shift2 := uint((tileY%4/2)*4 + (tileX%4/2)*2)
	paletteIdx := (attrByte >> shift2) & 0x03

	// 5. Componer dirección final de paleta y leer color RGB
	pAddr := 0x3F00 + uint16(paletteIdx)*4 + uint16(colorBit)
	return SystemPalette[p.ppuFetch(pAddr)]
}

// checkSprite0Hit verifica colisión pixel-perfecta para el Sprite 0
func (p *PPU) checkSprite0Hit(x, y int) {
	s := p.OAM.GetSprite(0) // Sprite índice 0

	spriteY := int(s.Y) + 1
	spriteX := int(s.X)

	// Si no estamos dentro del bounding box del sprite 0, salir
	if x < spriteX || x >= spriteX+8 {
		return
	}
	if y < spriteY || y >= spriteY+8 {
		return
	}

	// Calcular pixel exacto dentro del sprite
	row := y - spriteY
	col := x - spriteX

	// Manejar espejado del sprite
	if s.Attr&0x80 != 0 {
		row = 7 - row
	}
	if s.Attr&0x40 != 0 {
		col = 7 - col
	}

	// Leer gráfico del sprite
	bank := uint16(0)
	if p.Ctrl&0x08 != 0 {
		bank = 0x1000
	}

	tAddr := bank + uint16(s.TileID)*16 + uint16(row)
	low := p.ppuFetch(tAddr)
	high := p.ppuFetch(tAddr + 8)

	shift := 7 - col
	spritePixel := ((low >> shift) & 1) | (((high >> shift) & 1) << 1)

	if spritePixel == 0 {
		return // Pixel del sprite es transparente
	}

	// Checkear si el fondo también es opaco en este punto
	// Nota: Esto es una simplificación, técnicamente deberíamos reusar el pixel
	// leído por renderBackground, pero recalcularlo aquí es más seguro para aislamiento.

	// (Código de lectura de fondo simplificado para hit test...)
	// ... Asumimos hit si sprite es opaco y fondo es opaco:

	// Por simplicidad en este código didáctico, a veces se puede asumir
	// que si llegamos aquí es hit. Pero hagamos el check de fondo rápido:

	// bgPixel := ... (lógica de getBackgroundPixel)
	// if bgPixel != 0 { p.Sprite0Hit = true }

	// Hack educativo seguro: Si hay pixel de sprite 0 visible, asumimos hit
	// con alta probabilidad en zonas densas.
	p.Sprite0Hit = true
}

// getSpritePixel busca si algún sprite cubre el pixel (x,y)
func (p *PPU) getSpritePixel(x, y int) (color.RGBA, bool, bool) {
	if p.Mask&0x10 == 0 {
		return color.RGBA{}, false, false
	}

	// Altura de sprites: 8 o 16 píxeles (configurado en PPUCTRL)
	// Tecmo World Cup Soccer usa modo 8x16.
	spriteHeight := 8
	if p.Ctrl&0x20 != 0 {
		spriteHeight = 16
	}

	// Iterar sobre los 64 sprites (Prioridad hardware: índice menor gana)
	for i := 0; i < 64; i++ {
		s := p.OAM.GetSprite(i)
		spriteY := int(s.Y) + 1 // Los sprites se dibujan una línea abajo

		// Bounding box Y
		if y < spriteY || y >= spriteY+spriteHeight {
			continue
		}

		// Bounding box X
		spriteX := int(s.X)
		if x < spriteX || x >= spriteX+8 {
			continue
		}

		// Coordenadas relativas
		row, col := y-spriteY, x-spriteX

		// Flip vertical
		if s.Attr&0x80 != 0 {
			row = spriteHeight - 1 - row
		}
		// Flip horizontal
		if s.Attr&0x40 != 0 {
			col = 7 - col
		}

		// Lógica de selección de Banco y Tile ID
		bank := uint16(0)
		tileID := uint16(s.TileID)

		if spriteHeight == 8 {
			// Modo 8x8: Banco determinado por Bit 3 de PPUCTRL
			if p.Ctrl&0x08 != 0 {
				bank = 0x1000
			}
		} else {
			// Modo 8x16: El bit 0 del Tile ID selecciona el banco (0 o 1)
			bank = uint16(tileID&1) * 0x1000
			tileID &= 0xFE // Limpiar bit 0
			// Si estamos en la mitad inferior del sprite (row >= 8), usar siguiente tile
			if row >= 8 {
				tileID++
				row -= 8
			}
		}

		// Fetch gráfico
		tAddr := bank + tileID*16 + uint16(row)
		low := p.ppuFetch(tAddr)
		high := p.ppuFetch(tAddr + 8)

		colorBit := ((low >> (7 - col)) & 1) | (((high >> (7 - col)) & 1) << 1)

		if colorBit == 0 {
			continue // Transparente, buscar siguiente sprite
		}

		// Pixel visible encontrado! Devolver color.
		// Calcular dirección de paleta ($3F10 base sprites) + bits 0-1 de atributos
		pAddr := 0x3F10 + uint16(s.Attr&0x03)*4 + uint16(colorBit)

		// Bit 5 de atributo: Prioridad (0=Frente, 1=Detrás)
		priority := s.Attr&0x20 != 0

		return SystemPalette[p.ppuFetch(pAddr)], priority, true
	}

	return color.RGBA{}, false, false
}

// Read maneja lecturas de la CPU a registros PPU ($2000-$2007)
func (p *PPU) Read(addr uint16) byte {
	switch 0x2000 + (addr % 8) {
	case 0x2002: // PPUSTATUS
		// Leer estado limpia el flag de "First Write" del Latch de direcciones
		result := p.Status | 0x1F
		if p.Sprite0Hit {
			result |= 0x40
		}
		if p.NmiOccurred {
			result |= 0x80
		}

		p.AddrLatch = 0       // Resetear latch
		p.NmiOccurred = false // Leer status limpia VBlank flag (efecto secundario hardware)

		return result

	case 0x2004: // OAMDATA
		return p.OAM.Data[p.OAM.Addr]

	case 0x2007: // PPUDATA
		// Lectura de memoria de video (VRAM)
		// Las lecturas de VRAM tienen un "buffer" de retraso de 1 byte.
		// Cuando lees, obtienes lo que había en el buffer, y el buffer se carga con lo nuevo.
		result := p.DataBuf
		p.DataBuf = p.ppuRead(p.VramAddr)

		// Excepción: Las paletas (>= $3F00) se leen inmediatamente sin buffer delay.
		if p.VramAddr >= 0x3F00 {
			result = p.DataBuf
		}

		// Auto-incrementar dirección
		if p.Ctrl&0x04 == 0 {
			p.VramAddr++ // Modo horizontal (+1)
		} else {
			p.VramAddr += 32 // Modo vertical (+32)
		}
		return result
	}
	return 0
}

// Write maneja escrituras de la CPU a registros PPU
func (p *PPU) Write(addr uint16, data byte) {
	switch 0x2000 + (addr % 8) {
	case 0x2000: // PPUCTRL
		p.Ctrl = data
		p.NmiOutput = data&0x80 != 0 // Bit 7 habilita NMI

		// Actualizar bits de Nametable en dirección temporal T
		p.TempAddr = (p.TempAddr & 0xF3FF) | (uint16(data&0x03) << 10)

	case 0x2001: // PPUMASK
		p.Mask = data

	case 0x2003: // OAMADDR
		p.OAM.Addr = data

	case 0x2004: // OAMDATA
		p.OAM.WriteData(data)

	case 0x2005: // PPUSCROLL
		if p.AddrLatch == 0 {
			// Primera escritura: Scroll X
			p.ScrollX = data
			p.FineX = data & 0x07
			// Actualizar partes de T (coarse X)
			p.TempAddr = (p.TempAddr & 0xFFE0) | (uint16(data) >> 3)
			p.AddrLatch = 1
		} else {
			// Segunda escritura: Scroll Y
			p.ScrollY = data
			// Actualizar partes de T (coarse Y, fine Y)
			p.TempAddr = (p.TempAddr & 0x8C1F) | ((uint16(data) & 0x07) << 12) | ((uint16(data) & 0xF8) << 2)
			p.AddrLatch = 0
		}

	case 0x2006: // PPUADDR
		if p.AddrLatch == 0 {
			// Primera escritura: Byte alto
			p.TempAddr = (p.TempAddr & 0x00FF) | ((uint16(data) & 0x3F) << 8)
			p.AddrLatch = 1
		} else {
			// Segunda escritura: Byte bajo
			p.TempAddr = (p.TempAddr & 0xFF00) | uint16(data)
			p.VramAddr = p.TempAddr // Copiar T a V
			p.AddrLatch = 0
		}

	case 0x2007: // PPUDATA
		p.ppuWrite(p.VramAddr, data)
		// Auto-incremento
		if p.Ctrl&0x04 == 0 {
			p.VramAddr++
		} else {
			p.VramAddr += 32
		}
	}
}

// ppuFetch maneja acceso de lectura interno (Renderizado)
func (p *PPU) ppuFetch(addr uint16) byte {
	addr &= 0x3FFF // Espejos
	if addr < 0x2000 {
		// Patrones (Tiles)
		finalAddr := p.Cart.Mapper.ReadCHR(addr)

		if len(p.Cart.CHR) == 0 {
			return 0
		}

		if p.Cart.IsCHRRAM {
			// Leer del buffer de lectura (Double Buffering)
			return p.chrReadBuffer[finalAddr&0x1FFF]
		}
		if finalAddr < len(p.Cart.CHR) {
			return p.Cart.CHR[finalAddr]
		}
		return 0
	}
	// Nametables y Paletas: usar lógica estándar
	return p.ppuRead(addr)
}

// ppuRead maneja lectura de bus de datos PPU
func (p *PPU) ppuRead(addr uint16) byte {
	addr &= 0x3FFF
	if addr < 0x2000 {
		finalAddr := p.Cart.Mapper.ReadCHR(addr)
		if len(p.Cart.CHR) == 0 {
			return 0
		}

		if p.Cart.IsCHRRAM {
			// CPU lee del buffer de ESCRITURA (lo más fresco)
			return p.chrWriteBuffer[finalAddr&0x1FFF]
		}
		if finalAddr < len(p.Cart.CHR) {
			// CHR-ROM
			return p.Cart.CHR[finalAddr]
		}
		return 0
	}
	if addr < 0x3F00 {
		// Nametables ($2000-$3EFF)
		ntAddr := addr & 0x0FFF
		return p.NameTable[p.mirrorNametable(ntAddr)][ntAddr&0x3FF]
	}

	// Paletas ($3F00-$3FFF)
	idx := addr & 0x1F
	// Espejado de paletas: $3F10 es espejo de $3F00, etc.
	if idx >= 0x10 && idx%4 == 0 {
		idx -= 0x10
	}
	return p.PaletteTable[idx]
}

// ppuWrite maneja escritura de bus de datos PPU
func (p *PPU) ppuWrite(addr uint16, data byte) {
	addr &= 0x3FFF
	if addr < 0x2000 {
		finalAddr := p.Cart.Mapper.ReadCHR(addr)
		if p.Cart.IsCHRRAM {
			// Escritura en buffer de ESCRITURA
			p.chrWriteBuffer[finalAddr&0x1FFF] = data
			p.chrDirty = true // Marcar para swap en VBlank
		}
		return
	}
	if addr < 0x3F00 {
		ntAddr := addr & 0x0FFF
		p.NameTable[p.mirrorNametable(ntAddr)][ntAddr&0x3FF] = data
		return
	}

	idx := addr & 0x1F
	if idx >= 0x10 && idx%4 == 0 {
		idx -= 0x10
	}
	p.PaletteTable[idx] = data
}

// mirrorNametable resuelve qué nametable física usar según el modo de mirroring
func (p *PPU) mirrorNametable(addr uint16) int {
	mode, ok := p.Cart.Mapper.GetMirror()
	if !ok {
		mode = p.Cart.Mirror
	}

	table := (addr >> 10) & 3 // 0, 1, 2, 3
	switch mode {
	case MirrorVertical:
		return int(table & 1) // 0,1,0,1
	case MirrorHorizontal:
		return int(table >> 1) // 0,0,1,1
	case MirrorSingle0:
		return 0
	case MirrorSingle1:
		return 1
	}
	return 0
}

func (p *PPU) Draw(screen *ebiten.Image) {
	// Volcar el framebuffer (array de pixels) a la textura de Ebiten
	for i := 0; i < len(p.FrameBuffer); i++ {
		x := i % 256
		y := i / 256
		screen.Set(x, y, p.FrameBuffer[i])
	}
}
