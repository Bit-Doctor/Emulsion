package chip8

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

const (
	// FramePerSecond is the number of Frame update in Hertz
	FramePerSecond = 60.0
	// CyclePerSecond is the number of CPU cycle in Hertz
	CyclePerSecond = 600.0
	//CyclePerFrame is the number of CPU cycle per frame
	CyclePerFrame = CyclePerSecond / FramePerSecond
	// SamplingRate is the number of audio sample in Hertz
	SamplingRate = 44100.0
	// SamplePerFrame is the number of audio sample per frame
	SamplePerFrame = SamplingRate / FramePerSecond
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

// GetNextFrame takes in an input state run for one frame and return the video and audio data.
func (c *chip8) GetNextFrame(inputs [16]bool) ([]uint32, []int16, error) {
	c.keypad = inputs

	if c.dt != 0 {
		c.dt--
	}

	if c.st != 0 {
		c.st--
	}

	for i := 0; i < CyclePerFrame; i++ {
		op, err := c.fetch()
		if err != nil {
			return nil, nil, err
		}

		if err := c.decodeExecute(op); err != nil {
			return nil, nil, err
		}
	}

	return c.mapGraphic(), c.mapAudio(), nil
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

func (c *chip8) mapAudio() []int16 {
	const tone = 480.0
	const deltaPhase = 2 * math.Pi * tone / SamplingRate
	sb := make([]int16, SamplePerFrame*2)

	if c.st > 0 {
		var phase float64
		for i := 0; i < len(sb); i += 2 {
			phase += deltaPhase
			sample := int16(math.Sin(phase) * math.MaxInt16)
			sb[i] = sample   // Left channel
			sb[i+1] = sample // Right channel
		}
	}

	return sb
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
