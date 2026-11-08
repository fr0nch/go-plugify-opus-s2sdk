// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include "opus.h"

OpusEncoder* engine_encoder_create(opus_int32 rate, int channels, int application, int* error);
void engine_encoder_destroy(OpusEncoder* encoder);
opus_int32 engine_encode_float(OpusEncoder* encoder, const float* pcm, int frameSize, unsigned char* data, opus_int32 maxBytes);
int engine_encoder_ctl_set(OpusEncoder* encoder, int request, opus_int32 value);
*/
import "C"

import (
	"encoding/binary"
	"fmt"
)

type Application int

const (
	ApplicationVoIP               Application = C.OPUS_APPLICATION_VOIP
	ApplicationAudio              Application = C.OPUS_APPLICATION_AUDIO
	ApplicationRestrictedLowDelay Application = C.OPUS_APPLICATION_RESTRICTED_LOWDELAY
)

const defaultBitrate = 24000

type Encoder struct {
	encoder *C.OpusEncoder

	sampleRate     int
	channels       int
	application    Application
	bitrate        int
	complexity     int
	vbr            bool
	constrainedVBR bool
	lossRate       int
	inBandFEC      bool
	bandwidth      Bandwidth
	maxBandwidth   Bandwidth

	pcm []float32
}

type EncoderOption func(*Encoder) error

func WithSampleRate(rate int) EncoderOption {
	return func(e *Encoder) error {
		if !validSampleRate(rate) {
			return errInvalidSampleRate
		}
		e.sampleRate = rate

		return nil
	}
}

func WithChannels(channels int) EncoderOption {
	return func(e *Encoder) error {
		if !validChannels(channels) {
			return errInvalidChannelCount
		}
		e.channels = channels

		return nil
	}
}

func WithApplication(app Application) EncoderOption {
	return func(e *Encoder) error {
		switch app {
		case ApplicationVoIP, ApplicationAudio, ApplicationRestrictedLowDelay:
		default:
			return fmt.Errorf("%w: %d", errInvalidApplication, app)
		}
		e.application = app

		return nil
	}
}

func WithBitrate(bps int) EncoderOption {
	return func(e *Encoder) error {
		if bps < minBitrate || bps > maxBitrate {
			return fmt.Errorf("%w: %d", errBitrateOutOfRange, bps)
		}
		e.bitrate = bps

		return nil
	}
}

func WithComplexity(complexity int) EncoderOption {
	return func(e *Encoder) error {
		if complexity < 0 || complexity > 10 {
			return fmt.Errorf("%w: %d", errInvalidComplexity, complexity)
		}
		e.complexity = complexity

		return nil
	}
}

func WithVBR(vbr bool) EncoderOption {
	return func(e *Encoder) error {
		e.vbr = vbr

		return nil
	}
}

func WithConstrainedVBR(cvbr bool) EncoderOption {
	return func(e *Encoder) error {
		e.constrainedVBR = cvbr

		return nil
	}
}

func WithInBandFEC(fec bool) EncoderOption {
	return func(e *Encoder) error {
		e.inBandFEC = fec

		return nil
	}
}

func WithBandwidth(bw Bandwidth) EncoderOption {
	return func(e *Encoder) error {
		if bw > BandwidthFullband {
			return fmt.Errorf("%w: %d", errInvalidBandwidth, bw)
		}
		e.bandwidth = bw

		return nil
	}
}

func WithMaxBandwidth(bw Bandwidth) EncoderOption {
	return func(e *Encoder) error {
		if bw == BandwidthAuto {
			return fmt.Errorf("%w: max bandwidth must be explicit", errInvalidBandwidth)
		}
		if bw > BandwidthFullband {
			return fmt.Errorf("%w: %d", errInvalidBandwidth, bw)
		}
		e.maxBandwidth = bw

		return nil
	}
}

