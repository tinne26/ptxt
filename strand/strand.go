package strand

import "errors"
import "image/color"

import "github.com/tinne26/ptxt/internal"
import "github.com/tinne26/ptxt/core"

import "github.com/tinne26/ggfnt"
import "github.com/tinne26/ggfnt/rerules"

const strandShadowOffsetScalingOff uint8 = 0b0000_0100
const strandRewriteRulesDisabled   uint8 = 0b0001_0000
const strandFirstAppendIncoming    uint8 = 0b0010_0000
const strandLastAppendWasRune      uint8 = 0b0100_0000
const strandMainDyeColorActive     uint8 = 0b1000_0000

// In practice, fonts have many parameters that we might want to 
// configure before drawing: colors, settings, custom glyphs, etc.
// 
// In ptxt, these "font parametrizations" are represented as "strands".
// In general, you can mentally replace "strand" with "font", but
// they are not technically the same: fonts are static objects, while
// strands are modifiable parametrizations.
//
// To use a font in ptxt, you first create a strand for it and then
// link it to the renderer with [Renderer.SetStrand]().
//
// [Renderer.SetStrand]: https://pkg.go.dev/github.com/tinne26/ptxt#Renderer.SetStrand
type Strand struct {
	font *ggfnt.Font
	pickHandlers []glyphPickHandler

	// multi-purpose flags, see constants at the top of the file
	flags uint8

	// spacing
	interspacingShiftHorz int8
	interspacingShiftVert int8

	// settings and custom glyphs
	settings ggfnt.SettingsCache
	customGlyphs []core.GlyphMask

	// shadow
	shadowStrand *Strand
	shadowColor color.RGBA
	shadowOffsetX int8
	shadowOffsetY int8

	// coloring
	mainDyeKey ggfnt.DyeKey
	fontColorPalettesStartIndex uint8
	mainDyeRGBA8 color.RGBA
	fontColors rgbaSlice // including indices and palettes, up to 255 RGBA sets
	dyes rgbaSlice // indexed directly with ggfnt.DyeKey

	// mapping and rewrites
	utf8Tester rerules.Utf8Tester
	glyphTester rerules.GlyphTester
	tempGlyphBuffer []ggfnt.GlyphIndex // *
	// the tempGlyphBuffer is only used internally during operations, we don't hold to it.
	// notice that this makes the logic quite a bit harder to follow, but we need this to
	// send around functions as parameters to other functions. alternatively, we could also
	// only re-link the renderer buffer on StrandMapping.finish*() operations, but that
	// wouldn't work with twines, which need some stuff added at arbitrary points on the
	// renderer side.
	mappingCache *ggfnt.MappingCache

	// wrap glyphs
	spaceGlyph ggfnt.GlyphIndex
	wrapGlyphs [sentinelWrapModesCount][]ggfnt.GlyphIndex
	wrapGlyphRanges [sentinelWrapModesCount][]ggfnt.GlyphRange
	// NOTE: this is too many slices, so much overhead...

	// version dependent (gpu vs cpu) rendering data
	re renderData
}

// Creates a default strand of the given font.
func New(font *ggfnt.Font) *Strand {
	if font == nil { panic("nil font") }

	// initialize settings
	settings := ggfnt.NewSettingsCache(font)
	// (values already default to zero)

	// initialize colors
	numColors := font.Color().Count()
	if numColors == 0 { panic("invalid font") }
	fontColors := newRGBASlice(int(numColors))
	fontColorIndex := 0

	// initialize dyes
	mainDyeKey := NoDyeKey // regular dye keys can't reach this value
	dyes := newRGBASlice(int(font.Color().NumDyes()))
	font.Color().EachDye(func(key ggfnt.DyeKey, name string) {
		if name == "main" {
			if mainDyeKey != NoDyeKey { panic("font contains multiple 'main' dye keys") }
			mainDyeKey = key
		}
		dyes.Set(int(key), [4]float32{1.0, 1.0, 1.0, 1.0}) // default to white
		font.Color().EachDyeAlpha(key, func(alpha uint8) {
			alpha32 := float32(alpha)/255.0
			fontColors.Set(fontColorIndex, [4]float32{alpha32, alpha32, alpha32, alpha32})
			fontColorIndex += 1
		})
	})

	// initialize paletted colors
	font.Color().EachPalette(func(key ggfnt.PaletteKey, _ string) {
		font.Color().EachPaletteColor(key, func(rgba color.RGBA) {
			rgbaF32 := internal.RGBAToFloat32(rgba)
			fontColors.Set(fontColorIndex, rgbaF32)
			fontColorIndex += 1
		})
	})
	if fontColorIndex > 255 { panic(brokenCode) } // invalid font data

	// set up wrap glyphs (look for space only, by default)
	group, found := font.Mapping().Utf8(' ', settings.UnsafeSlice())
	var spaceGlyph ggfnt.GlyphIndex
	if found {
		if group.Size() != 1 {
			panic("expected font to have space ' ' mapped to a single glyph or not have it at all")
		}
		spaceGlyph = group.Select(0)
	} else {
		spaceGlyph = internal.MissingSpaceGlyph
	}	

	strand := &Strand{
		font: font,
		settings: *settings,
		fontColors: fontColors,
		dyes: dyes,
		spaceGlyph: spaceGlyph,
	}
	strand.renderDataInit()
	return strand
}

// Returns the underlying [*ggfnt.Font].
func (self *Strand) Font() *ggfnt.Font { return self.font }

