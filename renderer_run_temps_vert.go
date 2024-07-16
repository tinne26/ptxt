package ptxt

import "github.com/tinne26/ggfnt"
import "github.com/tinne26/ptxt/strand"

// --- layout line wrap ---

type vertLayoutWrapTempVariables struct {
	wrapPointFoundInCurrentLine bool
	lineCharCount int
	lastWrapSafeHeight int
	lastWrapSafeIndex int
	lastWrapType strand.WrapMode
}

func (self *vertLayoutWrapTempVariables) IncreaseLineCharCount() {
	self.lineCharCount += 1
}

// Note: pre and post Y's must have the bottom advance included already
func (self *vertLayoutWrapTempVariables) GlyphNonBreak(str *strand.Strand, glyphIndex ggfnt.GlyphIndex, index, preY, postY int) {
	if str.CanWrap(glyphIndex, strand.WrapAfter) {
		self.lastWrapSafeIndex = index + 1
		self.lastWrapType = strand.WrapAfter
		self.wrapPointFoundInCurrentLine = true
		self.lastWrapSafeHeight = postY
	} else if str.CanWrap(glyphIndex, strand.WrapElide) {
		self.lastWrapSafeIndex = index
		self.lastWrapType = strand.WrapElide
		self.wrapPointFoundInCurrentLine = true
		self.lastWrapSafeHeight = preY
	} else if str.CanWrap(glyphIndex, strand.WrapBefore) {
		self.lastWrapSafeIndex = index
		self.lastWrapType = strand.WrapBefore
		self.wrapPointFoundInCurrentLine = true
		self.lastWrapSafeHeight = preY
	}
}

// Returns the new index and y to continue on.
func (self *vertLayoutWrapTempVariables) GlyphBreak(renderer *Renderer, str *strand.Strand, glyphIndex ggfnt.GlyphIndex, index, preY, postY int) (newIndex, newY int) {
	if self.lineCharCount > 1 && str.CanWrap(glyphIndex, strand.WrapElide) { // easy case
		// NOTE: we could check lineCharCount == 1 here and if previous line
		//       had only one char or didn't have a safe wrap point, force
		//       elidable char to be visible... but it's a lot of pain for
		//       little gain on cases where everything went to **** already
		renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index) | 0x8000)
		index += 1
		postY = preY
	} else if self.lineCharCount > 1 && str.CanWrap(glyphIndex, strand.WrapBefore) {
		renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index))
		// index already correct
	} else if self.wrapPointFoundInCurrentLine {
		wrapIndex := uint16(self.lastWrapSafeIndex)
		if self.lastWrapType == strand.WrapElide {
			wrapIndex |= 0x8000
			self.lastWrapSafeIndex += 1
		}
		postY = self.lastWrapSafeHeight
		renderer.run.wrapIndices = append(renderer.run.wrapIndices, wrapIndex)
		index = self.lastWrapSafeIndex
	} else { // take as much of the first word as we can (or at least one char)
		// we have to discount the x increase unless we are on the first char anyways
		if self.lineCharCount == 1 {
			index += 1
			renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index))
		} else {
			postY = preY
			renderer.run.wrapIndices = append(renderer.run.wrapIndices, uint16(index))
		}
	}

	// update wrap variables
	self.PostBreakUpdate(index)
	return index, postY
}

func (self *vertLayoutWrapTempVariables) PostBreakUpdate(index int) {
	self.wrapPointFoundInCurrentLine = false
	self.lastWrapSafeIndex = index
	self.lastWrapSafeHeight = 0
	self.lastWrapType = strand.WrapBefore
	self.lineCharCount = 0
}

// --- layout line break ---

type vertLayoutLineBreakTempVariables struct {
	currentLineBreakWidth int
	consecutiveLineBreaks int
	maxLineAdvance int
	lineBreaksOnly bool
}

func (self *vertLayoutLineBreakTempVariables) Init(width int) {
	self.currentLineBreakWidth = width
	self.lineBreaksOnly = true
}

func (self *vertLayoutLineBreakTempVariables) NotifyNonBreak() {
	self.lineBreaksOnly = false
	self.consecutiveLineBreaks = 0
}

func (self *vertLayoutLineBreakTempVariables) UpdateMaxRightAdvance(advance int) {
	self.maxLineAdvance = max(self.maxLineAdvance, advance)
}

// Side effects: updates renderer.run.isMultiline, renderer.run.lineLengths,
//               renderer.run.bottom. You have to update renderer.run.right
//               with the returned value manually if necessary.
func (self *vertLayoutLineBreakTempVariables) NotifyBreak(renderer *Renderer, top, bottom, x int) int {
	// line break update
	if !renderer.run.isMultiline && (!self.lineBreaksOnly || len(renderer.run.glyphIndices) > 1) {
		renderer.run.isMultiline = true // NOTE: control glyphs might make ^ incorrect for twines
	}
	self.consecutiveLineBreaks += 1
	lineLen := bottom - top
	renderer.run.lineLengths = append(renderer.run.lineLengths, uint16(max(0, lineLen))) // the min is a big hack
	if bottom > renderer.run.bottom { renderer.run.bottom = bottom }
	if top    < renderer.run.top    { renderer.run.top    = top    }
	self.maxLineAdvance = 0
	return x + renderer.adjustParLineBreakHeightFor(self.currentLineBreakWidth, self.consecutiveLineBreaks)
}

// Side effects: updates renderer.run.lineLengths, renderer.run.bottom and renderer.run.right
func (self *vertLayoutLineBreakTempVariables) NotifyTextEnd(renderer *Renderer, y int) {
	if y > renderer.run.bottom { renderer.run.bottom = y }
	renderer.run.lineLengths = append(renderer.run.lineLengths, uint16(y))
	renderer.run.right += self.maxLineAdvance
	if self.lineBreaksOnly {
		//renderer.run.right -= renderer.run.firstRowAscent
	} else {
		//renderer.run.right += renderer.run.lastRowDescent
		// TODO: I don't know how do we translate this to vertical text.
		//       we might need to memorize and pass extra info during the process.
		// I feel it should be only the advance in multi-line case, but
		// I'm not sure about cases where line breaks happen at the end...
		// and the advance is variable for the whole line. so, basically, if
		// we end at this line, include the max advance, I guess..?
	}
}
