package main

import (
	"bufio"
	"fmt"
	"os"
)

type flagBytes struct {
	Z string
	N string
	H string
	C string
}

var Z uint8 = 128
var N uint8 = 64
var H uint8 = 32
var C uint8 = 16

var registers8 = map[string]uint8{
	"A": 0x0,
	"B": 0x0,
	"C": 0x0,
	"D": 0x0,
	"E": 0x0,
	"F": 0x0,
	"H": 0x0,
	"L": 0x0,
}

var registers16 = map[string]uint16{
	"SP": 0x0,
	"PC": 0x0,
}

var mmu = MMU{memory: [0x10000]uint8{}}

func PrintInstruction(location uint16, opcode uint8, name string, length, duration, shortDuration uint8) {
	fmt.Printf("Byte %X. %X %s: l: %d, d: %d, sd: %d\n", location, opcode, name, length, duration, shortDuration)
}

func u8PairToU16(argBytes [2]uint8) uint16 {
	out := uint16(0)
	out += uint16(argBytes[1]) << 8
	out += uint16(argBytes[0])
	return out
}

// Return as lower, higher byte
func splitUint16(num uint16) [2]uint8 {
	lower := uint8(num & 0xFF)
	higher := uint8((num & 0xFF00) >> 8)
	return [2]uint8{lower, higher}
}

func setZeroFlag() {
	registers8["F"] |= Z
}

func unsetZeroFlag() {
	registers8["F"] &^= Z
}

func setSubtractionFlag() {
	registers8["F"] |= N
}

func unsetSubtractionFlag() {
	registers8["F"] &^= N
}

func setCarryFlag() {
	registers8["F"] |= C
}

func unsetCarryFlag() {
	registers8["F"] &^= C
}

func setHalfCarryFlag() {
	registers8["F"] |= H
}

func unsetHalfCarryFlag() {
	registers8["F"] &^= H
}

func printFlagRegister() {
	fmt.Printf("Z: %t, N: %t, H: %t, C: %t\n",
		registers8["F"]&Z == Z,
		registers8["F"]&N == N,
		registers8["F"]&H == H,
		registers8["F"]&C == C,
	)
}

func RunFile(data []byte) {
	reader := bufio.NewReader(os.Stdin)
	nBytes := uint16(len(data))

	var i uint16
	cb := false
	for {
		i = registers16["PC"]
		argBytes := [2]uint8{0, 0}
		if i < nBytes-2 {
			argBytes = [2]uint8{data[i+1], data[i+2]}
		} else if i < nBytes-1 {
			argBytes = [2]uint8{data[i+1], 0}
		}

		i, cb = Instruction(data[i], i, argBytes, cb)
		registers16["PC"] = i
		reader.ReadString('\n')
	}
}

