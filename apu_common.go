package main

// Tabla de Length Counter (Indices 0-31)
// Valores representan duración en Frames (aprox).
var lengthTable = [32]byte{
	10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
}

// Estructuras auxiliares (Envelope, Sweep) podrían ir aquí si se comparten,
// pero Sweep es específico de canales. Envelope es compartido.

type Envelope struct {
	startFlag   bool
	loop        bool // (o Disable Decay)
	constantVol bool
	volCub      byte // Volume / Period

	decayLevel byte
	divider    byte
}

func (e *Envelope) Tick() {
	if e.startFlag {
		e.startFlag = false
		e.decayLevel = 15
		e.divider = e.volCub
	} else {
		if e.divider == 0 {
			e.divider = e.volCub
			if e.decayLevel > 0 {
				e.decayLevel--
			} else if e.loop {
				e.decayLevel = 15
			}
		} else {
			e.divider--
		}
	}
}

func (e *Envelope) Output() byte {
	if e.constantVol {
		return e.volCub
	}
	return e.decayLevel
}
