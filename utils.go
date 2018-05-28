package main

// Converts a pair of uint8 bytes to a single uint16 word.
// The pair is in low, high format.
func U8PairToU16(argBytes [2]uint8) uint16 {
	out := uint16(0)
	out += uint16(argBytes[1]) << 8
	out += uint16(argBytes[0])
	return out
}

// Converts a uint16 word to a pair of uint8 bytes.
// The pair is in low, high format.
func U16ToU8Pair(num uint16) [2]uint8 {
	lower := uint8(num & 0xFF)
	higher := uint8((num & 0xFF00) >> 8)
	return [2]uint8{lower, higher}
}

// CheckBit returns true if the bit of value is 1.
// Eg. CheckBit(0b10, 1) == true, CheckBit(0b10, 0) == false
func CheckBit(value *uint8, bit uint8) bool {
	return (*value & BitVal(bit)) == BitVal(bit)
}

// BitVal returns the value of the bit-th bit.
// Eg. BitVal(0) == 1, BitVal(7) == 128 etc.
func BitVal(bit uint8) uint8 {
	return (1 << bit)
}
