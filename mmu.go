package main

import (
	"math/rand"
)

// MEMORYSIZE is fixed at pow(2, 16) bytes
const MEMORYSIZE = 0x10000

// MMU is a struct with a fixed-size memory and access functions
type MMU struct {
	memory [MEMORYSIZE]uint8
}

// Reset initializes the memory of an MMU to random 8-bit integers.
func (m *MMU) Reset() {
	m.memory = [MEMORYSIZE]uint8{}
	for i := range m.memory {
		m.memory[i] = uint8(rand.Intn(0x100))
	}
}

// LoadCartridgeData reads into memory the data from a passed cartridge.
func (m *MMU) LoadCartridgeData(data []uint8) {
	for i := 0x0100; i < 0x8000; i++ {
		address := uint16(i)
		m.WriteByte(address, data[i])
	}
}

// ReadByte returns the byte of memory at a given address.
func (m *MMU) ReadByte(address uint16) uint8 {
	return m.memory[address]
}

// WriteByte writes a given byte to memory at a given address.
func (m *MMU) WriteByte(address uint16, value uint8) {
	m.memory[address] = value
}

// ReadWord reads a 16-bit word from memory starting at a given address.
// It returns a word in order lowByte, highByte.
func (m *MMU) ReadWord(address uint16) uint16 {
	lowByte := m.memory[address]
	highByte := m.memory[address+1]
	return U8PairToU16([2]uint8{lowByte, highByte})
}

// WriteWord writes a 16-bit word to memory starting at a given address.
// The byte is written lowByte, highByte.
func (m *MMU) WriteWord(address uint16, value uint16) {
	byteSlice := U16ToU8Pair(value)
	m.memory[address] = byteSlice[0]
	m.memory[address+1] = byteSlice[1]
}
