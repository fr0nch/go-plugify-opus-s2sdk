// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-FileCopyrightText: 2026 Fr4nch
// SPDX-License-Identifier: MIT

package opus

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include "opus.h"

const char* engine_strerror(int error);
*/
import "C"

import (
	"errors"
	"fmt"
)

var (
	errNotExported                      = errors.New("opus functions are not exported by the engine")
	errClosed                           = errors.New("opus object is closed")
	errTooShortForTableOfContentsHeader = errors.New("packet is too short to contain table of contents header")
	errInvalidSampleRate                = errors.New("invalid sample rate")
	errInvalidChannelCount              = errors.New("invalid channel count")
	errOutBufferTooSmall                = errors.New("out isn't large enough")
	errBitrateOutOfRange                = errors.New("bitrate out of range")
	errInvalidComplexity                = errors.New("invalid complexity")
	errInvalidInputLength               = errors.New("invalid input length")
	errInvalidFrameSize                 = errors.New("invalid frame size")
	errInvalidPLCFrameSize              = errors.New("invalid PLC frame size")
	errInvalidApplication               = errors.New("invalid application")
	errInvalidLossRate                  = errors.New("loss rate must be 0-100")
	errInvalidBandwidth                 = errors.New("invalid bandwidth")
)

func codeError(operation string, code C.int) error {
	return fmt.Errorf("%s: %s (%d)", operation, C.GoString(C.engine_strerror(code)), int(code))
}
