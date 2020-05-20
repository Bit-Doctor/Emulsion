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
		c.cls() // 00E0
	case op == 0x00EE:
		return c.ret() // 00EE
	case op&0xF000 == 0x0000:
		c.sys() // 0nnn
	case op&0xF000 == 0x1000:
		c.jump(nnn) // 1nnn
	case op&0xF000 == 0x2000:
		c.call(nnn) // 2nnn
	case op&0xF000 == 0x3000:
		c.skipIfVx(x, kk) // 3xkk
	case op&0xF000 == 0x4000:
		c.skipIfNotVx(x, kk) // 4xkk
	case op&0xF00F == 0x5000:
		c.skipIfVxVy(x, y) // 5xy0
	case op&0xF000 == 0x6000:
		c.setVx(x, kk) // 6xkk
	case op&0xF000 == 0x7000:
		c.addVx(x, kk) // 7xkk
	case op&0xF00F == 0x8000:
		c.setVxVy(x, y) // 8xy0
	case op&0xF00F == 0x8001:
		c.setVxOrVy(x, y) // 8xy1
	case op&0xF00F == 0x8002:
		c.setVxAndVy(x, y) //8xy2
	case op&0xF00F == 0x8003:
		c.setVxXorVy(x, y) //8xy3
	case op&0xF00F == 0x8004:
		c.addVxVy(x, y) //8xy4
	case op&0xF00F == 0x8005:
		c.subVxVy(x, y) //8xy5
	case op&0xF00F == 0x8006:
		c.shrVx(x) // 8xy6
	case op&0xF00F == 0x8007:
		c.subYX(x, y) // 8xy7
	case op&0xF00F == 0x800E:
		c.shlVx(x) // 8xyE
	case op&0xF00F == 0x9000:
		c.skipIfNotVcVy(x, y) // 9xy0
	case op&0xF000 == 0xA000:
		c.setI(nnn) // Annn
	case op&0xF000 == 0xB000:
		c.jumpV0(nnn) // Bnnn
	case op&0xF000 == 0xC000:
		c.rndVx(x, kk) // Cxkk
	case op&0xF000 == 0xD000:
		c.drawSprite(x, y, n) // Dxyn
	case op&0xF0FF == 0xE09E:
		c.skipIfPressed(x) // Ex9E
	case op&0xF0FF == 0xE0A1:
		c.skipIfNotPressed(x) // ExA1
	case op&0xF0FF == 0xF007:
		c.setVxDT(x) // Fx07
	case op&0xF0FF == 0xF00A:
		c.setVxKey(x) // Fx0A
	case op&0xF0FF == 0xF015:
		c.setDTVx(x) // Fx15
	case op&0xF0FF == 0xF018:
		c.setSTVx(x) // Fx18
	case op&0xF0FF == 0xF01E:
		c.addIVx(x) // Fx1E
	case op&0xF0FF == 0xF029:
		c.setIDigit(x) // Fx29
	case op&0xF0FF == 0xF033:
		c.bcd(x) // Fx33
	case op&0xF0FF == 0xF055:
		c.writeRegs(x) // Fx55
	case op&0xF0FF == 0xF065:
		c.readRegs(x) // Fx55
	default:
		return fmt.Errorf("opcode not supported: 0x%04X", op)
	}
	return nil
}

// Clear the display.
func (c *chip8) cls() {
	for i := range c.display {
		c.display[i] = 0
	}
}

// Return from a subroutine.
func (c *chip8) ret() error {
	if c.sp == 0 {
		return errors.New("stack underflow")
	}

	c.sp--
	c.pc = c.stack[c.sp]
	return nil
}

// Jump to a machine code routine at nnn.
// This instruction is only used on the old computers on which Chip-8 was originally implemented. It is ignored by modern interpreters.
func (c *chip8) sys() {}

// Jump to address.
func (c *chip8) jump(addr uint16) {
	c.pc = addr
}

// Call a subroutine at address.
func (c *chip8) call(addr uint16) error {
	if int(c.sp) == len(c.stack) {
		return errors.New("stack overflow")
	}

	c.stack[c.sp] = c.pc
	c.sp++
	c.pc = addr
	return nil
}

// Skip next instruction if Vx = kk.
func (c *chip8) skipIfVx(x, kk byte) {
	if c.v[x] == kk {
		c.pc += 2
	}
}

// Skip next instruction if Vx != kk.
func (c *chip8) skipIfNotVx(x, kk byte) {
	if c.v[x] != kk {
		c.pc += 2
	}
}

// Skip next instruction if Vx = Vy.
func (c *chip8) skipIfVxVy(x, y byte) {
	if c.v[x] == c.v[y] {
		c.pc += 2
	}
}

// Set Vx = kk.
func (c *chip8) setVx(x, kk byte) {
	c.v[x] = kk
}

// Set Vx = Vx + kk.
func (c *chip8) addVx(x, kk byte) {
	c.v[x] += kk
}

// Set Vx = Vy.
func (c *chip8) setVxVy(x, y byte) {
	c.v[x] = c.v[y]
}

