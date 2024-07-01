//go:build cputext

package strand

import "image/color"

import "github.com/tinne26/ptxt/core"
import "github.com/tinne26/ggfnt"

func (*Strand) notifyShaderNonMainDyeChange() {} // used to relink uniforms in gpu version
func (*Strand) notifyShaderPaletteChange() {} // used to relink uniforms in gpu version

type renderData struct {
	dyeMappings []uint8
	blend core.BlendMode
	
	prevTR, prevTG, prevTB, prevTA uint32
	prevRGBA8 color.RGBA
}

func (self *Strand) renderDataInit() {
	// dye mappings point from color index to dye tone index
	// the actual alphas are stored in the palette instead
	numDyeAlphas := self.font.Color().DyeAlphasCount()
	dyeMappings := make([]uint8, numDyeAlphas)
	minDyeIndex := (255 - numDyeAlphas) + 1
	for i := uint8(0); i < self.font.Color().NumDyes(); i++ {
		start, end := self.font.Color().GetDyeRange(ggfnt.DyeKey(i))
		if start == 0 && end == 0 { panic(brokenCode) } // or broken font
		if start < minDyeIndex { panic(brokenCode) } // or broken font
		start -= minDyeIndex
		end   -= minDyeIndex
		for j := start; j <= end; j++ {
			dyeMappings[j] = i
		}
	}
	self.re.dyeMappings = dyeMappings
}

func (self *renderData) GetBlendFunc() func(core.Target, int, int, [4]float32) color.RGBA {
	switch self.blend {
	case core.BlendOver: // glyphs drawn over target (default mode)
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			r, g, b, a := target.At(x, y).RGBA()
			if a == self.prevTA && r == self.prevTR && g == self.prevTG && b == self.prevTB {
				// fast case, same as previous
			} else if a == 0 {
				self.prevRGBA8 = f32toRGBA(rgba)
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			} else {
				self.prevRGBA8 = blendOverFunc(rgba, rgba32ToFloat32(r, g, b, a))
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			}
			return self.prevRGBA8
		}
	case core.BlendReplace: // glyph mask only (transparent pixels included!)
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			if rgba[3] == 0 { return color.RGBA{0, 0, 0, 0} }
			if self.prevTA != 111 {
				self.prevTA = 111
				self.prevRGBA8 = f32toRGBA(rgba)
			}
			return self.prevRGBA8
		}
	case core.BlendAdd: // add colors (black adds nothing, white stays white)
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			r, g, b, a := target.At(x, y).RGBA()
			if a == self.prevTA && r == self.prevTR && g == self.prevTG && b == self.prevTB {
				// fast case, same as previous
			} else {
				self.prevRGBA8 = addFunc(rgba, rgba32ToFloat32(r, g, b, a))
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			}
			return self.prevRGBA8
		}
	case core.BlendSub: // subtract colors (black removes nothing) (alpha = target)
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			r, g, b, a := target.At(x, y).RGBA()
			if a == self.prevTA && r == self.prevTR && g == self.prevTG && b == self.prevTB {
				// fast case, same as previous
			} else {
				self.prevRGBA8 = subFunc(rgba, rgba32ToFloat32(r, g, b, a))
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			}
			return self.prevRGBA8
		}
	case core.BlendMultiply: // multiply % of glyph and target colors and MixOver
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			r, g, b, a := target.At(x, y).RGBA()
			if a == self.prevTA && r == self.prevTR && g == self.prevTG && b == self.prevTB {
				// fast case, same as previous
			} else if a == 0 || rgba[3] == 0.0 {
				self.prevRGBA8 = color.RGBA{0, 0, 0, 0}
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			} else {
				self.prevRGBA8 = multiplyFunc(rgba, rgba32ToFloat32(r, g, b, a))
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			}
			return self.prevRGBA8
		}
	case core.BlendCut: // cut glyph shape hole based on alpha (cutout text)
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			r, g, b, a := target.At(x, y).RGBA()
			if a == self.prevTA && r == self.prevTR && g == self.prevTG && b == self.prevTB {
				// fast case, same as previous
			} else {
				self.prevRGBA8 = cutFunc(rgba, rgba32ToFloat32(r, g, b, a))
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			}
			return self.prevRGBA8
		}
	case core.BlendHue: // keep highest alpha, blend hues proportionally
		return func(target core.Target, x, y int, rgba [4]float32) color.RGBA {
			r, g, b, a := target.At(x, y).RGBA()
			if a == self.prevTA && r == self.prevTR && g == self.prevTG && b == self.prevTB {
				// fast case, same as previous
			} else if a == 0 {
				self.prevRGBA8 = f32toRGBA(rgba)
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			} else {
				self.prevRGBA8 = hueFunc(rgba, rgba32ToFloat32(r, g, b, a))
				self.prevTR, self.prevTG, self.prevTB, self.prevTA = r, g, b, a
			}
			return self.prevRGBA8
		}
	default:
		panic("invalid blend mode")
	}
}

