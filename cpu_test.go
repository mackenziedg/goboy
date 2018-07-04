package main

import (
	"testing"
)

func TestCPUInstructions(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	// notImplemented := []uint8{}
	// nonExistent := utils.Set{}
	// for _, v := range []uint8{0xD3, 0xE3, 0xE4, 0xF4, 0xDB, 0xEB, 0xEC, 0xFC, 0xDD, 0xED, 0xFD} {
	// 	nonExistent.Add(v)
	// }

	// in := true
	// for i := 0; i < 0x100; i++ {

	// 	if !nonExistent.Contains(uint8(i)) {
	// 		_, in = cpu.opcodeMap[uint8(i)]
	// 		if !in {
	// 			notImplemented = append(notImplemented, uint8(i))
	// 		}
	// 	}
	// }
	// if len(notImplemented) != 0 {
	// 	t.Errorf("The following opcodes have not been implemented:\n")
	// 	formatStr := strings.Repeat("%X, ", len(notImplemented))
	// 	tmp := make([]interface{}, len(notImplemented))
	// 	for i, val := range notImplemented {
	// 		tmp[i] = val
	// 	}
	// 	t.Skipf(formatStr, tmp...)
	// }

}

func TestRL(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		reg    uint8
		carry  bool
		output uint8
	}{
		{16, false, 32},
		{1, true, 3},
		{128, false, 0},
		{128, true, 1},
	}

	for _, table := range tables {
		if table.carry {
			cpu.SetCarryFlag()
		} else {
			cpu.UnsetCarryFlag()
		}
		orig := table.reg
		cpu.RotateLeft(&table.reg)

		if table.reg != table.output {
			t.Errorf("RL %d gave %b instead of %b", orig, table.reg, table.output)
		}
	}
}

func TestRR(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		reg    uint8
		carry  bool
		output uint8
	}{
		{16, false, 8},
		{2, true, 129},
		{1, false, 0},
		{1, true, 128},
	}

	for _, table := range tables {
		if table.carry {
			cpu.SetCarryFlag()
		} else {
			cpu.UnsetCarryFlag()
		}
		orig := table.reg
		cpu.RotateRight(&table.reg)

		if table.reg != table.output {
			t.Errorf("RR %d gave %b instead of %b", orig, table.reg, table.output)
		}
	}
}

func TestRLC(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		reg    uint8
		output uint8
		carry  bool
	}{
		{16, 16, false},
		{128, 129, true},
	}

	for _, table := range tables {
		orig := table.reg
		cpu.RotateLeftCarry(&table.reg)

		if table.reg != table.output {
			t.Errorf("RLC %d, gave %b instead of %b", orig, table.reg, table.output)
		}
		if cpu.GetCarryFlag() != table.carry {
			t.Errorf("RLC %d, gave carry = %v instead of %v", orig, cpu.GetCarryFlag(), table.carry)
		}
	}
}

func TestRRC(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		reg    uint8
		output uint8
		carry  bool
	}{
		{16, 16, false},
		{1, 129, true},
	}

	for _, table := range tables {
		if table.carry {
			cpu.SetCarryFlag()
		} else {
			cpu.UnsetCarryFlag()
		}
		orig := table.reg
		cpu.RotateRightCarry(&table.reg)

		if table.reg != table.output {
			t.Errorf("RRC %d, gave %b instead of %b", orig, table.reg, table.output)
		}
		if cpu.GetCarryFlag() != table.carry {
			t.Errorf("RRC %d, gave carry = %v instead of %v", orig, cpu.GetCarryFlag(), table.carry)

		}
	}
}

func TestStack(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	cpu.PushWord(0xFFFF)

	var reg uint16

	cpu.PopWord(&reg)

	if reg != 0xFFFF {
		t.Error("Popped value not the same as pushed.")
	}
}

