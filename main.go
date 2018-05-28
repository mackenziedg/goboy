package main

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Create a new GameBoy, clear it, and read in cartridge data.
	var gb = &(GameBoy{})
	gb.Reset()
	gb.LoadROMFromFile("./data/Tetris.gb")

	// Start the gameboy
	gb.Start()
}
