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
	NotifyA12(high bool)           // Notifica cambio en A12 de PPU (para IRQ del MMC3).
	Scanline()                     // Hook que la PPU llama cada línea (legacy, usar NotifyA12).
	Tick()                         // Hook que la CPU llama cada ciclo (para IRQ del FME-7).
	IRQState() bool                // Devuelve true si el Mapper pide interrumpir al CPU.
	GetMirror() (MirrorMode, bool) // Devuelve el modo de espejo si el Mapper lo controla dinámicamente.
}
