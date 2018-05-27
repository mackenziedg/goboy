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

type CPU struct {
	A uint8
	B uint8
	C uint8
	D uint8
	E uint8
	F uint8
	H uint8
	L uint8

	PC uint16
	SP uint16

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

	c.A = 0x0
	c.B = 0x0
	c.C = 0x0
	c.D = 0x0
	c.E = 0x0
	c.H = 0x0
	c.L = 0x0

	c.SP = 0x0
	c.PC = 0x0

	c.mmu = mmu
}

func (c *CPU) LoadCartridgeData(data []byte) {
	for i := 0x0100; i < 0x8000; i++ {
		address := uint16(i)
		c.mmu.WriteByte(address, data[i])
	}
	for i, v := range c.bootloader {
		address := uint16(i)
		c.mmu.WriteByte(address, v)
	}
}

func (c *CPU) Stepper() func() {
	reader := bufio.NewReader(os.Stdin)

	var i uint16
	cb := false
	breaking := false
	var insCount = 0

	return func() {
		insCount++
		i = c.PC
		argBytes := [2]uint8{0, 0}
		if i < MEMORYSIZE-2 {
			argBytes = splitUint16(c.mmu.ReadWord(i + 1))
		} else if i < MEMORYSIZE-1 {
			argBytes = [2]uint8{c.mmu.ReadByte(i + 1), 0}
		}

		i, cb, breaking = c.Instruction(c.mmu.ReadByte(i), i, argBytes, cb, breaking)
		c.PC = i
		if breaking {
			reader.ReadString('\n')
		}
	}
}

func (c *CPU) inc8(register *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	*register++
	c.unsetSubtractionFlag()
	if *register == 0 {
		c.setZeroFlag()
		c.setHalfCarryFlag()
	}
	return length, duration
}

func (c *CPU) dec8(register *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	*register--
	c.setSubtractionFlag()
	c.unsetZeroFlag()
	if *register == 0 {
		c.setZeroFlag()
	} else if *register == 255 {
		c.setHalfCarryFlag()
	}
	return length, duration
}

// pass low, high
func (c *CPU) inc16(lowReg, hiReg *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	combined := u8PairToU16([2]uint8{*lowReg, *hiReg})
	combined++
	vals := splitUint16(combined)
	*lowReg = vals[0]
	*hiReg = vals[1]
	return length, duration
}

// pass low, high
func (c *CPU) dec16(lowReg, hiReg *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	combined := u8PairToU16([2]uint8{*lowReg, *hiReg})
	combined--
	vals := splitUint16(combined)
	*lowReg = vals[0]
	*hiReg = vals[1]
	return length, duration
}

func (c *CPU) ldByte(register *uint8, byte_ uint8) (uint8, uint8) {
	var length uint8 = 2
	var duration uint8 = 8
	*register = byte_
	return length, duration
}

func (c *CPU) ldWord(lowReg, hiReg *uint8, word [2]uint8) (uint8, uint8) {
	var length uint8 = 3
	var duration uint8 = 12
	*lowReg = word[1]
	*hiReg = word[0]
	return length, duration
}

func (c *CPU) ldReg8(to *uint8, from *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	*to = *from
	return length, duration
}

func (c *CPU) ldReg8Adr(register *uint8, address uint16) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	*register = c.mmu.ReadByte(address)
	return length, duration
}

func (c *CPU) ldHLA() (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	address := u8PairToU16([2]uint8{c.L, c.H})
	c.mmu.WriteByte(address, c.A)
	return length, duration
}

func (c *CPU) subReg(register *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	c.setSubtractionFlag()
	c.unsetCarryFlag()
	c.unsetHalfCarryFlag()
	c.unsetZeroFlag()
	if c.A < *register {
		c.setCarryFlag()
		c.setHalfCarryFlag()
	} else if c.A == *register {
		c.setZeroFlag()
	}
	c.A -= *register
	return length, duration
}

