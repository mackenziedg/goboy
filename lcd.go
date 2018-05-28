package main

import (
	"fmt"
	"time"
)

type LCD struct {
	mmu *MMU
}

func (l *LCD) Reset(m *MMU) {
	l.mmu = m
}

func (l *LCD) CheckInterrupts() {

}

func (l *LCD) VBlankOn() {
	l.mmu.memory[0xFF0F] |= 1
}

func (l *LCD) VBlankOff() {
	l.mmu.memory[0xFF0F] &= 30
}

func (l *LCD) IncLY() {
	ly := l.mmu.memory[0xFF44]
	ly++

	if ly == 154 {
		ly = 0
	}
	l.mmu.memory[0xFF44] = ly

	if ly > 143 {
		l.VBlankOn()
	} else {
		l.VBlankOff()
	}
	time.Sleep(16667 * time.Microsecond) // About 60 Hz
}

// LoadTileFromAddress loads a tile sized chunk of memory at a given address and processes it into an 8x8 tile as a 64-length array of uint8s.
func (l *LCD) LoadTileFromAddress(address uint16) [64]uint8 {
	tileData := l.mmu.memory[address : address+16]
	tile := [64]uint8{}

	var row1, row2 uint8
	for i := uint8(0); i < 8; i++ {
		row1 = tileData[2*i]
		row2 = tileData[2*i+1]

		for b := uint8(0); b < 8; b++ {
			// TODO: Maybe do this with like shifting mask per loop?
			if CheckBit(&row1, b) {
				tile[i*8+b] |= 2
			}
			if CheckBit(&row2, b) {
				tile[i*8+b] |= 1
			}
		}
	}

	return tile
}

func (l *LCD) Start() func() {

	i := uint64(0)
	start := time.Now()
	return func() {
		i++

		l.IncLY()

		if i%60 == 0 {
			for j := uint16(0x8000); j < 0x8FFF; j += 0x10 {
				l.LoadTileFromAddress(0x8010 + j)
			}
			fmt.Println("60 screen updates in", time.Now().Sub(start))
			start = time.Now()
		}
	}
}
