package main

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var gb = &(GameBoy{})
	gb.Reset()
	gb.LoadROMFromFile("./data/Tetris.gb")
	gb.Start()
}
