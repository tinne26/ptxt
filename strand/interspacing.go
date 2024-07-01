package strand

// Returns the current horizontal interspacing shift.
func (self *Strand) HorzInterspacingShift() int8 {
	return self.interspacingShiftHorz
}

// Returns the current vertical interspacing shift.
func (self *Strand) VertInterspacingShift() int8 {
	return self.interspacingShiftVert
}

// Modifies the current horizontal interspacing shift.
func (self *Strand) SetHorzInterspacingShift(value int8) {
	self.interspacingShiftHorz = value
}

// Modifies the current vertical interspacing shift.
func (self *Strand) SetVertInterspacingShift(value int8) {
	self.interspacingShiftVert = value
}
