package main

/*
#include "libretro.h"

bool bridge_environment(retro_environment_t f, unsigned cmd, void *data) {
	return f(cmd, data);
}

void bridge_video_refresh(retro_video_refresh_t f, const void *data, unsigned width, unsigned height, size_t pitch) {
	f(data, width, height, pitch);
}

void bridge_audio_sample(retro_audio_sample_t f, int16_t left, int16_t right) {
	f(left, right);
}

size_t bridge_audio_sample_batch(retro_audio_sample_batch_t f, void *data, size_t frames) {
	return f(data, frames);
}

void bridge_input_poll(retro_input_poll_t f) {
	f();
}

int16_t bridge_input_state(retro_input_state_t f, unsigned port, unsigned device, unsigned index, unsigned id) {
	return f(port, device, index, id);
}

*/
import "C"
import "unsafe"

func environment(cmd uint, ptr unsafe.Pointer) bool {
	if envCb != nil {
		return bool(C.bridge_environment(envCb, C.unsigned(cmd), ptr))
	}
	return false
}

func videoRefresh(data []uint32, width, height, pitch uint) {
	if videoRefreshCb != nil {
		C.bridge_video_refresh(videoRefreshCb, unsafe.Pointer(&data[0]), C.unsigned(width), C.unsigned(height), C.size_t(pitch))
	}
}

func audioSample(left, right int16) {
	if audioSampleCb != nil {
		C.bridge_audio_sample(audioSampleCb, C.int16_t(left), C.int16_t(right))
	}
}

func audioSampleBatch(data []int16) int {
	if audioSampleBatchCb != nil {
		return int(C.bridge_audio_sample_batch(audioSampleBatchCb, unsafe.Pointer(&data[0]), C.size_t(len(data)/2)))
	}
	return 0
}

func inputPoll() {
	if inputPollCb != nil {
		C.bridge_input_poll(inputPollCb)
	}
}

func inputState(port, device, index, id uint) int16 {
	if inputStateCb != nil {
		return int16(C.bridge_input_state(inputStateCb, C.unsigned(port), C.unsigned(device), C.unsigned(index), C.unsigned(id)))
	}
	return 0
}