// Set Vx = Vx OR Vy.
func (c *chip8) setVxOrVy(x, y byte) {
	c.v[x] |= c.v[y]
}

// Set Vx = Vx AND Vy.
func (c *chip8) setVxAndVy(x, y byte) {
	c.v[x] &= c.v[y]
}

// Set Vx = Vx XOR Vy.
func (c *chip8) setVxXorVy(x, y byte) {
	c.v[x] ^= c.v[y]
}

// Set Vx = Vx + Vy, set VF = carry.
func (c *chip8) addVxVy(x, y byte) {
	c.v[x] += c.v[y]
	c.v[0xF] = 0
	if c.v[x] < c.v[y] {
		c.v[0xF] = 1
	}
}

// Set Vx = Vx - Vy, set VF = NOT borrow.
func (c *chip8) subVxVy(x, y byte) {
	c.v[0xF] = 1
	if c.v[x] < c.v[y] {
		c.v[0xF] = 0
	}
	c.v[x] -= c.v[y]
}

// Set Vx = Vx SHR 1, set VF = LSB of Vx.
func (c *chip8) shrVx(x byte) {
	c.v[0xF] = c.v[x] & 1
	c.v[x] >>= 1
}

// Set Vx = Vy - Vx, set VF = NOT borrow.
func (c *chip8) subYX(x, y byte) {
	c.v[0xF] = 1
	if c.v[y] < c.v[x] {
		c.v[0xF] = 0
	}
	c.v[x] = c.v[y] - c.v[x]
}

// Set Vx = Vx SHL 1, set VF = MSB of Vx.
func (c *chip8) shlVx(x byte) {
	c.v[0xF] = c.v[x] >> 7
	c.v[x] <<= 1
}

// Skip next instruction if Vx != Vy.
func (c *chip8) skipIfNotVcVy(x, y byte) {
	if c.v[x] != c.v[y] {
		c.pc += 2
	}
}

// Set I = addr.
func (c *chip8) setI(addr uint16) {
	c.i = addr
}

// Jump to addr + V0.
func (c *chip8) jumpV0(addr uint16) {
	c.pc = addr + uint16(c.v[0])
}

// Set Vx = random byte AND kk.
func (c *chip8) rndVx(x, kk byte) {
	c.v[x] = byte(c.rand.Intn(256)) & kk
}

// Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
// If the sprite is positioned so part of it is outside the coordinates of the display, it wraps around to the opposite side of the screen
func (c *chip8) drawSprite(x, y, n byte) {
	sprite := c.memory[c.i : c.i+uint16(n)]
	posX := c.v[x] % 64 // wraps around to the opposite side of the screen
	posY := c.v[y] % 32 // wraps around to the opposite side of the screen
	collision := false

	for i := range sprite {
		row := uint64(sprite[i]) << (64 - 8) // Move sprite to the MSB
		row = row>>posX | row<<(64-posX)     // Move sprite to correct position and wrap around

		or := c.display[posY] | row
		c.display[posY] ^= row

		if c.display[posY] != or {
			collision = true
		}

		posY = (posY + 1) % 32
	}

	c.v[0xF] = 0
	if collision {
		c.v[0xF] = 1
	}
}

// Skip next instruction if key with the value of Vx is pressed.
func (c *chip8) skipIfPressed(x byte) {
	if c.keypad[c.v[x]] {
		c.pc += 2
	}
}

// Skip next instruction if key with the value of Vx is not pressed.
func (c *chip8) skipIfNotPressed(x byte) {
	if !c.keypad[c.v[x]] {
		c.pc += 2
	}
}

// Set Vx = delay timer value.
func (c *chip8) setVxDT(x byte) {
	c.v[x] = c.dt
}

// Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
func (c *chip8) setVxKey(x byte) {
	for i, k := range c.keypad {
		if k {
			c.v[x] = byte(i)
			return
		}
	}

	c.pc -= 2 // Simulate halting if no key pressed
}

// Set delay timer = Vx.
func (c *chip8) setDTVx(x byte) {
	c.dt = c.v[x]
}

// Set sound timer = Vx.
func (c *chip8) setSTVx(x byte) {
	c.st = c.v[x]
}

// Set I = I + Vx.
func (c *chip8) addIVx(x byte) {
	c.i += uint16(c.v[x])
}

// Set I = location of sprite for digit Vx.
func (c *chip8) setIDigit(x byte) {
	c.i = uint16(c.v[x]) * 5
}

// Store BCD representation of Vx in memory locations I, I+1, and I+2.
func (c *chip8) bcd(x byte) {
	c.memory[c.i] = c.v[x] / 100
	c.memory[c.i+1] = (c.v[x] / 10) % 10
	c.memory[c.i+2] = c.v[x] % 10
}

// Store registers V0 through Vx in memory starting at location I.
func (c *chip8) writeRegs(x byte) {
	copy(c.memory[c.i:], c.v[:x+1])
}

// Read registers V0 through Vx from memory starting at location I.
func (c *chip8) readRegs(x byte) {
	copy(c.v[:x+1], c.memory[c.i:])
}
