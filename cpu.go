package main

import (
	"fmt"
	"io/ioutil"
	"math/bits"
	"time"
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

	opcodeMap   map[uint8]func() string
	cbOpcodeMap map[uint8]func() string

	bootloader [0x100]byte
	cycles     uint64

	breaking bool
}

// Only used for printing instruction information
type Instruction struct {
	name     string
	location uint16
	arg      uint16
	opcode   uint8
	length   uint8
	duration uint8
}

// Stupid bithacking used to have the hi and lo values be the
// high and low bytes of word so changing any will affect the other automatically
type Register struct {
	hi   *uint8
	lo   *uint8
	word uint16
}

// Reset links a new MMU to the CPU, clears the registers, sets up the opcode maps, and loads in the bootloader data from a file.
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

	c.SetupOpcodeMap()
}

// Fill in the opcodeMap and cbOpcodeMap
func (c *CPU) SetupOpcodeMap() {

	// INC/DEC
	c.opcodeMap = make(map[uint8]func() string)
	c.opcodeMap[0x3C] = func() string {
		c.Inc8(c.AF.hi)
		return "INC A"
	}
	c.opcodeMap[0x3D] = func() string {
		c.Dec8(c.AF.hi)
		return "DEC A"
	}
	c.opcodeMap[0x04] = func() string {
		c.Inc8(c.BC.hi)
		return "INC B"
	}
	c.opcodeMap[0x05] = func() string {
		c.Dec8(c.BC.hi)
		return "DEC B"
	}
	c.opcodeMap[0x0C] = func() string {
		c.Inc8(c.BC.lo)
		return "INC C"
	}
	c.opcodeMap[0x0D] = func() string {
		c.Dec8(c.BC.lo)
		return "DEC C"
	}
	c.opcodeMap[0x14] = func() string {
		c.Inc8(c.DE.hi)
		return "INC D"
	}
	c.opcodeMap[0x15] = func() string {
		c.Dec8(c.DE.hi)
		return "DEC D"
	}
	c.opcodeMap[0x1C] = func() string {
		c.Inc8(c.DE.lo)
		return "INC E"
	}
	c.opcodeMap[0x1D] = func() string {
		c.Dec8(c.DE.lo)
		return "DEC E"
	}
	c.opcodeMap[0x24] = func() string {
		c.Inc8(c.HL.hi)
		return "INC H"
	}
	c.opcodeMap[0x25] = func() string {
		c.Dec8(c.HL.hi)
		return "DEC H"
	}
	c.opcodeMap[0x2C] = func() string {
		c.Inc8(c.HL.lo)
		return "INC L"
	}
	c.opcodeMap[0x2D] = func() string {
		c.Inc8(c.HL.lo)
		return "DEC L"
	}
	c.opcodeMap[0x03] = func() string {
		c.Inc16(&c.BC.word)
		return "INC BC"
	}
	c.opcodeMap[0x0B] = func() string {
		c.Dec16(&c.BC.word)
		return "DEC BC"
	}
	c.opcodeMap[0x13] = func() string {
		c.Inc16(&c.DE.word)
		return "INC DE"
	}
	c.opcodeMap[0x1B] = func() string {
		c.Dec16(&c.DE.word)
		return "DEC DE"
	}
	c.opcodeMap[0x23] = func() string {
		c.Inc16(&c.HL.word)
		return "INC HL"
	}
	c.opcodeMap[0x2B] = func() string {
		c.Dec16(&c.HL.word)
		return "DEC HL"
	}

	// LD R,d8
	c.opcodeMap[0x3E] = func() string {
		c.LdByte(c.AF.hi)
		return "LD A,d8"
	}
	c.opcodeMap[0x06] = func() string {
		c.LdByte(c.BC.hi)
		return "LD B,d8"
	}
	c.opcodeMap[0x0E] = func() string {
		c.LdByte(c.BC.lo)
		return "LD C,d8"
	}
	c.opcodeMap[0x16] = func() string {
		c.LdByte(c.DE.hi)
		return "LD D,d8"
	}
	c.opcodeMap[0x1E] = func() string {
		c.LdByte(c.DE.lo)
		return "LD E,d8"
	}
	c.opcodeMap[0x26] = func() string {
		c.LdByte(c.HL.hi)
		return "LD H,d8"
	}
	c.opcodeMap[0x2E] = func() string {
		c.LdByte(c.HL.lo)
		return "LD L,d8"
	}
	// LD R,R
	c.opcodeMap[0x7F] = func() string {
		c.LdReg8(c.AF.hi, c.AF.hi)
		return "LD A,A"
	}
	c.opcodeMap[0x78] = func() string {
		c.LdReg8(c.AF.hi, c.BC.hi)
		return "LD A,B"
	}
	c.opcodeMap[0x79] = func() string {
		c.LdReg8(c.AF.hi, c.BC.lo)
		return "LD A,C"
	}
	c.opcodeMap[0x7A] = func() string {
		c.LdReg8(c.AF.hi, c.DE.hi)
		return "LD A,D"
	}
	c.opcodeMap[0x7B] = func() string {
		c.LdReg8(c.AF.hi, c.DE.lo)
		return "LD A,E"
	}
	c.opcodeMap[0x7C] = func() string {
		c.LdReg8(c.AF.hi, c.HL.hi)
		return "LD A,H"
	}
	c.opcodeMap[0x7D] = func() string {
		c.LdReg8(c.AF.hi, c.HL.lo)
		return "LD A,L"
	}
	c.opcodeMap[0x47] = func() string {
		c.LdReg8(c.BC.hi, c.AF.hi)
		return "LD B,A"
	}
	c.opcodeMap[0x40] = func() string {
		c.LdReg8(c.BC.hi, c.BC.hi)
		return "LD B,B"
	}
	c.opcodeMap[0x41] = func() string {
		c.LdReg8(c.BC.hi, c.BC.lo)
		return "LD B,C"
	}
	c.opcodeMap[0x42] = func() string {
		c.LdReg8(c.BC.hi, c.DE.hi)
		return "LD B,D"
	}
	c.opcodeMap[0x43] = func() string {
		c.LdReg8(c.BC.hi, c.DE.lo)
		return "LD B,E"
	}
	c.opcodeMap[0x44] = func() string {
		c.LdReg8(c.BC.hi, c.HL.hi)
		return "LD B,H"
	}
	c.opcodeMap[0x45] = func() string {
		c.LdReg8(c.BC.hi, c.HL.lo)
		return "LD B,L"
	}
	c.opcodeMap[0x4F] = func() string {
		c.LdReg8(c.BC.lo, c.AF.hi)
		return "LD C,A"
	}
	c.opcodeMap[0x48] = func() string {
		c.LdReg8(c.BC.lo, c.BC.hi)
		return "LD C,B"
	}
	c.opcodeMap[0x49] = func() string {
		c.LdReg8(c.BC.lo, c.BC.lo)
		return "LD C,C"
	}
	c.opcodeMap[0x4A] = func() string {
		c.LdReg8(c.BC.lo, c.DE.hi)
		return "LD C,D"
	}
	c.opcodeMap[0x4B] = func() string {
		c.LdReg8(c.BC.lo, c.DE.lo)
		return "LD C,E"
	}
	c.opcodeMap[0x4C] = func() string {
		c.LdReg8(c.BC.lo, c.HL.hi)
		return "LD C,H"
	}
	c.opcodeMap[0x4D] = func() string {
		c.LdReg8(c.BC.lo, c.HL.lo)
		return "LD C,L"
	}
	c.opcodeMap[0x57] = func() string {
		c.LdReg8(c.DE.hi, c.AF.hi)
		return "LD D,A"
	}
	c.opcodeMap[0x50] = func() string {
		c.LdReg8(c.DE.hi, c.BC.hi)
		return "LD D,B"
	}
	c.opcodeMap[0x51] = func() string {
		c.LdReg8(c.DE.hi, c.BC.lo)
		return "LD D,C"
	}
	c.opcodeMap[0x52] = func() string {
		c.LdReg8(c.DE.hi, c.DE.hi)
		return "LD D,D"
	}
	c.opcodeMap[0x53] = func() string {
		c.LdReg8(c.DE.hi, c.DE.lo)
		return "LD D,E"
	}
	c.opcodeMap[0x54] = func() string {
		c.LdReg8(c.DE.hi, c.HL.hi)
		return "LD D,H"
	}
	c.opcodeMap[0x55] = func() string {
		c.LdReg8(c.DE.hi, c.HL.lo)
		return "LD D,L"
	}
	c.opcodeMap[0x5F] = func() string {
		c.LdReg8(c.DE.lo, c.AF.hi)
		return "LD E,A"
	}
	c.opcodeMap[0x58] = func() string {
		c.LdReg8(c.DE.lo, c.BC.hi)
		return "LD E,B"
	}
	c.opcodeMap[0x59] = func() string {
		c.LdReg8(c.DE.lo, c.BC.lo)
		return "LD E,C"
	}
	c.opcodeMap[0x5A] = func() string {
		c.LdReg8(c.DE.lo, c.DE.hi)
		return "LD E,D"
	}
	c.opcodeMap[0x5B] = func() string {
		c.LdReg8(c.DE.lo, c.DE.lo)
		return "LD E,E"
	}
	c.opcodeMap[0x5C] = func() string {
		c.LdReg8(c.DE.lo, c.HL.hi)
		return "LD E,H"
	}
	c.opcodeMap[0x5D] = func() string {
		c.LdReg8(c.DE.lo, c.HL.lo)
		return "LD E,L"
	}
	c.opcodeMap[0x67] = func() string {
		c.LdReg8(c.HL.hi, c.AF.hi)
		return "LD H,A"
	}
	c.opcodeMap[0x60] = func() string {
		c.LdReg8(c.HL.hi, c.BC.hi)
		return "LD H,B"
	}
	c.opcodeMap[0x61] = func() string {
		c.LdReg8(c.HL.hi, c.BC.lo)
		return "LD H,C"
	}
	c.opcodeMap[0x62] = func() string {
		c.LdReg8(c.HL.hi, c.DE.hi)
		return "LD H,D"
	}
	c.opcodeMap[0x63] = func() string {
		c.LdReg8(c.HL.hi, c.DE.lo)
		return "LD H,E"
	}
	c.opcodeMap[0x64] = func() string {
		c.LdReg8(c.HL.hi, c.HL.hi)
		return "LD H,H"
	}
	c.opcodeMap[0x65] = func() string {
		c.LdReg8(c.HL.hi, c.HL.lo)
		return "LD H,L"
	}
	c.opcodeMap[0x6F] = func() string {
		c.LdReg8(c.HL.lo, c.AF.hi)
		return "LD L,A"
	}
	c.opcodeMap[0x68] = func() string {
		c.LdReg8(c.HL.lo, c.BC.hi)
		return "LD L,B"
	}
	c.opcodeMap[0x69] = func() string {
		c.LdReg8(c.HL.lo, c.BC.lo)
		return "LD L,C"
	}
	c.opcodeMap[0x6A] = func() string {
		c.LdReg8(c.HL.lo, c.DE.hi)
		return "LD L,D"
	}
	c.opcodeMap[0x6B] = func() string {
		c.LdReg8(c.HL.lo, c.DE.lo)
		return "LD L,E"
	}
	c.opcodeMap[0x6C] = func() string {
		c.LdReg8(c.HL.lo, c.HL.hi)
		return "LD L,H"
	}
	c.opcodeMap[0x6D] = func() string {
		c.LdReg8(c.HL.lo, c.HL.lo)
		return "LD L,L"
	}

	// LD R,a16
	c.opcodeMap[0x0A] = func() string {
		c.LdReg8Adr(c.AF.hi, c.BC.word)
		return "LD A,(BC)"
	}
	c.opcodeMap[0x1A] = func() string {
		c.LdReg8Adr(c.AF.hi, c.DE.word)
		return "LD A,(DE)"
	}
	c.opcodeMap[0x7E] = func() string {
		c.LdReg8Adr(c.AF.hi, c.HL.word)
		return "LD A,(HL)"
	}
	c.opcodeMap[0x2A] = func() string {
		c.LdReg8Adr(c.AF.hi, c.HL.word)
		c.Dec16(&c.HL.word)
		return "LD A,(HL+)"
	}
	c.opcodeMap[0x3A] = func() string {
		c.LdReg8Adr(c.AF.hi, c.HL.word)
		c.Dec16(&c.HL.word)
		return "LD A,(HL-)"
	}
	c.opcodeMap[0xF0] = func() string {
		*c.AF.hi = c.mmu.ReadByte(0xFF00 | uint16(c.mmu.ReadByte(c.PC.word+1)))
		c.PC.word += 2
		c.cycles += 12
		return "LDH A,a8"
	}

	// LD RR,d16
	c.opcodeMap[0x01] = func() string {
		c.LdWord(&c.BC.word)
		return "LD BC,d16"
	}
	c.opcodeMap[0x11] = func() string {
		c.LdWord(&c.DE.word)
		return "LD DE,d16"
	}
	c.opcodeMap[0x21] = func() string {
		c.LdWord(&c.HL.word)
		return "LD HL,d16"
	}
	c.opcodeMap[0x31] = func() string {
		c.SP.word = c.mmu.ReadWord(c.PC.word + 1)
		c.PC.word += 3
		c.cycles += 12
		return "LD SP,d16"
	}

	// LD address,R
	c.opcodeMap[0xE2] = func() string {
		c.mmu.WriteByte(0xFF00|uint16(*c.BC.lo), *c.AF.hi)
		c.PC.word++ // Opcode table says 2 but that seems wrong
		c.cycles += 8
		return "LD (C),A"
	}
	c.opcodeMap[0x02] = func() string {
		c.LdAdrA(c.BC.word)
		return "LD (BC),A"
	}
	c.opcodeMap[0x12] = func() string {
		c.LdAdrA(c.DE.word)
		return "LD (DE),A"
	}
	c.opcodeMap[0x77] = func() string {
		c.LdAdrA(c.HL.word)
		return "LD (HL),A"
	}
	c.opcodeMap[0x32] = func() string {
		// These are faster than the sum of their parts, so can't just call LdAddrA
		c.mmu.WriteByte(c.HL.word, *c.AF.hi)
		c.Dec16(&c.HL.word)
		return "LD (HL-),A"
	}
	c.opcodeMap[0x22] = func() string {
		// These are faster than the sum of their parts, so can't just call LdAddrA
		c.mmu.WriteByte(c.HL.word, *c.AF.hi)
		c.Inc16(&c.HL.word)
		return "LD (HL+),A"
	}
	c.opcodeMap[0xE0] = func() string {
		c.mmu.WriteByte(0xFF00|uint16(c.mmu.ReadByte(c.PC.word+1)), *c.AF.hi)
		c.PC.word += 2
		c.cycles += 12
		return "LDH (a8),A"
	}
	c.opcodeMap[0xEA] = func() string {
		c.mmu.WriteByte(c.mmu.ReadWord(c.PC.word+1), *c.AF.hi)
		c.PC.word += 3
		c.cycles += 16
		return "LD a16,A"
	}

	// Jump
	c.opcodeMap[0x18] = func() string {
		c.JRCond(true)
		return "JR r8"
	}
	c.opcodeMap[0x20] = func() string {
		c.JRCond(!c.GetZeroFlag())
		return "JR NZ,r8"
	}
	c.opcodeMap[0x28] = func() string {
		c.JRCond(c.GetZeroFlag())
		return "JR Z,r8"
	}
	c.opcodeMap[0xC3] = func() string {
		c.PC.word = c.mmu.ReadWord(c.PC.word + 1)
		c.cycles += 16
		return "JP a16"
	}

	// Stack ops
	c.opcodeMap[0xC5] = func() string {
		c.PushWord(c.BC.word)
		return "PUSH BC"
	}
	c.opcodeMap[0xC1] = func() string {
		c.PopWord(&c.BC.word)
		return "POP BC"
	}
	c.opcodeMap[0xC9] = func() string {
		c.SP.word += 2
		c.PC.word = c.mmu.ReadWord(c.SP.word)
		c.cycles += 16
		return "RET"
	}
	c.opcodeMap[0xCD] = func() string {
		c.mmu.WriteWord(c.SP.word, c.PC.word+3)
		c.SP.word -= 2
		c.PC.word = c.mmu.ReadWord(c.PC.word + 1)
		c.cycles += 24
		return "CALL a16"
	}

	// Misc
	c.opcodeMap[0x00] = func() string {
		c.PC.word++
		c.cycles += 4
		return "NOP"
	}
	c.opcodeMap[0xCB] = func() string {
		cbop := c.cbOpcodeMap[c.mmu.ReadByte(c.PC.word+1)]()
		c.PC.word++ // The length is 2 in total but the CB ins is one byte and the actual instruction is one byte. Since some of the CB instructions call functions which increment c.PC, setting this to increment 1 works best.
		c.cycles += 4
		return "CB " + cbop
	}

	// Arithmetic
	c.opcodeMap[0x07] = func() string {
		c.RotateCarry(c.AF.hi, 8)
		return "RLCA"
	}
	c.opcodeMap[0x17] = func() string {
		c.Rotate(c.AF.hi, 9)
		return "RLA"
	}

	c.opcodeMap[0x1F] = func() string {
		c.Rotate(c.AF.hi, -9)
		return "RRA"
	}
	c.opcodeMap[0x0F] = func() string {
		c.RotateCarry(c.AF.hi, -8)
		return "RRCA"
	}

	c.opcodeMap[0x09] = func() string {
		c.AddReg16(&c.BC.word)
		return "ADD HL,BC"
	}
	c.opcodeMap[0x19] = func() string {
		c.AddReg16(&c.DE.word)
		return "ADD HL,DE"
	}
	c.opcodeMap[0x29] = func() string {
		c.AddReg16(&c.HL.word)
		return "ADD HL,HL"
	}
	c.opcodeMap[0x39] = func() string {
		c.AddReg16(&c.SP.word)
		return "ADD HL,SP"
	}
	c.opcodeMap[0x80] = func() string {
		c.AddReg8(c.BC.hi)
		return "ADD B"
	}
	c.opcodeMap[0x81] = func() string {
		c.AddReg8(c.BC.lo)
		return "ADD C"
	}
	c.opcodeMap[0x82] = func() string {
		c.AddReg8(c.DE.hi)
		return "ADD D"
	}
	c.opcodeMap[0x83] = func() string {
		c.AddReg8(c.DE.lo)
		return "ADD E"
	}
	c.opcodeMap[0x84] = func() string {
		c.AddReg8(c.HL.hi)
		return "ADD H"
	}
	c.opcodeMap[0x85] = func() string {
		c.AddReg8(c.HL.lo)
		return "ADD L"
	}
	c.opcodeMap[0x86] = func() string {
		c.UnsetSubtractionFlag()
		c.UnsetCarryFlag()
		c.UnsetHalfCarryFlag()
		c.UnsetZeroFlag()

		byte_ := c.mmu.ReadByte(c.HL.word)
		*c.AF.hi += byte_

		if *c.AF.hi < byte_ {
			c.SetCarryFlag()
			c.SetHalfCarryFlag()
		}
		if *c.AF.hi == 0 {
			c.SetZeroFlag()
		}
		c.PC.word++
		c.cycles += 8
		return "ADD (HL)"
	}
	c.opcodeMap[0x87] = func() string {
		c.AddReg8(c.AF.hi)
		return "ADD A"
	}
	c.opcodeMap[0x90] = func() string {
		c.SubReg(c.BC.hi)
		return "SUB B"
	}
	c.opcodeMap[0x91] = func() string {
		c.SubReg(c.BC.lo)
		return "SUB C"
	}
	c.opcodeMap[0x92] = func() string {
		c.SubReg(c.DE.hi)
		return "SUB D"
	}
	c.opcodeMap[0x93] = func() string {
		c.SubReg(c.DE.lo)
		return "SUB E"
	}
	c.opcodeMap[0x94] = func() string {
		c.SubReg(c.HL.hi)
		return "SUB H"
	}
	c.opcodeMap[0x95] = func() string {
		c.SubReg(c.HL.lo)
		return "SUB L"
	}
	c.opcodeMap[0x96] = func() string {
		c.SetSubtractionFlag()
		c.UnsetCarryFlag()
		c.UnsetHalfCarryFlag()
		c.UnsetZeroFlag()

		byte_ := c.mmu.ReadByte(c.HL.word)
		if *c.AF.hi < byte_ {
			c.SetCarryFlag()
			c.SetHalfCarryFlag()
		} else if *c.AF.hi == byte_ {
			c.SetZeroFlag()
		}
		*c.AF.hi -= byte_
		c.PC.word++
		c.cycles += 8
		return "SUB (HL)"
	}
	c.opcodeMap[0x97] = func() string {
		c.SubReg(c.AF.hi)
		return "SUB A"
	}
	c.opcodeMap[0xA8] = func() string {
		c.XorReg(c.BC.hi)
		return "XOR B"
	}
	c.opcodeMap[0xA9] = func() string {
		c.XorReg(c.BC.lo)
		return "XOR C"
	}
	c.opcodeMap[0xAA] = func() string {
		c.XorReg(c.DE.hi)
		return "XOR D"
	}
	c.opcodeMap[0xAB] = func() string {
		c.XorReg(c.DE.lo)
		return "XOR E"
	}
	c.opcodeMap[0xAC] = func() string {
		c.XorReg(c.HL.hi)
		return "XOR H"
	}
	c.opcodeMap[0xAD] = func() string {
		c.XorReg(c.HL.lo)
		return "XOR L"
	}
	c.opcodeMap[0xAE] = func() string {
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
		c.PC.word++
		c.cycles += 4
		c.cycles += 4
		return "XOR (HL)"
	}
	c.opcodeMap[0xAF] = func() string {
		c.XorReg(c.AF.hi)
		return "XOR A"
	}
	c.opcodeMap[0xB0] = func() string {
		c.OrReg(c.BC.hi)
		return "OR B"
	}
	c.opcodeMap[0xB1] = func() string {
		c.OrReg(c.BC.lo)
		return "OR C"
	}
	c.opcodeMap[0xB2] = func() string {
		c.OrReg(c.DE.hi)
		return "OR D"
	}
	c.opcodeMap[0xB3] = func() string {
		c.OrReg(c.DE.lo)
		return "OR E"
	}
	c.opcodeMap[0xB4] = func() string {
		c.OrReg(c.HL.hi)
		return "OR H"
	}
	c.opcodeMap[0xB5] = func() string {
		c.OrReg(c.HL.lo)
		return "OR L"
	}
	c.opcodeMap[0xB6] = func() string {
		byte_ := c.mmu.ReadByte(c.HL.word)
		*c.AF.hi |= byte_
		if *c.AF.hi == 0 {
			c.SetZeroFlag()
		} else {
			c.UnsetZeroFlag()
		}
		c.UnsetSubtractionFlag()
		c.UnsetHalfCarryFlag()
		c.UnsetCarryFlag()
		c.PC.word++
		c.cycles += 8
		return "OR (HL)"
	}
	c.opcodeMap[0xB7] = func() string {
		c.OrReg(c.AF.hi)
		return "OR A"
	}
	c.opcodeMap[0xBE] = func() string {
		c.CPByte(c.mmu.ReadByte(c.HL.word))
		return "CP (HL)"
	}
	c.opcodeMap[0xFE] = func() string {
		c.CPByte(c.mmu.ReadByte(c.PC.word + 1))
		c.PC.word++ // CP d8 is length 2 and CPByte only increases by 1
		return "CP d8"
	}

	// CB opcode map setup here
	c.cbOpcodeMap = make(map[uint8]func() string)

	c.cbOpcodeMap[0x7C] = func() string {
		c.UnsetSubtractionFlag()
		c.SetHalfCarryFlag()
		if CheckBit(c.HL.hi, 7) {
			c.UnsetZeroFlag()
		} else {
			c.SetZeroFlag()
		}
		c.PC.word++
		c.cycles += 4
		return "BIT 7,H"
	}
	c.cbOpcodeMap[0x11] = func() string {
		c.Rotate(c.BC.lo, 9)
		return "RL C"
	}
}

