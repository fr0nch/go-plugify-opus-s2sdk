// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

import (
	"errors"
	"testing"
)

func TestConfigurationMode(t *testing.T) {
	cases := map[Configuration]ConfigurationMode{
		0:  ConfigurationModeSilkOnly,
		11: ConfigurationModeSilkOnly,
		12: ConfigurationModeHybrid,
		15: ConfigurationModeHybrid,
		16: ConfigurationModeCELTOnly,
		31: ConfigurationModeCELTOnly,
		32: 0,
	}

	for config, want := range cases {
		if got := config.Mode(); got != want {
			t.Errorf("Configuration(%d).Mode() = %v, want %v", config, got, want)
		}
	}
}

func TestConfigurationModeString(t *testing.T) {
	cases := map[ConfigurationMode]string{
		ConfigurationModeSilkOnly: "Silk-only",
		ConfigurationModeCELTOnly: "CELT-only",
		ConfigurationModeHybrid:   "Hybrid",
		0:                         "Invalid Configuration Mode",
	}

	for mode, want := range cases {
		if got := mode.String(); got != want {
			t.Errorf("ConfigurationMode(%d).String() = %q, want %q", mode, got, want)
		}
	}
}

func TestConfigurationFrameDuration(t *testing.T) {
	cases := map[Configuration]FrameDuration{
		16: FrameDuration2500us,
		17: FrameDuration5ms,
		0:  FrameDuration10ms,
		1:  FrameDuration20ms,
		10: FrameDuration40ms,
		11: FrameDuration60ms,
		31: FrameDuration20ms,
		32: 0,
	}

	for config, want := range cases {
		if got := config.FrameDuration(); got != want {
			t.Errorf("Configuration(%d).FrameDuration() = %v, want %v", config, got, want)
		}
	}
}

func TestFrameDuration(t *testing.T) {
	cases := []struct {
		duration    FrameDuration
		name        string
		nanoseconds int
		samples48k  int
	}{
		{FrameDuration2500us, "2.5ms", 2500000, 120},
		{FrameDuration5ms, "5ms", 5000000, 240},
		{FrameDuration10ms, "10ms", 10000000, 480},
		{FrameDuration20ms, "20ms", 20000000, 960},
		{FrameDuration40ms, "40ms", 40000000, 1920},
		{FrameDuration60ms, "60ms", 60000000, 2880},
		{0, "Invalid Frame Duration", 0, 0},
	}

	for _, c := range cases {
		if got := c.duration.String(); got != c.name {
			t.Errorf("FrameDuration(%d).String() = %q, want %q", c.duration, got, c.name)
		}
		if got := c.duration.Nanoseconds(); got != c.nanoseconds {
			t.Errorf("FrameDuration(%d).Nanoseconds() = %d, want %d", c.duration, got, c.nanoseconds)
		}
		if got := c.duration.Samples(48000); got != c.samples48k {
			t.Errorf("FrameDuration(%d).Samples(48000) = %d, want %d", c.duration, got, c.samples48k)
		}
	}
}

func TestConfigurationBandwidth(t *testing.T) {
	cases := map[Configuration]Bandwidth{
		0:  BandwidthNarrowband,
		4:  BandwidthMediumband,
		8:  BandwidthWideband,
		12: BandwidthSuperwideband,
		14: BandwidthFullband,
		16: BandwidthNarrowband,
		20: BandwidthWideband,
		24: BandwidthSuperwideband,
		28: BandwidthFullband,
		32: Bandwidth(255),
	}

	for config, want := range cases {
		if got := config.Bandwidth(); got != want {
			t.Errorf("Configuration(%d).Bandwidth() = %v, want %v", config, got, want)
		}
	}
}

