package main

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var done = make(chan int)

func main() {
	// Create a new GameBoy, clear it, and read in cartridge data.
	var gb = &(GameBoy{})
	var sdl = &(SDL{})
	sdl.Start(gb)

}
