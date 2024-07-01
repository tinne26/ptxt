package ptxt

import "image/color"

import "github.com/tinne26/ptxt/core"
import "github.com/tinne26/ptxt/strand"

import "github.com/tinne26/ggfnt"

// The [Renderer] is the heart of ptxt and the type around which everything
// else revolves.
// 
// Renderers have three groups of functions:
//  - Simple functions to adjust basic text properties like color, font,
//    align, scale...
//  - Simple functions to draw and measure text.
//  - Gateways to access more advanced or specific functionality.
//
// Gateways are auxiliary types that group specialized functions together
// and keep them out of the way for most workflows that won't require them.
// The following gateways are available:
//  - [Renderer.Strands](), to manage multiple font [*strand.Strand] objects.
//  - [Renderer.Twine](), to operate with multiple font strands, colors,
//    scales, effects and so on within a single block of text.
//  - [Renderer.Advanced](), for advanced options and configurations.
//
// To create a renderer, use [NewRenderer]() and then set a font [*strand.Strand]
// through [Renderer.SetStrand]().
type Renderer struct {
	strands []*strand.Strand
	
	align Align
	direction Direction
	scale uint8
	strandIndex StrandIndex
	boundingMode BoundingMode
	parBreakEnabled bool
	
	blendMode core.BlendMode
	fallbackMainDye color.RGBA // for strands with inactive main dye
	
	drawFunc func(core.Target, ggfnt.GlyphIndex, MaskDrawParameters)
	drawPassListener func(*Renderer, DrawPass)
	
	// operation buffers
	run struct {
		// measurings based on (0, 0) origin
		left int // can only be non-zero on mask bounding
		right int
		top int
		bottom int
		
		// main run data
		// NOTE: for sanity and safety, slices can't exceed 32k elements in size.
		//       this is checked while we generate the slices and so on.
		glyphIndices []ggfnt.GlyphIndex // for twine measuring and drawing, already on the relevant font
		lineLengths []uint16
		advances []uint16 // for twine measuring and drawing, already scaled
		kernings []int16 // for twine measuring and drawing, already scaled (int16 is such a waste...)
		wrapIndices []uint16 // indices before which we append wrap line breaks
		                     // (top bit [0x8000] used as replace bit flag [0x7FFF for value])
		// NOTE: (we could probably join advances + kernings into a single int16?)

		// aux data for some specific use-cases
		firstLineAscent int // we need this for reused draws. already scaled
		lastLineDescent int
		isMultiline bool // necessary for LastBaseline align
	}
}

// Creates a new [Renderer] with the following defaults:
//  - Text scale set to 1.
//  - Direction set to ptxt.[Horizontal].
//  - Bounding mode set to [LogicalBounding].
//  - Align set to (ptxt.[Left] | ptxt.[Baseline]).
//  - Fallback main dye color set to white.
//
// Beyond these properties, you must still set a font [*strand.Strand]
// through [Renderer.SetStrand]() before being able to operate with 
// the renderer.
func NewRenderer() *Renderer {
	var renderer Renderer
	renderer.strands = make([]*strand.Strand, 1)
	renderer.align = Left | Baseline
	renderer.direction = Horizontal
	renderer.scale = 1
	renderer.boundingMode = LogicalBounding
	renderer.fallbackMainDye = color.RGBA{255, 255, 255, 255}
	return &renderer
}

// ---- gateways ----

// Gateway to [RendererStrands]. For context on gateways, see [Renderer].
func (self *Renderer) Strands() *RendererStrands {
	return (*RendererStrands)(self)
}

// Gateway to [RendererTwine]. For context on gateways, see [Renderer].
func (self *Renderer) Twine() *RendererTwine {
	return (*RendererTwine)(self)
}

// Gateway to [RendererAdvanced]. For context on gateways, see [Renderer].
func (self *Renderer) Advanced() *RendererAdvanced {
	return (*RendererAdvanced)(self)
}

// ---- main getters / setters ----

