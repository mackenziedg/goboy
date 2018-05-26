package main

// Takes lower, higher pair
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
