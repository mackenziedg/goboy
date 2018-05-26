package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/bits"
	"os"
)

const Z = 7
const N = 6
const H = 5
const C = 4

const BLSIZE = 0x100

type cpu interface {
	Reset(mmu *MMU)
	Start()
	Instruction(opcode uint8, location uint16, argBytes [2]uint8, cbFlag bool) (uint16, bool)
	CBInstruction(opcode uint8, location uint16, argByte uint8) (string, uint8, uint8)

	printInstruction(location uint16, opcode uint8, name string, length, duration, shortDuration uint8)
	printRegisters()
	printFlagRegisters()
}

type CPU struct {
	registers8  map[string]uint8
	registers16 map[string]uint16

	mmu *MMU

	bootloader [0x100]byte
}

func (c *CPU) Reset(mmu *MMU) {
	dat, err := ioutil.ReadFile("./data/DMG_ROM.bin")
	check(err)
	fmt.Printf("Bootloader is %d bytes long.\n\n", len(dat))

	for i, v := range dat {
		c.bootloader[i] = v
	}

	c.registers8 = map[string]uint8{
		"A": 0x0,
		"B": 0x0,
		"C": 0x0,
		"D": 0x0,
		"E": 0x0,
		"F": 0x0,
		"H": 0x0,
		"L": 0x0,
	}

	c.registers16 = map[string]uint16{
		"SP": 0x0,
		"PC": 0x0,
	}

	c.mmu = mmu
}

func (c *CPU) RunBootloader() {
	var i uint16
	cb := false
	reader := bufio.NewReader(os.Stdin)
	for {
		i = c.registers16["PC"]
		argBytes := [2]uint8{0, 0}
		if i < BLSIZE-2 {
			argBytes = splitUint16(c.mmu.ReadWord(i + 1))
		} else if i < BLSIZE-1 {
			argBytes = [2]uint8{c.mmu.ReadByte(i + 1), 0}
		}

		i, cb = c.Instruction(c.mmu.ReadByte(i), i, argBytes, cb)
		c.registers16["PC"] = i
		reader.ReadString('\n')
	}
}

func (c *CPU) Start() {
	c.RunBootloader()
	reader := bufio.NewReader(os.Stdin)

	var i uint16
	cb := false
	for {
		i = c.registers16["PC"]
		argBytes := [2]uint8{0, 0}
		if i < MEMORYSIZE-2 {
			argBytes = splitUint16(c.mmu.ReadWord(i + 1))
		} else if i < MEMORYSIZE-1 {
			argBytes = [2]uint8{c.mmu.ReadByte(i + 1), 0}
		}

		i, cb = c.Instruction(c.mmu.ReadByte(i), i, argBytes, cb)
		c.registers16["PC"] = i
		reader.ReadString('\n')
	}
}

