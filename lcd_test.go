package main

import (
	"testing"
)

func TestLCD(t *testing.T) {
	lcd := &(LCD{})

	tile := []uint8{0x7D, 0x7C, 0x00, 0xC6, 0xC6, 0x00, 0x00, 0xFE, 0xC6, 0xC6, 0x00, 0xC6, 0xC6, 0x00, 0x00, 0x00}
	output := []uint8{0, 3, 3, 3, 3, 3, 0, 1, 2, 2, 0, 0, 0, 2, 2, 0, 1, 1, 0, 0, 0, 1, 1, 0, 2, 2, 2, 2, 2, 2, 2, 0, 3, 3, 0, 0, 0, 3, 3, 0, 2, 2, 0, 0, 0, 2, 2, 0, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	px := lcd.ConvertTileToPixels(tile)
	for i, p := range px {
		if p != output[i] {
			for x := 0; x < 8; x++ {
				t.Errorf("Row %d: Output: %v should be %v", x, px[8*x:8*x+8], output[8*x:8*x+8])
			}
			break
		}
	}

}
