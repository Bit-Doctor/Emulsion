# Emulsion: a go emulator
**Emulsion** is pet project meant to learn emulation and go during my free time.

As such is not meant to be competing with standard emulator for playing games.

The following systems are currently emulated: [CHIP-8](#chip-8)

## CHIP-8
![CHIP-8](https://raw.github.com/Bit-Doctor/Emulsion/master/media/chip8.gif)

[CHIP-8](https://en.wikipedia.org/wiki/CHIP-8) is an interpreted programming language, developed by Joseph Weisbecker. It was initially used on the COSMAC VIP and Telmac 1800 8-bit microcomputers in the mid-1970s.

### Usage

You'll need the [SDL](https://www.libsdl.org/) library. Then it is as simple as running:

```
$ go run ./cmd/chip8 <rom>
```

Alternatively a minimal [Libretro](https://www.libretro.com/) core is also available:

```
$ go build -buildmode=c-shared ./cmd/chip8_libretro`
$ retroarch -L chip8_libretro <rom>
```

### Inputs

On the standalone emulator the CHIP-8 keyboard is mapped following this diagram:

```
      CHIP-8                 Mapping
+---------------+       +---------------+
| 1 | 2 | 3 | C |       | 1 | 2 | 3 | 4 |
| 4 | 5 | 6 | D |       | Q | W | E | R |
| 7 | 8 | 9 | E |       | A | S | D | F |
| A | 0 | B | F |       | Z | X | C | V |
+---------------+       +---------------+
```
