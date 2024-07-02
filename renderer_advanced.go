package ptxt

import "strconv"

import "github.com/tinne26/ptxt/internal"
import "github.com/tinne26/ptxt/core"
import "github.com/tinne26/ptxt/strand"

import "github.com/tinne26/ggfnt"

// This type exists only for documentation and structuring purposes,
// acting as a [gateway] to advanced renderer functions and configurations
// that most users rarely need to touch.
//
// In general, this type is used through method chaining:
//   mask := renderer.Advanced().LoadMask(glyphIndex)
//
// [gateway]: https://pkg.go.dev/github.com/tinne26/ptxt#Renderer
type RendererAdvanced Renderer

// The mask draw parameters are used internally to draw glyphs.
// In terms of the API, these are only exposed if you are using
// custom draw functions through [RendererAdvanced.SetDrawFunc]().
type MaskDrawParameters struct {
	X, Y int
	Scale int
	RGBA [4]float32
}

// Sets a custom drawing function for the renderer.
// Can be set to nil to go back to the default drawing function.
func (self *RendererAdvanced) SetDrawFunc(fn func(core.Target, ggfnt.GlyphIndex, MaskDrawParameters)) {
	self.drawFunc = fn
}

// Related to [RendererAdvanced.SetDrawPassListener].
type DrawPass uint8
const (
	MainDrawPass DrawPass = iota
	ShadowDrawPass
)

// Allows the user to be notified of [MainDrawPass] and [ShadowDrawPass]
// draw passes right before they begin. Its main use is changing the draw
// function through [RendererAdvanced.SetDrawFunc]().
func (self *RendererAdvanced) SetDrawPassListener(fn func(*Renderer, DrawPass)) {
	self.drawPassListener = fn
}

// Loads the mask for the given glyph index. Mostly needed to implement
// custom drawing functions for [RendererAdvanced.SetDrawFunc]().
func (self *RendererAdvanced) LoadMask(glyphIndex ggfnt.GlyphIndex) core.GlyphMask {
	return (*Renderer)(self).loadMask(glyphIndex, (*Renderer)(self).Strand().Font())
}

func (self *Renderer) loadMask(glyphIndex ggfnt.GlyphIndex, font *ggfnt.Font) core.GlyphMask {
	fontKey := font.Header().ID()
	mask, found := internal.DefaultCache.GetGlyphMask(fontKey, glyphIndex)
	if found { return mask }

	// mask not found, obtain and cache
	alphaMask := font.Glyphs().RasterizeMask(glyphIndex)
	mask = alphaMaskToMask(alphaMask)
	internal.DefaultCache.SetGlyphMask(fontKey, glyphIndex, mask)
	return mask
}