// CPByte compares a byte with the value in the A register, setting whichever flags are relevant to the result.
func (c *CPU) CPByte(byte_ uint8) {
	c.SetSubtractionFlag()
	c.UnsetZeroFlag()
	c.UnsetCarryFlag()
	c.UnsetHalfCarryFlag()
	if *c.AF.hi-byte_ == 0 {
		c.SetZeroFlag()
	}
	if *c.AF.hi < byte_ {
		c.SetCarryFlag()
		c.SetHalfCarryFlag()
	}
	c.PC.word++
	c.cycles += 8
}

// Rotate rotates a byte by a given amount.
func (c *CPU) Rotate(register *uint8, amt int) {
	toCarry := false

	carryBit := uint8(0)
	if amt > 1 {
		carryBit = 7
	}

	if CheckBit(c.AF.hi, carryBit) {
		toCarry = true
	}

	*c.AF.hi = bits.RotateLeft8(*c.AF.hi, amt)

	if c.GetCarryFlag() {
		*c.AF.hi |= BitVal(carryBit)
	}

	if toCarry {
		c.SetCarryFlag()
	} else {
		c.UnsetCarryFlag()
	}
	c.UnsetZeroFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetSubtractionFlag()

	c.PC.word++
	c.cycles += 4
}

// RotateCarry rotates a byte by a given amount and carrys the overflow bit.
func (c *CPU) RotateCarry(register *uint8, amt int) {
	carryBit := uint8(0)
	if amt > 1 {
		carryBit = 7
	}
	if CheckBit(c.AF.hi, carryBit) {
		c.SetCarryFlag()
	} else {
		c.UnsetCarryFlag()
	}
	*c.AF.hi = bits.RotateLeft8(*c.AF.hi, amt)
	c.UnsetZeroFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetSubtractionFlag()
	c.PC.word++
	c.cycles += 4
}