func TestBandwidth(t *testing.T) {
	cases := []struct {
		bandwidth  Bandwidth
		name       string
		sampleRate int
	}{
		{BandwidthAuto, "Auto", 0},
		{BandwidthNarrowband, "Narrowband", 8000},
		{BandwidthMediumband, "Mediumband", 12000},
		{BandwidthWideband, "Wideband", 16000},
		{BandwidthSuperwideband, "Superwideband", 24000},
		{BandwidthFullband, "Fullband", 48000},
		{Bandwidth(255), "Invalid Bandwidth", 0},
	}

	for _, c := range cases {
		if got := c.bandwidth.String(); got != c.name {
			t.Errorf("Bandwidth(%d).String() = %q, want %q", c.bandwidth, got, c.name)
		}
		if got := c.bandwidth.SampleRate(); got != c.sampleRate {
			t.Errorf("Bandwidth(%d).SampleRate() = %d, want %d", c.bandwidth, got, c.sampleRate)
		}
	}
}

func TestTableOfContentsHeader(t *testing.T) {
	cases := []struct {
		toc       byte
		config    Configuration
		mode      ConfigurationMode
		bandwidth Bandwidth
		duration  FrameDuration
		stereo    bool
		frameCode FrameCode
	}{
		{0b11111000, 31, ConfigurationModeCELTOnly, BandwidthFullband, FrameDuration20ms, false, FrameCodeOneFrame},
		{0b11111101, 31, ConfigurationModeCELTOnly, BandwidthFullband, FrameDuration20ms, true, FrameCodeTwoEqualFrames},
		{0b00001010, 1, ConfigurationModeSilkOnly, BandwidthNarrowband, FrameDuration20ms, false, FrameCodeTwoDifferentFrames},
		{0b01111111, 15, ConfigurationModeHybrid, BandwidthFullband, FrameDuration20ms, true, FrameCodeArbitraryFrames},
		{0b10000000, 16, ConfigurationModeCELTOnly, BandwidthNarrowband, FrameDuration2500us, false, FrameCodeOneFrame},
	}

	for _, c := range cases {
		toc, err := ParseTableOfContentsHeader([]byte{c.toc, 0xff})
		if err != nil {
			t.Fatalf("ParseTableOfContentsHeader(%08b): %v", c.toc, err)
		}

		config := toc.Configuration()
		if config != c.config || config.Mode() != c.mode || config.Bandwidth() != c.bandwidth || config.FrameDuration() != c.duration {
			t.Errorf("TOC %08b: config %d %v %v %v, want %d %v %v %v", c.toc, config, config.Mode(), config.Bandwidth(),
				config.FrameDuration(), c.config, c.mode, c.bandwidth, c.duration)
		}
		if toc.IsStereo() != c.stereo {
			t.Errorf("TOC %08b: IsStereo() = %v, want %v", c.toc, toc.IsStereo(), c.stereo)
		}
		if toc.FrameCode() != c.frameCode {
			t.Errorf("TOC %08b: FrameCode() = %d, want %d", c.toc, toc.FrameCode(), c.frameCode)
		}
	}

	if _, err := ParseTableOfContentsHeader(nil); !errors.Is(err, errTooShortForTableOfContentsHeader) {
		t.Errorf("ParseTableOfContentsHeader(nil) error = %v", err)
	}
}

func TestEncoderOptions(t *testing.T) {
	cases := []struct {
		name   string
		option EncoderOption
		err    error
	}{
		{"sample rate", WithSampleRate(44100), errInvalidSampleRate},
		{"channels", WithChannels(3), errInvalidChannelCount},
		{"application", WithApplication(1), errInvalidApplication},
		{"bitrate low", WithBitrate(minBitrate - 1), errBitrateOutOfRange},
		{"bitrate high", WithBitrate(maxBitrate + 1), errBitrateOutOfRange},
		{"complexity", WithComplexity(11), errInvalidComplexity},
		{"bandwidth", WithBandwidth(6), errInvalidBandwidth},
		{"max bandwidth auto", WithMaxBandwidth(BandwidthAuto), errInvalidBandwidth},
		{"max bandwidth", WithMaxBandwidth(6), errInvalidBandwidth},
	}

	for _, c := range cases {
		if _, err := NewEncoder(c.option); !errors.Is(err, c.err) {
			t.Errorf("%s: error = %v, want %v", c.name, err, c.err)
		}
	}
}

