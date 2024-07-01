package ptxt

import "github.com/tinne26/ggfnt"
import "github.com/tinne26/ptxt/strand"

// --- layout line wrap ---
type layoutWrapTempVariables struct {
	wrapPointFoundInCurrentLine bool
	lineCharCount int
	lastWrapSafeWidth int
	lastWrapSafeIndex int
	lastWrapType strand.WrapMode
}

func (self *layoutWrapTempVariables) IncreaseLineCharCount() {
	self.lineCharCount += 1
}

func (self *layoutWrapTempVariables) GlyphNonBreak(str *strand.Strand, glyphIndex ggfnt.GlyphIndex, index, preX, postX int) {
	if str.CanWrap(glyphIndex, strand.WrapAfter) {
		self.lastWrapSafeIndex = index + 1
		self.lastWrapType = strand.WrapAfter
		self.wrapPointFoundInCurrentLine = true
		self.lastWrapSafeWidth = postX
	} else if str.CanWrap(glyphIndex, strand.WrapElide) {
		self.lastWrapSafeIndex = index
		self.lastWrapType = strand.WrapElide
		self.wrapPointFoundInCurrentLine = true
		self.lastWrapSafeWidth = preX
	} else if str.CanWrap(glyphIndex, strand.WrapBefore) {
		self.lastWrapSafeIndex = index
		self.lastWrapType = strand.WrapBefore
		self.wrapPointFoundInCurrentLine = true
		self.lastWrapSafeWidth = preX
	}
}

// Returns the new index and x to continue on.
func (self *layoutWrapTempVariables) GlyphBreak(renderer *Renderer, str *strand.Strand, glyphIndex ggfnt.GlyphIndex, index, preX, postX int) (newIndex, newX int) {
	if self.lineCharCount > 1 && str.CanWrap(glyphIndex, strand.WrapElide) { // easy case
		// NOTE: we could check lineCharCount == 1 here and if previous line
		//       had only one char or didn't have a safe wrap point, force
		//       elidable char to be visible... but it's a lot of pain for
		//       little gain on cases where everything went to **** already
		renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index) | 0x8000)
		index += 1
		postX = preX
	} else if self.lineCharCount > 1 && str.CanWrap(glyphIndex, strand.WrapBefore) {
		renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index))
		// index already correct
	} else if self.wrapPointFoundInCurrentLine {
		wrapIndex := uint16(self.lastWrapSafeIndex)
		if self.lastWrapType == strand.WrapElide {
			wrapIndex |= 0x8000
			self.lastWrapSafeIndex += 1
		}
		postX = self.lastWrapSafeWidth
		renderer.run.wrapIndices = append(renderer.run.wrapIndices, wrapIndex)
		index = self.lastWrapSafeIndex
	} else { // take as much of the first word as we can (or at least one char)
		// we have to discount the x increase unless we are on the first char anyways
		if self.lineCharCount == 1 {
			index += 1
			renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index))
		} else {
			postX = preX
			renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index))
		}
	}

	// update wrap variables
	self.PostBreakUpdate(index)
	return index, postX
}

func (self *layoutWrapTempVariables) PostBreakUpdate(index int) {
	self.wrapPointFoundInCurrentLine = false
	self.lastWrapSafeIndex = index
	self.lastWrapSafeWidth = 0
	self.lastWrapType = strand.WrapBefore
	self.lineCharCount = 0
}

// --- layout line break ---

type layoutLineBreakTempVariables struct {
	currentLineBreakHeight int
	consecutiveLineBreaks int
	lineBreaksOnly bool
}

func (self *layoutLineBreakTempVariables) Init(height int) {
	self.currentLineBreakHeight = height
	self.lineBreaksOnly = true
}

func (self *layoutLineBreakTempVariables) NotifyNonBreak() {
	self.lineBreaksOnly = false
	self.consecutiveLineBreaks = 0
}

// Side effects: updates renderer.run.isMultiline, renderer.run.lineLengths,
//               renderer.run.right. You have to update renderer.run.bottom
//               with the returned value manually if necessary.
func (self *layoutLineBreakTempVariables) NotifyBreak(renderer *Renderer, left, right, y int) int {
	// line break update
	if !renderer.run.isMultiline && (!self.lineBreaksOnly || len(renderer.run.glyphIndices) > 1) {
		renderer.run.isMultiline = true // NOTE: control glyphs might make ^ incorrect for twines
	}
	self.consecutiveLineBreaks += 1
	lineLen := right - left
	renderer.run.lineLengths = append(renderer.run.lineLengths, uint16(max(0, lineLen))) // the min is a big hack
	if right > renderer.run.right { renderer.run.right = right }
	if left  < renderer.run.left  { renderer.run.left  = left  }
	return y + renderer.adjustParLineBreakHeightFor(self.currentLineBreakHeight, self.consecutiveLineBreaks)
}