// JRCond jumps to a relative position if condition is true.
func (c *CPU) JRCond(condition bool) {
	if condition {
		c.cycles += 12
		// TODO: Probably a way to do this in one line (128 - arg or something)
		if arg := uint16(c.mmu.ReadByte(c.PC.word + 1)); arg > 127 {
			c.PC.word = c.PC.word - (255 - arg) + 1
		} else {
			c.PC.word = c.PC.word + arg + 2
		}
	} else {
		c.PC.word += 2
		c.cycles += 8
	}
}

// PushWord pushes a 16-bit word onto the stack
func (c *CPU) PushWord(word uint16) {
	c.mmu.WriteWord(c.SP.word, word)
	c.SP.word -= 2
	c.PC.word++
	c.cycles += 16
}

// PopWord pops a 16-bit word off of the stack into the location specified.
func (c *CPU) PopWord(word *uint16) {
	c.SP.word += 2
	*word = c.mmu.ReadWord(c.SP.word)
	c.PC.word++
	c.cycles += 12
}

// Inc8 increments an 8-bit register by 1.
func (c *CPU) Inc8(register *uint8) {
	*register++
	c.UnsetSubtractionFlag()
	if *register == 0 {
		c.SetZeroFlag()
		c.UnsetHalfCarryFlag()
	}
	c.PC.word++
	c.cycles += 4
}

