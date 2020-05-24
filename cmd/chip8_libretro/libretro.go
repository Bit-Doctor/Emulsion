package main

/*
#include "libretro.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

typedef const struct retro_game_info const_retro_game_info;
typedef const void const_void;
typedef const char const_char;
*/
import "C"

import (
	"unsafe"

	"github.com/Bit-Doctor/emulation/pkg/chip8"
)

var (
	envCb              C.retro_environment_t
	videoRefreshCb     C.retro_video_refresh_t
	audioSampleCb      C.retro_audio_sample_t
	audioSampleBatchCb C.retro_audio_sample_batch_t
	inputPollCb        C.retro_input_poll_t
	inputStateCb       C.retro_input_state_t
)

var (
	vm     *chip8.Chip8
	toFree []unsafe.Pointer
)

//export retro_set_environment
func retro_set_environment(cb C.retro_environment_t) {
	envCb = cb

	descriptors := []C.struct_retro_input_descriptor{
		{description: C.CString("0"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_SELECT},
		{description: C.CString("1"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_Y},
		{description: C.CString("2"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_UP},
		{description: C.CString("3"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_X},
		{description: C.CString("4"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_LEFT},
		{description: C.CString("5"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_START},
		{description: C.CString("6"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_RIGHT},
		{description: C.CString("7"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_B},
		{description: C.CString("8"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_DOWN},
		{description: C.CString("9"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_A},
		{description: C.CString("A"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_R},
		{description: C.CString("B"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_L},
		{description: C.CString("C"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_R2},
		{description: C.CString("D"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_L2},
		{description: C.CString("E"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_R3},
		{description: C.CString("F"), device: C.RETRO_DEVICE_JOYPAD, id: C.RETRO_DEVICE_ID_JOYPAD_L3},
		{},
	}

	for _, d := range descriptors {
		toFree = append(toFree, unsafe.Pointer(d.description))
	}

	environment(C.RETRO_ENVIRONMENT_SET_INPUT_DESCRIPTORS, unsafe.Pointer(&descriptors[0]))
}

//export retro_set_video_refresh
func retro_set_video_refresh(cb C.retro_video_refresh_t) { videoRefreshCb = cb }

//export retro_set_audio_sample
func retro_set_audio_sample(cb C.retro_audio_sample_t) { audioSampleCb = cb }

//export retro_set_audio_sample_batch
func retro_set_audio_sample_batch(cb C.retro_audio_sample_batch_t) { audioSampleBatchCb = cb }

//export retro_set_input_poll
func retro_set_input_poll(cb C.retro_input_poll_t) { inputPollCb = cb }

//export retro_set_input_state
func retro_set_input_state(cb C.retro_input_state_t) { inputStateCb = cb }

//export retro_init
func retro_init() {
	vm = chip8.New()
}

//export retro_deinit
func retro_deinit() {
	vm = nil

	for _, d := range toFree {
		C.free(d)
	}
}

//export retro_api_version
func retro_api_version() C.unsigned { return C.RETRO_API_VERSION }

//export retro_get_system_info
func retro_get_system_info(info *C.struct_retro_system_info) {
	info.library_name = C.CString("CHIP-8")
	info.library_version = C.CString("v0.1")
	info.need_fullpath = false
	info.valid_extensions = C.CString("ch8")

	toFree = append(toFree, unsafe.Pointer(info.library_name))
	toFree = append(toFree, unsafe.Pointer(info.library_version))
	toFree = append(toFree, unsafe.Pointer(info.valid_extensions))

	format := C.RETRO_PIXEL_FORMAT_XRGB8888
	environment(C.RETRO_ENVIRONMENT_SET_PIXEL_FORMAT, unsafe.Pointer(&format))
}

//export retro_get_system_av_info
func retro_get_system_av_info(info *C.struct_retro_system_av_info) {
	*info = C.struct_retro_system_av_info{
		timing: C.struct_retro_system_timing{
			fps:         chip8.FramePerSecond,
			sample_rate: chip8.SamplingRate,
		},
		geometry: C.struct_retro_game_geometry{
			base_width:  chip8.DisplayWidth,
			base_height: chip8.DisplayHeight,
			max_width:   chip8.DisplayWidth,
			max_height:  chip8.DisplayHeight,
		},
	}
}

//export retro_set_controller_port_device
func retro_set_controller_port_device(port, device C.unsigned) {}

//export retro_reset
func retro_reset() {}

//export retro_run
func retro_run() {
	inputPoll()
	inputs := [16]bool{
		0x0: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_SELECT) == 1,
		0x1: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_Y) == 1,
		0x2: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_UP) == 1,
		0x3: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_X) == 1,
		0x4: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_LEFT) == 1,
		0x5: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_START) == 1,
		0x6: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_RIGHT) == 1,
		0x7: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_B) == 1,
		0x8: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_DOWN) == 1,
		0x9: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_A) == 1,
		0xA: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_R) == 1,
		0xB: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_L) == 1,
		0xC: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_R2) == 1,
		0xD: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_L2) == 1,
		0xE: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_R3) == 1,
		0xF: inputState(0, C.RETRO_DEVICE_JOYPAD, 0, C.RETRO_DEVICE_ID_JOYPAD_L3) == 1,
	}

	fb, sb, _ := vm.GetNextFrame(inputs)

	videoRefresh(fb, chip8.DisplayWidth, chip8.DisplayHeight, chip8.DisplayWidth*4)
	audioSampleBatch(sb)
}

//export retro_serialize_size
func retro_serialize_size() C.size_t { return 0 }

//export retro_serialize
func retro_serialize(data unsafe.Pointer, size C.size_t) C.bool { return false }

//export retro_unserialize
func retro_unserialize(data *C.const_void, size C.size_t) C.bool { return false }

//export retro_cheat_reset
func retro_cheat_reset() {}

//export retro_cheat_set
func retro_cheat_set(index C.unsigned, enabled C.bool, code *C.const_char) {}

//export retro_load_game
func retro_load_game(info *C.const_retro_game_info) C.bool {
	if info == nil {
		return false
	}

	b := C.GoBytes(info.data, C.int(info.size))
	if err := vm.LoadGame(b); err != nil {
		return false
	}

	return true
}

//export retro_load_game_special
func retro_load_game_special(gameType C.unsigned, info *C.const_retro_game_info, numInfo C.size_t) C.bool {
	return false
}

//export retro_unload_game
func retro_unload_game() {}

//export retro_get_region
func retro_get_region() C.unsigned { return C.RETRO_REGION_PAL }

//export retro_get_memory_data
func retro_get_memory_data(id C.unsigned) unsafe.Pointer { return nil }

//export retro_get_memory_size
func retro_get_memory_size(id C.unsigned) C.size_t { return 0 }

func main() {}
