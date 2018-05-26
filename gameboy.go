package main

import (
	"fmt"
	"io/ioutil"
)

type gb interface {
	Reset()
	LoadROMFromFile(path string)
	Start()
}

type GameBoy struct {
	cpu *CPU
	mmu *MMU
	gpu *GPU
	apu *APU
}

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

func (g *GameBoy) LoadROMFromFile(path string) {
	dat, err := ioutil.ReadFile(path)
	check(err)
	fmt.Printf("Data is %d bytes long.\n\n", len(dat))

	g.Reset()

	g.cpu.LoadCartridgeData(dat)
	g.Start()
}

func (g *GameBoy) Start() {
	g.cpu.Start()
}