func (c *CPU) Instruction(opcode uint8, location uint16, argBytes [2]uint8, cbFlag bool) (uint16, bool) {
	var name string
	var length uint8
	var duration, shortDuration uint8
	var jump bool
	var jumpTo uint16

	if cbFlag {
		name, length, duration = c.CBInstruction(opcode, location, argBytes[0])
		cbFlag = false
	} else {

		switch opcode {

		// INC/DEC
		case 0x04:
			name = "INC B"
			length = 1
			duration = 4
			c.registers8["B"] += 1
			c.unsetSubtractionFlag()
			if c.registers8["B"] == 0 {
				c.setZeroFlag()
				c.setHalfCarryFlag()
			}
		case 0x05:
			name = "DEC B"
			length = 1
			duration = 4
			c.registers8["B"] -= 1
			c.setSubtractionFlag()
			c.unsetZeroFlag()
			if c.registers8["B"] == 0 {
				c.setZeroFlag()
			} else if c.registers8["B"] == 255 {
				c.setHalfCarryFlag()
			}
		case 0x0C:
			name = "INC C"
			length = 1
			duration = 4
			c.registers8["C"] += 1
			c.unsetSubtractionFlag()
			if c.registers8["C"] == 0 {
				c.setZeroFlag()
				c.setHalfCarryFlag()
			}
		case 0x0D:
			name = "DEC C"
			length = 1
			duration = 4
			c.setSubtractionFlag()
			c.unsetZeroFlag()
			if c.registers8["C"] == 0 {
				c.setZeroFlag()
			} else if c.registers8["C"] == 255 {
				c.setHalfCarryFlag()
			}
		case 0x15:
			name = "DEC D"
			length = 1
			duration = 4
			c.setSubtractionFlag()
			if c.registers8["D"] == 0 {
				c.setZeroFlag()
			} else if c.registers8["D"] == 255 {
				c.setHalfCarryFlag()
			}
		case 0x1D:
			name = "DEC E"
			length = 1
			duration = 4
			c.setSubtractionFlag()
			if c.registers8["E"] == 0 {
				c.setZeroFlag()
			} else if c.registers8["E"] == 255 {
				c.setHalfCarryFlag()
			}
		case 0x24:
			name = "INC H"
			length = 1
			duration = 4
			c.registers8["H"] += 1
			c.unsetSubtractionFlag()
			if c.registers8["H"] == 0 {
				c.setZeroFlag()
				c.setHalfCarryFlag()
			}

		case 0x13:
			name = "INC DE"
			length = 1
			duration = 8
			combined := u8PairToU16([2]uint8{c.registers8["E"], c.registers8["D"]})
			combined += 1
			vals := splitUint16(combined)
			c.registers8["D"] = vals[1]
			c.registers8["E"] = vals[0]
		case 0x23:
			name = "INC HL"
			length = 1
			duration = 8
			combined := u8PairToU16([2]uint8{c.registers8["L"], c.registers8["H"]})
			combined += 1
			vals := splitUint16(combined)
			c.registers8["H"] = vals[1]
			c.registers8["L"] = vals[0]

			// LD R,d8
		case 0x3E:
			name = "LD A,d8"
			length = 2
			duration = 8
			byte_ := argBytes[0]
			c.registers8["A"] = byte_
		case 0x06:
			name = "LD B,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			c.registers8["B"] = c.mmu.ReadByte(address)
		case 0x0E:
			name = "LD C,d8"
			length = 2
			duration = 8
			byte_ := argBytes[0]
			c.registers8["C"] = byte_
		case 0x16:
			name = "LD D,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			c.registers8["D"] = c.mmu.ReadByte(address)
		case 0x1E:
			name = "LD E,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			c.registers8["E"] = c.mmu.ReadByte(address)

			// LD R,R
		case 0x4F:
			name = "LD C,A"
			length = 1
			duration = 4
			c.registers8["C"] = c.registers8["A"]
		case 0x7B:
			name = "LD A,E"
			length = 1
			duration = 4
			c.registers8["A"] = c.registers8["E"]

			// LD R,a16
		case 0x1A:
			name = "LD A,(DE)"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{c.registers8["E"], c.registers8["D"]})
			c.registers8["A"] = c.mmu.ReadByte(address)

			// LD RR,d16
		case 0x11:
			name = "LD DE,d16"
			length = 3
			duration = 12
			c.registers8["E"] = argBytes[0]
			c.registers8["D"] = argBytes[1]
		case 0x21:
			name = "LD HL,d16"
			length = 3
			duration = 12
			c.registers8["L"] = argBytes[0]
			c.registers8["H"] = argBytes[1]
		case 0x32:
			name = "LD (HL-),A"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{c.registers8["L"], c.registers8["H"]})
			c.mmu.WriteByte(address, c.registers8["A"])
			combined := u8PairToU16([2]uint8{c.registers8["L"], c.registers8["H"]})
			combined -= 1
			vals := splitUint16(combined)
			c.registers8["H"] = vals[1]
			c.registers8["L"] = vals[0]
			c.setSubtractionFlag()
			if combined == 0 {
				c.setZeroFlag()
			} else if combined == 65535 {
				c.setCarryFlag()
			}
		case 0x22:
			name = "LD (HL+),A"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{c.registers8["L"], c.registers8["H"]})
			c.mmu.WriteByte(address, c.registers8["A"])
			combined := u8PairToU16([2]uint8{c.registers8["L"], c.registers8["H"]})
			combined += 1
			vals := splitUint16(combined)
			c.registers8["H"] = vals[1]
			c.registers8["L"] = vals[0]
		case 0x31:
			name = "LD SP,d16"
			length = 3
			duration = 12
			address := u8PairToU16(argBytes)
			c.registers16["SP"] = address

			// LD address,R
		case 0xE2:
			name = "LD (C),A"
			length = 2
			duration = 8
			address := 0xFF00 | uint16(c.registers8["C"])
			c.mmu.WriteByte(address, c.registers8["A"])
			length = 1 // Change later when figure out wtf
		case 0x77:
			name = "LD (HL), A"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{c.registers8["L"], c.registers8["H"]})
			c.mmu.WriteByte(address, c.registers8["A"])
		case 0xE0:
			name = "LDH (a8),A"
			length = 2
			duration = 12
			address := 0xFF00 | uint16(argBytes[0])
			c.mmu.WriteByte(address, c.registers8["A"])

		case 0x17:
			name = "RLA"
			length = 1
			duration = 4
			toCarry := false
			if c.checkBit(c.registers8["A"], 7) {
				toCarry = true
			}
			c.registers8["A"] = bits.RotateLeft8(c.registers8["A"], 9)
			if c.getCarryFlag() {
				c.registers8["A"] |= 1
			}
			if toCarry {
				c.setCarryFlag()
			} else {
				c.unsetCarryFlag()
			}
			c.unsetHalfCarryFlag()
			c.unsetSubtractionFlag()

			// Jump
		case 0x18:
			name = "JR r8"
			length = 2
			duration = 12
			arg := uint16(argBytes[0])
			jump = true
			if arg > 127 {
				jumpTo = location - (256 - arg)
			} else {
				jumpTo = location + arg
			}
		case 0x20:
			name = "JR NZ,r8"
			length = 2
			duration = 12
			shortDuration = 8
			arg := uint16(argBytes[0])
			if !c.getZeroFlag() {
				jump = true
				if arg > 127 {
					jumpTo = location - (256 - arg)
				} else {
					jumpTo = location + arg
				}
			}
		case 0x28:
			name = "JR Z,r8"
			length = 2
			duration = 12
			shortDuration = 8
			arg := uint16(argBytes[0])
			if c.getZeroFlag() {
				jump = true
				if arg > 127 {
					jumpTo = location - (256 - arg)
				} else {
					jumpTo = location + arg
				}
			}

			// Stack ops
		case 0xC5:
			name = "PUSH BC"
			length = 1
			duration = 16
			data := u8PairToU16([2]uint8{c.registers8["C"], c.registers8["B"]})
			c.mmu.WriteWord(c.registers16["SP"], data)
			c.registers16["SP"] -= 2
		case 0xC1:
			name = "POP BC"
			length = 1
			duration = 12
			c.registers16["SP"] += 2
			data := c.mmu.ReadWord(c.registers16["SP"])
			bytePair := splitUint16(data)
			c.registers8["C"] = bytePair[0]
			c.registers8["B"] = bytePair[1]
		case 0xC9:
			name = "RET"
			length = 1
			duration = 20
			shortDuration = 8
			jump = true
			c.registers16["SP"] += 2
			jumpTo = c.mmu.ReadWord(c.registers16["SP"])
		case 0xCD:
			name = "CALL a16"
			length = 3
			duration = 24
			c.mmu.WriteWord(c.registers16["SP"], location)
			c.registers16["SP"] -= 2
			jump = true
			jumpTo = u8PairToU16(argBytes)

			// Misc
		case 0xCB:
			name = "CB Prefix"
			length = 1
			duration = 4
			cbFlag = true

			// Arithmetic
		case 0xAF:
			name = "XOR A"
			length = 1
			duration = 4
			c.registers8["A"] ^= c.registers8["A"]
			if c.registers8["A"] == 0 {
				c.setZeroFlag()
			} else {
				c.unsetZeroFlag()
			}
		case 0xFE:
			name = "CP d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			val := c.mmu.ReadByte(address)
			c.setSubtractionFlag()
			if c.registers8["A"] == val {
				c.setZeroFlag()
			} else {
				c.unsetZeroFlag()
			}
			if c.registers8["A"] < val {
				c.setCarryFlag()
				c.setHalfCarryFlag()
			} else {
				c.unsetCarryFlag()
				c.unsetHalfCarryFlag()
			}
		case 0xFF:
			name = "RST 38H"
			length = 1
			duration = 16
		default:
			name = "Not implemented"
			length = 1
			duration = 1
		}
	}

	// Logging
	c.printInstruction(location, opcode, name, length, duration, shortDuration)
	c.printRegisters()
	c.printFlagRegister()
	fmt.Println()

	if !jump {
		location += uint16(length)
	} else {
		location = jumpTo
	}
	return location, cbFlag
}

