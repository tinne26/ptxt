package strand

// Returns the current glyph interspacing shift.
func (self *Strand) GlyphInterspacingShift() int8 {
	return self.interspacingShiftGlyph
}

// Returns the current line interspacing shift.
func (self *Strand) LineInterspacingShift() int8 {
	return self.interspacingShiftLine
}

// Modifies the current glyph interspacing shift.
func (self *Strand) SetGlyphInterspacingShift(value int8) {
	self.interspacingShiftGlyph = value
}

// Modifies the current line interspacing shift.
func (self *Strand) SetLineInterspacingShift(value int8) {
	self.interspacingShiftLine = value
}