// Dec8 increments an 8-bit register by 1.
func (c *CPU) Dec8(register *uint8) {
	*register--
	c.SetSubtractionFlag()
	c.UnsetZeroFlag()
	if *register == 0 {
		c.SetZeroFlag()
	} else if *register == 255 {
		c.SetHalfCarryFlag()
	}
	c.PC.word++
	c.cycles += 4
}

// Inc16 increments a 16-bit register pair by 1.
func (c *CPU) Inc16(registerPair *uint16) {
	*registerPair++
	c.PC.word++
	c.cycles += 8
}

// Dec16 increments a 16-bit register pair by 1.
func (c *CPU) Dec16(registerPair *uint16) {
	*registerPair--
	c.PC.word++
	c.cycles += 8
}

// LdByte reads a byte into a register.
func (c *CPU) LdByte(register *uint8) {
	*register = c.mmu.ReadByte(c.PC.word + 1)
	c.PC.word += 2
	c.cycles += 8
}

// LdWord loads a 16-bit word into a register pair.
func (c *CPU) LdWord(registerPair *uint16) {
	*registerPair = c.mmu.ReadWord(c.PC.word + 1)
	c.PC.word += 3
	c.cycles += 12
}

// LdReg8 copies the contents of a register into another.
func (c *CPU) LdReg8(to *uint8, from *uint8) {
	*to = *from
	c.PC.word++
	c.cycles += 4
}

