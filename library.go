// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo linux LDFLAGS: -ldl
#include "opus.h"

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

typedef OpusDecoder* (*decoder_create_fn)(opus_int32, int, int*);
typedef void (*decoder_destroy_fn)(OpusDecoder*);
typedef int (*decode_fn)(OpusDecoder*, const unsigned char*, opus_int32, opus_int16*, int, int);
typedef int (*decode_float_fn)(OpusDecoder*, const unsigned char*, opus_int32, float*, int, int);
typedef OpusEncoder* (*encoder_create_fn)(opus_int32, int, int, int*);
typedef void (*encoder_destroy_fn)(OpusEncoder*);
typedef opus_int32 (*encode_float_fn)(OpusEncoder*, const float*, int, unsigned char*, opus_int32);
typedef int (*encoder_ctl_fn)(OpusEncoder*, int, ...);
typedef int (*packet_get_nb_samples_fn)(const unsigned char*, opus_int32, opus_int32);
typedef const char* (*strerror_fn)(int);

static decoder_create_fn opus_decoder_create_ptr;
static decoder_destroy_fn opus_decoder_destroy_ptr;
static decode_fn opus_decode_ptr;
static decode_float_fn opus_decode_float_ptr;
static encoder_create_fn opus_encoder_create_ptr;
static encoder_destroy_fn opus_encoder_destroy_ptr;
static encode_float_fn opus_encode_float_ptr;
static encoder_ctl_fn opus_encoder_ctl_ptr;
static packet_get_nb_samples_fn opus_packet_get_nb_samples_ptr;
static strerror_fn opus_strerror_ptr;

int engine_opus_load(void) {
#ifdef _WIN32
	HMODULE module = GetModuleHandleA("engine2.dll");
	if (!module) return 0;
#define ENGINE_OPUS_SYMBOL(name) (void*)GetProcAddress(module, name)
#else
	void* module = dlopen("libengine2.so", RTLD_NOW | RTLD_NOLOAD);
	if (!module) return 0;
#define ENGINE_OPUS_SYMBOL(name) dlsym(module, name)
#endif
	opus_decoder_create_ptr = (decoder_create_fn)ENGINE_OPUS_SYMBOL("opus_decoder_create");
	opus_decoder_destroy_ptr = (decoder_destroy_fn)ENGINE_OPUS_SYMBOL("opus_decoder_destroy");
	opus_decode_ptr = (decode_fn)ENGINE_OPUS_SYMBOL("opus_decode");
	opus_decode_float_ptr = (decode_float_fn)ENGINE_OPUS_SYMBOL("opus_decode_float");
	opus_encoder_create_ptr = (encoder_create_fn)ENGINE_OPUS_SYMBOL("opus_encoder_create");
	opus_encoder_destroy_ptr = (encoder_destroy_fn)ENGINE_OPUS_SYMBOL("opus_encoder_destroy");
	opus_encode_float_ptr = (encode_float_fn)ENGINE_OPUS_SYMBOL("opus_encode_float");
	opus_encoder_ctl_ptr = (encoder_ctl_fn)ENGINE_OPUS_SYMBOL("opus_encoder_ctl");
	opus_packet_get_nb_samples_ptr = (packet_get_nb_samples_fn)ENGINE_OPUS_SYMBOL("opus_packet_get_nb_samples");
	opus_strerror_ptr = (strerror_fn)ENGINE_OPUS_SYMBOL("opus_strerror");
#undef ENGINE_OPUS_SYMBOL
	return opus_decoder_create_ptr && opus_decoder_destroy_ptr && opus_decode_ptr && opus_decode_float_ptr
		&& opus_encoder_create_ptr && opus_encoder_destroy_ptr && opus_encode_float_ptr && opus_encoder_ctl_ptr
		&& opus_packet_get_nb_samples_ptr && opus_strerror_ptr;
}

OpusDecoder* engine_decoder_create(opus_int32 rate, int channels, int* error) {
	return opus_decoder_create_ptr(rate, channels, error);
}

void engine_decoder_destroy(OpusDecoder* decoder) {
	opus_decoder_destroy_ptr(decoder);
}

int engine_decode(OpusDecoder* decoder, const unsigned char* data, opus_int32 len, opus_int16* pcm, int frameSize) {
	return opus_decode_ptr(decoder, data, len, pcm, frameSize, 0);
}

int engine_decode_float(OpusDecoder* decoder, const unsigned char* data, opus_int32 len, float* pcm, int frameSize) {
	return opus_decode_float_ptr(decoder, data, len, pcm, frameSize, 0);
}

OpusEncoder* engine_encoder_create(opus_int32 rate, int channels, int application, int* error) {
	return opus_encoder_create_ptr(rate, channels, application, error);
}

void engine_encoder_destroy(OpusEncoder* encoder) {
	opus_encoder_destroy_ptr(encoder);
}

opus_int32 engine_encode_float(OpusEncoder* encoder, const float* pcm, int frameSize, unsigned char* data, opus_int32 maxBytes) {
	return opus_encode_float_ptr(encoder, pcm, frameSize, data, maxBytes);
}

int engine_encoder_ctl_set(OpusEncoder* encoder, int request, opus_int32 value) {
	return opus_encoder_ctl_ptr(encoder, request, value);
}

int engine_packet_get_nb_samples(const unsigned char* data, opus_int32 len, opus_int32 rate) {
	return opus_packet_get_nb_samples_ptr(data, len, rate);
}

const char* engine_strerror(int error) {
	return opus_strerror_ptr(error);
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

var (
	loadOnce sync.Once
	loaded   bool
)

func load() error {
	loadOnce.Do(func() {
		loaded = C.engine_opus_load() != 0
	})

	if !loaded {
		return errNotExported
	}

	return nil
}

func bytesPointer(data []byte) *C.uchar {
	return (*C.uchar)(unsafe.Pointer(unsafe.SliceData(data)))
}

func floatsPointer(data []float32) *C.float {
	return (*C.float)(unsafe.Pointer(unsafe.SliceData(data)))
}

func int16sPointer(data []int16) *C.opus_int16 {
	return (*C.opus_int16)(unsafe.Pointer(unsafe.SliceData(data)))
}
