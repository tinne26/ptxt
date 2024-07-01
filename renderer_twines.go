package ptxt

import "github.com/tinne26/ptxt/core"

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Related to [RendererTwine.RegisterFunc](). Only up to 255
// functions can be registered.
type TwineFuncID uint8

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// This type exists only for documentation and structuring purposes,
// acting as a [gateway] to [Twine] operations and related configurations.
//
// In general, this type is used through method chaining:
//   renderer.Twine().Draw(canvas, twine, x, y)
//
// [gateway]: https://pkg.go.dev/github.com/tinne26/ptxt#Renderer
type RendererTwine Renderer

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Like [Renderer.Draw](), but accepting a twine instead of a string.
func (self *RendererTwine) Draw(target core.Target, twine Twine, x, y int) {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Like [Renderer.DrawWithWrap](), but accepting a twine instead of a string.
func (self *RendererTwine) DrawWithWrap(target core.Target, twine Twine, x, y int, maxLineLen int) {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Like [Renderer.Measure](), but accepting a twine instead of a string.
func (self *RendererTwine) Measure(twine Twine) {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Like [Renderer.MeasureWithWrap](), but accepting a twine instead of a string.
func (self *RendererTwine) MeasureWithWrap(twine Twine, maxLineLen int) {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Like [RendererAdvanced.Cache](), but accepting a twine instead of a string.
func (self *RendererTwine) Cache(twine Twine) {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Like [RendererAdvanced.AllGlyphsAvailable](), but accepting a twine instead of a string.
func (self *RendererTwine) AllGlyphsAvailable(text string) bool {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Registers a function for use with twine effects. Only a strict
// set of signatures are valid:
//  - func(args ptxt.TwineEffectArgs)
//  - func(args ptxt.TwineEffectArgs, target *ebiten.Image, targetArea image.Rectangle)
//  - ... (TODO)
// Nil functions are not allowed.
func (self *RendererTwine) RegisterFunc(fn any) TwineFuncID {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Releases a previously registered function. The given ID might be
// reused next time you register a new function.
func (self *RendererTwine) ReleaseFunc(id TwineFuncID) bool {
	panic("unimplemented")
}
