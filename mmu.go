package main

const MEMORYSIZE = 0x10000

type Memory interface {
	Reset(zero bool)
	ReadByte(address uint16) uint8
	ReadWord(address uint16) uint16
	WriteByte(address uint16, value uint8)
	WriteWord(address uint16, value uint16)
}

type MMU struct {
	memory [MEMORYSIZE]uint8
}

func (m *MMU) Reset() {
	m.memory = [MEMORYSIZE]uint8{}
}

func (m *MMU) ReadByte(address uint16) uint8 {
	return m.memory[address]
}

func (m *MMU) WriteByte(address uint16, value uint8) {
	m.memory[address] = value
}

func (m *MMU) ReadWord(address uint16) uint16 {
	lower := m.memory[address]
	higher := m.memory[address+1]
	return u8PairToU16([2]uint8{lower, higher})
}

func (m *MMU) WriteWord(address uint16, value uint16) {
	byteSlice := splitUint16(value)
	m.memory[address] = byteSlice[0]
	m.memory[address+1] = byteSlice[1]
}