func NewEncoder(opts ...EncoderOption) (*Encoder, error) {
	e := &Encoder{
		sampleRate:     48000,
		channels:       1,
		application:    ApplicationAudio,
		bitrate:        defaultBitrate,
		complexity:     5,
		vbr:            false,
		constrainedVBR: true,
		lossRate:       0,
		bandwidth:      BandwidthAuto,
		maxBandwidth:   BandwidthFullband,
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	if err := load(); err != nil {
		return nil, err
	}

	var code C.int
	e.encoder = C.engine_encoder_create(C.opus_int32(e.sampleRate), C.int(e.channels), C.int(e.application), &code)
	if code != C.OPUS_OK {
		return nil, codeError("create encoder", code)
	}

	if err := e.apply(); err != nil {
		e.Close()
		return nil, err
	}

	return e, nil
}

func (e *Encoder) apply() error {
	settings := []struct {
		name    string
		request C.int
		value   int
	}{
		{"set bitrate", C.OPUS_SET_BITRATE_REQUEST, e.bitrate},
		{"set complexity", C.OPUS_SET_COMPLEXITY_REQUEST, e.complexity},
		{"set vbr", C.OPUS_SET_VBR_REQUEST, boolValue(e.vbr)},
		{"set vbr constraint", C.OPUS_SET_VBR_CONSTRAINT_REQUEST, boolValue(e.constrainedVBR)},
		{"set packet loss", C.OPUS_SET_PACKET_LOSS_PERC_REQUEST, e.lossRate},
		{"set inband fec", C.OPUS_SET_INBAND_FEC_REQUEST, boolValue(e.inBandFEC)},
		{"set max bandwidth", C.OPUS_SET_MAX_BANDWIDTH_REQUEST, e.maxBandwidth.opusValue()},
		{"set bandwidth", C.OPUS_SET_BANDWIDTH_REQUEST, e.bandwidth.opusValue()},
	}

	for _, setting := range settings {
		if err := e.ctl(setting.name, setting.request, setting.value); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) ctl(name string, request C.int, value int) error {
	if e.encoder == nil {
		return errClosed
	}

	if code := C.engine_encoder_ctl_set(e.encoder, request, C.opus_int32(value)); code != C.OPUS_OK {
		return codeError(name, code)
	}

	return nil
}

func (e *Encoder) SetBitrate(bps int) error {
	if bps < minBitrate || bps > maxBitrate {
		return fmt.Errorf("%w: %d", errBitrateOutOfRange, bps)
	}

	if err := e.ctl("set bitrate", C.OPUS_SET_BITRATE_REQUEST, bps); err != nil {
		return err
	}
	e.bitrate = bps

	return nil
}

func (e *Encoder) SetComplexity(complexity int) error {
	if complexity < 0 || complexity > 10 {
		return fmt.Errorf("%w: %d", errInvalidComplexity, complexity)
	}

	if err := e.ctl("set complexity", C.OPUS_SET_COMPLEXITY_REQUEST, complexity); err != nil {
		return err
	}
	e.complexity = complexity

	return nil
}

func (e *Encoder) SetApplication(app Application) error {
	switch app {
	case ApplicationVoIP, ApplicationAudio, ApplicationRestrictedLowDelay:
	default:
		return fmt.Errorf("%w: %d", errInvalidApplication, app)
	}

	if err := e.ctl("set application", C.OPUS_SET_APPLICATION_REQUEST, int(app)); err != nil {
		return err
	}
	e.application = app

	return nil
}

func (e *Encoder) SetVBR(vbr bool) {
	if e.ctl("set vbr", C.OPUS_SET_VBR_REQUEST, boolValue(vbr)) == nil {
		e.vbr = vbr
	}
}

func (e *Encoder) SetConstrainedVBR(cvbr bool) {
	if e.ctl("set vbr constraint", C.OPUS_SET_VBR_CONSTRAINT_REQUEST, boolValue(cvbr)) == nil {
		e.constrainedVBR = cvbr
	}
}

func (e *Encoder) SetLossRate(rate int) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("%w: %d", errInvalidLossRate, rate)
	}

	if err := e.ctl("set packet loss", C.OPUS_SET_PACKET_LOSS_PERC_REQUEST, rate); err != nil {
		return err
	}
	e.lossRate = rate

	return nil
}

