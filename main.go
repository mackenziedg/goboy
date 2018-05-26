package main

import (
	"fmt"
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	dat, err := ioutil.ReadFile("./data/DMG_ROM.bin")
	check(err)
	fmt.Printf("Data is %d bytes long.\n", len(dat))
	RunFile(dat)
}
