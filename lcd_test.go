package main

import (
	"testing"
)

func TestLCD(t *testing.T) {
	lcd := &(LCD{})

	tile := []uint8{85, 51, 85, 51, 85, 51, 85, 51, 85, 51, 85, 51, 85, 51, 85, 51}

	px := lcd.ConvertTileToPixels(tile)
	firstRow := px[0:8]
	for i, p := range firstRow {
		if int(p) != (3 - i%4) {
			t.Errorf("Pixel %d with value %d does not equal %d.", i, p, 3-i%4)
		}
	}

}
