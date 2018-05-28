package main

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
	if ly == 153 {
		ly = 0
	} else {
		ly++
	}
	l.mmu.memory[0xFF44] = ly

	if ly < 154 && ly > 143 {
		l.VBlankOn()
	} else {
		l.VBlankOff()
	}

}

func (l *LCD) Start() func() {

	i := uint64(0)
	return func() {
		i++
		l.IncLY()
	}
}
