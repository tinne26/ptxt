//go:build !cputext

package strand

import "bytes"

import "github.com/tinne26/ptxt/core"
import "github.com/tinne26/ggfnt"
import "github.com/hajimehoshi/ebiten/v2"

// TODO: consider optimizing the single main dye case at 255 with a shared shader and so on
// TODO: consider minifying the shader source

var paletteDrawerShaderSource []byte = []byte(`
//kage:unit pixels
package main

const LenDyeMappings = {1}
const LenDyeColors   = {2}
const LenPalette     = {3}

// --- uniforms ---

// dye mappings and main dye key are static
var DyeMappings [LenDyeMappings]int // (because each dye can have multiple alpha values)
var MainDyeKey int // this is static per strand

// dye colors might change from time to time if we 
// have multiple dyes instead of only one main dye
var DyeColors [LenDyeColors]vec4 // rgba slice

// the main palette is unlikely to change more than once
var Palette [LenPalette]vec4 // rgba slice
var NumUsedPaletteEntries int // this is static per strand

// fragment shader entry point
func Fragment(targetCoords vec4, sourceCoords vec2, mainDye vec4) vec4 {
	fontColorIndex := int(imageSrc0UnsafeAt(sourceCoords).a*255.0)
	if fontColorIndex == 0 { discard() } // transparent
	fontColorIndex -= 1

	if fontColorIndex < len(DyeMappings) { // dye color
		dyeKey := DyeMappings[fontColorIndex]
		var dyeColor vec4
		if dyeKey == MainDyeKey {
			dyeColor = mainDye
		} else {
			dyeColor = DyeColors[dyeKey]
		}

		return dyeColor*Palette[fontColorIndex]
	} else { // static color
		return Palette[fontColorIndex]
	}
}
`)

// Shader uniforms, as inferred from palette_drawer.kage:
// > DyeMappings []uint8 // from dye alpha index to dye key
// > MainDyeKey uint8
// > DyeColors []vec4
// > Palette []vec4 // (dyes take x4 the required space, but we don't add unused colors)
// > NumUsedPaletteEntries uint8

type renderData struct {
	shaderOptions ebiten.DrawTrianglesShaderOptions
	shader *ebiten.Shader
	shaderVertices [4]ebiten.Vertex
}

// must be called at the end of strand init
func (self *Strand) renderDataInit() {
	var uint8To3Bytes = func(value uint8) (byte, byte, byte) {
		d3 := '0' + (value % 10)
		if value < 10 { return ' ', ' ', d3 }
		value /= 10
		d2 := '0' + (value % 10)
		if value < 10 { return ' ', d2, d3 }
		value /= 10
		d1 := '0' + (value % 10)
		return d1, d2, d3
	}

	// helper variable needed later
	numDyeIndices := self.font.Color().NumDyeIndices()
	
	// compile shader
	var err error
	arg1Pos := bytes.Index(paletteDrawerShaderSource, []byte{'{', '1', '}'})
	arg2Pos := bytes.Index(paletteDrawerShaderSource, []byte{'{', '2', '}'})
	arg3Pos := bytes.Index(paletteDrawerShaderSource, []byte{'{', '3', '}'})
	if arg1Pos == -1 || arg2Pos == -1 { panic(brokenCode) }
	
	b1, b2, b3 := uint8To3Bytes(numDyeIndices)
	paletteDrawerShaderSource[arg1Pos + 0] = b1
	paletteDrawerShaderSource[arg1Pos + 1] = b2
	paletteDrawerShaderSource[arg1Pos + 2] = b3

	b1, b2, b3  = uint8To3Bytes(uint8(self.dyes.Len()))
	paletteDrawerShaderSource[arg2Pos + 0] = b1
	paletteDrawerShaderSource[arg2Pos + 1] = b2
	paletteDrawerShaderSource[arg2Pos + 2] = b3

	b1, b2, b3  = uint8To3Bytes(uint8(self.fontColors.Len()))
	paletteDrawerShaderSource[arg3Pos + 0] = b1
	paletteDrawerShaderSource[arg3Pos + 1] = b2
	paletteDrawerShaderSource[arg3Pos + 2] = b3

	self.re.shader, err = ebiten.NewShader(paletteDrawerShaderSource)
	
	paletteDrawerShaderSource[arg1Pos + 0] = '{'
	paletteDrawerShaderSource[arg1Pos + 1] = '1'
	paletteDrawerShaderSource[arg1Pos + 2] = '}'
	paletteDrawerShaderSource[arg2Pos + 0] = '{'
	paletteDrawerShaderSource[arg2Pos + 1] = '2'
	paletteDrawerShaderSource[arg2Pos + 2] = '}'
	paletteDrawerShaderSource[arg3Pos + 0] = '{'
	paletteDrawerShaderSource[arg3Pos + 1] = '3'
	paletteDrawerShaderSource[arg3Pos + 2] = '}'
	
	if err != nil { panic("Kage shader compilation error:\n" + err.Error()) }

	// set uniforms
	self.re.shaderOptions.Uniforms = make(map[string]interface{}, 5)
	self.re.shaderOptions.Uniforms["MainDyeKey"] = self.mainDyeKey
	self.re.shaderOptions.Uniforms["NumUsedPaletteEntries"] = self.font.Color().Count()
	
	// dye mappings point from font color index to dye index.
	// the actual alphas are stored in the palette instead
	dyeMappings := make([]int32, numDyeIndices)
	var fontColorIndex uint8
	for i := uint8(0); i < self.font.Color().NumDyes(); i++ {
		alphaCount := self.font.Color().NumDyeAlphas(ggfnt.DyeKey(i))
		if alphaCount == 0 { panic(brokenCode) } // or broken font
		for j := uint8(0); j < alphaCount; j++ {
			dyeMappings[fontColorIndex] = int32(i)
			fontColorIndex += 1
		}
		if fontColorIndex > 255 { panic(brokenCode) }
	}
	self.re.shaderOptions.Uniforms["DyeMappings"] = dyeMappings

	self.notifyShaderNonMainDyeChange()
	self.notifyShaderPaletteChange()
}

