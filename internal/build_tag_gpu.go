//go:build !cputext

package internal

import "github.com/tinne26/ptxt/core"

// Based on Ebitengine internals.
const constMaskSizeFactor = 192

// With Ebitengine, the exact amount of mipmaps and helper fields is
// not known, so the values may not be completely accurate, and should
// be treated as a lower bound. With -tags cputext, the returned values
// are exact.
func glyphMaskByteSize(mask core.GlyphMask) uint32 {
	if mask == nil { return constMaskSizeFactor }
	w, h := mask.Size()
	return maskDimsByteSize(w, h)
}

func maskDimsByteSize(width, height int) uint32 {
	return uint32(width*height)*4 + constMaskSizeFactor
}
