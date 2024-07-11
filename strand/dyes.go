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
		// discretionary safety assertions
		if !isPremultiplied(rgba) { panic(nonPremultRGBA) }
		if int(dyeKey) >= self.dyes.Len() { panic("invalid dye key") }
		self.dyes.Set(int(dyeKey), internal.RGBAToFloat32(rgba))
		self.notifyShaderNonMainDyeChange()
	}
}

// Returns the color of the requested dye.
// Invalid dye keys will panic.
func (self *Strand) GetDye(dyeKey ggfnt.DyeKey) [4]float32 {
	// discretionary safety assertion
	if int(dyeKey) >= self.dyes.Len() { panic("invalid dye key") }
	return self.dyes.At(int(dyeKey))
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
	self.dyes.Set(int(self.mainDyeKey), internal.RGBAToFloat32(rgba))
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