// Draws a mask into the given target. Mostly needed to implement
// custom drawing functions for [RendererAdvanced.SetDrawFunc]().
//
// The drawing orientation will be automatically determined by the
// renderer's current direction.
func (self *RendererAdvanced) DrawMask(target core.Target, mask core.GlyphMask, fontStrand *strand.Strand, params MaskDrawParameters) {
	switch self.direction {
	case Horizontal:
		lnkDrawHorzMask(fontStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
	case Vertical:
		panic("unimplemented")
	case Sideways:
		lnkDrawSidewaysMask(fontStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
	case SidewaysRight:
		lnkDrawSidewaysRightMask(fontStrand, target, mask, params.X, params.Y, params.Scale, params.RGBA)
	default:
		panic("invalid text renderer direction")
	}
}

// See [RendererAdvanced.SetBoundingMode]().
type BoundingMode uint8
const (
	LogicalBounding       BoundingMode = 0b0000_0001 // font metrics based
	MaskBounding          BoundingMode = 0b0000_0010 // glyph mask rects
	NoDescLogicalBounding BoundingMode = LogicalBounding | noDescent // (very rarely used)
	NoDescMaskBounding    BoundingMode = MaskBounding    | noDescent // (very rarely used)
	
	noDescent BoundingMode = 0b1000_0000 // flag, can't be used in isolation
)

// Returns a string representation of the bounding mode.
func (self BoundingMode) String() string {
	switch self {
	case LogicalBounding       : return "LogicalBounding"
	case MaskBounding          : return "MaskBounding"
	case NoDescLogicalBounding : return "NoDescLogicalBounding"
	case NoDescMaskBounding    : return "NoDescMaskBounding"
	default:
		return "BoundingModeUndefined#" + strconv.Itoa(int(self))
	}
}

// Sets the bounding mode, which affects both measuring and drawing operations.
// 
// By default, measuring and drawing is done based on logical font metrics.
// In some very specific cases, though, you might prefer to operate using
// the raw glyph mask rects, which are more closely representative of the
// actual glyph sizes. As a rule of thumb, you don't want to change this
// setting unless you are doing some kind of "bitmap font art" where you
// need to create your own unique and unconventional layouts.
//
// Notice that glyph rect bounds require the full glyph masks to be
// retrieved, which aren't necessary on conventional measuring
// operations and can make operations more expensive.
func (self *RendererAdvanced) SetBoundingMode(mode BoundingMode) {
	self.boundingMode = mode
}

// See [RendererAdvanced.SetBoundingMode]().
func (self *RendererAdvanced) GetBoundingMode() BoundingMode {
	return self.boundingMode
}

// Each time we draw or measure, these values are updated. For most
// use-cases, you only care about total width and height of the text,
// but in some cases (e.g. on [MaskBounding] mode), you might also
// be interested in the offset of the text's top-left corner (relative
// to the text's baseline).
func (self *RendererAdvanced) LastBoundsOffset() (int, int) {
	return self.run.left, self.run.top
}

// Utility method to cache the glyphs of the given text in advance.
// Notice that in some extreme cases, if the cache is too small or
// the text maps to too many glyphs, not all glyphs might be cached.
//
// This method is virtually never strictly necessary, but in some very
// specific cases it may help smooth performance (e.g. pre-cache glyphs
// before switching to a new scene that uses a different font and draws
// a lot of text).
//
// TODO: string to glyphs depends on a lot of undocumented state.
// We are basically caching all glyphs in the rune's glyph set
// based on draw-like state.
func (self *RendererAdvanced) Cache(text string) {
	(*Renderer)(self).cache(text)
}

func (self *Renderer) cache(text string) {
	strand  := self.Strand()
	font    := strand.Font()
	mapping := font.Mapping()
	settings := strand.UnderlyingSettingsCache().UnsafeSlice()
	
	for _, codePoint := range text {
		group, found := mapping.Utf8(codePoint, settings)
		if !found { continue }
		for i := uint8(0); i < group.Size(); i++ {
			_ = self.loadMask(group.Select(i), font)
		}
	}
}

// Utility method that returns false if the current font [*strand.Strand]
// is missing any of the glyphs required to process the given text.
//
// TODO: should this do draw-like mapping first on the text, or is
// the current rune-by-rune approach better? Do we need a separate
// IsStringAvailable()?
func (self *RendererAdvanced) AllGlyphsAvailable(text string) bool {
	// NOTE: maybe it should be mentioned in the docs that strand
	// settings might affect the results of this operation? I'm not
	// sure of it myself, and it would be really uncommon and probably
	// point to a font design bad practice, but...
	// Also, rewrite rules are not taken into account.
	return (*Renderer)(self).allGlyphsAvailable(text)
}

func (self *Renderer) allGlyphsAvailable(text string) bool {
	strand  := self.Strand()
	font    := strand.Font()
	mapping := font.Mapping()
	settings := strand.UnderlyingSettingsCache().UnsafeSlice()
	for _, codePoint := range text {
		_, found := mapping.Utf8(codePoint, settings)
		if !found { return false }
	}
	return true
}

// Single-rune version of [RendererAdvanced.AllGlyphsAvailable]().
func (self *RendererAdvanced) IsRuneAvailable(codePoint rune) bool {
	return (*Renderer)(self).isRuneAvailable(codePoint)
}

func (self *Renderer) isRuneAvailable(codePoint rune) bool {
	strand  := self.Strand()
	settings := strand.UnderlyingSettingsCache().UnsafeSlice()
	_, found := strand.Font().Mapping().Utf8(codePoint, settings)
	return found
}

// func (self *RendererAdvanced) SetTabSpaces(n int) {}
// func (self *RendererAdvanced) GetTabSpaces() int {}

// A configuration to advance two consecutive line breaks as x1.5 line 
// height instead of x2. When it comes to long text, this tends to make
// the spacing between paragraphs look more natural.
//
// Three consecutive line breaks will be rendered as two full line breaks
// instead (and four as three and so on).
func (self *RendererAdvanced) SetParBreakEnabled(enabled bool) {
	self.parBreakEnabled = enabled
}

// Returns whether paragraph breaks are enabled or not. See also
// [RendererAdvanced.SetParBreakEnabled]().
func (self *RendererAdvanced) GetParBreakEnabled() bool {
	return self.parBreakEnabled
}

// Uses the data from the previous measure or draw operation to draw
// it directly without additional recomputations. This obviously
// makes this operation very low-level and unsafe.
//
// There are two main use-cases for this function are:
//  - Optimize Measure + Draw => Measure + DrawFromBuffer.
//  - Perform the same Draw repeatedly at different places or targets.
//
// You may change the target, horz/sideways text directions and colors
// if you want, but you can't change scale nor line wrap max length...
// or things will get messy. Well, maybe that counts as a use-case...
func (self *RendererAdvanced) DrawFromBuffer(target core.Target, x, y int) {
	(*Renderer)(self).drawFromBuffer(target, x, y)
}

func (self *Renderer) drawFromBuffer(target core.Target, x, y int) {
	if self.Strand() == nil {
		panic("ptxt.Renderer can't operate with a nil strand... maybe you forgot to Renderer.SetStrand()?")
	}

	mapping := self.Strand().Mapping()
	err := lnkBeginPass(mapping, strand.BufferPass)
	if err != nil { panic(err) }
	x, y = self.computeTextOrigin(x, y)
	self.drawText(target, x, y)
	lnkFinishPass(mapping, strand.BufferPass)
}

// func (self *RendererAdvanced) LastOpEndPos() (x, y int, outOfBounds bool) {}
// func (self *RendererAdvanced) SetGlyphMissPolicy(GlyphMissPolicy) {}
// func (self *RendererAdvanced) GetGlyphMissPolicy() GlyphMissPolicy {}
// type GlyphMissPolicy uint8
// const (
//    // NOTE: could also try uppercase if available..?
// 	GlyphMissPanic GlyphMissPolicy = iota // panic on missing glyph
// 	GlyphMissSkip // skip and ignore missing glyphs
// 	GlyphMissNotdef // draw a notdef glyph (use the font's "notdef" if present)
// 	GlyphMissEmptyRect // draw a standard notdef glyph, ignoring the font's "notdef"
// )

// func (self *RendererAdvanced) StoreState() {} // dubious due to necessarily patchy implementation
// func (self *RendererAdvanced) RestoreState() {} // same as above
// func (self *RendererAdvanced) NumStoredStates() int {} // same as above
