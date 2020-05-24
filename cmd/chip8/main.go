package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"
	"unsafe"

	"github.com/Bit-Doctor/emulation/pkg/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

var keyMap = map[sdl.Scancode]byte{
	sdl.SCANCODE_X: 0x0,
	sdl.SCANCODE_1: 0x1,
	sdl.SCANCODE_2: 0x2,
	sdl.SCANCODE_3: 0x3,
	sdl.SCANCODE_Q: 0x4,
	sdl.SCANCODE_W: 0x5,
	sdl.SCANCODE_E: 0x6,
	sdl.SCANCODE_A: 0x7,
	sdl.SCANCODE_S: 0x8,
	sdl.SCANCODE_D: 0x9,
	sdl.SCANCODE_Z: 0xA,
	sdl.SCANCODE_C: 0xB,
	sdl.SCANCODE_4: 0xC,
	sdl.SCANCODE_R: 0xD,
	sdl.SCANCODE_F: 0xE,
	sdl.SCANCODE_V: 0xF,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v <file>\n", os.Args[0])
		os.Exit(-1)
	}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Fprintln(os.Stderr, "cannot initialize SDL: ", err)
		os.Exit(-1)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(os.Args[0], sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 1280, 640, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create a window: ", err)
		os.Exit(-1)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create a renderer: ", err)
		os.Exit(-1)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, chip8.DisplayWidth, chip8.DisplayHeight)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create a texture: ", err)
	}
	defer texture.Destroy()

	wants := &sdl.AudioSpec{
		Freq:     chip8.SamplingRate,
		Format:   sdl.AUDIO_S16SYS,
		Channels: 2,
		Samples:  1024,
	}
	devID, err := sdl.OpenAudioDevice("", false, wants, nil, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot open audio device: ", err)
	}
	defer sdl.CloseAudioDevice(devID)
	sdl.PauseAudioDevice(devID, false)

	vm := chip8.New()
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot", err)
		os.Exit(-1)
	}

	if err := vm.LoadGame(data); err != nil {
		fmt.Fprintln(os.Stderr, "cannot load game data: ", err)
		os.Exit(-1)
	}

	running := true
	for running {
		var input [16]bool

		start := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if event.Keysym.Scancode == sdl.SCANCODE_ESCAPE {
					running = false
				} else if key, ok := keyMap[event.Keysym.Scancode]; ok {
					input[key] = event.Type == sdl.KEYDOWN
				}
			}
		}

		fb, sb, err := vm.GetNextFrame(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "system errored: ", err)
		}

		renderer.Clear()
		texture.UpdateRGBA(nil, fb, chip8.DisplayWidth)
		if err := renderer.Copy(texture, nil, nil); err != nil {
			fmt.Fprintln(os.Stderr, "cannot draw frame: ", err)
		}

		renderer.Present()

		var data []byte
		sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		sh.Len = len(sb) * 2
		sh.Cap = len(sb) * 2
		sh.Data = uintptr(unsafe.Pointer(&sb[0]))
		if err := sdl.QueueAudio(devID, data); err != nil {
			fmt.Fprintln(os.Stderr, "cannot queue audio: ", err)
		}

		time.Sleep(time.Second/chip8.FramePerSecond - time.Since(start))
	}
}
