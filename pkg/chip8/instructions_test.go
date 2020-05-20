package chip8

import (
	"math/rand"
	"testing"
)

func Test_chip8_fetch(t *testing.T) {
	type fields struct {
		memory [4096]byte
		pc     uint16
	}

	type wants struct {
		ret uint16
		pc  uint16
	}

	tests := []struct {
		name    string
		fields  fields
		wants   wants
		wantErr bool
	}{
		{name: "fetch first memory bytes", fields: fields{memory: [4096]byte{0xAB, 0xCD}}, wants: wants{ret: 0xABCD, pc: 2}},
		{name: "fetch last memory bytes", fields: fields{memory: [4096]byte{4094: 0xAB, 4095: 0xCD}, pc: 4094}, wants: wants{ret: 0xABCD, pc: 4096}},
		{name: "fetch out of memory", fields: fields{pc: 4096}, wants: wants{pc: 4096}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &chip8{
				memory: tt.fields.memory,
				pc:     tt.fields.pc,
			}
			got, err := c.fetch()
			if (err != nil) != tt.wantErr {
				t.Errorf("chip8.fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wants.ret {
				t.Errorf("chip8.fetch() = 0x%04X, want 0x%04X", got, tt.wants.ret)
			}
			if c.pc != tt.wants.pc {
				t.Errorf("c.pc = %v, want %v", c.pc, tt.wants.pc)
			}
		})
	}
}

