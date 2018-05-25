package main

import (
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func argsToUint16(args []uint8) uint16 {
	out := uint16(0)
	out += uint16(args[1]) << 8
	out += uint16(args[0])
	return out
}

func main() {

	dat, err := ioutil.ReadFile("./data/DMG_ROM.bin")
	check(err)
	for i := 0; i < len(dat); i++ {
		Instruction(dat[i])
	}
}
