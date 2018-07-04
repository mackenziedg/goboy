package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

// Define the pixel width and height of the GameBoy display
const (
	SCREENWIDTH  = 160
	SCREENHEIGHT = 144
)

// SDL is a struct which acts as the display for the GameBoy
type SDL struct{}

// Start acts as a wrapper for restarting and loading the GameBoy.
// TODO: This is not how it should work long term, but for now we're only ever loading one file so eh...
func (s *SDL) Start(gb *GameBoy) {
	gb.Reset()
	gb.LoadROMFromFile("./data/Tetris.gb")

	// Start the gameboy
	gbStepper := gb.Start()

	var winTitle = "goboy"
	var bgPixelWidth = 256
	var window *sdl.Window
	var renderer *sdl.Renderer
	var err error

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		if _, err = fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err); err != nil {
			panic(err)
		}
	}
	defer sdl.Quit()

	if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, SCREENWIDTH, SCREENHEIGHT, sdl.WINDOW_SHOWN); err != nil {
		if _, err = fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err); err != nil {
			panic(err)
		}
	}
	defer window.Destroy()

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		if _, err = fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err); err != nil {
			panic(err)
		}
	}
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()
	defer renderer.Destroy()

	renderer.Present()
	pxArray := [0x10000]uint8{}
	for {
		check(renderer.Clear())
		pxArray = gb.lcd.GetBGPixelArray()
		SCX := gb.mmu.ReadByte(0xFF43)
		SCY := gb.mmu.ReadByte(0xFF42)

		gbStepper()

		// Dump screen and crash when finished booting
		// if gb.cpu.PC.word > 0x100 {
		// 	f, err := os.Create("./BGScreenDump")
		// 	check(err)
		// 	defer f.Close()

		// 	var vramPixels bytes.Buffer

		// 	for y := 0; y < 256; y++ {
		// 		for x := 0; x < 256; x++ {
		// 			vramPixels.WriteString(strconv.Itoa(int(pxArray[x+SCREENWIDTH*y])))
		// 		}
		// 		vramPixels.WriteString("\n")
		// 	}
		// 	f.WriteString(vramPixels.String())

		// 	f3, err3 := os.Create("./Tiles")
		// 	check(err3)
		// 	defer f.Close()

		// 	var tilesBuf bytes.Buffer

		// 	for tid := 0; tid < 19; tid++ {
		// 		tile := gb.lcd.LoadTileFromAddress(0x8000 + uint16(tid*16))

		// 		tilesBuf.WriteString(strconv.Itoa(tid))
		// 		for i := 0; i < 64; i++ {
		// 			s := strconv.Itoa(int(tile[i]))

		// 			if i%8 == 0 {
		// 				tilesBuf.WriteString("\n")
		// 			}
		// 			tilesBuf.WriteString(s)
		// 		}
		// 		tilesBuf.WriteString("\n\n")
		// 	}
		// 	f3.WriteString(tilesBuf.String())

		// 	panic("finished booting")
		// }

		for x := uint8(0); x < SCREENWIDTH; x++ {
			for y := uint8(0); y < SCREENHEIGHT; y++ {
				gfx.PixelColor(renderer, int32(SCX+x), int32(SCY+y), s.ConvertColor(pxArray[int(SCX+x)+bgPixelWidth*int(SCX+y)]))
			}
		}

		renderer.Present()
	}
}

// ConvertColor is a helper class which converts a pixel value 0-3 into the corresponding display color for the screen.
func (s *SDL) ConvertColor(p uint8) sdl.Color {
	switch p {
	case 0:
		return sdl.Color{R: 255, G: 255, B: 255, A: 255}
	case 1:
		return sdl.Color{R: 170, G: 170, B: 170, A: 255}
	case 2:
		return sdl.Color{R: 80, G: 80, B: 80, A: 255}
	default:
		return sdl.Color{R: 0, G: 0, B: 0, A: 255}
	}
}
