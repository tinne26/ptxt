package strand

import "image/color"

import "github.com/tinne26/ptxt/internal"

import "github.com/tinne26/ggfnt"

// Sets a dye color. The method will panic if the underlying font
// doesn't have any such dye or if the dye key is not valid.
func (self *Strand) SetDye(dyeKey ggfnt.DyeKey, rgba color.RGBA) {
	if dyeKey == self.mainDyeKey {
		self.SetMainDye(rgba)
	} else {
		if !isPremultiplied(rgba) { panic(nonPremultRGBA) }
		dyeBaseIndex := int(dyeKey) << 2
		if dyeBaseIndex + 3 >= len(self.dyes) { // discretionary safety assertion
			panic("invalid dye key")
		}
		rgbaF32 := internal.RGBAToFloat32(rgba)
		self.dyes[dyeBaseIndex + 0] = rgbaF32[0]
		self.dyes[dyeBaseIndex + 1] = rgbaF32[1]
		self.dyes[dyeBaseIndex + 2] = rgbaF32[2]
		self.dyes[dyeBaseIndex + 3] = rgbaF32[3]
		self.notifyShaderNonMainDyeChange()
	}
}

// Returns the color of the requested dye.
// Invalid dye keys will panic.
func (self *Strand) GetDye(dyeKey ggfnt.DyeKey) [4]float32 {
	dyeBaseIndex := int(dyeKey) << 2
	if dyeBaseIndex + 3 >= len(self.dyes) { // discretionary safety assertion
		panic("invalid dye key")
	}
	return [4]float32{
		self.dyes[dyeBaseIndex + 0],
		self.dyes[dyeBaseIndex + 1],
		self.dyes[dyeBaseIndex + 2],
		self.dyes[dyeBaseIndex + 3],
	}
}

// Can be used to check whether the main dye of a font is
// defined or not.
const NoDyeKey = ggfnt.DyeKey(255)

// Returns the main dye key, [NoDyeKey] if none.
// 
// Most fonts have a main dye, but in some rare cases
// they might not (e.g. icon fonts that only use palettes).
func (self *Strand) MainDyeKey() ggfnt.DyeKey {
	return self.mainDyeKey
}

// Sets the strand's main dye color. This is sometimes
// called directly through the renderer, like:
//   renderer.Strand().SetMainDye(rgba)
// You can only set the main dye if [Strand.MainDyeKey]() != [NoDyeKey].
// Otherwise, the method will panic.
func (self *Strand) SetMainDye(rgba color.RGBA) {
	if self.mainDyeKey == NoDyeKey { panic("font doesn't have a \"main\" dye key") }
	if !isPremultiplied(rgba) { panic(nonPremultRGBA) }
	self.mainDyeRGBA8 = rgba
	self.setFlag(strandMainDyeColorActive, true)
	dyeBaseIndex := int(self.mainDyeKey) << 2
	rgbaF32 := internal.RGBAToFloat32(rgba)
	self.dyes[dyeBaseIndex + 0] = rgbaF32[0]
	self.dyes[dyeBaseIndex + 1] = rgbaF32[1]
	self.dyes[dyeBaseIndex + 2] = rgbaF32[2]
	self.dyes[dyeBaseIndex + 3] = rgbaF32[3]
}

// Returns the strand's main dye color. This is sometimes
// called directly through the renderer, like:
//   _ = renderer.Strand().GetMainDye(rgba)
// You can only query the main dye if [Strand.MainDyeKey]() != [NoDyeKey].
// Otherwise, the method will panic.
func (self *Strand) GetMainDye() color.RGBA {
	if self.mainDyeKey == NoDyeKey { panic("font doesn't have a \"main\" dye key") }
	return self.mainDyeRGBA8
}

// Related to [Renderer.SetColor](), see that for more details.
//
// By default, the strand's main dye is considered inactive.
// Setting the main dye explicitly through [Strand.SetDye]()
// or [Strand.SetMainDye]() will change the state to active.
// You can still set it to inactive afterwards through this
// method without losing the last value you set.
//
// [Renderer.SetColor]: https://pkg.go.dev/github.com/tinne26/ptxt#Renderer.SetColor
func (self *Strand) SetMainDyeActive(active bool) {
	if active && self.mainDyeKey == NoDyeKey {
		panic("font doesn't have a \"main\" dye key")
	}
	self.setFlag(strandMainDyeColorActive, active)
}

// Returns whether the main dye has been explicitly set.
// See also [Strand.SetMainDyeActive]().
func (self *Strand) IsMainDyeActive() bool {
	return self.getFlag(strandMainDyeColorActive)
}
