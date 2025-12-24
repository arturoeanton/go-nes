package main

import (
	"log"
)

// Niveles de log
const (
	LevelError = 0
	LevelInfo  = 1
	LevelDebug = 2
)

// CurrentLevel define el nivel de verbosidad actual
var CurrentLevel int = LevelError

// InitLogger configura el nivel de log
func InitLogger(level int) {
	CurrentLevel = level
}

// Error siempre se imprime
func LogError(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// Info se imprime si CurrentLevel >= LevelInfo
func LogInfo(format string, v ...interface{}) {
	if CurrentLevel >= LevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

// Debug se imprime si CurrentLevel >= LevelDebug
func LogDebug(format string, v ...interface{}) {
	if CurrentLevel >= LevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}
