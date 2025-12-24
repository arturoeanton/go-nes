package main

import (
	"testing"
)

func TestMapper1_Reset(t *testing.T) {
	// 128KB PRG (8 banks), 128KB CHR (16 banks)
	m := NewMapper1(8, 16)
	
	// Write with bit 7 set to reset
	// Address can be anything 0x8000+
	m.Write(0x8000, 0x80)

	if m.shiftReg != 0x10 {
		t.Errorf("Expected shiftReg 0x10 after reset, got 0x%X", m.shiftReg)
	}
	if m.writeCnt != 0 {
		t.Errorf("Expected writeCnt 0 after reset, got %d", m.writeCnt)
	}
	// Check default control mode (Mode 3: Fix Last Bank at $C000) -> Bits 2,3 = 11 -> 0x0C
	if m.control & 0x0C != 0x0C {
		t.Errorf("Expected control 0x0C set after reset, got 0x%X", m.control)
	}
}

func TestMapper1_SerialWrite(t *testing.T) {
	m := NewMapper1(8, 16)

	// We want to write value 0x0F (0000 1111) to Control Register (0x8000)
	// LSB first.
	// 1, 1, 1, 1, 0
	
	// Write 1
	m.Write(0x8000, 1)
	if m.shiftReg != 0x18 { // 10000 -> 11000
		t.Errorf("Step 1: Expected 0x18, got 0x%X", m.shiftReg)
	}
	
	// Write 1
	m.Write(0x8000, 1)
	if m.shiftReg != 0x1C { // 11000 -> 11100
		t.Errorf("Step 2: Expected 0x1C, got 0x%X", m.shiftReg)
	}

	// Write 1
	m.Write(0x8000, 1) // 11100 -> 11110
	
	// Write 1
	m.Write(0x8000, 1) // 11110 -> 11111 (0x1F)

	// Write 0 (5th write) -> Should trigger update
	m.control = 0 // Clear it first to ensure update happens
	m.Write(0x8000, 0)
	
	// Expected value: 01111 (0x0F)
	if m.control != 0x0F {
		t.Errorf("Expected control to be updated to 0x0F, got 0x%X", m.control)
	}
	// Shift reg should be reset
	if m.shiftReg != 0x10 {
		t.Errorf("ShiftReg not reset after 5th write, got 0x%X", m.shiftReg)
	}
}

func TestMapper1_PRGBanking(t *testing.T) {
	// 256KB PRG (16 banks). Indices 0-15.
	// 16384 bytes per bank.
	m := NewMapper1(16, 0) 

	// Set Control to 0x0C (Fix Last Bank at $C000) - Default.
	
	// Select Bank 6 for the switchable slot ($8000)
	// Register 3 ($E000). data = 6.
	// Sequence: 0, 1, 1, 0, 0 (LSB first: 0110 -> 6)
	
	val := 6 
	// Bits: 0, 1, 1, 0, 0
	m.Write(0xE000, byte((val >> 0) & 1))
	m.Write(0xE000, byte((val >> 1) & 1))
	m.Write(0xE000, byte((val >> 2) & 1))
	m.Write(0xE000, byte((val >> 3) & 1))
	m.Write(0xE000, byte((val >> 4) & 1))
	
	// Verify PRG Register
	if m.prgBank != 6 {
		t.Errorf("Expected prgBank register 6, got %d", m.prgBank)
	}
	
	// Verify Reads
	// $8000 should map to Bank 6 (Start 6 * 16384 = 98304)
	addr := uint16(0x8000)
	offset := m.Read(addr)
	expected := 6 * 16384
	if offset != expected {
		t.Errorf("Read(0x8000) expected offset %d, got %d", expected, offset)
	}
	
	// $C000 should map to Last Bank (15)
	addr = 0xC000
	offset = m.Read(addr)
	expected = 15 * 16384
	if offset != expected {
		t.Errorf("Read(0xC000) expected offset %d, got %d", expected, offset)
	}
}
