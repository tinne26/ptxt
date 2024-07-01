//go:build !cputext

package ptxt

import "image"

import "github.com/tinne26/ptxt/core"
import "github.com/hajimehoshi/ebiten/v2"

// ---- internal mask helper functions ----

// used for testing purposes
func newEmptyGlyphMask(width, height int) core.GlyphMask {
	return core.GlyphMask(ebiten.NewImage(width, height))
}

// ---- drawing ----

func alphaMaskToMask(alphaMask *image.Alpha) core.GlyphMask {
	if alphaMask == nil { return nil }

	// NOTE: since ebitengine doesn't have good support for alpha images,
	//       this is quite a pain, but not much we can do from here.
	rgba   := image.NewRGBA(alphaMask.Rect)
	pixels := rgba.Pix
	index  := 0
	for _, value := range alphaMask.Pix {
		pixels[index + 0] = value
		pixels[index + 1] = value
		pixels[index + 2] = value
		pixels[index + 3] = value
		index += 4
	}
	opts := ebiten.NewImageFromImageOptions{ PreserveBounds: true }
	out := ebiten.NewImageFromImageWithOptions(rgba, &opts)
	return out
}