// Sets a fallback main dye color for text rendering.
//
// For context, the color model for [*ggfnt.Font] objects is
// quite sophisticated. Fonts can have more than just one color:
//  - Up to 255 colors are supported per font, similar to an
//    indexed palette.
//  - There are two classes of colors: dyes and palette colors.
//    Palette colors are static colors that rarely change,
//    intended mainly for icons and pictograms. Dyes, on the other
//    hand, are user-customizable font colors. Fonts tend to have
//    only one main dye, but n-colored fonts are technically possible.
//  - Different instances or parametrizations of a font (see
//    [*strand.Strand]) can be configured to use different colors.
// You don't need to understand it all, just realize that a
// single "set color" method can't cover that much terrain.
//
// The [Renderer.SetColor]() method defines a color that will be
// used for [*strand.Strand] objects that don't have an active main dye.
// "Not active" means that either the main dye has never been set
// through [*strand.Strand.SetMainDye](), or that it has been explicitly
// deactivated afterwards.
//
// Back to the practical universe, there are two main ways to manage text
// color in ptxt:
//  - If all your fonts use *only one main dye color*, you can take
//    it easy and change colors exclusively through [Renderer.SetColor]().
//  - Otherwise, you really need to understand how the color
//    model works, understand that colors are actually stored on the
//    strands themselves and manage them directly. Mixing
//    [Renderer.SetColor](), [*strand.Strand.SetDye]() and others
//    is technically possible if you want to, but be warned that it
//    can get really ugly if you don't have a very clear understanding
//    of the precedences between color configurations.
func (self *Renderer) SetColor(rgba color.RGBA) {
	self.fallbackMainDye = rgba
}

// The renderer's [Align] defines how [Renderer.Draw]() and other operations
// interpret the coordinates passed to them. For example, let's assume that
// the renderer is using the [Horizontal] text [Direction]:
//  - If the align is set to (ptxt.[Top] | ptxt.[Left]), coordinates will be
//    interpreted as the top-left corner of the box that the text needs to
//    occupy.
//  - If the align is set to (ptxt.[Center]), coordinates will be interpreted
//    as the center of the box that the text needs to occupy.
// 
// (Run this [wasm example] for an interactive demonstration instead)
// 
// Notice that aligns have a horizontal and a vertical component, so you can
// use [Renderer.SetAlign](ptxt.[Right]) and similar to change only one of the
// components at a time.
// 
// [wasm example]: https://tinne26.github.io/ptxt-examples/align
func (self *Renderer) SetAlign(align Align) {
	self.align = self.align.Adjusted(align)
}

// Returns the current [Align]. See also [Renderer.SetAlign]().
func (self *Renderer) GetAlign() Align {
	return self.align
}

// Sets the text direction to be used on subsequent operations.
func (self *Renderer) SetDirection(direction Direction) {
	self.direction = direction
}

// Returns the current text [Direction]. See also [Renderer.SetDirection]().
func (self *Renderer) GetDirection() Direction {
	return self.direction
}

// Sets the [core.BlendMode] to be applied on subsequent drawing operations.
// The default mode is always regular source over target alpha blending.
func (self *Renderer) SetBlendMode(mode core.BlendMode) {
	self.blendMode = mode
}

// Returns the current blend mode. See also [Renderer.SetBlendMode]().
func (self *Renderer) GetBlendMode() core.BlendMode {
	return self.blendMode
}

// Sets the text scale to be applied on subsequent operations.
// Scale must be at least 1.
func (self *Renderer) SetScale(scale uint8) {
	if scale == 0 { panic("Renderer scale can't be zero") }
	self.scale = scale
}

// Returns the current text scale for the renderer.
// See also [Renderer.SetScale]().
func (self *Renderer) GetScale() uint8 {
	return self.scale
}

// Returns the currently active font [*strand.Strand]. Shorthand for:
//   index := renderer.Strands().Index()
//   return renderer.Strands().Get(index)
func (self *Renderer) Strand() *strand.Strand {
	return self.strands[self.strandIndex]
}

// Replaces the currently active font [*strand.Strand]. Shorthand for:
//   index := renderer.Strands().Index()
//   renderer.Strands().Replace(index, strand)
// Nil strands will cause the method to panic.
func (self *Renderer) SetStrand(fontStrand *strand.Strand) {
	if fontStrand == nil { panic("nil strand") }
	self.strands[self.strandIndex] = fontStrand
}

// ---- main operations ----

// Draws the given text with the current renderer's configuration.
// The drawing position depends on the given pixel coordinates and
// the renderer's align, as specified on [Renderer.SetAlign]().
//
// Text can't exceed 32k glyphs.
func (self *Renderer) Draw(target core.Target, text string, x, y int) {
	self.DrawWithWrap(target, text, x, y, maxInt32)
}