func Instruction(opcode uint8, location uint16, argBytes [2]uint8, cbFlag bool) (uint16, bool) {
	var name string
	var length uint8
	var duration, shortDuration uint8
	// var flags flagBytes
	var jump bool
	var jumpTo uint16

	if cbFlag {
		name, length, duration = CBInstruction(opcode, location, argBytes[0])
		cbFlag = false
	} else {

		switch opcode {
		case 0x00:
			name = "NOP"
			length = 1
			duration = 4
		case 0x01:
			name = "LD BC,d16"
			length = 3
			duration = 12
		case 0x02:
			name = "LD (BC),A"
			length = 1
			duration = 8
		case 0x03:
			name = "INC BC"
			length = 1
			duration = 8
		case 0x04:
			name = "INC B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x05:
			name = "DEC B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x06:
			name = "LD B,d8"
			length = 2
			duration = 8
		case 0x07:
			name = "RLCA"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x08:
			name = "LD (a16),SP"
			length = 3
			duration = 20
		case 0x09:
			name = "ADD HL,BC"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x0A:
			name = "LD A,(BC)"
			length = 1
			duration = 8
		case 0x0B:
			name = "DEC BC"
			length = 1
			duration = 8
		case 0x0C:
			name = "INC C"
			length = 1
			duration = 4
			registers8["C"] += 1
			unsetSubtractionFlag()
			if registers8["C"] == 0 {
				setZeroFlag()
				setHalfCarryFlag()
			}
		case 0x0D:
			name = "DEC C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x0E:
			name = "LD C,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			registers8["C"] = mmu.ReadByte(address)
		case 0x0F:
			name = "RRCA"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x10:
			name = "STOP 0"
			length = 2
			duration = 4
		case 0x11:
			name = "LD DE,d16"
			length = 3
			duration = 12
			address := u8PairToU16(argBytes)
			word := mmu.ReadWord(address)
			bytePair := splitUint16(word)
			registers8["E"] = bytePair[0]
			registers8["D"] = bytePair[1]
		case 0x12:
			name = "LD (DE),A"
			length = 1
			duration = 8
		case 0x13:
			name = "INC DE"
			length = 1
			duration = 8
			combined := u8PairToU16([2]uint8{registers8["D"], registers8["E"]})
			combined += 1
			vals := splitUint16(combined)
			registers8["D"] = vals[1]
			registers8["E"] = vals[0]
		case 0x14:
			name = "INC D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x15:
			name = "DEC D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x16:
			name = "LD D,d8"
			length = 2
			duration = 8
		case 0x17:
			name = "RLA"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x18:
			name = "JR r8"
			length = 2
			duration = 12
		case 0x19:
			name = "ADD HL,DE"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x1A:
			name = "LD A,(DE)"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{registers8["E"], registers8["D"]})
			registers8["A"] = mmu.ReadByte(address)
		case 0x1B:
			name = "DEC DE"
			length = 1
			duration = 8
		case 0x1C:
			name = "INC E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x1D:
			name = "DEC E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x1E:
			name = "LD E,d8"
			length = 2
			duration = 8
		case 0x1F:
			name = "RRA"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x20:
			name = "JR NZ,r8"
			length = 2
			duration = 12
			shortDuration = 8
			arg := uint16(argBytes[0])
			if registers8["F"]&Z != Z {
				jump = true
				if arg > 128 {
					jumpTo = location - (256 - arg)
				} else {
					jumpTo = location + arg
				}
			}
		case 0x21:
			name = "LD HL,d16"
			length = 3
			duration = 12
			address := u8PairToU16(argBytes)
			word := mmu.ReadWord(address)
			bytePair := splitUint16(word)
			registers8["L"] = bytePair[0]
			registers8["H"] = bytePair[1]
		case 0x22:
			name = "LD (HL+),A"
			length = 1
			duration = 8
		case 0x23:
			name = "INC HL"
			length = 1
			duration = 8
		case 0x24:
			name = "INC H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x25:
			name = "DEC H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x26:
			name = "LD H,d8"
			length = 2
			duration = 8
		case 0x27:
			name = "DAA"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x28:
			name = "JR Z,r8"
			length = 2
			duration = 12
			shortDuration = 8
		case 0x29:
			name = "ADD HL,HL"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x2A:
			name = "LD A,(HL+)"
			length = 1
			duration = 8
		case 0x2B:
			name = "DEC HL"
			length = 1
			duration = 8
		case 0x2C:
			name = "INC L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x2D:
			name = "DEC L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x2E:
			name = "LD L,d8"
			length = 2
			duration = 8
		case 0x2F:
			name = "CPL"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x30:
			name = "JR NC,r8"
			length = 2
			duration = 12
			shortDuration = 8
		case 0x31:
			name = "LD SP,d16"
			length = 3
			duration = 12
			address := u8PairToU16(argBytes)
			data := mmu.ReadWord(address)
			registers16["SP"] = data
		case 0x32:
			name = "LD (HL-),A"
			length = 1
			duration = 8
			registers8["L"] = registers8["A"]
		case 0x33:
			name = "INC SP"
			length = 1
			duration = 8
		case 0x34:
			name = "INC (HL)"
			length = 1
			duration = 12
			//flags = flagBytes{Z: "Z"}
		case 0x35:
			name = "DEC (HL)"
			length = 1
			duration = 12
			//flags = flagBytes{Z: "Z"}
		case 0x36:
			name = "LD (HL),d8"
			length = 2
			duration = 12
		case 0x37:
			name = "SCF"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x38:
			name = "JR C,r8"
			length = 2
			duration = 12
			shortDuration = 8
		case 0x39:
			name = "ADD HL,SP"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x3A:
			name = "LD A,(HL-)"
			length = 1
			duration = 8
		case 0x3B:
			name = "DEC SP"
			length = 1
			duration = 8
		case 0x3C:
			name = "INC A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x3D:
			name = "DEC A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x3E:
			name = "LD A,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			registers8["A"] = mmu.ReadByte(address)
		case 0x3F:
			name = "CCF"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x40:
			name = "LD B,B"
			length = 1
			duration = 4
		case 0x41:
			name = "LD B,C"
			length = 1
			duration = 4
		case 0x42:
			name = "LD B,D"
			length = 1
			duration = 4
		case 0x43:
			name = "LD B,E"
			length = 1
			duration = 4
		case 0x44:
			name = "LD B,H"
			length = 1
			duration = 4
		case 0x45:
			name = "LD B,L"
			length = 1
			duration = 4
		case 0x46:
			name = "LD B,(HL)"
			length = 1
			duration = 8
		case 0x47:
			name = "LD B,A"
			length = 1
			duration = 4
		case 0x48:
			name = "LD C,B"
			length = 1
			duration = 4
		case 0x49:
			name = "LD C,C"
			length = 1
			duration = 4
		case 0x4A:
			name = "LD C,D"
			length = 1
			duration = 4
		case 0x4B:
			name = "LD C,E"
			length = 1
			duration = 4
		case 0x4C:
			name = "LD C,H"
			length = 1
			duration = 4
		case 0x4D:
			name = "LD C,L"
			length = 1
			duration = 4
		case 0x4E:
			name = "LD C,(HL)"
			length = 1
			duration = 8
		case 0x4F:
			name = "LD C,A"
			length = 1
			duration = 4
			registers8["C"] = registers8["A"]
		case 0x50:
			name = "LD D,B"
			length = 1
			duration = 4
		case 0x51:
			name = "LD D,C"
			length = 1
			duration = 4
		case 0x52:
			name = "LD D,D"
			length = 1
			duration = 4
		case 0x53:
			name = "LD D,E"
			length = 1
			duration = 4
		case 0x54:
			name = "LD D,H"
			length = 1
			duration = 4
		case 0x55:
			name = "LD D,L"
			length = 1
			duration = 4
		case 0x56:
			name = "LD D,(HL)"
			length = 1
			duration = 8
		case 0x57:
			name = "LD D,A"
			length = 1
			duration = 4
		case 0x58:
			name = "LD E,B"
			length = 1
			duration = 4
		case 0x59:
			name = "LD E,C"
			length = 1
			duration = 4
		case 0x5A:
			name = "LD E,D"
			length = 1
			duration = 4
		case 0x5B:
			name = "LD E,E"
			length = 1
			duration = 4
		case 0x5C:
			name = "LD E,H"
			length = 1
			duration = 4
		case 0x5D:
			name = "LD E,L"
			length = 1
			duration = 4
		case 0x5E:
			name = "LD E,(HL)"
			length = 1
			duration = 8
		case 0x5F:
			name = "LD E,A"
			length = 1
			duration = 4
		case 0x60:
			name = "LD H,B"
			length = 1
			duration = 4
		case 0x61:
			name = "LD H,C"
			length = 1
			duration = 4
		case 0x62:
			name = "LD H,D"
			length = 1
			duration = 4
		case 0x63:
			name = "LD H,E"
			length = 1
			duration = 4
		case 0x64:
			name = "LD H,H"
			length = 1
			duration = 4
		case 0x65:
			name = "LD H,L"
			length = 1
			duration = 4
		case 0x66:
			name = "LD H,(HL)"
			length = 1
			duration = 8
		case 0x67:
			name = "LD H,A"
			length = 1
			duration = 4
		case 0x68:
			name = "LD L,B"
			length = 1
			duration = 4
		case 0x69:
			name = "LD L,C"
			length = 1
			duration = 4
		case 0x6A:
			name = "LD L,D"
			length = 1
			duration = 4
		case 0x6B:
			name = "LD L,E"
			length = 1
			duration = 4
		case 0x6C:
			name = "LD L,H"
			length = 1
			duration = 4
		case 0x6D:
			name = "LD L,L"
			length = 1
			duration = 4
		case 0x6E:
			name = "LD L,(HL)"
			length = 1
			duration = 8
		case 0x6F:
			name = "LD L,A"
			length = 1
			duration = 4
		case 0x70:
			name = "LD (HL),B"
			length = 1
			duration = 8
		case 0x71:
			name = "LD (HL),C"
			length = 1
			duration = 8
		case 0x72:
			name = "LD (HL),D"
			length = 1
			duration = 8
		case 0x73:
			name = "LD (HL),E"
			length = 1
			duration = 8
		case 0x74:
			name = "LD (HL),H"
			length = 1
			duration = 8
		case 0x75:
			name = "LD (HL),L"
			length = 1
			duration = 8
		case 0x76:
			name = "HALT"
			length = 1
			duration = 4
		case 0x77:
			name = "LD (HL),A"
			length = 1
			duration = 8
		case 0x78:
			name = "LD A,B"
			length = 1
			duration = 4
		case 0x79:
			name = "LD A,C"
			length = 1
			duration = 4
		case 0x7A:
			name = "LD A,D"
			length = 1
			duration = 4
		case 0x7B:
			name = "LD A,E"
			length = 1
			duration = 4
			registers8["A"] = registers8["E"]
		case 0x7C:
			name = "LD A,H"
			length = 1
			duration = 4
		case 0x7D:
			name = "LD A,L"
			length = 1
			duration = 4
		case 0x7E:
			name = "LD A,(HL)"
			length = 1
			duration = 8
		case 0x7F:
			name = "LD A,A"
			length = 1
			duration = 4
		case 0x80:
			name = "ADD A,B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x81:
			name = "ADD A,C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x82:
			name = "ADD A,D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x83:
			name = "ADD A,E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x84:
			name = "ADD A,H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x85:
			name = "ADD A,L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x86:
			name = "ADD A,(HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x87:
			name = "ADD A,A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x88:
			name = "ADC A,B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x89:
			name = "ADC A,C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x8A:
			name = "ADC A,D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x8B:
			name = "ADC A,E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x8C:
			name = "ADC A,H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x8D:
			name = "ADC A,L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x8E:
			name = "ADC A,(HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x8F:
			name = "ADC A,A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x90:
			name = "SUB B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x91:
			name = "SUB C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x92:
			name = "SUB D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x93:
			name = "SUB E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x94:
			name = "SUB H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x95:
			name = "SUB L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x96:
			name = "SUB (HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x97:
			name = "SUB A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x98:
			name = "SBC A,B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x99:
			name = "SBC A,C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x9A:
			name = "SBC A,D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x9B:
			name = "SBC A,E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x9C:
			name = "SBC A,H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x9D:
			name = "SBC A,L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0x9E:
			name = "SBC A,(HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0x9F:
			name = "SBC A,A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA0:
			name = "AND B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA1:
			name = "AND C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA2:
			name = "AND D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA3:
			name = "AND E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA4:
			name = "AND H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA5:
			name = "AND L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA6:
			name = "AND (HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xA7:
			name = "AND A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA8:
			name = "XOR B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xA9:
			name = "XOR C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xAA:
			name = "XOR D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xAB:
			name = "XOR E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xAC:
			name = "XOR H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xAD:
			name = "XOR L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xAE:
			name = "XOR (HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xAF:
			name = "XOR A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
			registers8["A"] ^= registers8["A"]
			if registers8["A"] == 0 {
				setZeroFlag()
			} else {
				unsetZeroFlag()
			}
		case 0xB0:
			name = "OR B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB1:
			name = "OR C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB2:
			name = "OR D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB3:
			name = "OR E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB4:
			name = "OR H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB5:
			name = "OR L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB6:
			name = "OR (HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xB7:
			name = "OR A"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB8:
			name = "CP B"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xB9:
			name = "CP C"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xBA:
			name = "CP D"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xBB:
			name = "CP E"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xBC:
			name = "CP H"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xBD:
			name = "CP L"
			length = 1
			duration = 4
			//flags = flagBytes{Z: "Z"}
		case 0xBE:
			name = "CP (HL)"
			length = 1
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xBF:
			name = "CP A"
			length = 1
			duration = 4
		case 0xC0:
			name = "RET NZ"
			length = 1
			duration = 20
			shortDuration = 8
		case 0xC1:
			name = "POP BC"
			length = 1
			duration = 12
		case 0xC2:
			name = "JP NZ,a16"
			length = 3
			duration = 16
			shortDuration = 12
		case 0xC3:
			name = "JP a16"
			length = 3
			duration = 16
		case 0xC4:
			name = "CALL NZ,a16"
			length = 3
			duration = 24
			shortDuration = 12
		case 0xC5:
			name = "PUSH BC"
			length = 1
			duration = 16
		case 0xC6:
			name = "ADD A,d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xC7:
			name = "RST 00H"
			length = 1
			duration = 16
		case 0xC8:
			name = "RET Z"
			length = 1
			duration = 20
			shortDuration = 8
		case 0xC9:
			name = "RET"
			length = 1
			duration = 16
		case 0xCA:
			name = "JP Z,a16"
			length = 3
			duration = 16
			shortDuration = 12
		case 0xCB:
			name = "PREFIX CB"
			length = 1
			duration = 4
			cbFlag = true
		case 0xCC:
			name = "CALL Z,a16"
			length = 3
			duration = 24
			shortDuration = 12
		case 0xCD:
			name = "CALL a16"
			length = 3
			duration = 24
			fmt.Println("Not implemented yet.")
		case 0xCE:
			name = "ADC A,d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xCF:
			name = "RST 08H"
			length = 1
			duration = 16
		case 0xD0:
			name = "RET NC"
			length = 1
			duration = 20
			shortDuration = 8
		case 0xD1:
			name = "POP DE"
			length = 1
			duration = 12
		case 0xD2:
			name = "JP NC,a16"
			length = 3
			duration = 16
			shortDuration = 12
		case 0xD4:
			name = "CALL NC,a16"
			length = 3
			duration = 24
			shortDuration = 12
		case 0xD5:
			name = "PUSH DE"
			length = 1
			duration = 16
		case 0xD6:
			name = "SUB d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xD7:
			name = "RST 10H"
			length = 1
			duration = 16
		case 0xD8:
			name = "RET C"
			length = 1
			duration = 20
			shortDuration = 8
		case 0xD9:
			name = "RETI"
			length = 1
			duration = 16
		case 0xDA:
			name = "JP C,a16"
			length = 3
			duration = 16
			shortDuration = 12
		case 0xDC:
			name = "CALL C,a16"
			length = 3
			duration = 24
			shortDuration = 12
		case 0xDE:
			name = "SBC A,d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xDF:
			name = "RST 18H"
			length = 1
			duration = 16
		case 0xE0:
			name = "LDH (a8),A"
			length = 2
			duration = 12
			address := 0xFF00 | uint16(argBytes[0])
			mmu.WriteByte(address, registers8["A"])
		case 0xE1:
			name = "POP HL"
			length = 1
			duration = 12
		case 0xE2:
			name = "LD (C),A"
			length = 2
			duration = 8
			address := 0xFF00 | uint16(registers8["C"])
			mmu.WriteByte(address, registers8["A"])
			length = 1 // Change later when figure out wtf
		case 0xE5:
			name = "PUSH HL"
			length = 1
			duration = 16
		case 0xE6:
			name = "AND d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xE7:
			name = "RST 20H"
			length = 1
			duration = 16
		case 0xE8:
			name = "ADD SP,r8"
			length = 2
			duration = 16
			//flags = flagBytes{Z: "Z"}
		case 0xE9:
			name = "JP (HL)"
			length = 1
			duration = 4
		case 0xEA:
			name = "LD (a16),A"
			length = 3
			duration = 16
		case 0xEE:
			name = "XOR d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xEF:
			name = "RST 28H"
			length = 1
			duration = 16
		case 0xF0:
			name = "LDH A,(a8)"
			length = 2
			duration = 12
		case 0xF1:
			name = "POP AF"
			length = 1
			duration = 12
			//flags = flagBytes{Z: "Z"}
		case 0xF2:
			name = "LD A,(C)"
			length = 2
			duration = 8
		case 0xF3:
			name = "DI"
			length = 1
			duration = 4
		case 0xF5:
			name = "PUSH AF"
			length = 1
			duration = 16
		case 0xF6:
			name = "OR d8"
			length = 2
			duration = 8
			//flags = flagBytes{Z: "Z"}
		case 0xF7:
			name = "RST 30H"
			length = 1
			duration = 16
		case 0xF8:
			name = "LD HL,SP+r8"
			length = 2
			duration = 12
			//flags = flagBytes{Z: "Z"}
		case 0xF9:
			name = "LD SP,HL"
			length = 1
			duration = 8
		case 0xFA:
			name = "LD A,(a16)"
			length = 3
			duration = 16
		case 0xFB:
			name = "EI"
			length = 1
			duration = 4
		case 0xFE:
			name = "CP d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			val := mmu.ReadByte(address)
			setSubtractionFlag()
			if registers8["A"] == val {
				setZeroFlag()
			} else {
				unsetZeroFlag()
			}
			if registers8["A"] < val {
				setCarryFlag()
				setHalfCarryFlag()
			} else {
				unsetCarryFlag()
				unsetHalfCarryFlag()
			}
		case 0xFF:
			name = "RST 38H"
			length = 1
			duration = 16
		}
	}

	fmt.Println("Instruction:")
	PrintInstruction(location, opcode, name, length, duration, shortDuration)
	fmt.Println("16-bit registers:", registers16)
	fmt.Println("8-bit registers:", registers8)
	fmt.Print("Flag register: ")
	printFlagRegister()
	fmt.Println()

	if !jump {
		location += uint16(length)
	} else {
		location = jumpTo
	}
	return location, cbFlag
}

func CBInstruction(opcode uint8, location uint16, argByte uint8) (string, uint8, uint8) {
	var name string
	var length uint8
	var duration uint8

	switch opcode {
	case 0x7C:
		name = "BIT 7,H"
		length = 2
		duration = 8

		unsetSubtractionFlag()
		setHalfCarryFlag()
		if registers8["H"]&uint8(64) == uint8(64) {
			setZeroFlag()
		} else {
			unsetZeroFlag()
		}
	}
	return name, length - 1, duration
}