func (e *Encoder) SetInBandFEC(fec bool) error {
	if err := e.ctl("set inband fec", C.OPUS_SET_INBAND_FEC_REQUEST, boolValue(fec)); err != nil {
		return err
	}
	e.inBandFEC = fec

	return nil
}

func (e *Encoder) SetBandwidth(bw Bandwidth) error {
	if bw > BandwidthFullband {
		return fmt.Errorf("%w: %d", errInvalidBandwidth, bw)
	}

	if err := e.ctl("set bandwidth", C.OPUS_SET_BANDWIDTH_REQUEST, bw.opusValue()); err != nil {
		return err
	}
	e.bandwidth = bw

	return nil
}

func (e *Encoder) SetMaxBandwidth(bw Bandwidth) error {
	if bw == BandwidthAuto {
		return fmt.Errorf("%w: max bandwidth must be explicit", errInvalidBandwidth)
	}
	if bw > BandwidthFullband {
		return fmt.Errorf("%w: %d", errInvalidBandwidth, bw)
	}

	if err := e.ctl("set max bandwidth", C.OPUS_SET_MAX_BANDWIDTH_REQUEST, bw.opusValue()); err != nil {
		return err
	}
	e.maxBandwidth = bw

	return nil
}

func (e *Encoder) Bitrate() int { return e.bitrate }

func (e *Encoder) Application() Application { return e.application }

func (e *Encoder) VBR() bool { return e.vbr }

func (e *Encoder) Complexity() int { return e.complexity }

func (e *Encoder) ConstrainedVBR() bool { return e.constrainedVBR }

func (e *Encoder) LossRate() int { return e.lossRate }

func (e *Encoder) InBandFEC() bool { return e.inBandFEC }

func (e *Encoder) Bandwidth() Bandwidth { return e.bandwidth }

func (e *Encoder) MaxBandwidth() Bandwidth { return e.maxBandwidth }

func (e *Encoder) Encode(in []byte, out []byte) (int, error) {
	if len(in)%2 != 0 {
		return 0, fmt.Errorf("%w: s16le length %d not a multiple of 2", errInvalidInputLength, len(in))
	}

	samples := len(in) / 2
	if cap(e.pcm) < samples {
		e.pcm = make([]float32, samples)
	}
	pcm := e.pcm[:samples]

	for i := range pcm {
		pcm[i] = float32(int16(binary.LittleEndian.Uint16(in[i*2:]))) / 32768
	}

	return e.EncodeFloat32(pcm, out)
}

func (e *Encoder) EncodeFloat32(in []float32, out []byte) (int, error) {
	if e.encoder == nil {
		return 0, errClosed
	}

	if len(in)%e.channels != 0 || !validEncodeFrameSize(e.sampleRate, len(in)/e.channels) {
		return 0, fmt.Errorf("%w: %d samples", errInvalidFrameSize, len(in))
	}

	if len(out) == 0 {
		return 0, errOutBufferTooSmall
	}

	bytes := C.engine_encode_float(e.encoder, floatsPointer(in), C.int(len(in)/e.channels), bytesPointer(out), C.opus_int32(len(out)))
	if bytes < 0 {
		return 0, codeError("encode", C.int(bytes))
	}

	return int(bytes), nil
}

func (e *Encoder) Close() {
	if e.encoder != nil {
		C.engine_encoder_destroy(e.encoder)
		e.encoder = nil
	}
}

func (b Bandwidth) opusValue() int {
	switch b {
	case BandwidthNarrowband:
		return C.OPUS_BANDWIDTH_NARROWBAND
	case BandwidthMediumband:
		return C.OPUS_BANDWIDTH_MEDIUMBAND
	case BandwidthWideband:
		return C.OPUS_BANDWIDTH_WIDEBAND
	case BandwidthSuperwideband:
		return C.OPUS_BANDWIDTH_SUPERWIDEBAND
	case BandwidthFullband:
		return C.OPUS_BANDWIDTH_FULLBAND
	}

	return C.OPUS_AUTO
}

func boolValue(value bool) int {
	if value {
		return 1
	}

	return 0
}
