package main

import (
	"fmt"
	"time"
)

type LCD struct {
	mmu *MMU
}

func (l *LCD) Reset(m *MMU) {
	l.mmu = m
}

func (l *LCD) CheckInterrupts() {

}

func (l *LCD) VBlankOn() {
	l.mmu.memory[0xFF0F] |= 1
}

func (l *LCD) VBlankOff() {
	l.mmu.memory[0xFF0F] &= 30
}

func (l *LCD) IncLY() {
	ly := l.mmu.memory[0xFF44]
	if ly == 154 {
		ly = 0
	} else {
		ly++
	}
	l.mmu.memory[0xFF44] = ly

	if ly > 143 {
		l.VBlankOn()
	} else {
		l.VBlankOff()
	}
	time.Sleep(16667 * time.Microsecond) // About 60 Hz

}

func (l *LCD) Start() func() {

	i := uint64(0)
	start := time.Now()
	return func() {
		i++

		l.IncLY()

		if i%60 == 0 {
			fmt.Println("60 screen updates in", time.Now().Sub(start))
			start = time.Now()
		}
	}
}