func TestIncDec8(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	reg := uint8(0)

	cpu.Inc8(&reg)
	if reg != 1 {
		t.Error("Inc8 not working properly.")
	}

	cpu.Dec8(&reg)
	if reg != 0 {
		t.Error("Dec8 not working properly.")
	}

	cpu.Dec8(&reg)
	if reg != 255 {
		t.Error("Dec8 not underflowing properly.")
	}

	cpu.Inc8(&reg)
	if reg != 0 {
		t.Error("Inc8 not underflowing properly.")
	}
}

func TestIncDec16(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	reg := uint16(0)

	cpu.Inc16(&reg)
	if reg != 1 {
		t.Error("Inc16 not working properly.")
	}

	cpu.Dec16(&reg)
	if reg != 0 {
		t.Error("Dec16 not working properly.")
	}

	cpu.Dec16(&reg)
	if reg != 0xFFFF {
		t.Error("Dec16 not underflowing properly.")
	}

	cpu.Inc16(&reg)
	if reg != 0 {
		t.Error("Inc16 not underflowing properly.")
	}
}

func TestAddReg8(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		AF  uint8
		reg uint8
		sum uint8
	}{
		{0, 1, 1},
		{255, 1, 0},
	}

	for _, table := range tables {
		*cpu.AF.hi = table.AF
		cpu.AddReg8(&table.reg)
		if *cpu.AF.hi != table.sum {
			t.Errorf("AddReg8 error. %X + %X = %X, is instead %X", table.AF, table.reg, table.sum, *cpu.AF.hi)
		}
	}
}

func TestAddReg16(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		HL  uint16
		reg uint16
		sum uint16
	}{
		{0, 1, 1},
		{0xFFFF, 1, 0},
	}

	for _, table := range tables {
		cpu.HL.word = table.HL
		cpu.AddReg16(&table.reg)
		if cpu.HL.word != table.sum {
			t.Errorf("AddReg16 error. %X + %X = %X, is instead %X", table.HL, table.reg, table.sum, cpu.HL.word)
		}
	}
}

func TestSubReg(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		AF  uint8
		reg uint8
		sum uint8
	}{
		{0, 1, 255},
		{255, 1, 254},
		{1, 1, 0},
	}

	for _, table := range tables {
		*cpu.AF.hi = table.AF
		cpu.SubReg(&table.reg)
		if *cpu.AF.hi != table.sum {
			t.Errorf("SubReg error. %X - %X = %X, is instead %X", table.AF, table.reg, table.sum, *cpu.AF.hi)
		}
	}
}

func TestXorReg(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		AF  uint8
		reg uint8
		xor uint8
	}{
		{5, 3, 6},
		{1, 1, 0},
	}

	for _, table := range tables {
		*cpu.AF.hi = table.AF
		cpu.XorReg(&table.reg)
		if *cpu.AF.hi != table.xor {
			t.Errorf("XorReg error. %X ^ %X = %X, is instead %X", table.AF, table.reg, table.xor, *cpu.AF.hi)
		}
	}
}

func TestOrReg(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		AF  uint8
		reg uint8
		or  uint8
	}{
		{1, 2, 3},
		{0, 0, 0},
	}

	for _, table := range tables {
		*cpu.AF.hi = table.AF
		cpu.OrReg(&table.reg)
		if *cpu.AF.hi != table.or {
			t.Errorf("OrReg error. %X ^ %X = %X, is instead %X", table.AF, table.reg, table.or, *cpu.AF.hi)
		}
	}
}

func TestAndReg(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	tables := []struct {
		AF  uint8
		reg uint8
		and uint8
	}{
		{1, 2, 0},
		{3, 1, 1},
	}

	for _, table := range tables {
		*cpu.AF.hi = table.AF
		cpu.AndReg(&table.reg)
		if *cpu.AF.hi != table.and {
			t.Errorf("AndReg error. %X ^ %X = %X, is instead %X", table.AF, table.reg, table.and, *cpu.AF.hi)
		}
	}
}
