//go:build cputext

package ptxt

import "image"

import "github.com/tinne26/ptxt/core"

// Note: good reference for alpha compositing:
// https://developer.android.com/reference/android/graphics/PorterDuff.Mode#alpha-compositing-modes

const (
	BlendOver     core.BlendMode = core.BlendOver     // glyphs drawn over target (default mode)
	BlendReplace  core.BlendMode = core.BlendReplace  // glyph mask only (transparent pixels included!)
	BlendAdd      core.BlendMode = core.BlendAdd      // add colors (black adds nothing, white stays white)
	BlendSub      core.BlendMode = core.BlendSub      // subtract colors (black removes nothing) (alpha = target)
	BlendMultiply core.BlendMode = core.BlendMultiply // multiply % of glyph and target colors and MixOver
	BlendCut      core.BlendMode = core.BlendCut      // cut glyph shape hole based on alpha (cutout text)
	BlendHue      core.BlendMode = core.BlendHue      // keep highest alpha, blend hues proportionally
)

// ---- internal mask helper functions ----

// used for testing purposes
func newEmptyGlyphMask(width, height int) core.GlyphMask {
	return core.GlyphMask(image.NewAlpha(image.Rect(0, 0, width, height)))
}

func alphaMaskToMask(alphaMask *image.Alpha) core.GlyphMask {
	return alphaMask
}

