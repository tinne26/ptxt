package ptxt

import "github.com/tinne26/ptxt/internal"
import "github.com/tinne26/ptxt/core"
import "github.com/tinne26/ptxt/strand"

import "github.com/tinne26/ggfnt"

func (self *Renderer) drawText(target core.Target, x, y int) {
	switch self.direction {
	case Horizontal:
		self.drawTextHorz(target, x, y)
	case Vertical:
		panic("unimplemented")
	case Sideways:
		self.drawTextSideways(target, x, y)
	case SidewaysRight:
		self.drawTextSidewaysRight(target, x, y)
	default:
		panic("unexpected direction '" + self.direction.String() + "'")
	}
}

// --- horz ---

func (self *Renderer) drawTextHorz(target core.Target, ox, oy int) {	
	fontStrand := self.Strand()
	drawParams := self.prepareDrawParams(ox, oy)

	// draw shadow
	shadowStrand := fontStrand.Shadow().GetStrand()
	if shadowStrand != nil {
		var offsetX, offsetY int
		drawParams.RGBA, offsetX, offsetY = self.prepareShadowDraw(fontStrand)
		lnkSetBlendMode(shadowStrand, self.blendMode)
		if self.drawFunc != nil {
			self.runHorzIterate(target, drawParams, offsetX, offsetY, self.drawFunc)
		} else {
			self.runHorzIterate(target, drawParams, offsetX, offsetY,  
				func(target core.Target, glyphIndex ggfnt.GlyphIndex, params MaskDrawParameters) {
					mask := self.loadMask(glyphIndex, shadowStrand.Font())
					if mask != nil {
						lnkDrawHorzMask(shadowStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
					}
				},
			)
		}
	}

	// draw main text
	drawParams.RGBA = self.prepareMainDraw(fontStrand)
	if self.drawFunc != nil {
		self.runHorzIterate(target, drawParams, 0, 0, self.drawFunc)
	} else {
		lnkSetBlendMode(fontStrand, self.blendMode)
		self.runHorzIterate(target, drawParams, 0, 0,
			func(target core.Target, glyphIndex ggfnt.GlyphIndex, params MaskDrawParameters) {
				mask := self.loadMask(glyphIndex, fontStrand.Font())
				if mask != nil {
					lnkDrawHorzMask(fontStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
				}
			},
		)
	}
}

func (self *Renderer) runHorzIterate(target core.Target, maskDrawParams MaskDrawParameters, offsetX, offsetY int, drawFunc func(core.Target, ggfnt.GlyphIndex, MaskDrawParameters)) {
	// helper variables
	ox := maskDrawParams.X
	currentStrand := self.Strand()
	currentHorzInterspacing := strandFullHorzInterspacing(currentStrand)*maskDrawParams.Scale

	// set up wrap info
	var drawWrapTemps drawWrapTempVariables
	drawWrapTemps.Init(self)
	var lineBreakTemps lineBreakTempVariables
	lineBreakTemps.SetBreakHeight(strandFullVertLineHeight(currentStrand)*maskDrawParams.Scale)

	// iteration
	var x, y int = self.computeLineStart(ox, 0), maskDrawParams.Y
	for index := 0; index < len(self.run.glyphIndices); index++ {
		// line wrap case
		if drawWrapTemps.IsLineWrapIndex(index) {
			elide := drawWrapTemps.WrapTypeIsElide()
			x, y = lineBreakTemps.ApplyHorzBreak(self, ox, y)
			drawWrapTemps.Update(self)
			if elide { continue }
		}

		// general drawing
		glyphIndex := self.run.glyphIndices[index]
		if glyphIndex < ggfnt.MaxGlyphs {
			lineBreakTemps.NotifyNonBreak()
			x += int(self.run.kernings[index])
			maskDrawParams.X = x + offsetX
			maskDrawParams.Y = y + offsetY
			drawFunc(target, glyphIndex, maskDrawParams)
			x += int(self.run.advances[index]) + currentHorzInterspacing
		} else { // control glyph
			switch glyphIndex {
			case ggfnt.GlyphNewLine:
				if !self.elideLineBreak(index) {
					x, y = lineBreakTemps.ApplyHorzBreak(self, ox, y)
				}
			case ggfnt.GlyphMissing:
				// should typically be triggered at an earlier point,
				// so I'm not even sure you can reach this normally
				panic("missing glyph")
			case internal.TwineEffectMarkerGlyph:
				panic("unimplemented")
			default:
				// other control glyphs to be fully ignored
				// TODO: or do we need to report any of these control glyphs,
				// somewhow, to the end user? maybe. maybe we need a separate
				// callback for it, and we can check the custom glyph range
			}
		}
	}
}

// --- sideways ---

func (self *Renderer) drawTextSideways(target core.Target, ox, oy int) {	
	fontStrand := self.Strand()
	drawParams := self.prepareDrawParams(ox, oy)

	// draw shadow
	shadowStrand := fontStrand.Shadow().GetStrand()
	if shadowStrand != nil {
		var offsetX, offsetY int
		drawParams.RGBA, offsetX, offsetY = self.prepareShadowDraw(fontStrand)
		lnkSetBlendMode(shadowStrand, self.blendMode)
		if self.drawFunc != nil {
			self.runSidewaysIterate(target, drawParams, offsetX, offsetY, self.drawFunc)
		} else {
			self.runSidewaysIterate(target, drawParams, offsetX, offsetY,  
				func(target core.Target, glyphIndex ggfnt.GlyphIndex, params MaskDrawParameters) {
					mask := self.loadMask(glyphIndex, shadowStrand.Font())
					if mask != nil {
						lnkDrawSidewaysMask(shadowStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
					}
				},
			)
		}
	}

	// draw main text
	drawParams.RGBA = self.prepareMainDraw(fontStrand)
	if self.drawFunc != nil {
		self.runSidewaysIterate(target, drawParams, 0, 0, self.drawFunc)
	} else {
		lnkSetBlendMode(fontStrand, self.blendMode)
		self.runSidewaysIterate(target, drawParams, 0, 0,
			func(target core.Target, glyphIndex ggfnt.GlyphIndex, params MaskDrawParameters) {
				mask := self.loadMask(glyphIndex, fontStrand.Font())
				if mask != nil {
					lnkDrawSidewaysMask(fontStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
				}
			},
		)
	}
}

func (self *Renderer) runSidewaysIterate(target core.Target, maskDrawParams MaskDrawParameters, offsetX, offsetY int, drawFunc func(core.Target, ggfnt.GlyphIndex, MaskDrawParameters)) {
	// helper variables
	oy := maskDrawParams.Y
	currentStrand := self.Strand()
	currentHorzInterspacing := strandFullHorzInterspacing(currentStrand)*maskDrawParams.Scale

	// set up wrap info
	var drawWrapTemps drawWrapTempVariables
	drawWrapTemps.Init(self)
	var lineBreakTemps lineBreakTempVariables
	lineBreakTemps.SetBreakHeight(strandFullVertLineHeight(currentStrand)*maskDrawParams.Scale)

	// iteration
	lsDiff := self.computeLineStart(oy, 0) - oy
	var x, y int = maskDrawParams.X, oy - lsDiff
	for index := 0; index < len(self.run.glyphIndices); index++ {
		// line wrap case
		if drawWrapTemps.IsLineWrapIndex(index) {
			elide := drawWrapTemps.WrapTypeIsElide()
			x, y = lineBreakTemps.ApplySidewaysBreak(self, x, oy)
			drawWrapTemps.Update(self)
			if elide { continue }
		}

		// general drawing
		glyphIndex := self.run.glyphIndices[index]
		if glyphIndex < ggfnt.MaxGlyphs {
			lineBreakTemps.NotifyNonBreak()
			y -= int(self.run.kernings[index])
			maskDrawParams.X = x + offsetY
			maskDrawParams.Y = y - offsetX
			drawFunc(target, glyphIndex, maskDrawParams)
			y -= int(self.run.advances[index]) + currentHorzInterspacing
		} else { // control glyph
			switch glyphIndex {
			case ggfnt.GlyphNewLine:
				if !self.elideLineBreak(index) {
					x, y = lineBreakTemps.ApplySidewaysBreak(self, x, oy)
				}
			case ggfnt.GlyphMissing:
				// should typically be triggered at an earlier point,
				// so I'm not even sure you can reach this normally
				panic("missing glyph")
			case internal.TwineEffectMarkerGlyph:
				panic("unimplemented")
			default:
				// other control glyphs to be fully ignored
				// TODO: or do we need to report any of these control glyphs,
				// somewhow, to the end user? maybe. maybe we need a separate
				// callback for it, and we can check the custom glyph range
			}
		}
	}
}

// --- sideways right ---

func (self *Renderer) drawTextSidewaysRight(target core.Target, ox, oy int) {	
	fontStrand := self.Strand()
	drawParams := self.prepareDrawParams(ox, oy)

	// draw shadow
	shadowStrand := fontStrand.Shadow().GetStrand()
	if shadowStrand != nil {
		var offsetX, offsetY int
		drawParams.RGBA, offsetX, offsetY = self.prepareShadowDraw(fontStrand)
		lnkSetBlendMode(shadowStrand, self.blendMode)
		if self.drawFunc != nil {
			self.runSidewaysRightIterate(target, drawParams, offsetX, offsetY, self.drawFunc)
		} else {
			self.runSidewaysRightIterate(target, drawParams, offsetX, offsetY,  
				func(target core.Target, glyphIndex ggfnt.GlyphIndex, params MaskDrawParameters) {
					mask := self.loadMask(glyphIndex, shadowStrand.Font())
					if mask != nil {
						lnkDrawSidewaysRightMask(shadowStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
					}
				},
			)
		}
	}

	// draw main text
	drawParams.RGBA = self.prepareMainDraw(fontStrand)
	if self.drawFunc != nil {
		self.runSidewaysRightIterate(target, drawParams, 0, 0, self.drawFunc)
	} else {
		lnkSetBlendMode(fontStrand, self.blendMode)
		self.runSidewaysRightIterate(target, drawParams, 0, 0,
			func(target core.Target, glyphIndex ggfnt.GlyphIndex, params MaskDrawParameters) {
				mask := self.loadMask(glyphIndex, fontStrand.Font())
				if mask != nil {
					lnkDrawSidewaysRightMask(fontStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
				}
			},
		)
	}
}

func (self *Renderer) runSidewaysRightIterate(target core.Target, maskDrawParams MaskDrawParameters, offsetX, offsetY int, drawFunc func(core.Target, ggfnt.GlyphIndex, MaskDrawParameters)) {
	// helper variables
	oy := maskDrawParams.Y
	currentStrand := self.Strand()
	currentHorzInterspacing := strandFullHorzInterspacing(currentStrand)*maskDrawParams.Scale
	
	// set up wrap info
	var drawWrapTemps drawWrapTempVariables
	drawWrapTemps.Init(self)
	var lineBreakTemps lineBreakTempVariables
	lineBreakTemps.SetBreakHeight(strandFullVertLineHeight(currentStrand)*maskDrawParams.Scale)

	// iteration
	var x, y int = maskDrawParams.X, self.computeLineStart(oy, 0)
	for index := 0; index < len(self.run.glyphIndices); index++ {
		// line wrap case
		if drawWrapTemps.IsLineWrapIndex(index) {
			elide := drawWrapTemps.WrapTypeIsElide()
			x, y = lineBreakTemps.ApplySidewaysRightBreak(self, x, oy)
			drawWrapTemps.Update(self)
			if elide { continue }
		}

		// general drawing
		glyphIndex := self.run.glyphIndices[index]
		if glyphIndex < ggfnt.MaxGlyphs {
			lineBreakTemps.NotifyNonBreak()
			y += int(self.run.kernings[index])
			maskDrawParams.X = x - offsetY
			maskDrawParams.Y = y + offsetX
			drawFunc(target, glyphIndex, maskDrawParams)
			y += int(self.run.advances[index]) + currentHorzInterspacing
		} else { // control glyph
			switch glyphIndex {
			case ggfnt.GlyphNewLine:
				if !self.elideLineBreak(index) {
					x, y = lineBreakTemps.ApplySidewaysRightBreak(self, x, oy)
				}
			case ggfnt.GlyphMissing:
				// should typically be triggered at an earlier point,
				// so I'm not even sure you can reach this normally
				panic("missing glyph")
			case internal.TwineEffectMarkerGlyph:
				panic("unimplemented")
			default:
				// other control glyphs to be fully ignored
				// TODO: or do we need to report any of these control glyphs,
				// somewhow, to the end user? maybe. maybe we need a separate
				// callback for it, and we can check the custom glyph range
			}
		}
	}
}

// ---- common helpers ----

func (self *Renderer) prepareDrawParams(ox, oy int) MaskDrawParameters {
	return MaskDrawParameters{
		X: ox,
		Y: oy,
		Scale: int(self.scale),
	}
}

// Returns the shadow color and offsets.
func (self *Renderer) prepareShadowDraw(fontStrand *strand.Strand) ([4]float32, int, int) {
	if self.drawPassListener != nil {
		self.drawPassListener(self, ShadowDrawPass)
	}
	rgba := internal.RGBAToFloat32(fontStrand.Shadow().GetColor())
	var offsetX8, offsetY8 int8 = fontStrand.Shadow().GetOffsets()
	var offsetX, offsetY int = int(offsetX8), int(offsetY8)
	if fontStrand.Shadow().IsOffsetScalingEnabled() {	
		offsetX *= int(self.scale)
		offsetY *= int(self.scale)
	}
	return rgba, offsetX, offsetY
}

func (self *Renderer) prepareMainDraw(strand *strand.Strand) [4]float32 {
	if self.drawPassListener != nil {
		self.drawPassListener(self, MainDrawPass)
	}
	if strand.IsMainDyeActive() {
		return strand.GetDye(strand.MainDyeKey())
	} else {
		return internal.RGBAToFloat32(self.fallbackMainDye)
	}
}
