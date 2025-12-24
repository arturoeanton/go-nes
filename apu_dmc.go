package main

// DMC (Delta Modulation Channel): Reproducción de samples.
// Es complejo, pero para evitar cuelgues lo vital es manejar el flag de IRQ y el bit de activo.

type DMCChannel struct {
	Enabled bool

	// Registros
	IrqEnabled bool // $4010 Bit 7
	Loop       bool // $4010 Bit 6
	RateIndex  byte // $4010 Bits 0-3

	DirectLoad byte // $4011

	SampleAddr uint16 // $4012 -> $C000 + (val * 64)
	SampleLen  uint16 // $4013 -> (val * 16) + 1

	// Estado Ejecución
	Active         bool
	BytesRemaining uint16
	CurrentAddr    uint16

	IrqPending bool
}

func NewDMCChannel() *DMCChannel {
	return &DMCChannel{}
}

func (d *DMCChannel) Write(addr uint16, data byte) {
	switch addr {
	case 0: // $4010
		// IL-- RRRR
		d.IrqEnabled = (data & 0x80) != 0
		d.Loop = (data & 0x40) != 0
		d.RateIndex = data & 0x0F

		if !d.IrqEnabled {
			d.IrqPending = false
		}

	case 1: // $4011
		// -DDD DDDD
		d.DirectLoad = data & 0x7F

	case 2: // $4012
		// AAAA AAAA
		d.SampleAddr = 0xC000 + (uint16(data) * 64)

	case 3: // $4013
		// LLLL LLLL
		d.SampleLen = (uint16(data) * 16) + 1
	}
}

// SetEnabled maneja escritura a $4015 Bit 4
func (d *DMCChannel) SetEnabled(enabled bool) {
	// Si pasa de 0 a 1 -> Restart sample si bytes remaining == 0
	if !d.Enabled && enabled {
		if d.BytesRemaining == 0 {
			d.CurrentAddr = d.SampleAddr
			d.BytesRemaining = d.SampleLen
		}
	}
	// Si pasa de 1 a 0 -> Stop immediately
	if !enabled {
		d.BytesRemaining = 0
	}

	d.Enabled = enabled
	d.IrqPending = false
}

// Status devuelve si hay bytes restantes (usado en $4015 bit 4)
func (d *DMCChannel) Status() bool {
	return d.BytesRemaining > 0
}

func (d *DMCChannel) Tick() {
	// Implementación dummy de consumo de bytes para simular fin de sample
	// En realidad esto depende del Rate Timer.
	// Si implementas audio real, aquí decrementas.
	// Para lógica simple, lo dejaremos estático o decremento lento si Active?
	// Batman podría esperar el bit 4? Mejor no tocar BytesRemaining automáticamente sin timer preciso,
	// o el juego pensará que terminó muy rápido.
	// DEJAR COMO ESTÁ: Solo start/stop manuales por ahora para evitar side effects no deseados.
}
