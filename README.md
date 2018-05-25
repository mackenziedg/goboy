# CPU Layout

## Registers

The GameBoy has internal 8-bit registers A, B, C, D, E, F, H, and L.
They can be used in pairs as 16-bit registers AF, BC, DE, and HL.
There are two additional 16-bit registers program counter (PC) and stack pointer (SP).

### CPU Flags

The F register holds the CPU flags in the form

| Bit | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
|-----|---|---|---|---|---|---|---|---|
| Abbreviation | Z | N | H | C | 0 | 0 | 0 | 0 |
| Meaning | Zero | Subtract | Half-carry | Carry | | | | |

The lower 4 bits are always 0 even if written to.