func (self *Strand) setBlendMode(blend core.BlendMode) {
	self.re.blend = blend
}

func (self *Strand) drawMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32, orientation uint8) {
	// helpers
	blendFunc := self.re.GetBlendFunc()
	firstDyeIndex := uint(256 - len(self.re.dyeMappings))
	numUsedPaletteEntries := uint(self.font.Color().Count())
	var prevColor [4]float32
	var prevIndex uint = 9999
	
	// extremely slow and naive approach
	srcRect := mask.Bounds()
	dstRect := target.Bounds()
	for iy := srcRect.Min.Y; iy < srcRect.Max.Y; iy++ {
		for ix := srcRect.Min.X; ix < srcRect.Max.X; ix++ {
			// get color at the given point
			var clr [4]float32
			colorIndex := uint(mask.AlphaAt(ix, iy).A)
			if colorIndex == 0 { // reserved for transparent
				if self.re.blend != core.BlendReplace { continue }
				clr = [4]float32{0, 0, 0, 0}
			} else {
				if prevIndex == colorIndex {
					clr = prevColor // fast repeated case
				} else if colorIndex < firstDyeIndex { // static color
					baseIndex := (colorIndex << 2)
					clr = ([4]float32)(self.palette[baseIndex : baseIndex + 4])
				} else { // dye color
					dyeIndex := self.re.dyeMappings[colorIndex - firstDyeIndex]
					if dyeIndex == uint8(self.mainDyeKey) {
						clr = rgba
					} else {
						baseIndex := (uint(dyeIndex) << 2)
						copy(clr[:], self.palette[baseIndex : baseIndex + 4])
					}
	
					// apply dye alpha
					alphaIndex := colorIndex - (256 - numUsedPaletteEntries)
					alpha := self.palette[alphaIndex << 2]
					if alpha != 1.0 {
						clr[0] *= alpha
						clr[1] *= alpha
						clr[2] *= alpha
						clr[3] *= alpha
					}
				}
			}

			// memorize results for potential reuse in successive iterations
			if colorIndex != prevIndex {
				prevColor = clr
				prevIndex = colorIndex

				// clear previous blend temp variables
				self.re.prevTR, self.re.prevTG, self.re.prevTB, self.re.prevTA = 2, 2, 2, 1 // invalid (non-premult)
			}
			
			// fill area
			var tx, ty int
			switch orientation {
			case 0: tx, ty = self.getHorzDrawScaledPixTopLeft(x, y, ix, iy, scale)
			case 1: tx, ty = self.getVertDrawScaledPixTopLeft(x, y, ix, iy, scale)
			case 2: tx, ty = self.getSidewaysDrawScaledPixTopLeft(x, y, ix, iy, scale)
			case 3: tx, ty = self.getSidewaysRightDrawScaledPixTopLeft(x, y, ix, iy, scale)
			default:
				panic("invalid orientation")
			}
			for zy := 0; zy < scale; zy++ {
				if ty + zy <  dstRect.Min.Y { continue }
				if ty + zy >= dstRect.Max.Y { break }

				for zx := 0; zx < scale; zx++ {
					if tx + zx <  dstRect.Min.X { continue }
					if tx + zx >= dstRect.Max.X { break }

					fx, fy := tx + zx, ty + zy
					target.Set(fx, fy, blendFunc(target, fx, fy, clr))
				}
			}
		}
	}
}

func (self *Strand) drawHorzMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	self.drawMask(target, mask, x, y, scale, rgba, 0)
}

func (self *Strand) drawVertMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	self.drawMask(target, mask, x, y, scale, rgba, 1)
}

func (self *Strand) drawSidewaysMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	self.drawMask(target, mask, x, y, scale, rgba, 2)
}

func (self *Strand) drawSidewaysRightMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	self.drawMask(target, mask, x, y, scale, rgba, 3)
}

func (self *Strand) getHorzDrawScaledPixTopLeft(x, y, sx, sy int, scale int) (int, int) {
	return x + sx*scale, y + sy*scale
}

func (self *Strand) getVertDrawScaledPixTopLeft(x, y, sx, sy int, scale int) (int, int) {
	panic("unimplemented")
}

