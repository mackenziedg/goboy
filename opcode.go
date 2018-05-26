package main

import (
	"bufio"
	"fmt"
	"math/bits"
	"os"
)

type flagBytes struct {
	Z string
	N string
	H string
	C string
}

var Z uint8 = 7
var N uint8 = 6
var H uint8 = 5
var C uint8 = 4

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
	fmt.Printf("Byte %X\n\t%X|%s: length: %d, duration: %d/%d\n", location, opcode, name, length, duration, shortDuration)
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
	registers8["F"] |= bitVal(Z)
}

func unsetZeroFlag() {
	registers8["F"] &^= bitVal(Z)
}

func setSubtractionFlag() {
	registers8["F"] |= bitVal(N)
}

func unsetSubtractionFlag() {
	registers8["F"] &^= bitVal(N)
}

func setCarryFlag() {
	registers8["F"] |= bitVal(C)
}

func unsetCarryFlag() {
	registers8["F"] &^= bitVal(C)
}

func setHalfCarryFlag() {
	registers8["F"] |= bitVal(H)
}

func unsetHalfCarryFlag() {
	registers8["F"] &^= bitVal(H)
}

func getZeroFlag() bool {
	return checkBit(registers8["F"], Z)
}

func getSubtractionFlag() bool {
	return checkBit(registers8["F"], N)
}

func getHalfCarryFlag() bool {
	return checkBit(registers8["F"], H)
}

func getCarryFlag() bool {
	return checkBit(registers8["F"], C)
}

func checkBit(value uint8, bit uint8) bool {
	return (value & bitVal(bit)) == bitVal(bit)
}

func bitVal(bit uint8) uint8 {
	return (1 << bit)
}

func printRegisters() {
	fmt.Printf("\tStack pointer: %X\n\t\tA: %X, B: %X, C: %X, D: %X, E: %X, H: %X, L: %X\n",
		registers16["SP"],
		registers8["A"],
		registers8["B"],
		registers8["C"],
		registers8["D"],
		registers8["E"],
		registers8["H"],
		registers8["L"],
	)
}

