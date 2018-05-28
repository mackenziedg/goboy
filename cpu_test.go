package main

import (
	"strings"
	"testing"

	"github.com/mackenziedg/utils"
)

func TestCPU(t *testing.T) {
	mmu := &(MMU{})
	cpu := &(CPU{})
	cpu.Reset(mmu)

	notImplemented := []uint8{}
	nonExistent := utils.Set{}
	for _, v := range []uint8{0xD3, 0xE3, 0xE4, 0xF4, 0xDB, 0xEB, 0xEC, 0xFC, 0xDD, 0xED, 0xFD} {
		nonExistent.Add(v)
	}

	for i := 0; i < 0x100; i++ {

		if !nonExistent.Contains(uint8(i)) {
			_, _, breaking := cpu.Instruction(uint8(i), false, false)
			if breaking {
				notImplemented = append(notImplemented, uint8(i))
			}
			breaking = false
		}
	}
	if len(notImplemented) != 0 {
		t.Errorf("The following opcodes have not been implemented:\n")
		formatStr := strings.Repeat("%X, ", len(notImplemented))
		tmp := make([]interface{}, len(notImplemented))
		for i, val := range notImplemented {
			tmp[i] = val
		}
		t.Skipf(formatStr, tmp...)
	}

}