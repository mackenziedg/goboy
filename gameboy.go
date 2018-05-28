package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// GameBoy is a wrapper for the hardware components.
// It controls the timing and linkage between the components.
type GameBoy struct {
	cpu *CPU
	mmu *MMU
	lcd *LCD
	apu *APU
}

// Reset() creates new hardware, links the memory to the processors, and resets each component.
func (g *GameBoy) Reset() {
	g.cpu = &(CPU{})
	g.mmu = &(MMU{})
	g.lcd = &(LCD{})
	g.apu = &(APU{})

	g.cpu.Reset(g.mmu)
	g.lcd.Reset(g.mmu)
	g.apu.Reset(g.mmu)
	g.mmu.Reset()
}

// LoadROMFromFile() loads a binary gameboy data file from a filepath string.
// Panics if any file read errors occur.
func (g *GameBoy) LoadROMFromFile(path string) {
	pathSplit := strings.Split(path, ",")
	fmt.Printf("Loading %s.\n", pathSplit[len(pathSplit)-1])
	dat, err := ioutil.ReadFile(path)
	check(err)
	fmt.Printf("Data is %d bytes long.\n", len(dat))

	g.Reset()

	g.mmu.LoadCartridgeData(dat)
	titleBytes := g.mmu.memory[0x0134:0x0142]

	fmt.Print("Now playing ")
	for _, v := range titleBytes {
		if v != 0x0 {
			fmt.Printf("%c", v)
		}
	}
	fmt.Print("!\n")
	g.Start()
}

// Starts the GameBoy. Returns a function which steps the CPU by one.
func (g *GameBoy) Start() {
	cpuStepper := g.cpu.Start()
	lcdStepper := g.lcd.Start()

	i := uint64(0)
	for {
		i++
		cpuStepper()
		lcdStepper()
	}

}
