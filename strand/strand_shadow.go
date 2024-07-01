package strand

import "image/color"

// See [Strand.Shadow]().
type StrandShadow Strand

// As a special feature, a strand can have another strand linked
// to it to be used as a "shadow" during draws. This can be used
// to create outlines or hard shadows.
//
// Shadows are rendered using the same glyph indices and baseline
// positions as the primary font strand, which means the two must
// be closely related. In general, a shadow will use the same
// font as the primary strand, possibly with an offset (hard shadow),
// or a derived font created from the primary one (e.g. an outline
// font, which can be easily generated using the ggfnt package).
// 
// The shadow strand can also be set to nil to remove the shadow.
func (self *Strand) Shadow() *StrandShadow {
	return (*StrandShadow)(self)
}

// Sets the strand to be used for the shadow.
// Can be cleared with nil.
func (self *StrandShadow) SetStrand(strand *Strand) {
	self.shadowStrand = strand
}

// Gets the current strand set as a shadow. Nil if none.
func (self *StrandShadow) GetStrand() *Strand {
	return self.shadowStrand
}

// Sets the shadow offsets. By default, offsets are scaled
// alongside the font size, but this behavior can be
// changed through [StrandShadow.SetOffsetScalingEnabled]().
func (self *StrandShadow) SetOffsets(x, y int8) {
	self.shadowOffsetX = x
	self.shadowOffsetY = y
}

// Returns the shadow offsets.
func (self *StrandShadow) GetOffsets() (int8, int8) {
	return self.shadowOffsetX, self.shadowOffsetY
}

// By default, shadow offsets will be scaled alongside the
// font size. In some cases, you might prefer disabling this
// behavior in order to get more precise control over the
// shadow positioning.
func (self *StrandShadow) SetOffsetScalingEnabled(enabled bool) {
	(*Strand)(self).setFlag(strandShadowOffsetScalingOff, !enabled)
}

// Returns whether the shadow offset scaling is enabled.
// See [StrandShadow.SetOffsetScalingEnabled]() for more details.
func (self *StrandShadow) IsOffsetScalingEnabled() bool {
	return !(*Strand)(self).getFlag(strandShadowOffsetScalingOff)
}

// Sets the strand's shadow color.
func (self *StrandShadow) SetColor(rgba color.RGBA) {
	self.shadowColor = rgba
}

// Returns the strand's shadow color.
func (self *StrandShadow) GetColor() color.RGBA {
	return self.shadowColor
}
