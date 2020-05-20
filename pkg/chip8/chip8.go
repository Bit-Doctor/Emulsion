package chip8

import (
	"math/rand"
)

// chip8 is based on Cowgod's Chip-8 Technical Reference v1.0
// Available at http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#1.0
type chip8 struct {
	// The 4096 bytes of memory.
	//
	// Memory Map:
	// +---------------+= 0xFFF (4095) End of Chip-8 RAM
	// |               |
	// |               |
	// |               |
	// |               |
	// |               |
	// | 0x200 to 0xFFF|
	// |     Chip-8    |
	// | Program / Data|
	// |     Space     |
	// |               |
	// |               |
	// |               |
	// +- - - - - - - -+= 0x600 (1536) Start of ETI 660 Chip-8 programs
	// |               |
	// |               |
	// |               |
	// +---------------+= 0x200 (512) Start of most Chip-8 programs
	// | 0x000 to 0x1FF|
	// | Reserved for  |
	// |  interpreter  |
	// +---------------+= 0x000 (0) Start of Chip-8 RAM
	memory [4096]byte

	// Chip-8 has 16 general purpose 8-bit registers, usually referred to as Vx, where x is a hexadecimal digit (0 through F)
	// The VF register should not be used by any program, as it is used as a flag by some instructions
	v [16]byte

	// Chip-8 has a 16-bit register called I.
	// This register is generally used to store memory addresses, so only the lowest (rightmost) 12 bits are usually used.
	i uint16

	// Chip-8 also has a special purpose 8-bit register, for the delay timer.
	// When this register is non-zero, it's automatically decremented at a rate of 60Hz.
	dt byte

	// Chip-8 also has a special purpose 8-bit register, for the sound timer.
	// When this register is non-zero, it's automatically decremented at a rate of 60Hz.
	st byte

	// The program counter (pc) should be 16-bit, and is used to store the currently executing address.
	pc uint16

	// The stack pointer (sp) can be 8-bit, it is used to point to the topmost level of the stack.
	sp byte

	// The stack is an array of 16-bit values, used to store the address that the interpreter shoud return to when finished with a subroutine.
	// Chip-8 allows for up to 16 levels of nested subroutines
	stack [16]uint16

	// The computers which originally used the Chip-8 Language had a 16-key hexadecimal keypad with the following layout:
	//
	// +-------+
	// |1|2|3|C|
	// |4|5|6|D|
	// |7|8|9|E|
	// |A|0|B|F|
	// +-------+
	keypad [16]bool

	// The original implementation of the Chip-8 language used a 64x32-pixel monochrome display with this format:
	// +----------------+
	// |(0,0)	  (63,0)|
	// |                |
	// |                |
	// |(0,31)	 (63,31)|
	// +----------------+
	//
	// Each bit will encode the status (on/off) of the pixel, hence the uint64 per line.
	display [32]uint64

	// Chip-8 has an instruction that generate a random number.
	rand *rand.Rand
}
