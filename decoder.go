// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include "opus.h"

OpusDecoder* engine_decoder_create(opus_int32 rate, int channels, int* error);
void engine_decoder_destroy(OpusDecoder* decoder);
int engine_decode(OpusDecoder* decoder, const unsigned char* data, opus_int32 len, opus_int16* pcm, int frameSize);
int engine_decode_float(OpusDecoder* decoder, const unsigned char* data, opus_int32 len, float* pcm, int frameSize);
*/
import "C"

import "encoding/binary"

type Decoder struct {
	decoder    *C.OpusDecoder
	err        error
	sampleRate int
	channels   int
	pcm        []int16
}

func NewDecoder() *Decoder {
	d := &Decoder{}
	_ = d.Init(BandwidthFullband.SampleRate(), 1)

	return d
}

func NewDecoderWithOutput(sampleRate, channels int) (*Decoder, error) {
	d := &Decoder{}
	if err := d.Init(sampleRate, channels); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Decoder) Init(sampleRate, channels int) error {
	d.err = d.init(sampleRate, channels)

	return d.err
}

func (d *Decoder) init(sampleRate, channels int) error {
	if !validSampleRate(sampleRate) {
		return errInvalidSampleRate
	}

	if !validChannels(channels) {
		return errInvalidChannelCount
	}

	if err := load(); err != nil {
		return err
	}

	var code C.int
	decoder := C.engine_decoder_create(C.opus_int32(sampleRate), C.int(channels), &code)
	if code != C.OPUS_OK {
		return codeError("create decoder", code)
	}

	d.Close()
	d.decoder = decoder
	d.sampleRate = sampleRate
	d.channels = channels

	return nil
}

func (d *Decoder) ready() error {
	if d.decoder != nil {
		return nil
	}

	if d.err != nil {
		return d.err
	}

	return errClosed
}

func (d *Decoder) decodeInt16(in []byte, out []int16) (int, error) {
	if err := d.ready(); err != nil {
		return 0, err
	}

	if len(in) == 0 {
		return 0, errTooShortForTableOfContentsHeader
	}

	frameSize := len(out) / d.channels
	if frameSize == 0 {
		return 0, errOutBufferTooSmall
	}

	samples := C.engine_decode(d.decoder, bytesPointer(in), C.opus_int32(len(in)), int16sPointer(out), C.int(frameSize))
	if samples < 0 {
		return 0, codeError("decode", samples)
	}

	return int(samples), nil
}

func (d *Decoder) decodeFloat32(in []byte, out []float32) (int, error) {
	if err := d.ready(); err != nil {
		return 0, err
	}

	if len(in) == 0 {
		return 0, errTooShortForTableOfContentsHeader
	}

	frameSize := len(out) / d.channels
	if frameSize == 0 {
		return 0, errOutBufferTooSmall
	}

	samples := C.engine_decode_float(d.decoder, bytesPointer(in), C.opus_int32(len(in)), floatsPointer(out), C.int(frameSize))
	if samples < 0 {
		return 0, codeError("decode", samples)
	}

	return int(samples), nil
}

func (d *Decoder) Decode(in, out []byte) (bandwidth Bandwidth, isStereo bool, err error) {
	if cap(d.pcm) < len(out)/2 {
		d.pcm = make([]int16, len(out)/2)
	}
	pcm := d.pcm[:len(out)/2]

	samples, err := d.decodeInt16(in, pcm)
	if err != nil {
		return 0, false, err
	}

	for i, sample := range pcm[:samples*d.channels] {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(sample))
	}

	toc := TableOfContentsHeader(in[0])

	return toc.Configuration().Bandwidth(), toc.IsStereo(), nil
}

func (d *Decoder) DecodeFloat32(in []byte, out []float32) (bandwidth Bandwidth, isStereo bool, err error) {
	if _, err = d.decodeFloat32(in, out); err != nil {
		return 0, false, err
	}

	toc := TableOfContentsHeader(in[0])

	return toc.Configuration().Bandwidth(), toc.IsStereo(), nil
}

func (d *Decoder) DecodeToInt16(in []byte, out []int16) (int, error) {
	return d.decodeInt16(in, out)
}

func (d *Decoder) DecodeToFloat32(in []byte, out []float32) (int, error) {
	return d.decodeFloat32(in, out)
}

func (d *Decoder) DecodePLC(out []int16) error {
	if err := d.ready(); err != nil {
		return err
	}

	frameSize := len(out) / d.channels
	if frameSize == 0 || len(out)%d.channels != 0 || frameSize%(d.sampleRate/400) != 0 || frameSize > MaxFrameSamples {
		return errInvalidPLCFrameSize
	}

	samples := C.engine_decode(d.decoder, nil, 0, int16sPointer(out), C.int(frameSize))
	if samples < 0 {
		return codeError("decode plc", samples)
	}

	return nil
}

func (d *Decoder) Close() {
	if d.decoder != nil {
		C.engine_decoder_destroy(d.decoder)
		d.decoder = nil
	}
}
