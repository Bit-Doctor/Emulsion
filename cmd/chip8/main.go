package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Bit-Doctor/emulation/pkg/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

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
	renderer.SetLogicalSize(chip8.DisplayWidth, chip8.DisplayHeight)

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, chip8.DisplayWidth, chip8.DisplayHeight)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create a texture: ", err)
	}
	defer texture.Destroy()

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
		start := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				switch event.Keysym.Scancode {
				case sdl.SCANCODE_ESCAPE:
					running = false
				}
			}
		}

		renderer.Clear()
		fb, err := vm.GetNextFrame()
		if err != nil {
			fmt.Fprintln(os.Stderr, "system errored: ", err)
		}
		texture.UpdateRGBA(nil, fb, chip8.DisplayWidth)
		if err := renderer.Copy(texture, nil, nil); err != nil {
			fmt.Fprintln(os.Stderr, "cannot draw frame: ", err)
		}
		renderer.Present()

		time.Sleep(time.Second/60 - time.Since(start))
	}
}