// Modifies the setting value. If the setting doesn't exist or
// the given value falls outside the valid range, the method will
// panic.
func (self *Strand) SetSetting(key ggfnt.SettingKey, option uint8) {
	numOptions := self.font.Settings().GetNumOptions(key)
	if option >= numOptions {
		panic("given setting option doesn't exist")
	}
	
	// TODO: what about mapping? what about state? if not running
	//       or rules disabled, I should set a pendingReConditionsRefresh
	//       bool = true
	mappingCasesAffected, rewriteConditionsAffected := self.settings.Set(key, option)
	if rewriteConditionsAffected {
		self.glyphTester.RefreshConditions(self.font, &self.settings)
		self.utf8Tester.RefreshConditions(self.font, &self.settings)
	}
	if mappingCasesAffected && self.mappingCache != nil {
		self.mappingCache.Drop()
	}
}

// Returns the current value of the setting. If the setting key
// is not valid, the method will panic.
func (self *Strand) GetSetting(key ggfnt.SettingKey) uint8 {
	return self.settings.Get(key)
}

// Unsafe, low level method to access the underlying settings cache.
func (self *Strand) UnderlyingSettingsCache() *ggfnt.SettingsCache {
	return &self.settings
}

// Related to [Strand.SetWrapGlyphs]() and [Strand.SetWrapGlyphRanges]().
type WrapMode uint8
const (
	WrapBefore  WrapMode = iota
	WrapAfter
	WrapElide
	sentinelWrapModesCount
)

// Defines the glyphs for which line wrapping is allowed.
// Spaces are always implicitly allowed as wrap glyphs.
// 
// See [Renderer.DrawWithWrap]() for further details.
//
// [Renderer.DrawWithWrap]: https://pkg.go.dev/github.com/tinne26/ptxt#Renderer.DrawWithWrap
func (self *Strand) SetWrapGlyphs(mode WrapMode, glyphIndices []ggfnt.GlyphIndex) {
	// NOTE: I should probably copy the data instead of reusing the slice.
	self.wrapGlyphs[mode] = glyphIndices
}

// Like [Strand.SetWrapGlyphs](), but using glyph ranges.
func (self *Strand) SetWrapGlyphRanges(mode WrapMode, glyphRanges []ggfnt.GlyphRange) {
	// NOTE: I should probably copy the data instead of reusing the slice.
	self.wrapGlyphRanges[mode] = glyphRanges
}

func (self *Strand) CanWrap(glyphIndex ggfnt.GlyphIndex, mode WrapMode) bool {
	// NOTE: I could have the data more organized and binary search
	if glyphIndex == self.spaceGlyph { return true }
	for _, index := range self.wrapGlyphs[mode] {
		if index == glyphIndex { return true }
	}
	for _, glyphRange := range self.wrapGlyphRanges[mode] {
		if glyphRange.Contains(glyphIndex) { return true }
	}
	return false
}

// For custom glyphs. Mask bounds are what determine the positioning.
// Some notable limitations:
//  - Glyphs that are too big simply can't be added.
//  - The font colors can't be arbitrarily changed or extended, so
//    you either use only the main dye... or you have to work with
//    the existing font's color palette.
func (self *Strand) AddGlyph(mask core.GlyphMask) (ggfnt.GlyphIndex, error) {
	bounds := mask.Bounds()
	placement := ggfnt.GlyphPlacement{
		Advance: uint8(min(255, bounds.Dx())),
		TopAdvance: self.font.Metrics().Ascent(),
		BottomAdvance: self.font.Metrics().Descent(),
		HorzCenter: uint8(min(255, bounds.Dx()/2)),
	}
	return self.AddGlyphWithPlacement(mask, placement)
}

// Like [Strand.AddGlyph](), but with customizable placement.
func (self *Strand) AddGlyphWithPlacement(mask core.GlyphMask, placement ggfnt.GlyphPlacement) (ggfnt.GlyphIndex, error) {
	index := int(ggfnt.GlyphCustomMin) + len(self.customGlyphs)
	if index > int(ggfnt.GlyphCustomMax) { return 0, errors.New("too many custom glyphs") }
	panic("unimplemented")
	// for i := 0; i < len(mask.Pix); i++ {
	// 	if mask.Pix[i] >= self.minColorIndex { continue }
	// 	if mask.Pix[i] == 0 { continue }
	// 	return 0, errors.New("mask uses values outside the range of the font colors")
	// }
	// self.customGlyphs = append(self.customGlyphs, mask)
	// return ggfnt.GlyphIndex(index), nil
}

// ---- color ----

// Replaces the colors of the given palette with custom ones.
//
// The method will panic if the given palette key is not valid
// or the number of given colors does not match the palette
// size.
func (self *Strand) Recolor(paletteKey ggfnt.PaletteKey, colors ...color.RGBA) {
	// get palette range and check sizes
	if self.font == nil { panic("font is nil, can't overwrite any palettes") }
	paletteSize := self.font.Color().NumPaletteColors(paletteKey)
	if len(colors) != int(paletteSize) {
		panic("number of colors does not match palette size")
	}

	// ensure all colors are premultiplied
	for i, _ := range colors { // discretionary safety check
		if isPremultiplied(colors[i]) { continue }
		panic(nonPremultRGBA)
	}

	// apply each color
	firstPaletteIndex := self.font.Color().NumDyeIndices()
	fontColorIndex := firstPaletteIndex + uint8(paletteKey)
	for i, _ := range colors {
		self.fontColors.Set(int(fontColorIndex), internal.RGBAToFloat32(colors[i]))
		fontColorIndex += 1
	}

	// notify changes
	self.notifyShaderPaletteChange()
}

// ---- helpers ----

func (self *Strand) setFlag(bit uint8, on bool) {
	if on { self.flags |= bit } else { self.flags &= ^bit }
}

func (self *Strand) getFlag(bit uint8) bool {
	return self.flags & bit != 0
}

