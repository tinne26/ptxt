package strand

import "github.com/tinne26/ggfnt"

// renderer internal use linkname target
func (self *StrandMapping) beginPass(pass GlyphPickerPass) error {
	for i, _ := range self.pickHandlers {
		self.pickHandlers[i].Picker.NotifyPass(pass, true)
	}
	err := self.utf8Tester.BeginSequence(self.font, &self.settings)
	if err != nil { return err }
	return self.glyphTester.BeginSequence(self.font, &self.settings)
}

// renderer internal use linkname target
func (self *StrandMapping) finishPass(pass GlyphPickerPass) {
	for i, _ := range self.pickHandlers {
		self.pickHandlers[i].Picker.NotifyPass(pass, false)
	}
}

// renderer internal use linkname target
//
// Precondition: neither utf8Tester rules nor rewriteRulesDisabled can be
// modified while a process is active.
//
// This function will panic if the glyph index is missing. Otherwise,
// I'd need to adjust and pass a glyph missing failure policy.
func (self *StrandMapping) appendCodePoint(codePoint rune, buffer []ggfnt.GlyphIndex) []ggfnt.GlyphIndex {
	self.tempGlyphBuffer = buffer

	// consider glyph - rune changes
	if !self.getFlag(strandLastAppendWasRune) && !self.getFlag(strandFirstAppendIncoming) {
		self.glyphTester.Break(self.testerAppendGlyphIndexFunc)
	}
	self.setFlag(strandFirstAppendIncoming, false)
	self.setFlag(strandLastAppendWasRune, true)

	if self.getFlag(strandRewriteRulesDisabled) || self.utf8Tester.NumRules() == 0 {
		// regular utf8 mapping without rules applied
		self.testerAppendCodePointFunc(codePoint)
	} else { // utf8Tester path
		err := self.utf8Tester.Feed(codePoint, self.testerAppendCodePointFunc)
		if err != nil { panic(err) }
	}
	return self.releaseTempGlyphBuffer()
}

// renderer internal use linkname target
func (self *StrandMapping) appendGlyphIndex(glyphIndex ggfnt.GlyphIndex, buffer []ggfnt.GlyphIndex) []ggfnt.GlyphIndex {
	self.tempGlyphBuffer = buffer

	// consider rune - glyph changes
	if self.getFlag(strandLastAppendWasRune) && !self.getFlag(strandFirstAppendIncoming) {
		self.utf8Tester.Break(self.testerAppendCodePointFunc)
	}
	self.setFlag(strandFirstAppendIncoming, false)
	self.setFlag(strandLastAppendWasRune, false)
	
	// if glyph rules disabled, this is a basic append
	if self.getFlag(strandRewriteRulesDisabled) || self.glyphTester.NumRules() == 0 {
		self.testerAppendGlyphIndexFunc(glyphIndex)
	} else {
		err := self.glyphTester.Feed(glyphIndex, self.testerAppendGlyphIndexFunc)
		if err != nil { panic(err) }
	}

	return self.releaseTempGlyphBuffer()
}	

// (internal)
func (self *StrandMapping) finishMapping(buffer []ggfnt.GlyphIndex) []ggfnt.GlyphIndex {
	self.tempGlyphBuffer = buffer
	self.utf8Tester.FinishSequence(self.testerAppendCodePointFunc)
	self.glyphTester.FinishSequence(self.testerAppendGlyphIndexFunc)
	return self.releaseTempGlyphBuffer()
}

// (internal)
func (self *StrandMapping) testerAppendGlyphIndexFunc(glyphIndex ggfnt.GlyphIndex) {	
	self.tempGlyphBuffer = append(self.tempGlyphBuffer, glyphIndex)
	for i, _ := range self.pickHandlers {
		self.pickHandlers[i].Picker.NotifyAddedGlyph(glyphIndex, NoCodePoint, 0, 0)
	}
}

// (internal)
func (self *StrandMapping) testerAppendCodePointFunc(codePoint rune) {
	// get glyph group for the code point, pick one glyph from it
	var glyphIndex ggfnt.GlyphIndex = ggfnt.GlyphMissing
	group, found := self.font.Mapping().Utf8(codePoint, self.settings.UnsafeSlice())
	if found {
		size := group.Size()
		flags := group.AnimationFlags()
		if size == 1 {
			glyphIndex = group.Select(0)
		} else {
			for i, _ := range self.pickHandlers {
				if !self.pickHandlers[i].IsCompatible(flags) { continue }
				numPendingGlyphs := self.glyphTester.NumPendingGlyphs()
				poolChoice := self.pickHandlers[i].Picker.Pick(codePoint, size, flags, numPendingGlyphs)
				glyphIndex = group.Select(poolChoice)
				break
			}
	
			// if no pick handlers or no compatible pick handlers, pick always the first glyph
			if glyphIndex == ggfnt.GlyphMissing {
				glyphIndex = group.Select(0)
			}
		}
	} else {
		if codePoint == '\n' { // manual line feed handling
			self.glyphTester.Break(self.testerAppendGlyphIndexFunc)
			self.testerAppendGlyphIndexFunc(ggfnt.GlyphNewLine)
			return
		} else if codePoint < 32 {
			panic("no glyph index for ASCII control code " + itoaRune(codePoint) + " [" + runeToUnicodeCode(codePoint) + "]")
		} else {
			panic("glyph index for '" + string(codePoint) + "' [" + runeToUnicodeCode(codePoint) + "] missing")
		}
	}

	// append or feed selected glyph
	if self.getFlag(strandRewriteRulesDisabled) || self.glyphTester.NumRules() == 0 {
		self.testerAppendGlyphIndexFunc(glyphIndex)
	} else {
		err := self.glyphTester.Feed(glyphIndex, self.testerAppendGlyphIndexFunc)
		if err != nil { panic(err) }
	}
}

func (self *StrandMapping) releaseTempGlyphBuffer() []ggfnt.GlyphIndex {
	buffer := self.tempGlyphBuffer
	self.tempGlyphBuffer = nil
	return buffer
}


// --- duplicate helpers for ease of use ---

func (self *StrandMapping) setFlag(bit uint8, on bool) {
	if on { self.flags |= bit } else { self.flags &= ^bit }
}

func (self *StrandMapping) getFlag(bit uint8) bool {
	return self.flags & bit != 0
}
