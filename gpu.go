package main

var COLORMAP = map[uint8][3]uint8{
	3: [3]uint8{0x0, 0x0, 0x0},
	2: [3]uint8{0x60, 0x60, 0x60},
	1: [3]uint8{0xC0, 0xC0, 0xC0},
	0: [3]uint8{0xFF, 0xFF, 0xFF},
}

type GPU struct {
	mmu *MMU
}

func (g *GPU) Reset(mmu *MMU) {
}
