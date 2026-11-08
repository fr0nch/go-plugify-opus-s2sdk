// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include "opus.h"

int engine_packet_get_nb_samples(const unsigned char* data, opus_int32 len, opus_int32 rate);
*/
import "C"

const (
	minBitrate = 6000
	maxBitrate = 510000

	MaxFrameSamples = 5760
	MaxPacketBytes  = 1275 * 3
)

func validSampleRate(rate int) bool {
	switch rate {
	case 8000, 12000, 16000, 24000, 48000:
		return true
	}

	return false
}

func validChannels(channels int) bool {
	return channels == 1 || channels == 2
}

func validEncodeFrameSize(sampleRate, samples int) bool {
	for _, tenthsMs := range []int{25, 50, 100, 200, 400, 600} {
		if samples == sampleRate*tenthsMs/10000 {
			return true
		}
	}

	return false
}

func PacketSamples(packet []byte, sampleRate int) (int, error) {
	if len(packet) == 0 {
		return 0, errTooShortForTableOfContentsHeader
	}

	if err := load(); err != nil {
		return 0, err
	}

	samples := C.engine_packet_get_nb_samples(bytesPointer(packet), C.opus_int32(len(packet)), C.opus_int32(sampleRate))
	if samples < 0 {
		return 0, codeError("packet samples", samples)
	}

	return int(samples), nil
}
