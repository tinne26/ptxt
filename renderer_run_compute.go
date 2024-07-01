package ptxt

import "image"

import "github.com/tinne26/ptxt/internal"
import "github.com/tinne26/ggfnt"

// Helper methods for drawing and measuring.

// Main axis centering is not accounted for, that's applied directly
// while drawing.
//
// The workingAscent and workingDescent parameters are necessary
// for those cases where we are using MaskBounding. These values
// will be obtained either from computeRunAdvances() or 
func (self *Renderer) computeTextOrigin(x, y int) (int, int) {
	const unexpectedAlign = "unexpected Renderer align value"

	strand := self.Strand()
	scale  := int(self.scale)
	font   := strand.Font()

	var shift int
	switch self.align.Vert() {
	case Top:
		shift = -self.run.top
	case CapLine:
		shift = int(font.Metrics().UppercaseAscent())*scale
		if shift == 0 {
			// if uppercase ascent is undefined, we use this heuristic.
			// this is of course very debatable, but I'd say that in 
			// practice this tends to be more helpful than nothing
			shift = -self.run.top
		}
	case Midline:
		shift = int(font.Metrics().MidlineAscent())*scale
		if shift == 0 {
			// if lowercase ascent is undefined, we use this heuristic.
			// this is of course very debatable, but I'd say that in 
			// practice this tends to be more helpful than nothing
			shift = (-self.run.top >> 1)
		}
	case VertCenter:
		shift = -self.run.top - ((self.run.bottom - self.run.top) >> 1)
	case Baseline:
		shift = 0
	case Bottom: 
		shift = -self.run.bottom
	case LastBaseline:
		if self.run.isMultiline { // multi-line text
			shift = -(self.run.bottom + self.run.lastLineDescent)
		} else { // single line text, last baseline == first baseline
			shift = 0
		}
	default:
		panic(unexpectedAlign)
	}

	switch self.direction {
	case Horizontal    : return x, y + shift
	case Vertical      : panic("unimplemented")
	case Sideways      : return x + shift, y
	case SidewaysRight : return x - shift, y
	default:
		panic("unexpected direction '" + self.direction.String() + "'")
	}
}

// - I can have a uint16 to mark "twine effect", and then a
//   second uint16 as the offset to it. but then maybe calling
//   the buffer "run.glyphIndices" is not ideal. otherwise, I
//   need to port everything to a twine model first, and operate
//   only with the twine. I don't like that because even in the
//   twine, when ligature replacements are needed, we can't
//   modify the twine itself. so, we kinda always need that
//   first conversion step. ok I guess.

// Spaces at the end of line are counted. They can be elided on wrap modes,
// but not here. TODO: separate vertical text direction handling.
func (self *Renderer) computeRunLayout(maxLineLen int) {
	// resize advances buffer, clear metrics
	self.run.isMultiline = false
	self.run.advances = setBufferSize(self.run.advances, len(self.run.glyphIndices))
	self.run.kernings = setBufferSize(self.run.kernings, len(self.run.glyphIndices))
	self.run.wrapIndices = self.run.wrapIndices[ : 0]
	self.run.lineLengths = self.run.lineLengths[ : 0]
	self.run.top, self.run.bottom, self.run.left, self.run.right = 0, 0, 0, 0
	self.run.firstLineAscent = 0
	self.run.lastLineDescent = 0
	if self.boundingMode == 0 { self.boundingMode = LogicalBounding } // default init
	if len(self.run.glyphIndices) == 0 { return } // trivial case
	switch self.boundingMode & ^noDescent {
	case LogicalBounding:
		self.computeRunLogicalLayout(maxLineLen)
	case MaskBounding:
		self.computeRunMaskLayout(maxLineLen)
	default:
		panic("unexpected bounding mode")
	}
}