func Test_chip8_decodeExecute(t *testing.T) {
	type fields struct {
		memory  [4096]byte
		v       [16]byte
		i       uint16
		dt      byte
		st      byte
		pc      uint16
		sp      byte
		stack   [16]uint16
		keypad  [16]bool
		display [32]uint64
		rand    *rand.Rand
	}
	tests := []struct {
		name    string
		args    uint16
		fields  fields
		wants   fields
		wantErr bool
	}{
		{name: "00E0", args: 0x00E0, fields: fields{display: [32]uint64{0: 1, 16: 1, 31: 1}}},
		{name: "00EE", args: 0x00EE, fields: fields{sp: 1, stack: [16]uint16{0xFF}}, wants: fields{pc: 0xFF}},
		{name: "00EE stack underflow", args: 0x00EE, wantErr: true},
		{name: "00EE stack underflow", args: 0x0000},
		{name: "1nnn", args: 0x1123, wants: fields{pc: 0x123}},
		{name: "2nnn", args: 0x2123, fields: fields{pc: 0xFF}, wants: fields{pc: 0x123, sp: 1, stack: [16]uint16{0xFF}}},
		{name: "2nnn stack overflow", args: 0x2000, fields: fields{sp: 16}, wants: fields{sp: 16}, wantErr: true},
		{name: "3xkk skip", args: 0x32FF, fields: fields{v: [16]byte{2: 0xFF}}, wants: fields{v: [16]byte{2: 0xFF}, pc: 2}},
		{name: "3xkk no skip", args: 0x32FF},
		{name: "4xkk skip", args: 0x42FF, wants: fields{pc: 2}},
		{name: "4xkk no skip", args: 0x42FF, fields: fields{v: [16]byte{2: 0xFF}}, wants: fields{v: [16]byte{2: 0xFF}}},
		{name: "5xy0 skip", args: 0x5120, fields: fields{v: [16]byte{1: 0xFF, 2: 0xFF}}, wants: fields{v: [16]byte{1: 0xFF, 2: 0xFF}, pc: 2}},
		{name: "5xy0 no skip", args: 0x5120, fields: fields{v: [16]byte{1: 0xFF}}, wants: fields{v: [16]byte{1: 0xFF}}},
		{name: "6xkk", args: 0x6212, fields: fields{v: [16]byte{2: 0xF0}}, wants: fields{v: [16]byte{2: 0x12}}},
		{name: "7xkk", args: 0x7203, fields: fields{v: [16]byte{2: 10}}, wants: fields{v: [16]byte{2: 13}}},
		{name: "8xy0", args: 0x8120, fields: fields{v: [16]byte{2: 0xFF}}, wants: fields{v: [16]byte{1: 0xFF, 2: 0xFF}}},
		{name: "8xy1", args: 0x8121, fields: fields{v: [16]byte{1: 0xF0, 2: 0x0F}}, wants: fields{v: [16]byte{1: 0xFF, 2: 0x0F}}},
		{name: "8xy2", args: 0x8122, fields: fields{v: [16]byte{1: 0x0F, 2: 0xFF}}, wants: fields{v: [16]byte{1: 0x0F, 2: 0xFF}}},
		{name: "8xy3", args: 0x8123, fields: fields{v: [16]byte{1: 0x0F, 2: 0xFF}}, wants: fields{v: [16]byte{1: 0xF0, 2: 0xFF}}},
		{name: "8xy4 no carry", args: 0x8124, fields: fields{v: [16]byte{1: 1, 2: 3, 0xF: 1}}, wants: fields{v: [16]byte{1: 4, 2: 3}}},
		{name: "8xy4 carry", args: 0x8124, fields: fields{v: [16]byte{1: 3, 2: 0xFF, 0xF: 0}}, wants: fields{v: [16]byte{1: 2, 2: 0xFF, 0xF: 1}}},
		{name: "8xy5 no borrow", args: 0x8125, fields: fields{v: [16]byte{1: 3, 2: 1, 0xF: 0}}, wants: fields{v: [16]byte{1: 2, 2: 1, 0xF: 1}}},
		{name: "8xy5 borrow", args: 0x8125, fields: fields{v: [16]byte{1: 3, 2: 0xFF, 0xF: 1}}, wants: fields{v: [16]byte{1: 4, 2: 0xFF, 0xF: 0}}},
		{name: "8xy6 LSB", args: 0x8106, fields: fields{v: [16]byte{1: 0xFF, 0xF: 0}}, wants: fields{v: [16]byte{1: 0x7F, 0xF: 1}}},
		{name: "8xy6 no LSB", args: 0x8106, fields: fields{v: [16]byte{1: 0xF0, 0xF: 1}}, wants: fields{v: [16]byte{1: 0x78, 0xF: 0}}},
		{name: "8xy7 no borrow", args: 0x8127, fields: fields{v: [16]byte{1: 1, 2: 3, 0xF: 0}}, wants: fields{v: [16]byte{1: 2, 2: 3, 0xF: 1}}},
		{name: "8xy7 borrow", args: 0x8127, fields: fields{v: [16]byte{1: 0xFF, 2: 3, 0xF: 1}}, wants: fields{v: [16]byte{1: 4, 2: 3, 0xF: 0}}},
		{name: "8xyE MSB", args: 0x810E, fields: fields{v: [16]byte{1: 0xFF, 0xF: 0}}, wants: fields{v: [16]byte{1: 0xFE, 0xF: 1}}},
		{name: "8xyE no MSB", args: 0x810E, fields: fields{v: [16]byte{1: 0x0F, 0xF: 1}}, wants: fields{v: [16]byte{1: 0x1E, 0xF: 0}}},
		{name: "9xy0 skip", args: 0x9120, fields: fields{v: [16]byte{1: 0xFF}}, wants: fields{v: [16]byte{1: 0xFF}, pc: 2}},
		{name: "9xy0 no skip", args: 0x9120, fields: fields{v: [16]byte{1: 0xFF, 2: 0xFF}}, wants: fields{v: [16]byte{1: 0xFF, 2: 0xFF}}},
		{name: "Annn", args: 0xA123, wants: fields{i: 0x123}},
		{name: "Bnnn", args: 0xB123, fields: fields{v: [16]byte{1}}, wants: fields{v: [16]byte{1}, pc: 0x124}},
		{name: "Cxkk", args: 0xC1FF, fields: fields{rand: rand.New(rand.NewSource(0))}, wants: fields{v: [16]byte{1: 250}}},
		{name: "Dxyn draw on top right", args: 0xD011, fields: fields{v: [16]byte{0, 0}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{0xF0 << 56}, v: [16]byte{0, 0}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn draw on top left", args: 0xD011, fields: fields{v: [16]byte{56, 0}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{0xF0}, v: [16]byte{56, 0}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn draw on bottom right", args: 0xD011, fields: fields{v: [16]byte{0, 31}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{31: 0xF0 << 56}, v: [16]byte{0, 31}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn draw on bottom left", args: 0xD011, fields: fields{v: [16]byte{56, 31}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{31: 0xF0}, v: [16]byte{56, 31}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn warp horizontaly", args: 0xD011, fields: fields{v: [16]byte{64, 0}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{0xF0 << 56}, v: [16]byte{64, 0}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn warp verticaly", args: 0xD011, fields: fields{v: [16]byte{0, 32}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{0xF0 << 56}, v: [16]byte{0, 32}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn partial warp horizontaly", args: 0xD011, fields: fields{v: [16]byte{63, 0}, memory: [4096]byte{0xF0}}, wants: fields{display: [32]uint64{0xE0<<56 | 1}, v: [16]byte{63, 0}, memory: [4096]byte{0xF0}}},
		{name: "Dxyn partial warp verticaly", args: 0xD012, fields: fields{v: [16]byte{0, 31}, memory: [4096]byte{0xF0, 0x90}}, wants: fields{display: [32]uint64{0: 0x90 << 56, 31: 0xF0 << 56}, v: [16]byte{0, 31}, memory: [4096]byte{0xF0, 0x90}}},
		{name: "Dxyn collision", args: 0xD011, fields: fields{v: [16]byte{0, 0}, memory: [4096]byte{0xF0}, display: [32]uint64{^uint64(0)}}, wants: fields{display: [32]uint64{^uint64(0xF0 << 56)}, v: [16]byte{0, 0, 0xF: 1}, memory: [4096]byte{0xF0}}},
		{name: "Ex9E skip", args: 0xE09E, fields: fields{v: [16]byte{1}, keypad: [16]bool{1: true}}, wants: fields{pc: 2, v: [16]byte{1}}},
		{name: "Ex9E no skip", args: 0xE09E, fields: fields{v: [16]byte{1}}, wants: fields{v: [16]byte{1}}},
		{name: "ExA1 skip", args: 0xE0A1, fields: fields{v: [16]byte{1}}, wants: fields{pc: 2, v: [16]byte{1}}},
		{name: "ExA1 no skip", args: 0xE0A1, fields: fields{v: [16]byte{1}, keypad: [16]bool{1: true}}, wants: fields{v: [16]byte{1}}},
		{name: "Fx07", args: 0xF107, fields: fields{dt: 0xFF}, wants: fields{v: [16]byte{1: 0xFF}, dt: 0xFF}},
		{name: "Fx0A key pressed", args: 0xF10A, fields: fields{keypad: [16]bool{3: true}}, wants: fields{v: [16]byte{1: 3}}},
		{name: "Fx0A no key pressed", args: 0xF10A, fields: fields{pc: 2}, wants: fields{pc: 0}},
		{name: "Fx15", args: 0xF115, fields: fields{v: [16]byte{1: 0xFF}}, wants: fields{dt: 0xFF, v: [16]byte{1: 0xFF}}},
		{name: "Fx18", args: 0xF118, fields: fields{v: [16]byte{1: 0xFF}}, wants: fields{st: 0xFF, v: [16]byte{1: 0xFF}}},
		{name: "Fx1E", args: 0xF11E, fields: fields{i: 1, v: [16]byte{1: 0xFE}}, wants: fields{i: 0xFF, v: [16]byte{1: 0xFE}}},
		{name: "Fx29", args: 0xF129, fields: fields{v: [16]byte{1: 1}}, wants: fields{i: 5, v: [16]byte{1: 1}}},
		{name: "Fx33", args: 0xF133, fields: fields{v: [16]byte{1: 123}}, wants: fields{memory: [4096]byte{1, 2, 3}, v: [16]byte{1: 123}}},
		{name: "Fx55", args: 0xF255, fields: fields{i: 1, v: [16]byte{1, 2, 3, 4}}, wants: fields{memory: [4096]byte{0, 1, 2, 3}, i: 1, v: [16]byte{1, 2, 3, 4}}},
		{name: "Fx65", args: 0xF265, fields: fields{memory: [4096]byte{0, 1, 2, 3, 4}, i: 1}, wants: fields{memory: [4096]byte{0, 1, 2, 3, 4}, i: 1, v: [16]byte{1, 2, 3}}},
		{name: "Unkown opcode", args: 0xF088, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &chip8{
				memory:  tt.fields.memory,
				v:       tt.fields.v,
				i:       tt.fields.i,
				dt:      tt.fields.dt,
				st:      tt.fields.st,
				pc:      tt.fields.pc,
				sp:      tt.fields.sp,
				stack:   tt.fields.stack,
				keypad:  tt.fields.keypad,
				display: tt.fields.display,
				rand:    tt.fields.rand,
			}
			if err := c.decodeExecute(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("chip8.decodeExecute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if c.v != tt.wants.v {
				t.Errorf("c.v = %v, want %v", c.v, tt.wants.v)
			}
			if c.i != tt.wants.i {
				t.Errorf("c.i = %v, want %v", c.i, tt.wants.i)
			}
			if c.dt != tt.wants.dt {
				t.Errorf("c.dt = %v, want %v", c.dt, tt.wants.dt)
			}
			if c.st != tt.wants.st {
				t.Errorf("c.st = %v, want %v", c.st, tt.wants.st)
			}
			if c.pc != tt.wants.pc {
				t.Errorf("c.pc = %v, want %v", c.pc, tt.wants.pc)
			}
			if c.sp != tt.wants.sp {
				t.Errorf("c.sp = %v, want %v", c.sp, tt.wants.sp)
			}
			if c.memory != tt.wants.memory {
				t.Errorf("c.memory = %v, want %v", c.memory, tt.wants.memory)
			}
			if c.display != tt.wants.display {
				t.Errorf("c.display = %v, want %v", c.display, tt.wants.display)
			}
		})
	}
}
