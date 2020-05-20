package chip8

import (
	"errors"
	"math/rand"
	"time"
)

const (
	// FramePerSecond is the number of Frame update in Hertz
	FramePerSecond = 60
	// CyclePerSecond is the number of CPU cycle in Hertz
	CyclePerSecond = 600
	// DisplayWidth is the number of pixels in a row
	DisplayWidth = 64
	// DisplayHeight is the number of pixels in a column
	DisplayHeight = 32
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
	display [DisplayHeight]uint64

	// Chip-8 has an instruction that generate a random number.
	rand *rand.Rand
}

// New return a fully initialized instance of the CHIP-8 system.
func New() *chip8 {
	m := &chip8{
		pc: 0x200,
		memory: [4096]byte{
			// Programs may also refer to a group of sprites representing the hexadecimal digits 0 through F.
			// These sprites are 5 bytes long, or 8x5 pixels.
			// The data should be stored in the interpreter area of Chip-8 memory (0x000 to 0x1FF).
			0xF0, 0x90, 0x90, 0x90, 0xF0, //0
			0x20, 0x60, 0x20, 0x20, 0x70, //1
			0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
			0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
			0x90, 0x90, 0xF0, 0x10, 0x10, //4
			0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
			0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
			0xF0, 0x10, 0x20, 0x40, 0x40, //7
			0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
			0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
			0xF0, 0x90, 0xF0, 0x90, 0x90, //A
			0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
			0xF0, 0x80, 0x80, 0x80, 0xF0, //C
			0xE0, 0x90, 0x90, 0x90, 0xE0, //D
			0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
			0xF0, 0x80, 0xF0, 0x80, 0x80, //F

		},
		rand: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
	return m
}

// GetNextFrame is cool
func (c *chip8) GetNextFrame() ([]uint32, error) {
	for i := 0; i < CyclePerSecond/FramePerSecond; i++ {
		op, err := c.fetch()
		if err != nil {
			return nil, err
		}

		if err := c.decodeExecute(op); err != nil {
			return nil, err
		}
	}

	if c.dt != 0 {
		c.dt--
	}

	if c.st != 0 {
		c.st--
	}

	return c.mapGraphic(), nil
}

func (c *chip8) mapGraphic() []uint32 {
	fb := make([]uint32, DisplayWidth*DisplayHeight)
	for y, row := range c.display {
		for x := 0; x < DisplayWidth; x++ {
			if row&(1<<(63-x)) > 0 {
				fb[x+(y*DisplayWidth)] = 0xF2F4F3
			} else {
				fb[x+(y*DisplayWidth)] = 0x00171F
			}
		}
	}

	return fb
}

// LoadGame load game data in the memory.
// If the data cannot fit in the memory it will return an error.
func (c *chip8) LoadGame(data []byte) error {
	if len(data) >= len(c.memory)-int(c.pc) {
		return errors.New("the ROM cannot fit in memory")
	}
	copy(c.memory[c.pc:], data)
	return nil
}