// Precondition: except glyph indices, run data has been cleared 
// and slices resized to len(glyphIndices), and we have at least
// one glyph index.
func (self *Renderer) computeRunLogicalLayout(maxLineLen int) {
	if len(self.run.glyphIndices) > 32000 { panic(preViolation) }

	currentStrand := self.Strand()
	currentFont   := currentStrand.Font()
	currentScale  := int(self.scale)
	currentHorzInterspacing := strandFullHorzInterspacing(currentStrand)*currentScale
	var layoutBreak layoutLineBreakTempVariables
	layoutBreak.Init(strandFullVertLineHeight(currentStrand)*currentScale)
	self.run.firstLineAscent = int(currentFont.Metrics().Ascent())*currentScale
	self.run.top = -self.run.firstLineAscent
	if self.boundingMode & noDescent == 0 {
		self.run.lastLineDescent = int(currentFont.Metrics().Descent())*currentScale
	}

	var prevEffectiveGlyph ggfnt.GlyphIndex = ggfnt.GlyphMissing
	var prevInterspacing int
	var x, index int
	var layoutWrap layoutWrapTempVariables
	for index < len(self.run.glyphIndices) {
		glyphIndex := self.run.glyphIndices[index]
		if glyphIndex < ggfnt.MaxGlyphs {
			layoutBreak.NotifyNonBreak()
			layoutWrap.IncreaseLineCharCount()

			memoX := x
			kerning := int(currentFont.Kerning().Get(prevEffectiveGlyph, glyphIndex))*currentScale
			self.run.kernings[index] = int16(kerning)
			advance := int(currentFont.Glyphs().Advance(glyphIndex))*currentScale
			if advance < 0 || advance > 65535 { panic("advance > 65535") } // discretional assertion
			self.run.advances[index] = uint16(advance)
			x += prevInterspacing + kerning + advance
			prevInterspacing = currentHorzInterspacing
			prevEffectiveGlyph = glyphIndex

			// line wrapping pain
			if x <= maxLineLen {
				layoutWrap.GlyphNonBreak(currentStrand, glyphIndex, index, memoX, x)
			} else {
				index, x = layoutWrap.GlyphBreak(self, currentStrand, glyphIndex, index, memoX, x)
				self.run.bottom = layoutBreak.NotifyBreak(self, 0, x, self.run.bottom)
				x, prevInterspacing = 0, 0
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
					self.run.bottom = layoutBreak.NotifyBreak(self, 0, x, self.run.bottom)
					x, prevInterspacing = 0, 0
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
			}
		}
		index += 1
	}

	// take last x and descent into account
	layoutBreak.NotifyTextEnd(self, x)
}

