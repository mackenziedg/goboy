package main

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var gb = new(GameBoy)
	gb.Reset()
	// gb.LoadROMFromFile("./data/DMG_ROM.bin")
	gb.Start()
}
