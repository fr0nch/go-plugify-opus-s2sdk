# go-plugify-opus-s2sdk

Opus encoder/decoder for Go plugins running inside a Counter-Strike 2 server (Plugify / s2sdk).

The package does not ship its own libopus. It uses the Opus build that is already linked into the CS2 engine: the
functions are resolved at runtime from `libengine2.so` (Linux, `dlopen(..., RTLD_NOLOAD)`) or `engine2.dll` (Windows,
`GetModuleHandleA`). **It only works inside the CS2 process**; elsewhere `NewEncoder` / `NewDecoderWithOutput` return
the error `opus functions are not exported by the engine`.

## Based on pion/opus

The public API, option names, error messages, defaults and the TOC parser (`table_of_contents_header.go`) are taken
from [pion/opus](https://github.com/pion/opus) (master branch, MIT License, © The Pion community). Its license is
included in [LICENSES/pion-opus.txt](LICENSES/pion-opus.txt). The pure-Go SILK/CELT codec of pion is replaced by calls
into the engine's libopus.

The Opus headers in [include/](include) come from Opus 1.3.1 (BSD license, see [include/COPYING](include/COPYING)).

## Install

```sh
go get github.com/fr0nch/go-plugify-opus-s2sdk
```

cgo is required.

## Example

```go
import opus "github.com/fr0nch/go-plugify-opus-s2sdk"

encoder, err := opus.NewEncoder(
	opus.WithSampleRate(48000),
	opus.WithChannels(1),
	opus.WithApplication(opus.ApplicationAudio),
	opus.WithBitrate(64000),
	opus.WithComplexity(10),
)
if err != nil {
	return err
}
defer encoder.Close()

packet := make([]byte, opus.MaxPacketBytes)
n, err := encoder.EncodeFloat32(pcm[:960], packet)

decoder, err := opus.NewDecoderWithOutput(48000, 1)
if err != nil {
	return err
}
defer decoder.Close()

out := make([]float32, opus.MaxFrameSamples)
bandwidth, isStereo, err := decoder.DecodeFloat32(packet[:n], out)
```

## Differences from pion/opus

- `Encoder` and `Decoder` hold libopus state in C memory: always use them through pointers (`NewDecoder` and
  `NewDecoderWithOutput` return `*Decoder`), never copy them by value, and call `Close()` when done.
- `NewDecoder()` returns a decoder for 48 kHz mono like pion; if the engine is not available, the error is returned by
  the first `Decode*` call.
- `EncodeSILK` is not provided: it is pion's internal SILK encoder and public libopus has no equivalent.
- Extra: `WithInBandFEC` / `SetInBandFEC` / `InBandFEC` (`OPUS_SET_INBAND_FEC`).
- All sample rates supported by libopus (8, 12, 16, 24, 48 kHz) and frame sizes from 2.5 to 60 ms are accepted;
  `WithBandwidth` also accepts `BandwidthAuto` and `BandwidthMediumband`.
- `DecodePLC` accepts any multiple of 2.5 ms up to 120 ms, as libopus does.
- `Configuration`, `TableOfContentsHeader`, `ConfigurationMode`, `FrameDuration` and `FrameCode` are exported
  (`ParseTableOfContentsHeader(packet)`).

Not safe for concurrent use: one `Encoder` / `Decoder` per goroutine.

## License

[MIT](LICENSE) © Fr4nch. Code derived from pion/opus: MIT © The Pion community.
