package ptxt

import "github.com/tinne26/ggfnt"
import "github.com/tinne26/ptxt/internal"

// Precondition: except glyph indices, run data has been cleared 
// and slices resized to len(glyphIndices), and we have at least
// one glyph index.
func (self *Renderer) computeVerticalRunLogicalLayout(maxLineLen int) {
	if len(self.run.glyphIndices) > 32000 { panic(preViolation) }

	currentStrand := self.Strand()
	currentFont   := currentStrand.Font()
	currentScale  := int(self.scale)
	currentGlyphInterspacing := strandFullVertGlyphInterspacing(currentStrand)*currentScale
	var layoutBreak vertLayoutLineBreakTempVariables
	layoutBreak.Init(strandFullLineWidth(currentStrand)*currentScale)
	self.run.firstRowAscent = int(currentFont.Metrics().Ascent())*currentScale
	self.run.top = -self.run.firstRowAscent
	if self.boundingMode & noDescent == 0 {
		self.run.lastRowDescent = int(currentFont.Metrics().Descent())*currentScale
	}

	var prevEffectiveGlyph ggfnt.GlyphIndex = ggfnt.GlyphMissing
	var prevInterspacing int
	var prevBottomAdvance, prevTopAdvance int
	var y, index int = self.run.top, 0
	var layoutWrap vertLayoutWrapTempVariables
	for index < len(self.run.glyphIndices) {
		glyphIndex := self.run.glyphIndices[index]
		if glyphIndex < ggfnt.MaxGlyphs {
			layoutBreak.NotifyNonBreak()
			layoutWrap.IncreaseLineCharCount()
			memoY := y + prevBottomAdvance
			kerning := int(currentFont.Kerning().GetVert(prevEffectiveGlyph, glyphIndex))*currentScale
			self.run.kernings[index] = int16(kerning)
			
			self.run.advances[index] = uint16(prevBottomAdvance + prevTopAdvance)
			placement := currentFont.Glyphs().Placement(glyphIndex)
			horzShift := int(placement.HorzCenter)*currentScale
			self.run.horzShifts[index] = uint16(horzShift)
			if self.run.right - horzShift < self.run.left {
				self.run.left = self.run.right - horzShift
			}
			prevTopAdvance = int(placement.TopAdvance)*currentScale
			y += prevBottomAdvance + prevInterspacing + kerning + prevTopAdvance
			prevInterspacing = currentGlyphInterspacing
			prevBottomAdvance = int(placement.BottomAdvance)*currentScale
			prevEffectiveGlyph = glyphIndex

			// line wrapping pain
			yWrap := y + prevBottomAdvance
			if yWrap <= maxLineLen {
				layoutBreak.UpdateMaxRightAdvance(int(placement.Advance)*currentScale - horzShift)
				layoutWrap.GlyphNonBreak(currentStrand, glyphIndex, index, memoY, yWrap)
			} else {
				//var yWithBottom int
				index, _ = layoutWrap.GlyphBreak(self, currentStrand, glyphIndex, index, memoY, yWrap)
				self.run.right = layoutBreak.NotifyBreak(self, self.run.top, memoY, self.run.right)
				y, prevInterspacing, prevBottomAdvance, prevTopAdvance = 0, 0, 0, 0
				prevEffectiveGlyph = ggfnt.GlyphMissing
				continue
			}
		} else { // control glyph
			switch glyphIndex {
			case ggfnt.GlyphNewLine:
				if self.elideLineBreak(index) {
					// line break should be elided, absorbed by immediately previous line wrapping break
				} else {
					// apply break
					self.run.right = layoutBreak.NotifyBreak(self, self.run.top, y, self.run.right)
					y, prevInterspacing, prevBottomAdvance, prevTopAdvance = 0, 0, 0, 0
					prevEffectiveGlyph = ggfnt.GlyphMissing
					layoutWrap.PostBreakUpdate(index + 1)
				}
			case ggfnt.GlyphMissing:
				// should typically be triggered at an earlier point,
				// so I'm not even sure you can reach this normally
				panic("missing glyph")
			case ggfnt.GlyphZilch: // we don't change the prevEffectiveGlyph here
				self.run.advances[index] = 0
				self.run.kernings[index] = 0 
				self.run.horzShifts[index] = 0
			case internal.TwineEffectMarkerGlyph:
				panic("unimplemented")
				// lineBreaksOnly = false // yes or no? it depends?
				// prevEffectiveGlyph = glyphIndex
				// TODO: whether this breaks or doesn't break prev effective glyph
				// should probably depend on the effect. color changes shouldn't break
				// anything. other effects might break. basically, effects with padding
				// should break kerning, but maybe we want to leave it to the user,
				// as some cases might be ambiguous
			default:
				// ... some other control glyph, possibly a custom control glyph
				// for the font or user code. we are not breaking kerning nor
				// interrupting the previous glyph, but that could be discussed (TODO)
				self.run.advances[index] = 0
				self.run.kernings[index] = 0
				self.run.horzShifts[index] = 0
			}
		}
		index += 1
	}

	// take last x and descent into account
	layoutBreak.NotifyTextEnd(self, y + prevBottomAdvance)
}

func (self *Renderer) computeVerticalRunMaskLayout(maxLineLen int) {
	panic("sorry, Vertical rendering only supports LogicalBounding at the moment, MaskBounding is still unimplemented")
}

func (self *Renderer) computeVertLineStart(oy int, lineIndex uint16) int {
	switch self.align.Vert() {
	case VertCenter : return oy - int(self.run.lineLengths[lineIndex] >> 1)
	case Bottom     : return oy - int(self.run.lineLengths[lineIndex])
	default:
		return oy
	}
}
