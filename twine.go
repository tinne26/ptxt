package ptxt

import "image/color"

import "github.com/tinne26/ggfnt"

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// A flexible type that can have text content added as utf8, raw
// glyphs or a mix of both, with some styling directives also being
// supported through control codes and custom functions.
//
// Twines are an alternative to strings relevant for text formatting,
// custom effects and direct glyph encoding.
//
// Users might wrap twines into their own types for more advanced
// formatting directives.
type Twine struct {
	contents []byte
}

// --- twine creation ---

type twinePopSpecialDirective uint8

// Constants for popping special directives when working with [Weave]()
// and [Twine.Weave](). 
const (
	Pop    twinePopSpecialDirective = 66 // pop last effect function still active
	PopAll twinePopSpecialDirective = 67 // pop all effect functions still active
	Stop   twinePopSpecialDirective = 68 // pop last motion function still active
)


// NOTICE: TWINES ARE UNIMPLEMENTED
//
// Creates a [Twine] from the given arguments. For example:
//   rgba  := color.RGBA{ 80, 200, 120, 255 }
//   twine := ptxt.Weave("NICE ", rgba, "EMERALD", ptxt.Pop, '!')
// You can also pass a twine as the first argument to append to it
// instead of creating a new one. To pop fonts, colors, effects or
// motions, you can use the ptxt.[Pop] and ptxt.[PopAll] constants. 
func Weave(args ...any) Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) Weave(args ...any) *Twine {
	panic("unimplemented")
}

// --- twine basic text content addition ---

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) Add(text string) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) AddRunes(codePoints ...rune) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) AddGlyphs(indices ...ggfnt.GlyphIndex) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) AddUtf8(bytes ...byte) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) AddLineBreak() *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) AddLineMetricsRefresh() *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) Reset() {
	panic("unimplemented")
}

// --- push / pop ---

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) Pop() *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PopAll() *Twine {
	panic("unimplemented")
}

// --- effects ---

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushColor(textColor color.RGBA) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
//
// For changing the font strand at any point in the text. Unlike with
// vectorial fonts, where more arbitrary changes might be used, pixel
// fonts typically will only change between variants of the same font
// in a single block of text. If you are using multiple pixel art fonts
// in the same paragraph, chances are that the fonts have been designed
// to work together from the start.
func (self *Twine) PushStrand(index StrandIndex) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushScale(scale uint8) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushScaleShift(scaleShift int8) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushPadder(spacer TwinePadder) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushSettingChange(key ggfnt.SettingKey, value uint8) {
	// NOTE: this is for changing a setting in the current font strand.
	// This probably can't be done arbitrarily outside this function
	// because sometimes we might need to backtrack on twine formatting
	// or something...

	// TODO: doesn't make much sense for this to be a push, at least in public
	// API terms. To be fair, strands, colors, scales and shifts should
	// all be able to be changed "permanently". But at least with those
	// you can just push and never pop in a fairly reasonable way. But
	// with setting changes, it feels much more icky.
	// Ok, there are some use cases for scope limited changes, but it's
	// not particularly common.

	// TODO: maybe document that in some cases it's actually safe to
	// change settings at any point? like, with a custom function if
	// we are not using any padding or back graphics effect, and no
	// wrap?
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
//
// Registers the current horizontal position in the text and sets it as the new
// line restart position. This is useful to create itemized lists or any other
// kind of text block that requires indentation for multiple lines.
func (self *Twine) PushLineRestartMarker() *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushColorChanger(fn TwineFuncID, payload ...byte) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
//
// For adding text movement (jumpy, wavy, etc). Some motion functions
// and structs are already implemented in tinne26/ptxt/twine (TODO).
func (self *Twine) PushMotion(fn TwineFuncID, payload ...byte) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
//
// For custom rendering over the underlying text.
// Some examples:
//  - Crude strikethrough effect.
//  - Crude underline effect.
//  - Spoiler cover.
//  - Wrap within a [TwinePadder] to draw your own graphics at an
//    arbitrary point in the text.
func (self *Twine) PushGfxFront(fn TwineFuncID, payload ...byte) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
func (self *Twine) PushGfxBack(fn TwineFuncID, payload ...byte) *Twine {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
//
// For custom user configurations, callbacks and notifications.
// Some examples:
//  - Change user-side configuration for custom glyph drawing function
//    (see [RendererAdvanced.SetDrawFunc]()).
//  - ...
// func (self *Twine) PushCallback(fn TwineFuncID, payload ...byte) *Twine {
// 	panic("unimplemented")
// }

// TODO: what else...?
// - some people might want to change shadows midways. I don't think that's
//   too nice looking, but I could consider it.
