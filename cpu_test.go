package main

import (
	"testing"
)

func TestCPU(t *testing.T) {
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

	// Test rotations

	// RL
	reg := uint8(1)
	cpu.UnsetCarryFlag()

	cpu.Rotate(&reg, 9)
	if reg != 2 {
		t.Errorf("rotate gave %b instead of %b", reg, 2)
	}
	if cpu.GetCarryFlag() {
		t.Errorf("carry flag should not be set")
	}

	cpu.SetCarryFlag()

	cpu.Rotate(&reg, 9)
	if reg != 5 {
		t.Errorf("rotate gave %b instead of %b", reg, 5)
	}
	if cpu.GetCarryFlag() {
		t.Errorf("carry flag should not be set")
	}

	reg = 128
	cpu.Rotate(&reg, 9)
	if reg != 0 {
		t.Errorf("rotate gave %b instead of %b", reg, 1)
	}
	if !cpu.GetCarryFlag() {
		t.Errorf("carry flag should be set")
	}

	// RR
	reg = 1
	cpu.UnsetCarryFlag()
	cpu.Rotate(&reg, -9)
	if reg != 0 {
		t.Errorf("rotate gave %b instead of %b", reg, 0)
	}
	if !cpu.GetCarryFlag() {
		t.Errorf("carry flag should be set")
	}

	reg = 2
	cpu.SetCarryFlag()
	cpu.Rotate(&reg, -9)
	if reg != 129 {
		t.Errorf("rotate gave %b instead of %b", reg, 129)
	}
	if cpu.GetCarryFlag() {
		t.Errorf("carry flag should not be set")
	}

	// RLC
	reg = 1
	cpu.UnsetCarryFlag()
	cpu.RotateCarry(&reg, 9)
	if reg != 2 {
		t.Errorf("rotate gave %b instead of %b", reg, 129)
	}

	reg = 128
	cpu.UnsetCarryFlag()
	cpu.RotateCarry(&reg, 9)
	if reg != 1 {
		t.Errorf("rotate gave %b instead of %b", reg, 129)
	}
	if !cpu.GetCarryFlag() {
		t.Errorf("carry flag should be set")
	}

	// RRC
	reg = 1
	cpu.UnsetCarryFlag()
	cpu.RotateCarry(&reg, -9)
	if reg != 127 {
		t.Errorf("rotate gave %b instead of %b", reg, 129)
	}
	if !cpu.GetCarryFlag() {
		t.Errorf("carry flag should be set")
	}

	reg = 128
	cpu.UnsetCarryFlag()
	cpu.RotateCarry(&reg, 9)
	if reg != 1 {
		t.Errorf("rotate gave %b instead of %b", reg, 129)
	}
	if !cpu.GetCarryFlag() {
		t.Errorf("carry flag should be set")
	}

}
