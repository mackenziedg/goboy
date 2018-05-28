package main

import (
	"fmt"
	"io/ioutil"
	"math/bits"
	"unsafe"
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
	AF Register
	BC Register
	DE Register
	HL Register

	PC Register
	SP Register

	mmu *MMU

	bootloader [0x100]byte
}

type Instruction struct {
	name          string
	location      uint16
	arg           uint16
	opcode        uint8
	length        uint8
	duration      uint8
	shortDuration uint8
}

type Register struct {
	hi   *uint8
	lo   *uint8
	word uint16
}

// Reset links a new MMU to the CPU, clears the registers, and loads in the bootloader data from a file.
func (c *CPU) Reset(mmu *MMU) {
	dat, err := ioutil.ReadFile("./data/DMG_ROM.bin")
	check(err)

	for i, v := range dat {
		c.bootloader[i] = v
	}

	c.AF.word = 0x0
	c.AF.lo = (*uint8)(unsafe.Pointer(&c.AF.word))
	c.AF.hi = (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&c.AF.word)) + unsafe.Sizeof(uint8(0))))

	c.BC.word = 0x0
	c.BC.lo = (*uint8)(unsafe.Pointer(&c.BC.word))
	c.BC.hi = (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&c.BC.word)) + unsafe.Sizeof(uint8(0))))

	c.DE.word = 0x0
	c.DE.lo = (*uint8)(unsafe.Pointer(&c.DE.word))
	c.DE.hi = (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&c.DE.word)) + unsafe.Sizeof(uint8(0))))

	c.HL.word = 0x0
	c.HL.lo = (*uint8)(unsafe.Pointer(&c.HL.word))
	c.HL.hi = (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&c.HL.word)) + unsafe.Sizeof(uint8(0))))

	c.SP.word = 0x0
	c.PC.word = 0x0

	c.mmu = mmu
}

// Inc8 increments an 8-bit register by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Inc8(register *uint8) (uint8, uint8) {
	*register++
	c.UnsetSubtractionFlag()
	if *register == 0 {
		c.SetZeroFlag()
		c.UnsetHalfCarryFlag()
	}
	return 1, 4
}

// Dec8 increments an 8-bit register by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Dec8(register *uint8) (uint8, uint8) {
	*register--
	c.SetSubtractionFlag()
	c.UnsetZeroFlag()
	if *register == 0 {
		c.SetZeroFlag()
	} else if *register == 255 {
		c.SetHalfCarryFlag()
	}
	return 1, 4
}

// Inc16 increments a 16-bit register pair by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Inc16(registerPair *uint16) (uint8, uint8) {
	*registerPair++
	return 1, 8
}

// Dec16 increments a 16-bit register pair by 1.
// It returns the length and duration of the instruction.
func (c *CPU) Dec16(registerPair *uint16) (uint8, uint8) {
	*registerPair--
	return 1, 8
}

// LdByte reads a byte into a register.
// It returns the length and duration of the instruction.
func (c *CPU) LdByte(register *uint8, byte_ uint8) (uint8, uint8) {
	*register = byte_
	return 2, 8
}

// LdWord loads a 16-bit word into a register pair.
// It returns the length and duration of the instruction.
func (c *CPU) LdWord(registerPair *uint16, word uint16) (uint8, uint8) {
	*registerPair = word
	return 3, 12
}

// LdReg8 copies the contents of a register into another.
// It returns the length and duration of the instruction.
func (c *CPU) LdReg8(to *uint8, from *uint8) (uint8, uint8) {
	*to = *from
	return 1, 4
}

// LdReg8Adr copies the contents of a memory address into a register.
// It returns the length and duration of the instruction.
func (c *CPU) LdReg8Adr(register *uint8, address uint16) (uint8, uint8) {
	*register = c.mmu.ReadByte(address)
	return 1, 8
}

// LdHLA copies the value of register A into the memory address specified.
// It returns the length and duration of the instruction.
func (c *CPU) LdAdrA(address uint16) (uint8, uint8) {
	c.mmu.WriteByte(address, *c.AF.hi)
	return 1, 8
}

