package chip8

import (
	"errors"
)

// Fetch return the next opCode and increment the program counter.
// All instructions are 2 bytes long and are stored most-significant-byte first.
// In memory, the first byte of each instruction should be located at an even addresses.
// If a program includes sprite data, it should be padded so any instructions following it will be properly situated in RAM.
func (c *chip8) fetch() (uint16, error) {
	if int(c.pc) == len(c.memory) {
		return 0x0000, errors.New("segmentation fault")
	}

	op := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	c.pc += 2

	return op, nil
}