func (c *CPU) Instruction(opcode uint8, location uint16, argBytes [2]uint8, cbFlag bool, breaking bool) (uint16, bool, bool) {
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
			length, duration = c.inc8(&c.B)
		case 0x05:
			name = "DEC B"
			length, duration = c.dec8(&c.B)
		case 0x0C:
			name = "INC C"
			length, duration = c.inc8(&c.C)
		case 0x0D:
			name = "DEC C"
			length, duration = c.dec8(&c.C)
		case 0x15:
			name = "DEC D"
			length, duration = c.dec8(&c.D)
		case 0x1D:
			name = "DEC E"
			length, duration = c.dec8(&c.E)
		case 0x24:
			name = "INC H"
			length, duration = c.inc8(&c.H)
		case 0x2C:
			name = "INC L"
			length, duration = c.inc8(&c.L)

		case 0x13:
			name = "INC DE"
			length, duration = c.inc16(&c.E, &c.D)
		case 0x23:
			name = "INC HL"
			length, duration = c.inc16(&c.L, &c.H)

			// LD R,d8
		case 0x3E:
			name = "LD A,d8"
			length, duration = c.ldByte(&c.A, argBytes[0])
		case 0x06:
			name = "LD B,d8"
			length, duration = c.ldByte(&c.B, argBytes[0])
		case 0x0E:
			name = "LD C,d8"
			length, duration = c.ldByte(&c.C, argBytes[0])
		case 0x16:
			name = "LD D,d8"
			length, duration = c.ldByte(&c.D, argBytes[0])
		case 0x1E:
			name = "LD E,d8"
			length, duration = c.ldByte(&c.E, argBytes[0])

			// LD R,R
		case 0x4F:
			name = "LD C,A"
			length, duration = c.ldReg8(&c.C, &c.A)
		case 0x7B:
			name = "LD A,E"
			length, duration = c.ldReg8(&c.A, &c.E)

			// LD R,a16
		case 0x1A:
			name = "LD A,(DE)"
			length, duration = c.ldReg8Adr(&c.A, u8PairToU16([2]uint8{c.E, c.D}))

			// LD RR,d16
		case 0x11:
			name = "LD DE,d16"
			length, duration = c.ldWord(&c.E, &c.D, argBytes)
		case 0x21:
			name = "LD HL,d16"
			length, duration = c.ldWord(&c.L, &c.H, argBytes)

		case 0x32:
			name = "LD (HL-),A"
			length, duration = c.ldHLA()
			_, _ = c.dec16(&c.L, &c.H)
		case 0x22:
			name = "LD (HL+),A"
			length, duration = c.ldHLA()
			_, _ = c.inc16(&c.L, &c.H)
		case 0x31:
			name = "LD SP,d16"
			length = 3
			duration = 12
			address := u8PairToU16(argBytes)
			c.SP = address

			// LD address,R
		case 0xE2:
			name = "LD (C),A"
			length = 2
			duration = 8
			address := 0xFF00 | uint16(c.C)
			c.mmu.WriteByte(address, c.A)
			length = 1 // Change later when figure out wtf
		case 0x77:
			name = "LD (HL),A"
			length, duration = c.ldHLA()
		case 0xE0:
			name = "LDH (a8),A"
			length = 2
			duration = 12
			address := 0xFF00 | uint16(argBytes[0])
			c.mmu.WriteByte(address, c.A)

		case 0x17:
			name = "RLA"
			length = 1
			duration = 4
			toCarry := false
			if checkBit(c.A, 7) {
				toCarry = true
			}
			c.A = bits.RotateLeft8(c.A, 9)
			if c.getCarryFlag() {
				c.A |= 1
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
					jumpTo = location - (255 - arg) + 1
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
		case 0xC3:
			name = "JP a16"
			length = 3
			duration = 16
			arg := u8PairToU16(argBytes)
			jump = true
			jumpTo = arg

			// Stack ops
		case 0xC5:
			name = "PUSH BC"
			length = 1
			duration = 16
			data := u8PairToU16([2]uint8{c.C, c.B})
			c.mmu.WriteWord(c.SP, data)
			c.SP -= 2
		case 0xC1:
			name = "POP BC"
			length = 1
			duration = 12
			c.SP += 2
			data := c.mmu.ReadWord(c.SP)
			bytePair := splitUint16(data)
			c.C = bytePair[0]
			c.B = bytePair[1]
		case 0xC9:
			name = "RET"
			length = 1
			duration = 20
			shortDuration = 8
			jump = true
			c.SP += 2
			// fmt.Printf("Returning to: %X \n", c.mmu.ReadWord(c.SP))
			jumpTo = c.mmu.ReadWord(c.SP)
		case 0xCD:
			name = "CALL a16"
			length = 3
			duration = 24
			c.mmu.WriteWord(c.SP, location)
			c.SP -= 2
			jump = true
			jumpTo = u8PairToU16(argBytes)

			// Misc
		case 0x00:
			name = "NOP"
			length = 1
			duration = 4
		case 0xCB:
			name = "CB Prefix"
			length = 1
			duration = 4
			cbFlag = true

			// Arithmetic
		case 0x95:
			name = "SUB L"
			length, duration = c.subReg(&c.L)

		case 0x96:
			name = "SUB (HL)"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{c.L, c.H})
			sub := c.mmu.ReadByte(address)
			c.setSubtractionFlag()
			c.unsetCarryFlag()
			c.unsetHalfCarryFlag()
			c.unsetZeroFlag()
			if c.A < sub {
				c.setCarryFlag()
				c.setHalfCarryFlag()
			} else if c.A == sub {
				c.setZeroFlag()
			}
			c.A -= sub

		case 0xAF:
			name = "XOR A"
			length = 1
			duration = 4
			c.A ^= c.A
			if c.A == 0 {
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
			if c.A == val {
				c.setZeroFlag()
			} else {
				c.unsetZeroFlag()
			}
			if c.A < val {
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
			breaking = true
			// for i := 0x8000; i < 0x97FF; i++ {
			// 	fmt.Printf("%X ", c.mmu.ReadByte(uint16(i)))
			// }
			// fmt.Println()
		}
	}

	// Logging
	if breaking {
		c.printInstruction(location, opcode, name, length, duration, shortDuration)
		c.printRegisters()
		c.printFlagRegister()
		fmt.Println()
	}

	if !jump {
		location += uint16(length)
	} else {
		location = jumpTo
	}
	return location, cbFlag, breaking
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
		if checkBit(c.H, 7) {
			c.unsetZeroFlag()
		} else {
			c.setZeroFlag()
		}
	case 0x11:
		name = "RL C"
		length = 2
		duration = 8
		toCarry := false
		if checkBit(c.C, 7) {
			toCarry = true
		}
		c.C = bits.RotateLeft8(c.C, 9)
		if c.getCarryFlag() {
			c.C |= 1
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
	c.F |= bitVal(Z)
}

func (c *CPU) unsetZeroFlag() {
	c.F &^= bitVal(Z)
}

func (c *CPU) setSubtractionFlag() {
	c.F |= bitVal(N)
}

func (c *CPU) unsetSubtractionFlag() {
	c.F &^= bitVal(N)
}

func (c *CPU) setCarryFlag() {
	c.F |= bitVal(C)
}

func (c *CPU) unsetCarryFlag() {
	c.F &^= bitVal(C)
}

func (c *CPU) setHalfCarryFlag() {
	c.F |= bitVal(H)
}

func (c *CPU) unsetHalfCarryFlag() {
	c.F &^= bitVal(H)
}

func (c *CPU) getZeroFlag() bool {
	return checkBit(c.F, Z)
}

func (c *CPU) getSubtractionFlag() bool {
	return checkBit(c.F, N)
}

func (c *CPU) getHalfCarryFlag() bool {
	return checkBit(c.F, H)
}

func (c *CPU) getCarryFlag() bool {
	return checkBit(c.F, C)
}

func (c *CPU) printInstruction(location uint16, opcode uint8, name string, length, duration, shortDuration uint8) {
	fmt.Printf("Byte %X\n\t%X|%s: length: %d, duration: %d/%d\n", location, opcode, name, length, duration, shortDuration)
}

func (c *CPU) printRegisters() {
	fmt.Printf("\tStack pointer: %X ($%X) \n\t\tA: %X, B: %X, C: %X, D: %X, E: %X, H: %X, L: %X\n",
		c.SP,
		c.mmu.ReadWord(c.SP),
		c.A,
		c.B,
		c.C,
		c.D,
		c.E,
		c.H,
		c.L,
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