func TestEncoderOptionsApply(t *testing.T) {
	e := &Encoder{}
	opts := []EncoderOption{
		WithSampleRate(24000),
		WithChannels(2),
		WithApplication(ApplicationVoIP),
		WithBitrate(64000),
		WithComplexity(7),
		WithVBR(true),
		WithConstrainedVBR(true),
		WithInBandFEC(true),
		WithBandwidth(BandwidthMediumband),
		WithMaxBandwidth(BandwidthWideband),
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			t.Fatal(err)
		}
	}

	if e.sampleRate != 24000 || e.channels != 2 || e.Application() != ApplicationVoIP || e.Bitrate() != 64000 ||
		e.Complexity() != 7 || !e.VBR() || !e.ConstrainedVBR() || !e.InBandFEC() ||
		e.Bandwidth() != BandwidthMediumband || e.MaxBandwidth() != BandwidthWideband {
		t.Errorf("unexpected encoder state: %+v", e)
	}
}

func TestEncoderValidationWithoutEngine(t *testing.T) {
	e := &Encoder{}

	if err := e.SetLossRate(101); !errors.Is(err, errInvalidLossRate) {
		t.Errorf("SetLossRate(101) error = %v", err)
	}
	if err := e.SetLossRate(10); !errors.Is(err, errClosed) {
		t.Errorf("SetLossRate(10) on closed encoder error = %v", err)
	}
	if err := e.SetMaxBandwidth(BandwidthAuto); !errors.Is(err, errInvalidBandwidth) {
		t.Errorf("SetMaxBandwidth(Auto) error = %v", err)
	}
	if _, err := e.Encode([]byte{0}, make([]byte, MaxPacketBytes)); !errors.Is(err, errInvalidInputLength) {
		t.Errorf("Encode odd length error = %v", err)
	}
	if _, err := e.EncodeFloat32(make([]float32, 960), make([]byte, MaxPacketBytes)); !errors.Is(err, errClosed) {
		t.Errorf("EncodeFloat32 on closed encoder error = %v", err)
	}
}

func TestBandwidthOpusValue(t *testing.T) {
	cases := map[Bandwidth]int{
		BandwidthAuto:          -1000,
		BandwidthNarrowband:    1101,
		BandwidthMediumband:    1102,
		BandwidthWideband:      1103,
		BandwidthSuperwideband: 1104,
		BandwidthFullband:      1105,
	}

	for bandwidth, want := range cases {
		if got := bandwidth.opusValue(); got != want {
			t.Errorf("Bandwidth(%d).opusValue() = %d, want %d", bandwidth, got, want)
		}
	}
}

func TestDecoderInitValidation(t *testing.T) {
	if _, err := NewDecoderWithOutput(44100, 1); !errors.Is(err, errInvalidSampleRate) {
		t.Errorf("NewDecoderWithOutput(44100, 1) error = %v", err)
	}
	if _, err := NewDecoderWithOutput(48000, 0); !errors.Is(err, errInvalidChannelCount) {
		t.Errorf("NewDecoderWithOutput(48000, 0) error = %v", err)
	}
}

func TestWithoutEngine(t *testing.T) {
	if _, err := NewEncoder(); !errors.Is(err, errNotExported) {
		t.Errorf("NewEncoder() error = %v, want %v", err, errNotExported)
	}
	if _, err := NewDecoderWithOutput(48000, 2); !errors.Is(err, errNotExported) {
		t.Errorf("NewDecoderWithOutput() error = %v, want %v", err, errNotExported)
	}
	if _, err := PacketSamples([]byte{0xf8}, 48000); !errors.Is(err, errNotExported) {
		t.Errorf("PacketSamples() error = %v, want %v", err, errNotExported)
	}

	d := NewDecoder()
	if _, _, err := d.Decode([]byte{0xf8}, make([]byte, 1920)); !errors.Is(err, errNotExported) {
		t.Errorf("Decode() error = %v, want %v", err, errNotExported)
	}
	if err := d.DecodePLC(make([]int16, 960)); !errors.Is(err, errNotExported) {
		t.Errorf("DecodePLC() error = %v, want %v", err, errNotExported)
	}
	d.Close()
}