func (self *Strand) getSidewaysDrawScaledPixTopLeft(x, y, sx, sy int, scale int) (int, int) {
	return x + sy*scale, y - (sx + 1)*scale // TODO: I don't know where the +1 comes from
}

func (self *Strand) getSidewaysRightDrawScaledPixTopLeft(x, y, sx, sy int, scale int) (int, int) {
	return x - (sy + 1)*scale, y + sx*scale // TODO: I don't know where the +1 comes from
}

// ---- helper functions for blending and color operations ----

func blendOverFunc(new, curr [4]float32) color.RGBA {
	if new[3]  == 1.0 { return f32toRGBA(new)  }
	if new[3]  == 0.0 { return f32toRGBA(curr) }
	if curr[3] == 0.0 { return f32toRGBA(new)  }
	oma := 1.0 - new[3] // one minus alpha
	return color.RGBA {
		R: uint8((new[0] + curr[0]*oma)*255.0),
		G: uint8((new[1] + curr[1]*oma)*255.0),
		B: uint8((new[2] + curr[2]*oma)*255.0),
		A: uint8((new[3] + curr[3]*oma)*255.0),
	}
}

func cutFunc(new, curr [4]float32) color.RGBA {
	newAlpha := new[3]
	if newAlpha == 0 { return f32toRGBA(curr) }
	alpha := curr[3] - newAlpha
	if alpha < 0 { alpha = 0 }
	return color.RGBA {
		R: uint8(min(curr[0], alpha)*255.0),
		G: uint8(min(curr[1], alpha)*255.0),
		B: uint8(min(curr[2], alpha)*255.0),
		A: uint8(alpha*255.0),
	}
}

func hueFunc(new, curr [4]float32) color.RGBA {
	if new[3]  == 0 { return f32toRGBA(curr) }
	if curr[3] == 0 { return f32toRGBA(new) }

	// hue contribution is proportional to alpha.
	// if both alphas are equal, hue contributions are 50/50
	ta := new[3] + curr[3] // alpha sum (total)
	ma := max(new[3], curr[3]) // max alpha
	r := ((new[0] + curr[0])*ma)/ta
	g := ((new[1] + curr[1])*ma)/ta
	b := ((new[2] + curr[2])*ma)/ta
	return blendOverFunc([4]float32{r, g, b, ma}, curr)
}

func subFunc(new, curr [4]float32) color.RGBA {
	if new[3] == 0 { return f32toRGBA(curr) }
	return color.RGBA{
		R: uint8(max(curr[0] - new[0], 0)*255.0),
		G: uint8(max(curr[1] - new[1], 0)*255.0),
		B: uint8(max(curr[2] - new[2], 0)*255.0),
		A: uint8(curr[3]*255.0),
	}
}

func addFunc(new, curr [4]float32) color.RGBA {
	return color.RGBA{
		R: uint8(min(curr[0] + new[0], 1)*255.0),
		G: uint8(min(curr[1] + new[1], 1)*255.0),
		B: uint8(min(curr[2] + new[2], 1)*255.0),
		A: uint8(min(curr[3] + new[3], 1)*255.0),
	}
}

func multiplyFunc(new, curr [4]float32) color.RGBA {
	return color.RGBA{
		R: uint8(min(curr[0]*new[0], 1)*255.0),
		G: uint8(min(curr[1]*new[1], 1)*255.0),
		B: uint8(min(curr[2]*new[2], 1)*255.0),
		A: uint8(min(curr[3]*new[3], 1)*255.0),
	}
}

func f32toRGBA(rgba [4]float32) color.RGBA {
	return color.RGBA{
		R: uint8(rgba[0]*255.0),
		G: uint8(rgba[1]*255.0),
		B: uint8(rgba[2]*255.0),
		A: uint8(rgba[3]*255.0),
	}
}

func colorToFloat32(clr color.Color) [4]float32 {
	r, g, b, a := clr.RGBA()
	return rgba32ToFloat32(r, g, b, a)
}

func rgba32ToFloat32(r, g, b, a uint32) [4]float32 {
	return [4]float32{
		float32(r)/65535.0,
		float32(g)/65535.0,
		float32(b)/65535.0,
		float32(a)/65535.0,
	}
}

func rescaleAlphaRGBA(rgba color.RGBA, newAlpha uint8) color.RGBA {
	if rgba.A == newAlpha { return rgba }

	factor := float64(newAlpha)/float64(rgba.A)
	return color.RGBA{
		R: uint8(float64(rgba.R)*factor),
		G: uint8(float64(rgba.G)*factor),
		B: uint8(float64(rgba.B)*factor),
		A: newAlpha,
	}
}
