//go:build cputext

package core

import "image"
import "image/draw"

// See documentation on build_tag_gpu.go instead.
// This is the fallback mode.

type Target = draw.Image

type GlyphMask = *image.Alpha

type BlendMode uint8

const (
	BlendOver     BlendMode = 0 // glyphs drawn over target (default mode)
	BlendReplace  BlendMode = 1 // glyph mask only (transparent pixels included!)
	BlendAdd      BlendMode = 2 // add colors (black adds nothing, white stays white)
	BlendSub      BlendMode = 3 // subtract colors (black removes nothing) (alpha = target)
	BlendMultiply BlendMode = 4 // multiply % of glyph and target colors and MixOver
	BlendCut      BlendMode = 5 // cut glyph shape hole based on alpha (cutout text)
	BlendHue      BlendMode = 6 // keep highest alpha, blend hues proportionally
)
