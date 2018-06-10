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

// Reset creates new hardware, links the memory to the processors, and resets each component.
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

// CheckCartridgeHeader checks and prints the cartridge header information,
//including game title, memory type, and size.
func (g *GameBoy) CheckCartridgeHeader() {

	// Game title in upper-case ASCII always here
	titleBytes := g.mmu.memory[0x0134:0x0142]

	// Load information about the cartridge memory type etc.
	// to make sure we can run it with this crappy emulator
	memInfoBytes := g.mmu.memory[0x0147:0x014B]
	switch memInfoBytes[0] {
	case 0x0:
		fmt.Println("Cartridge uses ROM only, good to go!")
	default:
		panic(fmt.Errorf("Cartridge uses ROM plus other stuff. Abort!\n(Cartridge type %X)", memInfoBytes[0]))
	}

	romSizeString := ""
	switch memInfoBytes[1] {
	case 0x0:
		romSizeString = "32 KB"
	case 0x1:
		romSizeString = "64 KB"
	default:
		panic(fmt.Errorf("cartridge ROM size %X not supported", memInfoBytes[1]))

	}
	fmt.Printf("ROM Size is %s.\n", romSizeString)

	ramSizeString := ""
	switch memInfoBytes[2] {
	case 0x0:
		ramSizeString = "No cartridge RAM."
	default:
		panic(fmt.Errorf("cartridge RAM not supported"))
	}
	fmt.Printf("%s\n", ramSizeString)

	destination := "Japanese"
	if memInfoBytes[3] == 1 {
		destination = "non-Japanese"
	}
	fmt.Printf("Cartridge is for %s destination.\n", destination)

	fmt.Print("\nNow playing ")
	for _, v := range titleBytes {
		if v != 0x0 {
			fmt.Printf("%c", v)
		}
	}
	fmt.Print("!\n========================================\n")
}

// LoadROMFromFile loads a binary GameBoy data file from a filepath string.
// Panics if any file read errors occur.
func (g *GameBoy) LoadROMFromFile(path string) {
	pathSplit := strings.Split(path, ",")
	fmt.Printf("Loading %s.\n", pathSplit[len(pathSplit)-1])
	dat, err := ioutil.ReadFile(path)
	check(err)
	fmt.Printf("Loaded 0x%X bytes of data.\n", len(dat))

	g.Reset()

	g.mmu.LoadCartridgeData(dat)
	g.CheckCartridgeHeader()
}

// Start starts the GameBoy.
func (g *GameBoy) Start() {
	cpuStepper := g.cpu.Start()
	lcdStepper := g.lcd.Start()

	// i := uint64(0)
	go func() {
		for {
			cpuStepper()
		}
		<-done
	}()
	go func() {
		for {
			lcdStepper()
		}
		<-done
	}()
}