// Precondition: except glyph indices, run data has been cleared 
// and slices resized to len(glyphIndices), and we have at least
// one glyph index.
func (self *Renderer) computeRunMaskLayout(maxLineLen int) {
	if len(self.run.glyphIndices) > 32000 { panic(preViolation) }

	self.run.bottom = -9999
	self.run.left   = +9999

	currentStrand := self.Strand()
	currentFont   := currentStrand.Font()
	currentScale  := int(self.scale)
	currentHorzInterspacing := strandFullHorzInterspacing(currentStrand)*currentScale
	var layoutBreak layoutLineBreakTempVariables
	layoutBreak.Init(strandFullVertLineHeight(currentStrand)*currentScale)

	var prevEffectiveGlyph ggfnt.GlyphIndex = ggfnt.GlyphMissing
	var prevInterspacing, prevMaskRight, maskLeft int = 0, -9999, +9999
	var x, y, index int
	var firstNonEmptyLine uint8 // 0 for first non empty not reached, 1 for reached, 2 for exceeded
	var layoutWrap layoutWrapTempVariables
	for index < len(self.run.glyphIndices) {
		glyphIndex := self.run.glyphIndices[index]
		if glyphIndex < ggfnt.MaxGlyphs {
			layoutBreak.NotifyNonBreak()
			layoutWrap.IncreaseLineCharCount()

			// get glyph mask for bounds and adjust run bounds
			mask := self.loadMask(glyphIndex, currentFont)
			var bounds image.Rectangle
			if mask != nil {
				bounds = mask.Bounds()
				if x + bounds.Min.X*currentScale < maskLeft {
					maskLeft = x + bounds.Min.X*currentScale
				}
			}

			kerning := int(currentFont.Kerning().Get(prevEffectiveGlyph, glyphIndex))*currentScale
			self.run.kernings[index] = int16(kerning)
			advance := int(currentFont.Glyphs().Advance(glyphIndex))*currentScale
			if advance < 0 || advance > 65535 { panic("advance > 65535") } // discretional assertion
			self.run.advances[index] = uint16(advance)
			maskRight := x + bounds.Max.X*currentScale + prevInterspacing + kerning
			
			// line wrapping pain
			if x <= maxLineLen {
				layoutWrap.GlyphNonBreak(currentStrand, glyphIndex, index, prevMaskRight, maskRight)
				if mask != nil {
					if firstNonEmptyLine == 0 {
						firstNonEmptyLine = 1
						self.run.top = y // this is not correct here yet, but it's corrected with first line ascent later
					}
					if firstNonEmptyLine == 1 {
						self.run.firstLineAscent = max(self.run.firstLineAscent, -bounds.Min.Y*currentScale)
					}
					if self.boundingMode & noDescent == 0 { // descent case
						self.run.lastLineDescent = max(self.run.lastLineDescent, bounds.Max.Y*currentScale)
					}
					elevation := min(bounds.Max.Y*currentScale, 0)
					self.run.bottom = max(self.run.bottom, y + elevation, y + elevation + self.run.lastLineDescent)
					
					prevMaskRight = maskRight // this must stay inside this if
				}
				x += prevInterspacing + kerning + advance
				prevInterspacing = currentHorzInterspacing
				prevEffectiveGlyph = glyphIndex
			} else {
				index, x = layoutWrap.GlyphBreak(self, currentStrand, glyphIndex, index, prevMaskRight, maskRight)
				y = layoutBreak.NotifyBreak(self, maskLeft, x, y)
				self.run.lastLineDescent = 0
				x, prevInterspacing, prevMaskRight, maskLeft = 0, 0, -9999, 9999
				prevEffectiveGlyph = ggfnt.GlyphMissing
				if firstNonEmptyLine == 1 { firstNonEmptyLine = 2 }
				continue
			}
		} else { // control glyph
			switch glyphIndex {
			case ggfnt.GlyphNewLine:
				if self.elideLineBreak(index) {
					// line break should be elided, absorbed by immediately previous line wrapping break
				} else {
					// apply break
					y = layoutBreak.NotifyBreak(self, maskLeft, prevMaskRight, y)
					layoutWrap.PostBreakUpdate(index + 1)
					self.run.lastLineDescent = 0
					x, prevInterspacing, prevMaskRight, maskLeft = 0, 0, -9999, 9999
					prevEffectiveGlyph = ggfnt.GlyphMissing
					if firstNonEmptyLine == 1 { firstNonEmptyLine = 2 }
				}
			case ggfnt.GlyphMissing:
				// should typically be triggered at an earlier point,
				// so I'm not even sure you can reach this normally
				panic("missing glyph")
			case ggfnt.GlyphZilch: // we don't change the prevEffectiveGlyph here
				self.run.advances[index] = 0
				self.run.kernings[index] = 0 
			case internal.TwineEffectMarkerGlyph:
				panic("unimplemented")
			default:
				self.run.advances[index] = 0
				self.run.kernings[index] = 0
			}
		}
		index += 1
	}

	// final adjustments
	self.run.top -= self.run.firstLineAscent
	self.run.left = min(self.run.left, maskLeft)
	if prevMaskRight > self.run.right { self.run.right = prevMaskRight }
	lineLen := self.run.right - self.run.left
	self.run.lineLengths = append(self.run.lineLengths, uint16(max(0, lineLen))) // TODO: big hack max
	
	self.run.bottom = max(self.run.bottom, self.run.top)
	self.run.left   = min(self.run.right, self.run.left)
}

func (self *Renderer) computeRunAdvancesWithWrap(maxLineLen int) (width, height int) {
	panic("unimplemented")
}

func (self *Renderer) computeLineStart(o int, lineIndex uint16) int {
	switch self.align.Horz() {
	case Left       : return o
	case HorzCenter : return o - int(self.run.lineLengths[lineIndex] >> 1)
	case Right      : return o - int(self.run.lineLengths[lineIndex])
	default:
		panic(brokenCode)
	}
}

func (self *Renderer) elideLineBreak(index int) bool {
	wrapsLen := len(self.run.wrapIndices)
	return wrapsLen > 0 && (self.run.wrapIndices[wrapsLen - 1] & 0x7FFF) + 1 == uint16(index)
}

// Basically, if we are on the second line break and par break is enabled,
// we either return half the height or none at all.
func (self *Renderer) adjustParLineBreakHeightFor(lineBreakHeight, consecutiveLineBreaks int) int {
	// no par break case
	if !self.parBreakEnabled { return lineBreakHeight }

	// par break case
	switch consecutiveLineBreaks {
	case 2: return (lineBreakHeight >> 1)
	case 3: return lineBreakHeight - (lineBreakHeight >> 1) // (complete prev half break)
	default:
		return lineBreakHeight
	}
}