func (self *Strand) setBlendMode(blend ebiten.Blend) {
	self.re.shaderOptions.Blend = blend
}

// must be called any time a dye color is changed
func (self *Strand) notifyShaderNonMainDyeChange() {
	self.re.shaderOptions.Uniforms["DyeColors"] = self.dyes.data
}

// must be called any time a palette range is changed
func (self *Strand) notifyShaderPaletteChange() {
	self.re.shaderOptions.Uniforms["Palette"] = self.fontColors.data
}

// Precondition: scale is already overriden from the strand if necessary.
func (self *Strand) drawHorzMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	// configure the vertices
	srcRect := mask.Bounds()
	scMinX, scMinY := srcRect.Min.X*scale, srcRect.Min.Y*scale
	scMaxX, scMaxY := srcRect.Max.X*scale, srcRect.Max.Y*scale
	self.setShaderDstVertices(float32(x + scMinX), float32(y + scMinY), float32(x + scMaxX), float32(y + scMaxY))
	self.setShaderSrcVertices(float32(srcRect.Min.X), float32(srcRect.Min.Y), float32(srcRect.Max.X), float32(srcRect.Max.Y))
	
	// configure the main dye color
	self.setShaderVertColors(rgba)

	// invoke the shader
	self.re.shaderOptions.Images[0] = mask
	target.DrawTrianglesShader(self.re.shaderVertices[:], []uint16{0, 1, 2, 2, 1, 3}, self.re.shader, &self.re.shaderOptions)
}

// Precondition: scale is already overriden from the strand if necessary.
func (self *Strand) drawSidewaysMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	// configure the vertices
	srcRect := mask.Bounds()
	scMinX, scMinY := srcRect.Min.X*scale, srcRect.Min.Y*scale
	scMaxX, scMaxY := srcRect.Max.X*scale, srcRect.Max.Y*scale
	self.setShaderSidewaysDstVertices(float32(x + scMinY), float32(y - scMinX), float32(x + scMaxY), float32(y - scMaxX))
	self.setShaderSrcVertices(float32(srcRect.Min.X), float32(srcRect.Min.Y), float32(srcRect.Max.X), float32(srcRect.Max.Y))
	
	// configure the main dye color
	self.setShaderVertColors(rgba)

	// invoke the shader
	self.re.shaderOptions.Images[0] = mask
	target.DrawTrianglesShader(self.re.shaderVertices[:], []uint16{0, 1, 2, 2, 1, 3}, self.re.shader, &self.re.shaderOptions)
}