// Preconditions: renderer.run.firstLineAscent and renderer.run.LastLineDescent are set
// Side effects: updates renderer.run.lineLengths, renderer.run.right and renderer.run.bottom
func (self *layoutLineBreakTempVariables) NotifyTextEnd(renderer *Renderer, x int) {
	if x > renderer.run.right { renderer.run.right = x }
	renderer.run.lineLengths = append(renderer.run.lineLengths, uint16(x))
	if self.lineBreaksOnly {
		renderer.run.bottom -= renderer.run.firstLineAscent
	} else {
		renderer.run.bottom += renderer.run.lastLineDescent
	}
}

// --- draw line break ---

type lineBreakTempVariables struct {
	lineBreakHeight int
	consecutiveLineBreaks uint16
	lineIndex uint16
}

func (self *lineBreakTempVariables) SetBreakHeight(height int) {
	self.lineBreakHeight = height
}

func (self *lineBreakTempVariables) NotifyNonBreak() {
	self.consecutiveLineBreaks = 0
}

func (self *lineBreakTempVariables) ApplyHorzBreak(renderer *Renderer, ox, y int) (newX, newY int) {
	self.lineIndex += 1
	self.consecutiveLineBreaks += 1
	x := renderer.computeLineStart(ox, self.lineIndex)
	y += self.getLineBreakHeight(renderer)
	return x, y
}

func (self *lineBreakTempVariables) ApplySidewaysBreak(renderer *Renderer, x, oy int) (int, int) {
	self.lineIndex += 1
	self.consecutiveLineBreaks += 1
	y := oy - (renderer.computeLineStart(oy, self.lineIndex) - oy)
	x += self.getLineBreakHeight(renderer)
	return x, y
}

func (self *lineBreakTempVariables) ApplySidewaysRightBreak(renderer *Renderer, x, oy int) (int, int) {
	self.lineIndex += 1
	self.consecutiveLineBreaks += 1
	y := renderer.computeLineStart(oy, self.lineIndex)
	x -= self.getLineBreakHeight(renderer)
	return x, y
}

func (self *lineBreakTempVariables) getLineBreakHeight(renderer *Renderer) int {
	if !renderer.parBreakEnabled { return self.lineBreakHeight }
	switch self.consecutiveLineBreaks {
	case 2: return (self.lineBreakHeight >> 1)
	case 3: return (self.lineBreakHeight - (self.lineBreakHeight >> 1)) // (complete prev half break)
	default:
		return self.lineBreakHeight
	}
}

// --- draw line wrap ---
type drawWrapTempVariables struct {
	nextWrapIndex uint16
	nextSliceIndex uint16
	nextWrapType strand.WrapMode
}

func (self *drawWrapTempVariables) Init(renderer *Renderer) {
	if len(renderer.run.wrapIndices) == 0 {
		self.nextWrapIndex = 65535
	} else {
		nextWrapIndex := renderer.run.wrapIndices[self.nextSliceIndex]
		self.nextWrapType = strand.WrapBefore
		if (nextWrapIndex & 0x8000) != 0 {
			self.nextWrapType = strand.WrapElide
		}
		self.nextWrapIndex = nextWrapIndex & 0x7FFF
		self.nextSliceIndex += 1
	}
}

func (self *drawWrapTempVariables) Update(renderer *Renderer) {
	if uint16(len(renderer.run.wrapIndices)) <= self.nextSliceIndex {
		self.nextWrapIndex = 65535
	} else {
		nextWrapIndex := renderer.run.wrapIndices[self.nextSliceIndex]
		self.nextWrapType = strand.WrapBefore
		if (nextWrapIndex & 0x8000) != 0 {
			self.nextWrapType = strand.WrapElide
		}
		self.nextWrapIndex = nextWrapIndex & 0x7FFF
		self.nextSliceIndex += 1
	}
}
func (self *drawWrapTempVariables) IsLineWrapIndex(index int) bool {
	return uint16(index) == self.nextWrapIndex
}
func (self *drawWrapTempVariables) WrapTypeIsElide() bool {
	return self.nextWrapType == strand.WrapElide
}
