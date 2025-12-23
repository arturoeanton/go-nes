package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// InputDevice define la interfaz para cualquier dispositivo de entrada (mando, pistola, teclado).
type InputDevice interface {
	Write(data byte) // Escribir al puerto del dispositivo (Strobe)
	Read() byte      // Leer estado serialmente (un bit a la vez)
}

// KeyboardController simula un mando de NES estándar usando el teclado del PC.
type KeyboardController struct {
	strobe  bool // Mantiene el estado de la señal "Strobe" (Latch)
	index   byte // Qué botón estamos leyendo actualmente (0-7)
	buttons byte // Snapshot del estado de los 8 botones capturado durante el Strobe
}

func NewController() *KeyboardController {
	return &KeyboardController{}
}

// Write maneja la señal de "Strobe" en el puerto $4016.
// Protocolo:
// 1. CPU escribe 1: El mando empieza a "escuchar".
// 2. CPU escribe 0: El mando "congela" (latches) el estado actual de todos los botones.
// Esto evita que el estado cambie a mitad de lectura si el jugador suelta un botón muy rápido.
func (k *KeyboardController) Write(data byte) {
	newStrobe := data&1 == 1

	// Si el strobe pasa de 1 a 0, capturamos el estado real de los inputs
	if k.strobe && !newStrobe {
		k.buttons = k.getSnapshot()
		k.index = 0 // Reseteamos el lector al primer botón (A)
	}
	k.strobe = newStrobe
}

// Read devuelve el estado de un botón a la vez (Serial).
// La CPU llama a esto 8 veces consecutivas para leer A, B, Select, Start, Up, Down, Left, Right.
func (k *KeyboardController) Read() byte {
	// Si leemos más de 8 veces, la NES devuelve 1 habitualmente (circuito abierto/pull-up)
	if k.index > 7 {
		return 1
	}

	// Comportamiento especial: Si strobe se mantiene en 1, siempre devuelve el estado en vivo del botón A.
	if k.strobe {
		k.buttons = k.getSnapshot()
		return k.buttons & 1
	}

	// Extraemos el bit correspondiente al botón actual (index)
	res := (k.buttons >> k.index) & 1
	k.index++ // Preparamos para leer el siguiente botón en la próxima llamada
	return res
}

// getSnapshot lee el teclado real (Ebiten) y empaqueta los bits en un byte.
// Orden de bits estándar NES:
// Bit 0: A
// Bit 1: B
// Bit 2: Select
// Bit 3: Start
// Bit 4: Up
// Bit 5: Down
// Bit 6: Left
// Bit 7: Right
func (k *KeyboardController) getSnapshot() byte {
	var state byte
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		state |= 1 << 0
		log.Println("A presionado")
	} // Tecla Z = Botón A
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		state |= 1 << 1
		log.Println("B presionado")
	} // Tecla X = Botón B
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		state |= 1 << 2
		log.Println("Select presionado")
	} // Shift = Select
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		state |= 1 << 3
		log.Println("Start presionado")
	} // Enter = Start
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		state |= 1 << 4
		log.Println("Up presionado")
	} // Flecha Arriba
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		state |= 1 << 5
		log.Println("Down presionado")
	} // Flecha Abajo
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		state |= 1 << 6
		log.Println("Left presionado")
	} // Flecha Izquierda
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		state |= 1 << 7
		log.Println("Right presionado")
	} // Flecha Derecha
	return state
}