// LdReg8Adr copies the contents of a memory address into a register.
func (c *CPU) LdReg8Adr(register *uint8, address uint16) {
	*register = c.mmu.ReadByte(address)
	c.PC.word += 1
	c.cycles += 8
}

// LdHLA copies the value of register A into the memory address specified.
func (c *CPU) LdAdrA(address uint16) {
	c.mmu.WriteByte(address, *c.AF.hi)
	c.PC.word++
	c.cycles += 8
}

// AddReg8 adds a register to A.
func (c *CPU) AddReg8(register *uint8) {
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
	c.PC.word++
	c.cycles += 4
}

// AddReg16 adds a register to HL, storing the result in HL.
func (c *CPU) AddReg16(word *uint16) {
	c.UnsetSubtractionFlag()
	c.UnsetCarryFlag()
	c.UnsetHalfCarryFlag()
	c.HL.word += *word

	c.PC.word++
	c.cycles += 8
}

// SubReg subtracts a register from A.
func (c *CPU) SubReg(register *uint8) {
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
	c.PC.word++
	c.cycles += 4
}

// Xors a register with register A and stores the result in A.
func (c *CPU) XorReg(register *uint8) {
	*c.AF.hi ^= *register
	if *c.AF.hi == 0 {
		c.SetZeroFlag()
	} else {
		c.UnsetZeroFlag()
	}
	c.UnsetSubtractionFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetCarryFlag()
	c.PC.word++
	c.cycles += 4
}