// Precondition: scale is already overriden from the strand if necessary.
func (self *Strand) drawSidewaysRightMask(target core.Target, mask core.GlyphMask, x, y int, scale int, rgba [4]float32) {
	// configure the vertices
	srcRect := mask.Bounds()
	scMinX, scMinY := srcRect.Min.X*scale, srcRect.Min.Y*scale
	scMaxX, scMaxY := srcRect.Max.X*scale, srcRect.Max.Y*scale
	self.setShaderSidewaysDstVertices(float32(x - scMinY), float32(y + scMinX), float32(x - scMaxY), float32(y + scMaxX))
	self.setShaderSrcVertices(float32(srcRect.Min.X), float32(srcRect.Min.Y), float32(srcRect.Max.X), float32(srcRect.Max.Y))
	
	// configure the main dye color
	self.setShaderVertColors(rgba)

	// invoke the shader
	self.re.shaderOptions.Images[0] = mask
	target.DrawTrianglesShader(self.re.shaderVertices[:], []uint16{0, 1, 2, 2, 1, 3}, self.re.shader, &self.re.shaderOptions)
}

func (self *Strand) setShaderSrcVertices(minX, minY, maxX, maxY float32) {
	// (0 = top-left, 1 = top-right, 2 = bottom-left, 3 = bottom-right)
	self.re.shaderVertices[0].SrcX = minX
	self.re.shaderVertices[0].SrcY = minY
	self.re.shaderVertices[1].SrcX = maxX
	self.re.shaderVertices[1].SrcY = minY
	self.re.shaderVertices[2].SrcX = minX
	self.re.shaderVertices[2].SrcY = maxY
	self.re.shaderVertices[3].SrcX = maxX
	self.re.shaderVertices[3].SrcY = maxY
}

func (self *Strand) setShaderDstVertices(minX, minY, maxX, maxY float32) {
	// (0 = top-left, 1 = top-right, 2 = bottom-left, 3 = bottom-right)
	self.re.shaderVertices[0].DstX = minX
	self.re.shaderVertices[0].DstY = minY
	self.re.shaderVertices[1].DstX = maxX
	self.re.shaderVertices[1].DstY = minY
	self.re.shaderVertices[2].DstX = minX
	self.re.shaderVertices[2].DstY = maxY
	self.re.shaderVertices[3].DstX = maxX
	self.re.shaderVertices[3].DstY = maxY
}

func (self *Strand) setShaderSidewaysDstVertices(minX, minY, maxX, maxY float32) {
	self.re.shaderVertices[0].DstX = minX
	self.re.shaderVertices[0].DstY = minY
	self.re.shaderVertices[1].DstX = minX
	self.re.shaderVertices[1].DstY = maxY
	self.re.shaderVertices[2].DstX = maxX
	self.re.shaderVertices[2].DstY = minY
	self.re.shaderVertices[3].DstX = maxX
	self.re.shaderVertices[3].DstY = maxY
}

func (self *Strand) setShaderVertColors(rgba [4]float32) {
	for i := 0; i < 4; i++ {
		self.re.shaderVertices[i].ColorR = rgba[0]
		self.re.shaderVertices[i].ColorG = rgba[1]
		self.re.shaderVertices[i].ColorB = rgba[2]
		self.re.shaderVertices[i].ColorA = rgba[3]
	}

	// useful for debug
	// self.re.shaderVertices[0].ColorR = 1.0
	// self.re.shaderVertices[0].ColorG = 0.0
	// self.re.shaderVertices[0].ColorB = 0.0
	// self.re.shaderVertices[0].ColorA = 1.0
	// self.re.shaderVertices[1].ColorR = 0.0
	// self.re.shaderVertices[1].ColorG = 1.0
	// self.re.shaderVertices[1].ColorB = 0.0
	// self.re.shaderVertices[1].ColorA = 1.0
	// self.re.shaderVertices[2].ColorR = 0.0
	// self.re.shaderVertices[2].ColorG = 0.0
	// self.re.shaderVertices[2].ColorB = 1.0
	// self.re.shaderVertices[2].ColorA = 1.0
	// self.re.shaderVertices[3].ColorR = 0.0
	// self.re.shaderVertices[3].ColorG = 1.0
	// self.re.shaderVertices[3].ColorB = 1.0
	// self.re.shaderVertices[3].ColorA = 1.0
}