// AddReg8 adds a register to A.
// It returns the length and duration of the instruction.
func (c *CPU) AddReg8(register *uint8) (uint8, uint8) {
	c.UnsetSubtractionFlag()
	c.UnsetCarryFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetZeroFlag()

	*c.AF.hi += *register

	if *c.AF.hi < *register {
		c.SetCarryFlag()
		c.SetHalfCarryFlag()
	}
	if *c.AF.hi == 0 {
		c.SetZeroFlag()
	}
	return 1, 4
}

// AddReg16 adds a register to HL, storing the result in HL.
// It returns the length and duration of the instruction.
func (c *CPU) AddReg16(word *uint16) (uint8, uint8) {
	c.UnsetSubtractionFlag()
	c.UnsetCarryFlag()
	c.UnsetHalfCarryFlag()

	c.HL.word += *word
	return 1, 8
}

// SubReg subtracts a register from A.
// It returns the length and duration of the instruction.
func (c *CPU) SubReg(register *uint8) (uint8, uint8) {
	c.SetSubtractionFlag()
	c.UnsetCarryFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetZeroFlag()
	if *c.AF.hi < *register {
		c.SetCarryFlag()
		c.SetHalfCarryFlag()
	} else if *c.AF.hi == *register {
		c.SetZeroFlag()
	}
	*c.AF.hi -= *register
	return 1, 4
}

// Xors a register with register A and stores the result in A.
// Returns the length and duration of the instruction.
func (c *CPU) XorReg(register *uint8) (uint8, uint8) {
	*c.AF.hi ^= *register
	if *c.AF.hi == 0 {
		c.SetZeroFlag()
	} else {
		c.UnsetZeroFlag()
	}
	c.UnsetSubtractionFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetCarryFlag()
	return 1, 4
}

// Ors a register with register A and stores the result in A.
// Returns the length and duration of the instruction.
func (c *CPU) OrReg(register *uint8) (uint8, uint8) {
	*c.AF.hi |= *register
	if *c.AF.hi == 0 {
		c.SetZeroFlag()
	} else {
		c.UnsetZeroFlag()
	}
	c.UnsetSubtractionFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetCarryFlag()
	return 1, 4
}

// Start writes the bootloader data into the 0x100-0xFFF range of the MMU and returns a stepping function.
// This function takes one CPU step each time it is called.
func (c *CPU) Start() func() uint8 {
	for i, v := range c.bootloader {
		address := uint16(i)
		c.mmu.WriteByte(address, v)
	}

	cb := false
	breaking := true
	curIns := Instruction{}
	var insCount uint64 = 0

	return func() uint8 {
		insCount++

		curIns, cb, breaking = c.Instruction(c.mmu.ReadByte(c.PC.word), cb, breaking)
		if curIns.location > 0x54 && curIns.location < 0x94 {
			c.PrintInstruction(insCount, curIns)
			c.PrintRegisters()
			c.PrintFlagRegister()
			fmt.Scanln()
		}
		return curIns.duration
	}
}