// Ors a register with register A and stores the result in A.
// Returns the length and duration of the instruction.
func (c *CPU) OrReg(register *uint8) {
	*c.AF.hi |= *register
	if *c.AF.hi == 0 {
		c.SetZeroFlag()
	} else {
		c.UnsetZeroFlag()
	}
	c.UnsetSubtractionFlag()
	c.UnsetHalfCarryFlag()
	c.UnsetCarryFlag()
	c.PC.word++
	c.cycles += 4
}

// Start writes the bootloader data into the 0x100-0xFFF range of the MMU and returns a stepping function.
// This returned function takes one CPU step each time it is called.
func (c *CPU) Start() func() {
	for i, v := range c.bootloader {
		address := uint16(i)
		c.mmu.WriteByte(address, v)
	}

	var lastIns string

	var insCount = uint64(0)
	start := time.Now()
	var timeDelay time.Duration
	var dt time.Duration
	oldTime := time.Now()
	oldCycles := uint64(0)

	return func() {

		oldCycles = c.cycles
		oldTime = time.Now()
		// Procces the current opcode
		lastIns = c.opcodeMap[c.mmu.ReadByte(c.PC.word)]()
		dt = time.Now().Sub(oldTime)

		timeDelay = time.Duration(c.cycles-oldCycles) * (65 * time.Nanosecond)
		if dt < timeDelay {
			// fmt.Println(timeDelay - dt)
			time.Sleep(timeDelay - dt)
		}

		if c.cycles > 5000000 {
			fmt.Println("5M CPU ops in", time.Now().Sub(start))
			c.cycles = 0
			start = time.Now()
		}
		insCount++

		// fmt.Printf("%X ", c.PC.word)
		if c.PC.word == 0x100 {
			c.breaking = true
		}

		if c.breaking {
			fmt.Printf("%X\t", c.PC.word)
			c.PrintInstruction(lastIns)
			c.PrintRegisters()
			c.PrintFlagRegister()
			fmt.Scanln()
		}
	}
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

// PrintInstruction prints the byte location, opcode, opcode name, length, and duration of the passed instruction.
func (c *CPU) PrintInstruction(name string) {
	fmt.Printf("Step %d, %s", c.cycles, name)
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
