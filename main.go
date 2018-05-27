package main

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const PIXELWIDTH, PIXELHEIGHT int32 = 160, 154
const SCALEFACTOR = 1

func main() {
	var gb = &(GameBoy{})
	gb.Reset()
	gb.LoadROMFromFile("./data/Tetris.gb")
	GBStep, mmu := gb.Start()

	var winTitle string = "GoBoy"
	var winWidth, winHeight int32 = PIXELWIDTH * SCALEFACTOR, PIXELHEIGHT * SCALEFACTOR
	var window *sdl.Window
	var renderer *sdl.Renderer
	var err error

	check(sdl.Init(sdl.INIT_EVERYTHING))
	defer sdl.Quit()

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
	check(err)
	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	check(err)
	renderer.Clear()
	defer renderer.Destroy()

	const VRAMSTART, VRAMEND uint16 = 0x8000, 0x97FF
	for {
		renderer.Clear()
		data := mmu.memory[VRAMSTART:VRAMEND]
		frameData := LoadTiles(data)

		for pY := int32(0); pY < PIXELHEIGHT; pY++ {
			for pX := int32(0); pX < PIXELWIDTH; pX++ {
				pC := sdl.Color{frameData[pY*PIXELWIDTH+pX],
					frameData[pY*PIXELWIDTH+pX],
					frameData[pY*PIXELWIDTH+pX],
					255}
				gfx.PixelColor(renderer, pX, pY, pC)
			}
		}
		renderer.Present()

		GBStep()
	}
}

func LoadTiles(vram []uint8) [PIXELHEIGHT * PIXELWIDTH]uint8 {
	frame := [PIXELHEIGHT * PIXELWIDTH]uint8{}

	for i := 0; i < len(vram); i += 128 {
		d := LoadTile(vram, uint16(i))
		for j, v := range d {
			frame[i+j] = v
		}
	}

	return frame
}

func LoadTile(vram []uint8, address uint16) [64]uint8 {
	tile := [64]uint8{}
	var line uint8

	for i := uint16(0); i < 16; i += 2 {
		lowByte := vram[address]
		highByte := vram[address+1]

		for b := uint8(0); b < 8; b++ {
			low1 := checkBit(lowByte, b)
			high1 := checkBit(highByte, b)

			if !high1 {
				if low1 {
					tile[line*8+b] = 85
				}
			} else {
				if !low1 {
					tile[line*8+b] = 85
				} else {
					tile[line*8+b] = 255
				}
			}
		}
		line++
	}
	return tile
}
