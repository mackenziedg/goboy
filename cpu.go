package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/bits"
	"os"
)

// Flag constants are the bit number of the F register corresponding to that flag.
const Z = 7
const N = 6
const H = 5
const C = 4

// The bootloader is 0x100 bytes long.
const BLSIZE = 0x100

// CPU consists of a set of registers, a pointer to the MMU unit, and bootloader data
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

// Reset links a new MMU to the CPU, clears the registers, and loads in the bootloader data from a file.
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

// Start writes the bootloader data into the 0x100-0xFFF range of the MMU and returns a stepping function.
// This function takes one CPU step each time it is called.
func (c *CPU) Start() func() {
	for i, v := range c.bootloader {
		address := uint16(i)
		c.mmu.WriteByte(address, v)
	}

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
			argBytes = U16ToU8Pair(c.mmu.ReadWord(i + 1))
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

// Inc8 increments an 8-bit register by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Inc8(register *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	*register++
	c.UnsetSubtractionFlag()
	if *register == 0 {
		c.SetZeroFlag()
		c.UnsetHalfCarryFlag()
	}
	return length, duration
}

// Dec8 increments an 8-bit register by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Dec8(register *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	*register--
	c.SetSubtractionFlag()
	c.UnsetZeroFlag()
	if *register == 0 {
		c.SetZeroFlag()
	} else if *register == 255 {
		c.SetHalfCarryFlag()
	}
	return length, duration
}

// Inc16 increments a 16-bit register pair by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Inc16(lowReg, hiReg *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	combined := U8PairToU16([2]uint8{*lowReg, *hiReg})
	combined++
	vals := U16ToU8Pair(combined)
	*lowReg = vals[0]
	*hiReg = vals[1]
	return length, duration
}

// Dec16 increments a 16-bit register pair by 1.
// It returns the length and duration of the instruction.
func (c *CPU) dec16(lowReg, hiReg *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	combined := U8PairToU16([2]uint8{*lowReg, *hiReg})
	combined--
	vals := U16ToU8Pair(combined)
	*lowReg = vals[0]
	*hiReg = vals[1]
	return length, duration
}

// LdByte reads a byte into a register.
// It returns the length and duration of the instruction.
func (c *CPU) LdByte(register *uint8, byte_ uint8) (uint8, uint8) {
	var length uint8 = 2
	var duration uint8 = 8
	*register = byte_
	return length, duration
}

// LdWord loads a 16-bit word into a register pair.
// It returns the length and duration of the instruction.
func (c *CPU) LdWord(lowReg, hiReg *uint8, word [2]uint8) (uint8, uint8) {
	var length uint8 = 3
	var duration uint8 = 12
	*lowReg = word[1]
	*hiReg = word[0]
	return length, duration
}

// LdReg8 copies the contents of a register into another.
// It returns the length and duration of the instruction.
func (c *CPU) LdReg8(to *uint8, from *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	*to = *from
	return length, duration
}

// LdReg8Adr copies the contents of a memory address into a register.
// It returns the length and duration of the instruction.
func (c *CPU) LdReg8Adr(register *uint8, address uint16) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	*register = c.mmu.ReadByte(address)
	return length, duration
}

// LdHLA copies the value of register A into the memory address specified by register pair HL.
// It returns the length and duration of the instruction.
func (c *CPU) LdHLA() (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 8
	address := U8PairToU16([2]uint8{c.L, c.H})
	c.mmu.WriteByte(address, c.A)
	return length, duration
}

// SubReg subtracts a register from A.
// It returns the length and duration of the instruction.
func (c *CPU) SubReg(register *uint8) (uint8, uint8) {
	var length uint8 = 1
	var duration uint8 = 4
	c.SetSubtractionFlag()
	c.UnsetCarryFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetZeroFlag()
	if c.A < *register {
		c.SetCarryFlag()
		c.SetHalfCarryFlag()
	} else if c.A == *register {
		c.SetZeroFlag()
	}
	c.A -= *register
	return length, duration
}

