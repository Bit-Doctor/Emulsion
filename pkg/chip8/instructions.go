package chip8

import (
	"errors"
	"fmt"
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

// Decode and execute the provided opCode.
// The original implementation of the Chip-8 language includes 36 different instructions.
func (c *chip8) decodeExecute(op uint16) error {
	// nnn or addr - A 12-bit value, the lowest 12 bits of the instruction
	nnn := op & 0xFFF
	// kk or byte - An 8-bit value, the lowest 8 bits of the instruction
	kk := byte(op & 0xFF)
	// n or nibble - A 4-bit value, the lowest 4 bits of the instruction
	n := byte(op & 0xF)
	// x - A 4-bit value, the lower 4 bits of the high byte of the instruction
	x := byte(op >> 8 & 0xF)
	// y - A 4-bit value, the upper 4 bits of the low byte of the instruction
	y := byte(op >> 4 & 0xF)

	switch {
	case op == 0x00E0:
		fallthrough
	case op == 0x00EE:
		fallthrough
	case op&0xF000 == 0x0000:
		fallthrough
	case op&0xF000 == 0x1000:
		fallthrough
	case op&0xF000 == 0x2000:
		fallthrough
	case op&0xF000 == 0x3000:
		fallthrough
	case op&0xF000 == 0x4000:
		fallthrough
	case op&0xF00F == 0x5000:
		fallthrough
	case op&0xF000 == 0x6000:
		fallthrough
	case op&0xF000 == 0x7000:
		fallthrough
	case op&0xF00F == 0x8000:
		fallthrough
	case op&0xF00F == 0x8001:
		fallthrough
	case op&0xF00F == 0x8002:
		fallthrough
	case op&0xF00F == 0x8003:
		fallthrough
	case op&0xF00F == 0x8004:
		fallthrough
	case op&0xF00F == 0x8005:
		fallthrough
	case op&0xF00F == 0x8006:
		fallthrough
	case op&0xF00F == 0x8007:
		fallthrough
	case op&0xF00F == 0x800E:
		fallthrough
	case op&0xF00F == 0x9000:
		fallthrough
	case op&0xF000 == 0xA000:
		fallthrough
	case op&0xF000 == 0xB000:
		fallthrough
	case op&0xF000 == 0xC000:
		fallthrough
	case op&0xF000 == 0xD000:
		fallthrough
	case op&0xF0FF == 0xE09E:
		fallthrough
	case op&0xF0FF == 0xE0A1:
		fallthrough
	case op&0xF0FF == 0xF007:
		fallthrough
	case op&0xF0FF == 0xF00A:
		fallthrough
	case op&0xF0FF == 0xF015:
		fallthrough
	case op&0xF0FF == 0xF018:
		fallthrough
	case op&0xF0FF == 0xF01E:
		fallthrough
	case op&0xF0FF == 0xF029:
		fallthrough
	case op&0xF0FF == 0xF033:
		fallthrough
	case op&0xF0FF == 0xF055:
		fallthrough
	case op&0xF0FF == 0xF065:
		fallthrough
	default:
		return fmt.Errorf("opcode not supported: 0x%04X", op)
	}
	return nil
}
