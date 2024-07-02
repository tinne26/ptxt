package strand

import "github.com/tinne26/ggfnt"

// Special value for [GlyphPicker] code points.
const NoCodePoint rune = -1_234_567_890

// In [*ggfnt.Font] objects, code points can be mapped to more than one
// glyph at the same time. This is most often used to add animated
// characters to fonts, but other uses are also possible. This is an
// advanced feature that most fonts do not use; you may ignore glyph
// pickers if that's your case.
//
// Otherwise, glyph pickers are interfaces that select the right glyphs
// to use for given code points, glyph pools, positions and instants.
// In general, glyph picking is extremely context dependent, so font
// creators are encouraged to provide explicit rules or concrete glyph
// picker implementations on their own, if necessary.
//
// Glyph pickers are configured directly at the font strand level
// through [StrandGlyphPickers.Add]().
type GlyphPicker interface {
	// This method is invoked whenever we need to pick a glyph index for a rune.
	// We are given the size of the glyph pool, and must return a value < groupSize.
	// Group size can never be zero.
	//
	// Finally, numQueuedGlyphs is relevant when rewrite glyph rules exist. In this
	// case, it's possible that we have to map a code point to a rune while we are
	// still in the middle of the detection of a glyph rewrite rule. When this happens,
	// numQueuedGlyphs can be greater than zero, and the last NotifyAddedGlyph()
	// call doesn't necessarily correspond to the previous glyph index in the text.
	// This is admittedly complicated; in most cases, it doesn't even matter, as
	// the glyph picker is not interested in the absolute glyph position within the
	// text... otherwise, you can often assume that the font won't contain glyph
	// rewrite rules or that they won't collide with the utf8 content, and panic
	// if numQueuedGlyphs is not zero, or assume that the final position will be
	// offset by numQueuedGlyphs. Very context dependent.
	Pick(codePoint rune, groupSize uint8, flags ggfnt.AnimationFlags, numQueuedGlyphs int) uint8
	
	// If the glyph doesn't correspond directly to a code point (e.g. was encoded
	// directly as a glyph or is the result of a glyph rewrite rule), then the code
	// point will be [NoCodePoint], the group size will be 0 and the flags will be
	// empty.
	NotifyAddedGlyph(glyphIndex ggfnt.GlyphIndex, codePoint rune, groupSize uint8, flags ggfnt.AnimationFlags)
	
	// Notifies the glyph picker at the start and end of operation passes.
	// Can be relevant for setups and cleanups.
	NotifyPass(pass GlyphPickerPass, start bool)
}

// Related to [GlyphPicker].
type GlyphPickerPass = uint8
const (
	MeasurePass GlyphPickerPass = iota
	DrawPass
	BufferPass
)

// In some cases you will need to use multiple [GlyphPicker] interfaces on
// a single font strand (see [StrandGlyphPickers.AddWithFlags]()). In those
// cases, you can use the glyph picker flags to specify when to use or not
// use each glyph picker.
// 
// In particular, a glyph picker will only be used for a glyph if the glyph's
// animation flags contain all the 'Required' and none of the 'Rejected' flags
// associated to to the picker. Otherwise, the next glyph picker in the list
// will be considered.
type GlyphPickerFlags struct {
	Required ggfnt.AnimationFlags // picker not applied if flags & group != RequiredAnimFlags
	Rejected ggfnt.AnimationFlags // picker not applied if flags & group != 0
}

func (self *GlyphPickerFlags) compatible(flags ggfnt.AnimationFlags) bool {
	return (self.Required & flags == self.Required) && (self.Rejected & flags == 0)
}

// Internal type for wrapping glyph pickers with animation flags.
type glyphPickHandler struct {
	Picker GlyphPicker
	Flags GlyphPickerFlags
}

func (self *glyphPickHandler) IsCompatible(flags ggfnt.AnimationFlags) bool {
	return self.Flags.compatible(flags)
}