// Instruction executes one instruction, depending on the location, arguments, and if the previous byte was 0xCB.
// If breaking = true, the instruction and register data will be printed.
// Returns the new location (to load into the PC register), cbFlag if the byte is 0xCB, and whether to begin breaking.
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
			length, duration = c.Inc8(&c.B)
		case 0x05:
			name = "DEC B"
			length, duration = c.Dec8(&c.B)
		case 0x0C:
			name = "INC C"
			length, duration = c.Inc8(&c.C)
		case 0x0D:
			name = "DEC C"
			length, duration = c.Dec8(&c.C)
		case 0x15:
			name = "DEC D"
			length, duration = c.Dec8(&c.D)
		case 0x1D:
			name = "DEC E"
			length, duration = c.Dec8(&c.E)
		case 0x24:
			name = "INC H"
			length, duration = c.Inc8(&c.H)
		case 0x2C:
			name = "INC L"
			length, duration = c.Inc8(&c.L)

		case 0x13:
			name = "INC DE"
			length, duration = c.Inc16(&c.E, &c.D)
		case 0x23:
			name = "INC HL"
			length, duration = c.Inc16(&c.L, &c.H)

			// LD R,d8
		case 0x3E:
			name = "LD A,d8"
			length, duration = c.LdByte(&c.A, argBytes[0])
		case 0x06:
			name = "LD B,d8"
			length, duration = c.LdByte(&c.B, argBytes[0])
		case 0x0E:
			name = "LD C,d8"
			length, duration = c.LdByte(&c.C, argBytes[0])
		case 0x16:
			name = "LD D,d8"
			length, duration = c.LdByte(&c.D, argBytes[0])
		case 0x1E:
			name = "LD E,d8"
			length, duration = c.LdByte(&c.E, argBytes[0])

			// LD R,R
		case 0x4F:
			name = "LD C,A"
			length, duration = c.LdReg8(&c.C, &c.A)
		case 0x7B:
			name = "LD A,E"
			length, duration = c.LdReg8(&c.A, &c.E)

			// LD R,a16
		case 0x1A:
			name = "LD A,(DE)"
			length, duration = c.LdReg8Adr(&c.A, U8PairToU16([2]uint8{c.E, c.D}))

			// LD RR,d16
		case 0x11:
			name = "LD DE,d16"
			length, duration = c.LdWord(&c.E, &c.D, argBytes)
		case 0x21:
			name = "LD HL,d16"
			length, duration = c.LdWord(&c.L, &c.H, argBytes)

		case 0x32:
			name = "LD (HL-),A"
			length, duration = c.LdHLA()
			_, _ = c.dec16(&c.L, &c.H)
		case 0x22:
			name = "LD (HL+),A"
			length, duration = c.LdHLA()
			_, _ = c.Inc16(&c.L, &c.H)
		case 0x31:
			name = "LD SP,d16"
			length = 3
			duration = 12
			address := U8PairToU16(argBytes)
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
			length, duration = c.LdHLA()
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
			if CheckBit(c.A, 7) {
				toCarry = true
			}
			c.A = bits.RotateLeft8(c.A, 9)
			if c.GetCarryFlag() {
				c.A |= 1
			}
			if toCarry {
				c.SetCarryFlag()
			} else {
				c.UnsetCarryFlag()
			}
			c.UnsetHalfCarryFlag()
			c.UnsetSubtractionFlag()

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
			if !c.GetZeroFlag() {
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
			if c.GetZeroFlag() {
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
			arg := U8PairToU16(argBytes)
			jump = true
			jumpTo = arg

			// Stack ops
		case 0xC5:
			name = "PUSH BC"
			length = 1
			duration = 16
			data := U8PairToU16([2]uint8{c.C, c.B})
			c.mmu.WriteWord(c.SP, data)
			c.SP -= 2
		case 0xC1:
			name = "POP BC"
			length = 1
			duration = 12
			c.SP += 2
			data := c.mmu.ReadWord(c.SP)
			bytePair := U16ToU8Pair(data)
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
			jumpTo = U8PairToU16(argBytes)

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
			length, duration = c.SubReg(&c.L)

		case 0x96:
			name = "SUB (HL)"
			length = 1
			duration = 8
			address := U8PairToU16([2]uint8{c.L, c.H})
			sub := c.mmu.ReadByte(address)
			c.UnsetSubtractionFlag()
			c.UnsetCarryFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetZeroFlag()
			if c.A < sub {
				c.SetCarryFlag()
				c.SetHalfCarryFlag()
			} else if c.A == sub {
				c.SetZeroFlag()
			}
			c.A -= sub

		case 0xAF:
			name = "XOR A"
			length = 1
			duration = 4
			c.A ^= c.A
			if c.A == 0 {
				c.SetZeroFlag()
			} else {
				c.UnsetZeroFlag()
			}
		case 0xFE:
			name = "CP d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			val := c.mmu.ReadByte(address)
			c.SetSubtractionFlag()
			if c.A == val {
				c.SetZeroFlag()
			} else {
				c.UnsetZeroFlag()
			}
			if c.A < val {
				c.SetCarryFlag()
				c.SetHalfCarryFlag()
			} else {
				c.UnsetCarryFlag()
				c.UnsetHalfCarryFlag()
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
		c.PrintInstruction(location, opcode, name, length, duration, shortDuration)
		c.PrintRegisters()
		c.PrintFlagRegister()
		fmt.Println()
	}

	if !jump {
		location += uint16(length)
	} else {
		location = jumpTo
	}
	return location, cbFlag, breaking
}

// CBInstruction processes an instruction in the same manner as Instruction. It is called if the previous byte is 0xCB.
// It returns the opcode name, the length of the instruction minus one (need to figure out why), and the instruction duration.
func (c *CPU) CBInstruction(opcode uint8, location uint16, argByte uint8) (string, uint8, uint8) {
	var name string
	var length uint8
	var duration uint8

	switch opcode {
	case 0x7C:
		name = "BIT 7,H"
		length = 2
		duration = 8
		c.UnsetSubtractionFlag()
		c.SetHalfCarryFlag()
		if CheckBit(c.H, 7) {
			c.UnsetZeroFlag()
		} else {
			c.SetZeroFlag()
		}
	case 0x11:
		name = "RL C"
		length = 2
		duration = 8
		toCarry := false
		if CheckBit(c.C, 7) {
			toCarry = true
		}
		c.C = bits.RotateLeft8(c.C, 9)
		if c.GetCarryFlag() {
			c.C |= 1
		}
		if toCarry {
			c.SetCarryFlag()
		} else {
			c.UnsetCarryFlag()
		}
		c.UnsetHalfCarryFlag()
		c.UnsetSubtractionFlag()
	}
	return name, length - 1, duration
}

// SetZeroFlag sets the zero flag to 1.
func (c *CPU) SetZeroFlag() {
	c.F |= BitVal(Z)
}

// UnsetZeroFlag sets the zero flag to 0.
func (c *CPU) UnsetZeroFlag() {
	c.F &^= BitVal(Z)
}

// SetSubtractionFlag sets the subtraction flag to 1.
func (c *CPU) SetSubtractionFlag() {
	c.F |= BitVal(N)
}

// UnsetSubtractionFlag sets the subtraction flag to 0.
func (c *CPU) UnsetSubtractionFlag() {
	c.F &^= BitVal(N)
}

// SetCarryFlag sets the carry flag to 1.
func (c *CPU) SetCarryFlag() {
	c.F |= BitVal(C)
}

// UnsetCarryFlag sets the carry flag to 0.
func (c *CPU) UnsetCarryFlag() {
	c.F &^= BitVal(C)
}

// SetHalfCarryFlag sets the half-carry flag to 1.
func (c *CPU) SetHalfCarryFlag() {
	c.F |= BitVal(H)
}

// UnsetHalfCarryFlag sets the half-carry flag to 0.
func (c *CPU) UnsetHalfCarryFlag() {
	c.F &^= BitVal(H)
}

// GetZeroFlag returns true if the zero flag is set.
func (c *CPU) GetZeroFlag() bool {
	return CheckBit(c.F, Z)
}

// GetSubtractionFlag returns true if the subtraction flag is set.
func (c *CPU) GetSubtractionFlag() bool {
	return CheckBit(c.F, N)
}

// GetHalfCarryFlag returns true if the half-carry flag is set.
func (c *CPU) GetHalfCarryFlag() bool {
	return CheckBit(c.F, H)
}

// GetCarryFlag returns true if the carry flag is set.
func (c *CPU) GetCarryFlag() bool {
	return CheckBit(c.F, C)
}

// PrintInstruction prints the byte location, opcode, opcode name, length, duration, and shortDuration of the passed instruction.
func (c *CPU) PrintInstruction(location uint16, opcode uint8, name string, length, duration, shortDuration uint8) {
	fmt.Printf("Byte %X\n\t%X|%s: length: %d, duration: %d/%d\n", location, opcode, name, length, duration, shortDuration)
}

// PrintRegisters prints the stack pointer location and data, and the values in each register except the flag register.
func (c *CPU) PrintRegisters() {
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

// PrintFlagRegister prints whether each flag in the register is set.
func (c *CPU) PrintFlagRegister() {
	fmt.Printf("\tFlag Register:\n\t\tZ: %t, N: %t, H: %t, C: %t\n",
		c.GetZeroFlag(),
		c.GetSubtractionFlag(),
		c.GetHalfCarryFlag(),
		c.GetCarryFlag(),
	)
}
