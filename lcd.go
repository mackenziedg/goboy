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
}

// ConvertTileToPixels converts an array of tile data into an array of pixel values
func (l *LCD) ConvertTileToPixels(tileData []uint8) [64]uint8 {
	var row1, row2 uint8
	tile := [64]uint8{}
	for i := uint8(0); i < 8; i++ {
		row1 = tileData[2*i]
		row2 = tileData[2*i+1]

		for b := uint8(0); b < 8; b++ {
			// TODO: Maybe do this with like shifting mask per loop?
			if CheckBit(&row1, b) {
				tile[i*8+(7-b)] |= 1
			}
			if CheckBit(&row2, b) {
				tile[i*8+(7-b)] |= 2
			}
		}
	}
	return tile
}

// LoadTileFromAddress loads a tile sized chunk of memory at a given address and processes it into an 8x8 tile as a 64-length array of uint8s.
func (l *LCD) LoadTileFromAddress(address uint16) [64]uint8 {
	tileData := l.mmu.memory[address : address+16]

	return l.ConvertTileToPixels(tileData)
}

// GetBGPixelArray returns an array of pixels which constitute the background of the GameBoy display.
func (l *LCD) GetBGPixelArray() [0x10000]uint8 {

	//Tile map in 0x9800-0x9BFF
	bgPixels := [0x10000]uint8{}

	// Loop over tilemap. Each index in the map points to an 8x8 tile.
	for ix, v := range l.mmu.memory[0x9800:0x9C00] {
		ULX := (ix * 8) % 256
		ULY := (ix / 32) * 2048
		curTile := l.LoadTileFromAddress(0x8000 + uint16(16*v))
		for j, px := range curTile {
			pxid := (j/8)*256 + j%8
			bgPixels[ULX+ULY+pxid] = px
		}
	}

	return bgPixels
}

func (l *LCD) Start() func() {

	i := uint64(0)
	start := time.Now()

	return func() {
		i++

		l.IncLY()

		if i%60 == 0 {
			fmt.Println("60 screen updates in", time.Now().Sub(start))
			start = time.Now()
		}
	}
}
