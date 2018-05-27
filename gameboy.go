package main

import (
	"fmt"
	"io/ioutil"
)

// GameBoy is a wrapper for the hardware components.
// It controls the timing and linkage between the components.
type GameBoy struct {
	cpu *CPU
	mmu *MMU
	gpu *GPU
	apu *APU
}

// Reset() creates new hardware, links the memory to the processors, and resets each component.
func (g *GameBoy) Reset() {
	g.cpu = &(CPU{})
	g.mmu = &(MMU{})
	g.gpu = &(GPU{})
	g.apu = &(APU{})

	g.cpu.Reset(g.mmu)
	g.gpu.Reset(g.mmu)
	g.apu.Reset(g.mmu)
	g.mmu.Reset()
}

// LoadROMFromFile() loads a binary gameboy data file from a filepath string.
// Panics if any file read errors occur.
func (g *GameBoy) LoadROMFromFile(path string) {
	dat, err := ioutil.ReadFile(path)
	check(err)
	fmt.Printf("Data is %d bytes long.\n\n", len(dat))

	g.Reset()

	g.mmu.LoadCartridgeData(dat)
	g.Start()
}

// Starts the GameBoy. Returns a function which steps the CPU by one.
func (g *GameBoy) Start() func() {
	stepper := g.cpu.Start()
	return stepper
}
