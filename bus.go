package main

// ==========================================
// BUS DEL SISTEMA (Interconexión General)
// ==========================================
// El Bus es como la placa madre de la NES. Conecta todos los componentes
// (CPU, PPU, RAM, Cartucho, Mandos) y dirige el tráfico de datos
// según las direcciones de memoria (Memory Mapping).
type Bus struct {
	RAM  [2048]byte  // RAM interna del sistema (2KB).
	ROM  []byte      // Referencia a la ROM del programa (PRG-ROM) para acceso rápido.
	PPU  *PPU        // Puntero a la Unidad de Procesamiento de Gráficos.
	Joy1 InputDevice // Puntero al controlador 1 (Joystick).
}

// Read maneja las lecturas de la CPU desde el bus de direcciones (0x0000 - 0xFFFF).
// La CPU pone una dirección en el bus, y este método decide qué componente responde.
func (b *Bus) Read(addr uint16) byte {
	switch {
	// $0000 - $1FFF: RAM del Sistema (2KB repetidos 4 veces)
	case addr < 0x2000:
		// La NES tiene solo 2KB de RAM física ($0000-$07FF).
		// Las direcciones $0800-$1FFF son "espejos" (mirrors) de esos 2KB.
		// Usamos el operador módulo (%) para simular esto.
		return b.RAM[addr%0x0800]

	// $2000 - $3FFF: Registros PPU (8 bytes repetidos)
	case addr < 0x4000:
		// La CPU se comunica con la PPU a través de puertos mapeados aquí.
		// Aunque el rango es grande, solo los primeros 8 bytes son registros reales.
		// El resto son espejos. La PPU maneja el módulo internamente.
		return b.PPU.Read(addr)

	// $4016: Puerto de lectura del Joystick 1
	case addr == 0x4016:
		if b.Joy1 != nil {
			return b.Joy1.Read()
		}
		return 0

	// $4017: Puerto del Joystick 2 (no implementado completamente)
	case addr == 0x4017:
		// En hardware real devuelve estado del Joy 2. Aquí devolvemos 0x40 (frame irq flag off).
		return 0x40

	// $8000 - $FFFF: Espacio del Cartucho (PRG-ROM)
	case addr >= 0x8000:
		// Aquí está el código del juego. Como el espacio de direcciones es limitado (32KB),
		// pero los juegos pueden ser enormes (Megabytes), usamos un "Mapper".
		// El Mapper traduce la dirección lógica de la CPU a la dirección física en la ROM.

		// 1. Preguntar al Mapper qué índice del array ROM corresponde a esta dirección
		index := b.PPU.Cart.Mapper.Read(addr)

		// 2. Si el mapper devuelve un índice válido, leer el dato
		if index >= 0 && index < len(b.PPU.Cart.PRG) {
			return b.PPU.Cart.PRG[index]
		}
		return 0

	default:
		// Direcciones no mapeadas o APU (no implementada) devuelven 0 habitualmente (open bus).
		return 0
	}
}

// Write maneja las escrituras de la CPU en el bus.
func (b *Bus) Write(addr uint16, data byte) {
	switch {
	// $0000 - $1FFF: Escritura en RAM
	case addr < 0x2000:
		b.RAM[addr%0x0800] = data

	// $2000 - $3FFF: Escritura en Registros PPU
	case addr < 0x4000:
		b.PPU.Write(addr, data)

	// $4014: OAM DMA (Direct Memory Access)
	// Este es un evento especial rápido. Copia 256 bytes desde la RAM de la CPU
	// a la memoria de sprites (OAM) de la PPU de una sola vez.
	case addr == 0x4014:
		base := uint16(data) << 8 // Dirección base en RAM (página XX00)
		for i := 0; i < 256; i++ {
			// Leer de RAM y escribir directamente en OAM
			// Nota: En hardware real esto congela la CPU por 513 ciclos.
			b.PPU.OAM.Data[i] = b.Read(base + uint16(i))
		}

	// $4016: Escritura en Joystick (Strobe)
	// Escribir 1 y luego 0 le dice al mando "captura los botones presionados ahora".
	case addr == 0x4016:
		if b.Joy1 != nil {
			b.Joy1.Write(data)
		}

	// $8000 - $FFFF: Escritura en Cartucho (Mapper)
	// La ROM es de solo lectura, PERO escribir aquí se usa para enviar
	// comandos al chip Mapper (ej: "Cambia al banco de gráficos 5").
	case addr >= 0x8000:
		b.PPU.Cart.Mapper.Write(addr, data)

	default:
		// APU y otros registros de I/O no implementados se ignoran.
		return
	}
}