func (c *CPU) CBInstruction(opcode uint8, location uint16, argByte uint8) (string, uint8, uint8) {
	var name string
	var length uint8
	var duration uint8

	switch opcode {
	case 0x7C:
		name = "BIT 7,H"
		length = 2
		duration = 8
		c.unsetSubtractionFlag()
		c.setHalfCarryFlag()
		if c.checkBit(c.registers8["H"], 7) {
			c.setZeroFlag()
		} else {
			c.unsetZeroFlag()
		}
	case 0x11:
		name = "RL C"
		length = 2
		duration = 8
		toCarry := false
		if c.checkBit(c.registers8["C"], 7) {
			toCarry = true
		}
		c.registers8["C"] = bits.RotateLeft8(c.registers8["C"], 9)
		if c.getCarryFlag() {
			c.registers8["C"] |= 1
		}
		if toCarry {
			c.setCarryFlag()
		} else {
			c.unsetCarryFlag()
		}
		c.unsetHalfCarryFlag()
		c.unsetSubtractionFlag()
	}
	return name, length - 1, duration
}

func (c *CPU) setZeroFlag() {
	c.registers8["F"] |= c.bitVal(Z)
}

func (c *CPU) unsetZeroFlag() {
	c.registers8["F"] &^= c.bitVal(Z)
}