// Like [Renderer.Draw](), but automatically wraps a line and jumps to the next 
// one if line length would exceed the given 'maxLineLen'.
//
// Text can't exceed 32k glyphs.
func (self *Renderer) DrawWithWrap(target core.Target, text string, x, y int, maxLineLen int) {
	// convert the input from code points to glyphs
	// (this includes rewrite rules and glyph selection)
	mapping := self.Strand().Mapping()
	err := lnkBeginPass(mapping, strand.DrawPass)
	if err != nil { panic(err) }
	if self.Strand() == nil {
		panic("ptxt.Renderer can't operate with a nil strand... maybe you forgot to Renderer.SetStrand()?")
	}
	self.run.glyphIndices = self.run.glyphIndices[ : 0]
	for _, codePoint := range text {
		self.run.glyphIndices = lnkAppendCodePoint(mapping, codePoint, self.run.glyphIndices)
		if len(self.run.glyphIndices) > 32000 {
			panic("text run exceeding 32k glyph indices")
		}
	}
	self.run.glyphIndices = lnkFinishMapping(mapping, self.run.glyphIndices)
	
	// compute text advances and metrics
	self.computeRunLayout(maxLineLen)

	// determine text baseline origin (x matters if text is vertical or sideways)
	x, y = self.computeTextOrigin(x, y)

	// draw text
	self.drawText(target, x, y)

	// cleanup
	lnkFinishPass(mapping, strand.DrawPass)
}

// In the default [LogicalBounding] mode, this method returns the dimensions
// of the "highlight rectangle" of the given text. If you have ever selected
// text on a text editor or browser, the total width and height of the shaded
// area would be the "highlight rectangle".
//
// Notice that some glyphs could potentially spill outside the "highlight
// rectangle". You should generally leave some padding around the text instead
// of trying to compensate for spilling in hacky ways; spills, overshoot, side
// bearings and so on are part of typographic design, and you should trust
// font designers to know better than you if they decided to let something
// spill (sometimes compromises have to be made).
//
// Text can't exceed 32k glyphs.
func (self *Renderer) Measure(text string) (width, height int) {
	return self.MeasureWithWrap(text, maxInt32)
}

// Like [Renderer.Measure](), but considering automatic line wrapping 
// at the given 'maxLineLen'.
//
// Text can't exceed 32k glyphs.
func (self *Renderer) MeasureWithWrap(text string, maxLineLen int) (width, height int) {
	// (INTERNAL NOTE): There's a difference in behavior between etxt and
	// ptxt. etxt tends to always include the line gap, while ptxt doesn't
	// include it if it's only a single line. There are pros and cons to
	// both:
	// - Not including line gap for a single line is technically more
	//   accurate, improving single-line centering precision.
	// - Including line gap even for a single line can make centering more
	//   consistent across single vs multiline text and makes independent
	//   measures more composable.
	// The reason why ptxt goes with the first and etxt with the second is
	// simply that in the case of bitmap fonts, centering errors become too
	// evident otherwise. With etxt, instead, this doesn't tend to happen,
	// as text is much higher resolution and even line gaps tend to be much
	// smaller than bitmap font line gaps, so the trade offs are different.
	// Concerned users might depend on adding font.Metrics().LineGap() to
	// single lines if their use-case demands it. We could technically make
	// it an Advanced() flag too, but it seems a bit too much to me.

	// convert the input from code points to glyphs
	// (this includes rewrite rules and glyph selection)
	mapping := self.Strand().Mapping()
	err := lnkBeginPass(mapping, strand.MeasurePass)
	if err != nil { panic(err) }
	if self.Strand() == nil {
		panic("ptxt.Renderer can't operate with a nil strand... maybe you forgot to Renderer.SetStrand()?")
	}
	self.run.glyphIndices = self.run.glyphIndices[ : 0]
	for _, codePoint := range text {
		self.run.glyphIndices = lnkAppendCodePoint(mapping, codePoint, self.run.glyphIndices)
		if len(self.run.glyphIndices) > 32000 {
			panic("text run exceeding 32k glyph indices")
		}
	}
	self.run.glyphIndices = lnkFinishMapping(mapping, self.run.glyphIndices)

	// get text bounding box and advances
	self.computeRunLayout(maxLineLen)

	// cleanup and return
	lnkFinishPass(mapping, strand.MeasurePass)
	return self.run.right - self.run.left, self.run.bottom - self.run.top
}
