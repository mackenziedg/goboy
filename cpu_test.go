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