func (c *CPU) setSubtractionFlag() {
	c.registers8["F"] |= c.bitVal(N)
}

func (c *CPU) unsetSubtractionFlag() {
	c.registers8["F"] &^= c.bitVal(N)
}

func (c *CPU) setCarryFlag() {
	c.registers8["F"] |= c.bitVal(C)
}

func (c *CPU) unsetCarryFlag() {
	c.registers8["F"] &^= c.bitVal(C)
}

func (c *CPU) setHalfCarryFlag() {
	c.registers8["F"] |= c.bitVal(H)
}

func (c *CPU) unsetHalfCarryFlag() {
	c.registers8["F"] &^= c.bitVal(H)
}

func (c *CPU) getZeroFlag() bool {
	return c.checkBit(c.registers8["F"], Z)
}

func (c *CPU) getSubtractionFlag() bool {
	return c.checkBit(c.registers8["F"], N)
}

func (c *CPU) getHalfCarryFlag() bool {
	return c.checkBit(c.registers8["F"], H)
}

func (c *CPU) getCarryFlag() bool {
	return c.checkBit(c.registers8["F"], C)
}

func (c *CPU) checkBit(value uint8, bit uint8) bool {
	return (value & c.bitVal(bit)) == c.bitVal(bit)
}

func (c *CPU) bitVal(bit uint8) uint8 {
	return (1 << bit)
}

func (c *CPU) printInstruction(location uint16, opcode uint8, name string, length, duration, shortDuration uint8) {
	fmt.Printf("Byte %X\n\t%X|%s: length: %d, duration: %d/%d\n", location, opcode, name, length, duration, shortDuration)
}

func (c *CPU) printRegisters() {
	fmt.Printf("\tStack pointer: %X\n\t\tA: %X, B: %X, C: %X, D: %X, E: %X, H: %X, L: %X\n",
		c.registers16["SP"],
		c.registers8["A"],
		c.registers8["B"],
		c.registers8["C"],
		c.registers8["D"],
		c.registers8["E"],
		c.registers8["H"],
		c.registers8["L"],
	)
}

func (c *CPU) printFlagRegister() {
	fmt.Printf("\tFlag Register:\n\t\tZ: %t, N: %t, H: %t, C: %t\n",
		c.getZeroFlag(),
		c.getSubtractionFlag(),
		c.getHalfCarryFlag(),
		c.getCarryFlag(),
	)
}
