// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

type (
	TableOfContentsHeader byte
	Configuration         byte
	ConfigurationMode     byte
	FrameDuration         byte
	FrameCode             byte
	Bandwidth             byte
)

func ParseTableOfContentsHeader(packet []byte) (TableOfContentsHeader, error) {
	if len(packet) == 0 {
		return 0, errTooShortForTableOfContentsHeader
	}

	return TableOfContentsHeader(packet[0]), nil
}

func (t TableOfContentsHeader) Configuration() Configuration {
	return Configuration(t >> 3)
}

func (t TableOfContentsHeader) IsStereo() bool {
	return (t & 0b00000100) != 0
}

const (
	FrameCodeOneFrame           FrameCode = 0
	FrameCodeTwoEqualFrames     FrameCode = 1
	FrameCodeTwoDifferentFrames FrameCode = 2
	FrameCodeArbitraryFrames    FrameCode = 3
)

func (t TableOfContentsHeader) FrameCode() FrameCode {
	return FrameCode(t & 0b00000011)
}

const (
	ConfigurationModeSilkOnly ConfigurationMode = iota + 1
	ConfigurationModeCELTOnly
	ConfigurationModeHybrid
)

func (c ConfigurationMode) String() string {
	switch c {
	case ConfigurationModeSilkOnly:
		return "Silk-only"
	case ConfigurationModeCELTOnly:
		return "CELT-only"
	case ConfigurationModeHybrid:
		return "Hybrid"
	}

	return "Invalid Configuration Mode"
}

func (c Configuration) Mode() ConfigurationMode {
	switch {
	case c <= 11:
		return ConfigurationModeSilkOnly
	case c >= 12 && c <= 15:
		return ConfigurationModeHybrid
	case c >= 16 && c <= 31:
		return ConfigurationModeCELTOnly
	default:
		return 0
	}
}

const (
	FrameDuration2500us FrameDuration = iota + 1
	FrameDuration5ms
	FrameDuration10ms
	FrameDuration20ms
	FrameDuration40ms
	FrameDuration60ms
)

func (f FrameDuration) String() string {
	switch f {
	case FrameDuration2500us:
		return "2.5ms"
	case FrameDuration5ms:
		return "5ms"
	case FrameDuration10ms:
		return "10ms"
	case FrameDuration20ms:
		return "20ms"
	case FrameDuration40ms:
		return "40ms"
	case FrameDuration60ms:
		return "60ms"
	}

	return "Invalid Frame Duration"
}

func (f FrameDuration) Nanoseconds() int {
	switch f {
	case FrameDuration2500us:
		return 2500000
	case FrameDuration5ms:
		return 5000000
	case FrameDuration10ms:
		return 10000000
	case FrameDuration20ms:
		return 20000000
	case FrameDuration40ms:
		return 40000000
	case FrameDuration60ms:
		return 60000000
	}

	return 0
}

func (f FrameDuration) Samples(sampleRate int) int {
	return int(int64(sampleRate) * int64(f.Nanoseconds()) / 1000000000)
}

func (c Configuration) FrameDuration() FrameDuration {
	switch c {
	case 16, 20, 24, 28:
		return FrameDuration2500us
	case 17, 21, 25, 29:
		return FrameDuration5ms
	case 0, 4, 8, 12, 14, 18, 22, 26, 30:
		return FrameDuration10ms
	case 1, 5, 9, 13, 15, 19, 23, 27, 31:
		return FrameDuration20ms
	case 2, 6, 10:
		return FrameDuration40ms
	case 3, 7, 11:
		return FrameDuration60ms
	}

	return 0
}

const (
	BandwidthAuto          Bandwidth = 0
	BandwidthNarrowband    Bandwidth = 1
	BandwidthMediumband    Bandwidth = 2
	BandwidthWideband      Bandwidth = 3
	BandwidthSuperwideband Bandwidth = 4
	BandwidthFullband      Bandwidth = 5
)

func (c Configuration) Bandwidth() Bandwidth {
	switch {
	case c <= 3:
		return BandwidthNarrowband
	case c <= 7:
		return BandwidthMediumband
	case c <= 11:
		return BandwidthWideband
	case c <= 13:
		return BandwidthSuperwideband
	case c <= 15:
		return BandwidthFullband
	case c <= 19:
		return BandwidthNarrowband
	case c <= 23:
		return BandwidthWideband
	case c <= 27:
		return BandwidthSuperwideband
	case c <= 31:
		return BandwidthFullband
	}

	return Bandwidth(255)
}

func (b Bandwidth) String() string {
	switch b {
	case BandwidthAuto:
		return "Auto"
	case BandwidthNarrowband:
		return "Narrowband"
	case BandwidthMediumband:
		return "Mediumband"
	case BandwidthWideband:
		return "Wideband"
	case BandwidthSuperwideband:
		return "Superwideband"
	case BandwidthFullband:
		return "Fullband"
	}

	return "Invalid Bandwidth"
}

func (b Bandwidth) SampleRate() int {
	switch b {
	case BandwidthAuto:
		return 0
	case BandwidthNarrowband:
		return 8000
	case BandwidthMediumband:
		return 12000
	case BandwidthWideband:
		return 16000
	case BandwidthSuperwideband:
		return 24000
	case BandwidthFullband:
		return 48000
	}

	return 0
}