// Instruction executes one instruction, depending on the location, arguments, and if the previous byte was 0xCB.
// If breaking = true, the instruction and register data will be printed.
// Returns the current Instruction, cbFlag if the byte is 0xCB, and whether to begin breaking.
func (c *CPU) Instruction(opcode uint8, cbFlag bool, breaking bool) (Instruction, bool, bool) {
	var name string
	var length uint8
	var location = c.PC.word
	var duration, shortDuration uint8
	var jump bool
	var jumpTo uint16

	// If the previous byte was 0xCB we process the instruction from the CB table
	if cbFlag {
		name, length, duration = c.CBInstruction(opcode, c.mmu.ReadByte(c.PC.word+1))
		cbFlag = false
	} else {

		// TODO: Probably should be a map[uint8]func for speed or array of func pointers
		switch opcode {

		// INC/DEC
		case 0x3C:
			name = "INC A"
			length, duration = c.Inc8(c.AF.hi)
		case 0x3D:
			name = "DEC A"
			length, duration = c.Dec8(c.AF.hi)
		case 0x04:
			name = "INC B"
			length, duration = c.Inc8(c.BC.hi)
		case 0x05:
			name = "DEC B"
			length, duration = c.Dec8(c.BC.hi)
		case 0x0C:
			name = "INC C"
			length, duration = c.Inc8(c.BC.lo)
		case 0x0D:
			name = "DEC C"
			length, duration = c.Dec8(c.BC.lo)
		case 0x14:
			name = "INC D"
			length, duration = c.Inc8(c.DE.hi)
		case 0x15:
			name = "DEC D"
			length, duration = c.Dec8(c.DE.hi)
		case 0x1C:
			name = "INC E"
			length, duration = c.Inc8(c.DE.lo)
		case 0x1D:
			name = "DEC E"
			length, duration = c.Dec8(c.DE.lo)
		case 0x24:
			name = "INC H"
			length, duration = c.Inc8(c.HL.hi)
		case 0x25:
			name = "DEC H"
			length, duration = c.Dec8(c.HL.hi)
		case 0x2C:
			name = "INC L"
			length, duration = c.Inc8(c.HL.lo)
		case 0x2D:
			name = "DEC L"
			length, duration = c.Inc8(c.HL.lo)
		case 0x03:
			name = "INC BC"
			length, duration = c.Inc16(&c.BC.word)
		case 0x13:
			name = "DEC DE"
			length, duration = c.Inc16(&c.DE.word)
		case 0x23:
			name = "INC HL"
			length, duration = c.Inc16(&c.HL.word)
		case 0x0B:
			name = "DEC BC"
			length, duration = c.Dec16(&c.BC.word)
		case 0x1B:
			name = "DEC DE"
			length, duration = c.Dec16(&c.DE.word)
		case 0x2B:
			name = "DEC HL"
			length, duration = c.Dec16(&c.HL.word)

			// LD R,d8
		case 0x3E:
			name = "LD A,d8"
			length, duration = c.LdByte(c.AF.hi, c.mmu.ReadByte(c.PC.word+1))
		case 0x06:
			name = "LD B,d8"
			length, duration = c.LdByte(c.BC.hi, c.mmu.ReadByte(c.PC.word+1))
		case 0x0E:
			name = "LD C,d8"
			length, duration = c.LdByte(c.BC.lo, c.mmu.ReadByte(c.PC.word+1))
		case 0x16:
			name = "LD D,d8"
			length, duration = c.LdByte(c.DE.hi, c.mmu.ReadByte(c.PC.word+1))
		case 0x1E:
			name = "LD E,d8"
			length, duration = c.LdByte(c.DE.lo, c.mmu.ReadByte(c.PC.word+1))
		case 0x26:
			name = "LD H,d8"
			length, duration = c.LdByte(c.HL.hi, c.mmu.ReadByte(c.PC.word+1))
		case 0x2E:
			name = "LD L,d8"
			length, duration = c.LdByte(c.HL.lo, c.mmu.ReadByte(c.PC.word+1))

			// LD R,R
		case 0x7F:
			name = "LD A,A"
			length, duration = c.LdReg8(c.AF.hi, c.AF.hi)
		case 0x78:
			name = "LD A,B"
			length, duration = c.LdReg8(c.AF.hi, c.BC.hi)
		case 0x79:
			name = "LD A,C"
			length, duration = c.LdReg8(c.AF.hi, c.BC.lo)
		case 0x7A:
			name = "LD A,D"
			length, duration = c.LdReg8(c.AF.hi, c.DE.hi)
		case 0x7B:
			name = "LD A,E"
			length, duration = c.LdReg8(c.AF.hi, c.DE.lo)
		case 0x7C:
			name = "LD A,H"
			length, duration = c.LdReg8(c.AF.hi, c.HL.hi)
		case 0x7D:
			name = "LD A,L"
			length, duration = c.LdReg8(c.AF.hi, c.HL.lo)
		case 0x47:
			name = "LD B,A"
			length, duration = c.LdReg8(c.BC.hi, c.AF.hi)
		case 0x40:
			name = "LD B,B"
			length, duration = c.LdReg8(c.BC.hi, c.BC.hi)
		case 0x41:
			name = "LD B,C"
			length, duration = c.LdReg8(c.BC.hi, c.BC.lo)
		case 0x42:
			name = "LD B,D"
			length, duration = c.LdReg8(c.BC.hi, c.DE.hi)
		case 0x43:
			name = "LD B,E"
			length, duration = c.LdReg8(c.BC.hi, c.DE.lo)
		case 0x44:
			name = "LD B,H"
			length, duration = c.LdReg8(c.BC.hi, c.HL.hi)
		case 0x45:
			name = "LD B,L"
			length, duration = c.LdReg8(c.BC.hi, c.HL.lo)
		case 0x4F:
			name = "LD C,A"
			length, duration = c.LdReg8(c.BC.lo, c.AF.hi)
		case 0x48:
			name = "LD C,B"
			length, duration = c.LdReg8(c.BC.lo, c.BC.hi)
		case 0x49:
			name = "LD C,C"
			length, duration = c.LdReg8(c.BC.lo, c.BC.lo)
		case 0x4A:
			name = "LD C,D"
			length, duration = c.LdReg8(c.BC.lo, c.DE.hi)
		case 0x4B:
			name = "LD C,E"
			length, duration = c.LdReg8(c.BC.lo, c.DE.lo)
		case 0x4C:
			name = "LD C,H"
			length, duration = c.LdReg8(c.BC.lo, c.HL.hi)
		case 0x4D:
			name = "LD C,L"
			length, duration = c.LdReg8(c.BC.lo, c.HL.lo)
		case 0x57:
			name = "LD D,A"
			length, duration = c.LdReg8(c.DE.hi, c.AF.hi)
		case 0x50:
			name = "LD D,B"
			length, duration = c.LdReg8(c.DE.hi, c.BC.hi)
		case 0x51:
			name = "LD D,C"
			length, duration = c.LdReg8(c.DE.hi, c.BC.lo)
		case 0x52:
			name = "LD D,D"
			length, duration = c.LdReg8(c.DE.hi, c.DE.hi)
		case 0x53:
			name = "LD D,E"
			length, duration = c.LdReg8(c.DE.hi, c.DE.lo)
		case 0x54:
			name = "LD D,H"
			length, duration = c.LdReg8(c.DE.hi, c.HL.hi)
		case 0x55:
			name = "LD D,L"
			length, duration = c.LdReg8(c.DE.hi, c.HL.lo)
		case 0x5F:
			name = "LD E,A"
			length, duration = c.LdReg8(c.DE.lo, c.AF.hi)
		case 0x58:
			name = "LD E,B"
			length, duration = c.LdReg8(c.DE.lo, c.BC.hi)
		case 0x59:
			name = "LD E,C"
			length, duration = c.LdReg8(c.DE.lo, c.BC.lo)
		case 0x5A:
			name = "LD E,D"
			length, duration = c.LdReg8(c.DE.lo, c.DE.hi)
		case 0x5B:
			name = "LD E,E"
			length, duration = c.LdReg8(c.DE.lo, c.DE.lo)
		case 0x5C:
			name = "LD E,H"
			length, duration = c.LdReg8(c.DE.lo, c.HL.hi)
		case 0x5D:
			name = "LD E,L"
			length, duration = c.LdReg8(c.DE.lo, c.HL.lo)
		case 0x67:
			name = "LD H,A"
			length, duration = c.LdReg8(c.HL.hi, c.AF.hi)
		case 0x60:
			name = "LD H,B"
			length, duration = c.LdReg8(c.HL.hi, c.BC.hi)
		case 0x61:
			name = "LD H,C"
			length, duration = c.LdReg8(c.HL.hi, c.BC.lo)
		case 0x62:
			name = "LD H,D"
			length, duration = c.LdReg8(c.HL.hi, c.DE.hi)
		case 0x63:
			name = "LD H,E"
			length, duration = c.LdReg8(c.HL.hi, c.DE.lo)
		case 0x64:
			name = "LD H,H"
			length, duration = c.LdReg8(c.HL.hi, c.HL.hi)
		case 0x65:
			name = "LD H,L"
			length, duration = c.LdReg8(c.HL.hi, c.HL.lo)
		case 0x6F:
			name = "LD L,A"
			length, duration = c.LdReg8(c.HL.lo, c.AF.hi)
		case 0x68:
			name = "LD L,B"
			length, duration = c.LdReg8(c.HL.lo, c.BC.hi)
		case 0x69:
			name = "LD L,C"
			length, duration = c.LdReg8(c.HL.lo, c.BC.lo)
		case 0x6A:
			name = "LD L,D"
			length, duration = c.LdReg8(c.HL.lo, c.DE.hi)
		case 0x6B:
			name = "LD L,E"
			length, duration = c.LdReg8(c.HL.lo, c.DE.lo)
		case 0x6C:
			name = "LD L,H"
			length, duration = c.LdReg8(c.HL.lo, c.HL.hi)
		case 0x6D:
			name = "LD L,L"
			length, duration = c.LdReg8(c.HL.lo, c.HL.lo)

			// LD R,a16
		case 0x0A:
			name = "LD A,(BC)"
			length, duration = c.LdReg8Adr(c.AF.hi, c.BC.word)
		case 0x1A:
			name = "LD A,(DE)"
			length, duration = c.LdReg8Adr(c.AF.hi, c.DE.word)
		case 0x7E:
			name = "LD A,(HL)"
			length, duration = c.LdReg8Adr(c.AF.hi, c.HL.word)
		case 0x2A:
			name = "LD A,(HL+)"
			length, duration = c.LdReg8Adr(c.AF.hi, c.HL.word)
			_, _ = c.Dec16(&c.HL.word)
		case 0x3A:
			name = "LD A,(HL-)"
			length, duration = c.LdReg8Adr(c.AF.hi, c.HL.word)
			_, _ = c.Dec16(&c.HL.word)
		case 0xF0:
			name = "LDH A,a8"
			length = 2
			duration = 12
			address := 0xFF00 | uint16(c.mmu.ReadByte(c.PC.word+1))
			*c.AF.hi = c.mmu.ReadByte(address)

			// LD RR,d16
		case 0x01:
			name = "LD BC,d16"
			length, duration = c.LdWord(&c.BC.word, c.mmu.ReadWord(c.PC.word+1))
		case 0x11:
			name = "LD DE,d16"
			length, duration = c.LdWord(&c.DE.word, c.mmu.ReadWord(c.PC.word+1))
		case 0x21:
			name = "LD HL,d16"
			length, duration = c.LdWord(&c.HL.word, c.mmu.ReadWord(c.PC.word+1))

		case 0x31:
			name = "LD SP,d16"
			length = 3
			duration = 12
			c.SP.word = c.mmu.ReadWord(c.PC.word + 1)

			// LD address,R
		case 0xE2:
			name = "LD (C),A"
			length = 2
			duration = 8
			address := 0xFF00 | uint16(*c.BC.lo)
			c.mmu.WriteByte(address, *c.AF.hi)
			length = 1 // Change later when figure out wtf
		case 0x02:
			name = "LD (BC),A"
			length, duration = c.LdAdrA(c.BC.word)
		case 0x12:
			name = "LD (DE),A"
			length, duration = c.LdAdrA(c.DE.word)
		case 0x77:
			name = "LD (HL),A"
			length, duration = c.LdAdrA(c.HL.word)
		case 0x32:
			name = "LD (HL-),A"
			length, duration = c.LdAdrA(c.HL.word)
			_, _ = c.Dec16(&c.HL.word)
		case 0x22:
			name = "LD (HL+),A"
			length, duration = c.LdAdrA(c.HL.word)
			_, _ = c.Inc16(&c.HL.word)
		case 0xE0:
			name = "LDH (a8),A"
			length = 2
			duration = 12
			address := 0xFF00 | uint16(c.mmu.ReadByte(c.PC.word+1))
			c.mmu.WriteByte(address, *c.AF.hi)
		case 0xEA:
			name = "LD a16,A"
			length = 3
			duration = 16
			c.mmu.WriteByte(c.mmu.ReadWord(c.PC.word+1), *c.AF.hi)

			// Jump
		case 0x18:
			name = "JR r8"
			length = 2
			duration = 12
			arg := uint16(c.mmu.ReadByte(c.PC.word + 1))
			jump = true
			if arg > 127 {
				jumpTo = c.PC.word - (255 - arg) + 1
			} else {
				jumpTo = c.PC.word + arg + uint16(length)
			}
		case 0x20:
			name = "JR NZ,r8"
			length = 2
			duration = 12
			shortDuration = 8
			arg := uint16(c.mmu.ReadByte(c.PC.word + 1))
			if !c.GetZeroFlag() {
				jump = true
				if arg > 127 {
					jumpTo = c.PC.word - (255 - arg) + 1
				} else {
					jumpTo = c.PC.word + arg + uint16(length)
				}
			}
		case 0x28:
			name = "JR Z,r8"
			length = 2
			duration = 12
			shortDuration = 8
			arg := uint16(c.mmu.ReadByte(c.PC.word + 1))
			if c.GetZeroFlag() {
				jump = true
				if arg > 127 {
					jumpTo = c.PC.word - (255 - arg) + 1
				} else {
					jumpTo = c.PC.word + arg + uint16(length)
				}
			}
		case 0xC3:
			name = "JP a16"
			length = 3
			duration = 16
			jump = true
			jumpTo = c.mmu.ReadWord(c.PC.word + 1)

			// Stack ops
		case 0xC5:
			name = "PUSH BC"
			length = 1
			duration = 16
			c.mmu.WriteWord(c.SP.word, c.BC.word)
			c.SP.word -= 2
		case 0xC1:
			name = "POP BC"
			length = 1
			duration = 12
			c.SP.word += 2
			c.BC.word = c.mmu.ReadWord(c.SP.word)
		case 0xC9:
			name = "RET"
			length = 1
			duration = 20
			shortDuration = 8
			jump = true
			c.SP.word += 2
			jumpTo = c.mmu.ReadWord(c.SP.word)
		case 0xCD:
			name = "CALL a16"
			length = 3
			duration = 24
			c.mmu.WriteWord(c.SP.word, c.PC.word+uint16(length))
			c.SP.word -= 2
			jump = true
			jumpTo = c.mmu.ReadWord(c.PC.word + 1)

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
		case 0x07:
			name = "RLCA"
			length = 1
			duration = 4
			if CheckBit(c.AF.hi, 7) {
				c.SetCarryFlag()
			} else {
				c.UnsetCarryFlag()
			}
			*c.AF.hi = bits.RotateLeft8(*c.AF.hi, 8)
			c.UnsetZeroFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetSubtractionFlag()
		case 0x17:
			name = "RLA"
			length = 1
			duration = 4
			toCarry := false
			if CheckBit(c.AF.hi, 7) {
				toCarry = true
			}
			*c.AF.hi = bits.RotateLeft8(*c.AF.hi, 9)
			if c.GetCarryFlag() {
				*c.AF.hi |= 1
			}
			if toCarry {
				c.SetCarryFlag()
			} else {
				c.UnsetCarryFlag()
			}
			c.UnsetHalfCarryFlag()
			c.UnsetSubtractionFlag()
		case 0x0F:
			name = "RRCA"
			length = 1
			duration = 4
			if CheckBit(c.AF.hi, 0) {
				c.SetCarryFlag()
			} else {
				c.UnsetCarryFlag()
			}
			*c.AF.hi = bits.RotateLeft8(*c.AF.hi, -8)
			c.UnsetZeroFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetSubtractionFlag()
		case 0x1F:
			name = "RRA"
			length = 1
			duration = 4
			toCarry := false
			if CheckBit(c.AF.hi, 1) {
				toCarry = true
			}
			*c.AF.hi = bits.RotateLeft8(*c.AF.hi, -9)
			if c.GetCarryFlag() {
				*c.AF.hi |= BitVal(7)
			}
			if toCarry {
				c.SetCarryFlag()
			} else {
				c.UnsetCarryFlag()
			}
			c.UnsetHalfCarryFlag()
			c.UnsetSubtractionFlag()

		case 0x09:
			name = "ADD HL,BC"
			length, duration = c.AddReg16(&c.BC.word)
		case 0x19:
			name = "ADD HL,DE"
			length, duration = c.AddReg16(&c.DE.word)
		case 0x29:
			name = "ADD HL,HL"
			length, duration = c.AddReg16(&c.HL.word)
		case 0x39:
			name = "ADD HL,SP"
			length, duration = c.AddReg16(&c.SP.word)
		case 0x80:
			name = "ADD B"
			length, duration = c.AddReg8(c.BC.hi)
		case 0x81:
			name = "ADD C"
			length, duration = c.AddReg8(c.BC.lo)
		case 0x82:
			name = "ADD D"
			length, duration = c.AddReg8(c.DE.hi)
		case 0x83:
			name = "ADD E"
			length, duration = c.AddReg8(c.DE.lo)
		case 0x84:
			name = "ADD H"
			length, duration = c.AddReg8(c.HL.hi)
		case 0x85:
			name = "ADD L"
			length, duration = c.AddReg8(c.HL.lo)
		case 0x86:
			name = "ADD (HL)"
			length = 1
			duration = 8
			add := c.mmu.ReadByte(c.HL.word)
			c.UnsetSubtractionFlag()
			c.UnsetCarryFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetZeroFlag()
			*c.AF.hi += add
			if *c.AF.hi == 0 {
				c.SetZeroFlag()
			}
			if *c.AF.hi < add {
				c.SetCarryFlag()
				c.SetHalfCarryFlag()
			}
		case 0x87:
			name = "ADD A"
			length, duration = c.AddReg8(c.AF.hi)

		case 0x90:
			name = "SUB B"
			length, duration = c.SubReg(c.BC.hi)
		case 0x91:
			name = "SUB C"
			length, duration = c.SubReg(c.BC.lo)
		case 0x92:
			name = "SUB D"
			length, duration = c.SubReg(c.DE.hi)
		case 0x93:
			name = "SUB E"
			length, duration = c.SubReg(c.DE.lo)
		case 0x94:
			name = "SUB H"
			length, duration = c.SubReg(c.HL.hi)
		case 0x95:
			name = "SUB L"
			length, duration = c.SubReg(c.HL.lo)
		case 0x96:
			name = "SUB (HL)"
			length = 1
			duration = 8
			sub := c.mmu.ReadByte(c.HL.word)
			c.UnsetSubtractionFlag()
			c.UnsetCarryFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetZeroFlag()
			if *c.AF.hi < sub {
				c.SetCarryFlag()
				c.SetHalfCarryFlag()
			} else if *c.AF.hi == sub {
				c.SetZeroFlag()
			}
			*c.AF.hi -= sub
		case 0x97:
			name = "SUB A"
			length, duration = c.SubReg(c.AF.hi)

		case 0xA8:
			name = "XOR B"
			length, duration = c.XorReg(c.BC.hi)
		case 0xA9:
			name = "XOR C"
			length, duration = c.XorReg(c.BC.lo)
		case 0xAA:
			name = "XOR D"
			length, duration = c.XorReg(c.DE.hi)
		case 0xAB:
			name = "XOR E"
			length, duration = c.XorReg(c.DE.lo)
		case 0xAC:
			name = "XOR H"
			length, duration = c.XorReg(c.HL.hi)
		case 0xAD:
			name = "XOR L"
			length, duration = c.XorReg(c.HL.lo)
		case 0xAE:
			name = "XOR (HL)"
			length = 1
			duration = 8
			byte_ := c.mmu.ReadByte(c.HL.word)
			*c.AF.hi ^= byte_
			if *c.AF.hi == 0 {
				c.SetZeroFlag()
			} else {
				c.UnsetZeroFlag()
			}
			c.UnsetSubtractionFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetCarryFlag()
		case 0xAF:
			name = "XOR A"
			length, duration = c.XorReg(c.AF.hi)

		case 0xB0:
			name = "OR B"
			length, duration = c.OrReg(c.BC.hi)
		case 0xB1:
			name = "OR C"
			length, duration = c.OrReg(c.BC.lo)
		case 0xB2:
			name = "OR D"
			length, duration = c.OrReg(c.DE.hi)
		case 0xB3:
			name = "OR E"
			length, duration = c.OrReg(c.DE.lo)
		case 0xB4:
			name = "OR H"
			length, duration = c.OrReg(c.HL.hi)
		case 0xB5:
			name = "OR L"
			length, duration = c.OrReg(c.HL.lo)
		case 0xB6:
			name = "OR (HL)"
			length = 1
			duration = 8
			byte_ := c.mmu.ReadByte(c.HL.word)
			*c.AF.hi ^= byte_
			if *c.AF.hi == 0 {
				c.SetZeroFlag()
			} else {
				c.UnsetZeroFlag()
			}
			c.UnsetSubtractionFlag()
			c.UnsetHalfCarryFlag()
			c.UnsetCarryFlag()
		case 0xB7:
			name = "OR A"
			length, duration = c.OrReg(c.AF.hi)
		case 0xFE:
			name = "CP d8"
			length = 2
			duration = 8
			address := uint16(c.mmu.ReadByte(c.PC.word + 1))
			val := c.mmu.ReadByte(address)
			c.SetSubtractionFlag()
			if *c.AF.hi == val {
				c.SetZeroFlag()
			} else {
				c.UnsetZeroFlag()
			}
			if *c.AF.hi < val {
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
		}
	}

	arg := uint16(0)
	if length == 2 {
		arg = uint16(c.mmu.ReadByte(c.PC.word + 1))
	} else if length == 3 {
		arg = c.mmu.ReadWord(c.PC.word + 1)
	}

	if !jump {
		c.PC.word += uint16(length)
	} else {
		c.PC.word = jumpTo
	}
	return Instruction{name, location, arg, opcode, length, duration, shortDuration}, cbFlag, breaking
}

// CBInstruction processes an instruction in the same manner as Instruction. It is called if the previous byte is 0xCB.
// It returns the opcode name, the length of the instruction minus one (need to figure out why), and the instruction duration.
func (c *CPU) CBInstruction(opcode uint8, argByte uint8) (string, uint8, uint8) {
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
		if CheckBit(c.HL.hi, 7) {
			c.UnsetZeroFlag()
		} else {
			c.SetZeroFlag()
		}
	case 0x11:
		name = "RL C"
		length = 2
		duration = 8
		toCarry := false
		if CheckBit(c.BC.lo, 7) {
			toCarry = true
		}
		*c.BC.lo = bits.RotateLeft8(*c.BC.lo, 9)
		if c.GetCarryFlag() {
			*c.BC.lo |= 1
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
	*c.AF.lo |= BitVal(Z)
}

// UnsetZeroFlag sets the zero flag to 0.
func (c *CPU) UnsetZeroFlag() {
	*c.AF.lo &^= BitVal(Z)
}

// SetSubtractionFlag sets the subtraction flag to 1.
func (c *CPU) SetSubtractionFlag() {
	*c.AF.lo |= BitVal(N)
}

// UnsetSubtractionFlag sets the subtraction flag to 0.
func (c *CPU) UnsetSubtractionFlag() {
	*c.AF.lo &^= BitVal(N)
}

// SetCarryFlag sets the carry flag to 1.
func (c *CPU) SetCarryFlag() {
	*c.AF.lo |= BitVal(C)
}

// UnsetCarryFlag sets the carry flag to 0.
func (c *CPU) UnsetCarryFlag() {
	*c.AF.lo &^= BitVal(C)
}

// SetHalfCarryFlag sets the half-carry flag to 1.
func (c *CPU) SetHalfCarryFlag() {
	*c.AF.lo |= BitVal(H)
}

// UnsetHalfCarryFlag sets the half-carry flag to 0.
func (c *CPU) UnsetHalfCarryFlag() {
	*c.AF.lo &^= BitVal(H)
}

// GetZeroFlag returns true if the zero flag is set.
func (c *CPU) GetZeroFlag() bool {
	return CheckBit(c.AF.lo, Z)
}

// GetSubtractionFlag returns true if the subtraction flag is set.
func (c *CPU) GetSubtractionFlag() bool {
	return CheckBit(c.AF.lo, N)
}

// GetHalfCarryFlag returns true if the half-carry flag is set.
func (c *CPU) GetHalfCarryFlag() bool {
	return CheckBit(c.AF.lo, H)
}

// GetCarryFlag returns true if the carry flag is set.
func (c *CPU) GetCarryFlag() bool {
	return CheckBit(c.AF.lo, C)
}

// PrintInstruction prints the byte location, opcode, opcode name, length, duration, and shortDuration of the passed instruction.
func (c *CPU) PrintInstruction(insCount uint64, i Instruction) {
	fmt.Printf("Step %d, Byte %X\n\t%X|%s $%X: length: %d, duration: %d/%d\n", insCount, i.location, i.opcode, i.name, i.arg, i.length, i.duration, i.shortDuration)
}

// PrintRegisters prints the stack pointer location and data, and the values in each register except the flag register.
func (c *CPU) PrintRegisters() {
	fmt.Printf("\tStack pointer: %X ($%X) \n\t\tA: %X, B: %X, C: %X, D: %X, E: %X, H: %X, L: %X\n",
		c.SP.word,
		c.mmu.ReadWord(c.SP.word+2),
		*c.AF.hi,
		*c.BC.hi,
		*c.BC.lo,
		*c.DE.hi,
		*c.DE.lo,
		*c.HL.hi,
		*c.HL.lo,
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
