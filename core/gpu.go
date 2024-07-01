//go:build !cputext

package core

import _ "image"

import "github.com/hajimehoshi/ebiten/v2"

// Alias to allow compiling the package without Ebitengine (-tags cputext).
// 
// Without Ebitengine, [Target] defaults to [image/draw.Image].
type Target = *ebiten.Image

// A GlyphMask is the image that results from rasterizing a glyph. You
// rarely need to use glyph masks directly unless you are working with
// advanced functions.
// 
// Without Ebitengine, [GlyphMask] defaults to [*image.Alpha]. The image
// bounds are adjusted to allow drawing the glyph at its intended position.
// In particular, bounds.Min.Y is typically negative, with y = 0 corresponding
// to the glyph's baseline, y < 0 to the ascending portions and y > 0 to
// the descending ones.
// 
// Notice that masks only use alpha values, but these don't represent
// opacity or luminance, but rather indices from the font's color set.
type GlyphMask = *ebiten.Image

// The blend mode specifies how to compose colors when drawing glyphs:
//  - Without Ebitengine, the blend mode can be BlendOver, BlendReplace, BlendAdd, BlendSub, BlendMultiply, BlendCut and BlendHue.
//  - With Ebitengine, the blend mode is [ebiten.Blend].
type BlendMode = ebiten.Blend
