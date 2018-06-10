package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

// SDL is a struct which acts as the display for the GameBoy
type SDL struct{}

// Start acts as a wrapper for restarting and loading the GameBoy.
// TODO: This is not how it should work long term, but for now we're only ever loading one file so eh...
func (s *SDL) Start(gb *GameBoy) func() {
	gb.Reset()
	gb.LoadROMFromFile("./data/Tetris.gb")

	check(sdl.Init(sdl.INIT_EVERYTHING))
	defer sdl.Quit()

	window, err := sdl.CreateWindow("GoBoy", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	check(err)
	defer window.Destroy()

	surface, err := window.GetSurface()
	check(err)
	surface.FillRect(nil, 0)

	rect := sdl.Rect{0, 0, 200, 200}
	surface.FillRect(&rect, 0xffff0000)
	window.UpdateSurface()

	// Start the gameboy
	gb.Start()

	return func() {

		var winTitle = "SDL2 GFX"
		var winWidth, winHeight int32 = 256, 256
		var window *sdl.Window
		var renderer *sdl.Renderer
		var err error

		if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err)
		}
		defer sdl.Quit()

		if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		}
		defer window.Destroy()

		if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		}
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		defer renderer.Destroy()

		renderer.Present()
		pxArray := [0x10000]uint8{}
		for {
			renderer.Clear()
			pxArray = gb.lcd.GetTileMap()
			for i, v := range pxArray {
				// fmt.Println(i/256, i%256)
				gfx.PixelColor(renderer, int32(i%256), int32(i/256), s.ConvertColor(v))
			}
			renderer.Present()
		}
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