func printFlagRegister() {
	fmt.Printf("\tFlag Register:\n\t\tZ: %t, N: %t, H: %t, C: %t\n",
		getZeroFlag(),
		getSubtractionFlag(),
		getHalfCarryFlag(),
		getCarryFlag(),
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
		// INC/DEC

		case 0x04:
			name = "INC B"
			length = 1
			duration = 4
			registers8["B"] += 1
			unsetSubtractionFlag()
			if registers8["B"] == 0 {
				setZeroFlag()
				setHalfCarryFlag()
			}
		case 0x05:
			name = "DEC B"
			length = 1
			duration = 4
			registers8["B"] -= 1
			setSubtractionFlag()
			if registers8["B"] == 0 {
				setZeroFlag()
			} else if registers8["B"] == 255 {
				setHalfCarryFlag()
			}
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
			setSubtractionFlag()
			if registers8["C"] == 0 {
				setZeroFlag()
			} else if registers8["C"] == 255 {
				setHalfCarryFlag()
			}
		case 0x15:
			name = "DEC D"
			length = 1
			duration = 4
			setSubtractionFlag()
			if registers8["D"] == 0 {
				setZeroFlag()
			} else if registers8["D"] == 255 {
				setHalfCarryFlag()
			}
		case 0x1D:
			name = "DEC E"
			length = 1
			duration = 4
			setSubtractionFlag()
			if registers8["E"] == 0 {
				setZeroFlag()
			} else if registers8["E"] == 255 {
				setHalfCarryFlag()
			}
		case 0x24:
			name = "INC H"
			length = 1
			duration = 4
			registers8["H"] += 1
			unsetSubtractionFlag()
			if registers8["H"] == 0 {
				setZeroFlag()
				setHalfCarryFlag()
			}

		case 0x13:
			name = "INC DE"
			length = 1
			duration = 8
			combined := u8PairToU16([2]uint8{registers8["E"], registers8["D"]})
			combined += 1
			vals := splitUint16(combined)
			registers8["D"] = vals[1]
			registers8["E"] = vals[0]
		case 0x23:
			name = "INC HL"
			length = 1
			duration = 8
			combined := u8PairToU16([2]uint8{registers8["L"], registers8["H"]})
			combined += 1
			vals := splitUint16(combined)
			registers8["H"] = vals[1]
			registers8["L"] = vals[0]

			// LD R,d8
		case 0x3E:
			name = "LD A,d8"
			length = 2
			duration = 8
			byte_ := argBytes[0]
			registers8["A"] = byte_
		case 0x06:
			name = "LD B,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			registers8["B"] = mmu.ReadByte(address)
		case 0x0E:
			name = "LD C,d8"
			length = 2
			duration = 8
			byte_ := argBytes[0]
			registers8["C"] = byte_
		case 0x16:
			name = "LD D,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			registers8["D"] = mmu.ReadByte(address)
		case 0x1E:
			name = "LD E,d8"
			length = 2
			duration = 8
			address := uint16(argBytes[0])
			registers8["E"] = mmu.ReadByte(address)

			// LD R,R
		case 0x4F:
			name = "LD C,A"
			length = 1
			duration = 4
			registers8["C"] = registers8["A"]
		case 0x7B:
			name = "LD A,E"
			length = 1
			duration = 4
			registers8["A"] = registers8["E"]

			// LD R,a16
		case 0x1A:
			name = "LD A,(DE)"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{registers8["E"], registers8["D"]})
			registers8["A"] = mmu.ReadByte(address)

			// LD RR,d16
		case 0x11:
			name = "LD DE,d16"
			length = 3
			duration = 12
			registers8["E"] = argBytes[0]
			registers8["D"] = argBytes[1]
		case 0x21:
			name = "LD HL,d16"
			length = 3
			duration = 12
			registers8["L"] = argBytes[0]
			registers8["H"] = argBytes[1]
		case 0x32:
			name = "LD (HL-),A"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{registers8["L"], registers8["H"]})
			mmu.WriteByte(address, registers8["A"])
			combined := u8PairToU16([2]uint8{registers8["L"], registers8["H"]})
			combined -= 1
			vals := splitUint16(combined)
			registers8["H"] = vals[1]
			registers8["L"] = vals[0]
			setSubtractionFlag()
			if combined == 0 {
				setZeroFlag()
			} else if combined == 65535 {
				setCarryFlag()
			}
		case 0x22:
			name = "LD (HL+),A"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{registers8["L"], registers8["H"]})
			mmu.WriteByte(address, registers8["A"])
			combined := u8PairToU16([2]uint8{registers8["L"], registers8["H"]})
			combined += 1
			vals := splitUint16(combined)
			registers8["H"] = vals[1]
			registers8["L"] = vals[0]
		case 0x31:
			name = "LD SP,d16"
			length = 3
			duration = 12
			address := u8PairToU16(argBytes)
			registers16["SP"] = address

			// LD address,R
		case 0xE2:
			name = "LD (C),A"
			length = 2
			duration = 8
			address := 0xFF00 | uint16(registers8["C"])
			mmu.WriteByte(address, registers8["A"])
			length = 1 // Change later when figure out wtf
		case 0x77:
			name = "LD (HL), A"
			length = 1
			duration = 8
			address := u8PairToU16([2]uint8{registers8["L"], registers8["H"]})
			mmu.WriteByte(address, registers8["A"])
		case 0xE0:
			name = "LDH (a8),A"
			length = 2
			duration = 12
			address := 0xFF00 | uint16(argBytes[0])
			mmu.WriteByte(address, registers8["A"])

		case 0x17:
			name = "RLA"
			length = 1
			duration = 4
			toCarry := false
			if checkBit(registers8["A"], 7) {
				toCarry = true
			}
			registers8["A"] = bits.RotateLeft8(registers8["A"], 9)
			if getCarryFlag() {
				registers8["A"] |= 1
			}
			if toCarry {
				setCarryFlag()
			} else {
				unsetCarryFlag()
			}
			unsetHalfCarryFlag()
			unsetSubtractionFlag()

			// Jump
		case 0x18:
			name = "JR r8"
			length = 2
			duration = 12
			fmt.Println("Not implemented")
		case 0x20:
			name = "JR NZ,r8"
			length = 2
			duration = 12
			shortDuration = 8
			arg := uint16(argBytes[0])
			if !getZeroFlag() {
				jump = true
				if arg > 128 {
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
			fmt.Println("Not implemented")

			// Stack ops
		case 0xC5:
			name = "PUSH BC"
			length = 1
			duration = 16
			data := u8PairToU16([2]uint8{registers8["C"], registers8["B"]})
			mmu.WriteWord(registers16["SP"], data)
			registers16["SP"] -= 2
		case 0xC1:
			name = "POP BC"
			length = 1
			duration = 12
			registers16["SP"] += 2
			data := mmu.ReadWord(registers16["SP"])
			bytePair := splitUint16(data)
			registers8["C"] = bytePair[0]
			registers8["B"] = bytePair[1]
		case 0xC9:
			name = "RET"
			length = 1
			duration = 20
			shortDuration = 8
			jump = true
			registers16["SP"] += 2
			jumpTo = mmu.ReadWord(registers16["SP"])
		case 0xCD:
			name = "CALL a16"
			length = 3
			duration = 24
			mmu.WriteWord(registers16["SP"], location)
			registers16["SP"] -= 2
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
			registers8["A"] ^= registers8["A"]
			if registers8["A"] == 0 {
				setZeroFlag()
			} else {
				unsetZeroFlag()
			}
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
		default:
			name = "Not implemented"
			length = 1
			duration = 1
		}
	}

	// Logging
	PrintInstruction(location, opcode, name, length, duration, shortDuration)
	printRegisters()
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
		if checkBit(registers8["H"], 7) {
			setZeroFlag()
		} else {
			unsetZeroFlag()
		}
	case 0x11:
		name = "RL C"
		length = 2
		duration = 8
		toCarry := false
		if checkBit(registers8["C"], 7) {
			toCarry = true
		}
		registers8["C"] = bits.RotateLeft8(registers8["C"], 9)
		if getCarryFlag() {
			registers8["C"] |= 1
		}
		if toCarry {
			setCarryFlag()
		} else {
			unsetCarryFlag()
		}
		unsetHalfCarryFlag()
		unsetSubtractionFlag()
	}
	return name, length - 1, duration
}
