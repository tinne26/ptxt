package strand

// [*ggfnt.Font] objects can map a rune to multiple glyphs (the most
// common use of this are character animations). When this happens,
// its up to the renderer code to decide which exact glyph to use.
//
// The [StrandGlyphPickers] struct allows us to configure the [GlyphPicker]
// entities of a given [Strand] as we need to.
//
// In general, this type is used through method chaining: 
//   strand.GlyphPickers().Add(glyphPicker)
type StrandGlyphPickers Strand

// Gateway to [StrandGlyphPickers].
func (self *Strand) GlyphPickers() *StrandGlyphPickers {
	return (*StrandGlyphPickers)(self)
}

// Returns the number of glyph pickers configured for the strand.
func (self *StrandGlyphPickers) Count() int {
	return len(self.pickHandlers)
}

// Adds a [GlyphPicker] to the strand. 
//
// Using multiple glyph pickers is possible, but if you want to
// do that you will also need to use [GlyphPickerFlags] and
// [StrandGlyphPickers.AddWithFlags](). See those for more details.
func (self *StrandGlyphPickers) Add(picker GlyphPicker) {
	if picker == nil { panic("nil GlyphPicker") }
	self.pickHandlers = append(
		self.pickHandlers,
		glyphPickHandler{ Picker: picker },
	)
}

// Similar to [StrandGlyphPickers.Add](), but associating
// [GlyphPickerFlags] to the newly added picker.
//
// Using many glyph pick handlers can impact rendering performance,
// as they need to be updated and evaluated for each glyph being drawn.
//
// For more context, please refer to the documentation of [GlyphPickerFlags].
func (self *StrandGlyphPickers) AddWithFlags(picker GlyphPicker, flags GlyphPickerFlags) {
	if picker == nil { panic("nil GlyphPicker") }
	self.pickHandlers = append(
		self.pickHandlers,
		glyphPickHandler{ Picker: picker, Flags: flags },
	)
}

// Overrides the glyph picker flags of the nth glyph picker.
// If the index falls outside [0..[StrandGlyphPickers.Count]() - 1],
// the method will panic.
//
// For more details about glyph picker flags, see the documentation
// of the [GlyphPickerFlags] type.
func (self *StrandGlyphPickers) SetFlags(index int, flags GlyphPickerFlags) {
	self.pickHandlers[index].Flags = flags
}

// Clears all glyph pickers from the strand.
func (self *StrandGlyphPickers) ClearAll() {
	self.pickHandlers = self.pickHandlers[ : 0]
}

// Removes the most recently added [GlyphPicker].
func (self *StrandGlyphPickers) Pop() GlyphPicker {
	if len(self.pickHandlers) == 0 { return nil }
	last := self.pickHandlers[len(self.pickHandlers) - 1].Picker
	self.pickHandlers = self.pickHandlers[ : len(self.pickHandlers) - 1]
	return last
}

// Invokes the given callback for each glyph pick handler
// configured on the strand. Iteration is done in order of
// addition, which matches internal use order.
func (self *StrandGlyphPickers) Each(fn func(GlyphPicker, GlyphPickerFlags)) {
	for i, _ := range self.pickHandlers {
		fn(self.pickHandlers[i].Picker, self.pickHandlers[i].Flags)
	}
}
