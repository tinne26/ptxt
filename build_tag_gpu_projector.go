//go:build !cputext

package ptxt

import "math"
import "image"
import "strconv"

import "github.com/hajimehoshi/ebiten/v2"

var projectShaderOpts ebiten.DrawTrianglesShaderOptions
var projectShaderVerts []ebiten.Vertex
var projectShaderIndices []uint16 = []uint16{0, 1, 2, 2, 1, 3}
var projectBilinearShader *ebiten.Shader
var projectBilinearShaderSrc []byte = []byte(`
//kage:unit pixels
package main

func Fragment(_ vec4, sourceCoords vec2, _ vec4) vec4 {
	unit := fwidth(sourceCoords)
	halfUnit := unit/2.0
	minCoords, maxCoords := getMinMaxSourceCoords()
	tl := imageSrc0UnsafeAt(clamp(sourceCoords - halfUnit, minCoords, maxCoords))
	tr := imageSrc0UnsafeAt(clamp(sourceCoords + vec2(+halfUnit.x, -halfUnit.y), minCoords, maxCoords))
	bl := imageSrc0UnsafeAt(clamp(sourceCoords + vec2(-halfUnit.x, +halfUnit.y), minCoords, maxCoords))
	br := imageSrc0UnsafeAt(clamp(sourceCoords + halfUnit, minCoords, maxCoords))
	delta  := min(fract(sourceCoords + halfUnit), unit)/unit
	return mix(mix(tl, tr, delta.x), mix(bl, br, delta.x), delta.y)
}

func getMinMaxSourceCoords() (vec2, vec2) {
	origin := imageSrc0Origin()
	return origin, origin + imageSrc0Size() - vec2(1.0/65536.0)
}
`)

// A helper type for projection between logical and high
// solution canvases. See [Projector.Project]() for more
// context.
type Projector uint8
const (
	// Proportional scaling respects the aspect ratio of the logical
	// canvas, which means that unfilled borders might be left on the
	// high resolution canvas.
	Proportional Projector = iota

	// Integer scaling with no deformation nor distortions (unless
	// minification is required, in which case the algorithm falls
	// back to the same behavior as Proportional).
	PixelPerfect

	// Fills the whole high resolution canvas without any regards for
	// aspect ratio. No borders are left on the high resolution canvas.
	Stretched
)

// Returns a textual representation of the projector type.
func (self Projector) String() string {
	switch self {
	case Proportional: return "Proportional"
	case PixelPerfect: return "PixelPerfect"
	case Stretched: return "Stretched"
	default:
		return "ProjectorUndefined#" + strconv.Itoa(int(self))
	}
}

// Utility method, commonly used to remap cursor coordinates from
// a high resolution layout to the logical space for your game.
//
// This function panics if fromWidth or fromHeight are zero and
// the corresponding to* value is non-zero (typically indicative
// of programmer mistake).
func (self Projector) Remap(x, y int, fromWidth, fromHeight, toWidth, toHeight int) (int, int) {
	// safety checks
	if fromWidth == 0 && toWidth != 0 {
		panic("Projector.Remap(): fromWidth == 0 && toWidth != 0")
	}
	if fromHeight == 0 && toHeight != 0 {
		panic("Projector.Remap(): fromHeight == 0 && toHeight != 0")
	}

	// convert parameters to float64
	fx, fy     := float64(x), float64(y)
	fw64, fh64 := float64(fromWidth), float64(fromHeight)
	tw64, th64 := float64(  toWidth), float64(  toHeight)

	var outX, outY float64
	switch self {
	case Proportional:
		outX, outY = remapProportional(fx, fy, fw64, fh64, tw64, th64)
	case PixelPerfect:
		outX, outY = remapPixelPerfect(fx, fy, fw64, fh64, tw64, th64)
	case Stretched:
		outX, outY = remapStretched(fx, fy, fw64, fh64, tw64, th64)
	default:
		panic("invalid " + self.String())
	}

	return int(outX), int(outY)
}

func remapProportional(x, y, fw64, fh64, tw64, th64 float64) (float64, float64) {	
	// compute margins
	var xMargin, yMargin float64
	fromAspectRatio, toAspectRatio := fw64/fh64, tw64/th64
	switch {
	case fromAspectRatio > toAspectRatio: // horz margins
		xMargin = (fw64 - toAspectRatio*fh64)/2.0
	case toAspectRatio > fromAspectRatio: // vert margins
		yMargin = (fh64 - fw64/toAspectRatio)/2.0
	}
	
	// compute relative x/y and project
	relX := (x - xMargin)/(fw64 - xMargin*2.0)
	relY := (y - yMargin)/(fh64 - yMargin*2.0)
	return clamp(relX, 0.0, 1.0)*tw64, clamp(relY, 0.0, 1.0)*th64
}

func remapPixelPerfect(x, y, fw64, fh64, tw64, th64 float64) (float64, float64) {	
	// get zoom factors
	xZoom, yZoom := fw64/tw64, fh64/th64
	zoomLevel := math.Min(xZoom, yZoom)
	if zoomLevel < 1.0 { // proportional case
		return remapProportional(x, y, fw64, fh64, tw64, th64)
	}
	zoomLevel = math.Floor(zoomLevel)

	// compute margins
	xMargin, yMargin := (fw64 - tw64*zoomLevel)/2.0, (fh64 - th64*zoomLevel)/2.0
	
	// compute relative x/y and project
	relX := (x - xMargin)/(fw64 - xMargin*2.0)
	relY := (y - yMargin)/(fh64 - yMargin*2.0)
	return clamp(relX, 0.0, 1.0)*tw64, clamp(relY, 0.0, 1.0)*th64
}

func remapStretched(x, y, fw64, fh64, tw64, th64 float64) (float64, float64) {
	return (tw64*x)/fw64, (th64*x)/fh64
}

// The projector type is intended to make life easier when implementing
// simple pixel art games. In general, ptxt should only be used on logical
// canvases of known sizes; once you are done drawing the text, you want
// to project your logical canvas onto the higher resolution screen. While
// Ebitengine can already do that for you automatically if you are using a
// fixed layout, doing it manually has some advantages:
//  - You can render additional high resolution content behind and in front
//    of the logical canvas.
//  - You can let the player select different scaling models and switch
//    between them more easily.
//
// The returned value is a subimage of 'fullResCanvas' containing the area
// of the screen where the logical canvas has been projected to. In some
// cases it can be 'fullCanvas' itself.
// 
// Example of basic usage on [ebiten.Game].Draw():
//   func (self *Game) Draw(hiResCanvas *ebiten.Image) {
//       self.canvas.Fill(color.RGBA{128, 128, 128, 255}) // logical canvas background
//       self.text.Draw(self.canvas, "HELLO WORLD!", GameWidth/2, GameHeight/2) // text
//       ptxt.Proportional.Project(self.canvas, hiResCanvas) // scaling
//   }
func (self Projector) Project(logicalCanvas, fullResCanvas *ebiten.Image) *ebiten.Image {
	switch self {
	case Proportional:
		return projectProportional(logicalCanvas, fullResCanvas)
	case PixelPerfect:
		return projectPixelPerfect(logicalCanvas, fullResCanvas)
	case Stretched:
		projectStretched(logicalCanvas, fullResCanvas)
		return fullResCanvas
	default:
		panic("invalid Projector '" + self.String() + "'")
	}
}

func projectProportional(logicalCanvas, canvas *ebiten.Image) *ebiten.Image {
	logicalBounds, canvasBounds := logicalCanvas.Bounds(), canvas.Bounds()
	logicalWidth, logicalHeight := logicalBounds.Dx(), logicalBounds.Dy()
	canvasWidth, canvasHeight := canvasBounds.Dx(), canvasBounds.Dy()

	// trivial case: both screens have the same size
	if logicalWidth == canvasWidth && logicalHeight == canvasHeight {
		projectNearest(logicalCanvas, canvas)
		return canvas
	}

	// get aspect ratios
	logicalAspectRatio := float64(logicalWidth)/float64(logicalHeight)
	canvasAspectRatio  := float64(canvasWidth)/float64(canvasHeight)

	// compare aspect ratios	
	if logicalAspectRatio == canvasAspectRatio {
		// simple case, aspect ratios match, only scaling is necessary
		scalingFactor := float64(canvasWidth)/float64(logicalWidth)
		if scalingFactor - float64(int(scalingFactor)) == 0 {
			projectNearest(logicalCanvas, canvas)
		} else {
			projectBilinear(logicalCanvas, canvas, 0, 0)
		}
		return canvas
	} else {
		// aspect ratios don't match, must also apply translation
		if canvasAspectRatio < logicalAspectRatio {
			// (we have excess canvas height)
			adjustedCanvasHeight := float64(canvasWidth)/logicalAspectRatio
			yMargin := (float64(canvasHeight) - adjustedCanvasHeight)/2.0
			yMarginWhole, yMarginFract := math.Modf(yMargin)
			minY := canvasBounds.Min.Y + int(yMarginWhole)
			maxY := canvasBounds.Max.Y - int(yMarginWhole)
			subRect := image.Rect(canvasBounds.Min.X, minY, canvasBounds.Max.X, maxY)
			subTarget := canvas.SubImage(subRect).(*ebiten.Image)
			projectBilinear(logicalCanvas, subTarget, 0, float32(yMarginFract))
			return subTarget
		} else { // canvasAspectRatio > logicalAspectRatio
			// (we have excess canvas width)
			adjustedCanvasWidth := float64(canvasHeight)*logicalAspectRatio
			xMargin := (float64(canvasWidth) - adjustedCanvasWidth)/2.0
			xMarginWhole, xMarginFract := math.Modf(xMargin)
			minX := canvasBounds.Min.X + int(xMarginWhole)
			maxX := canvasBounds.Max.X - int(xMarginWhole)
			subRect := image.Rect(minX, canvasBounds.Min.Y, maxX, canvasBounds.Max.Y)
			subTarget := canvas.SubImage(subRect).(*ebiten.Image)
			projectBilinear(logicalCanvas, subTarget, float32(xMarginFract), 0)
			return subTarget
		}
	}
}

// Unless you are on macOS, of course.
func projectPixelPerfect(logicalCanvas, canvas *ebiten.Image) *ebiten.Image {
	logicalBounds, canvasBounds := logicalCanvas.Bounds(), canvas.Bounds()
	logicalWidth, logicalHeight := logicalBounds.Dx(), logicalBounds.Dy()
	canvasWidth , canvasHeight  := canvasBounds.Dx(), canvasBounds.Dy()

	// trivial case: both screens have the same size
	if logicalWidth == canvasWidth && logicalHeight == canvasHeight {
		projectNearest(logicalCanvas, canvas)
		return canvas
	}

	// get zoom levels
	var tx, ty int = canvasBounds.Min.X, canvasBounds.Min.Y
	xZoom := float64(canvasWidth)/float64(logicalWidth)
	yZoom := float64(canvasHeight)/float64(logicalHeight)
	zoomLevel := math.Min(xZoom, yZoom)
	var outWidth, outHeight int
	if zoomLevel < 1.0 {
		// minification (we switch to bilinear filtering)
		outWidth  = int(float64(logicalWidth)*zoomLevel)
		outHeight = int(float64(logicalHeight)*zoomLevel)
		tx += (canvasWidth - outWidth) >> 1
		ty += (canvasHeight - outHeight) >> 1
		subRect := image.Rect(tx, ty, tx + outWidth, ty + outHeight)
		targetSubCanvas := canvas.SubImage(subRect).(*ebiten.Image)
		projectBilinear(logicalCanvas, targetSubCanvas, 0, 0)
		return targetSubCanvas
	} else {
		// integer scaling
		intZoomLevel := int(zoomLevel)
		outWidth  = logicalWidth*intZoomLevel
		outHeight = logicalHeight*intZoomLevel
		tx += (canvasWidth - outWidth) >> 1
		ty += (canvasHeight - outHeight) >> 1
		subRect := image.Rect(tx, ty, tx + outWidth, ty + outHeight)
		targetSubCanvas := canvas.SubImage(subRect).(*ebiten.Image)
		projectNearest(logicalCanvas, targetSubCanvas)
		return targetSubCanvas
	}
}

func projectStretched(logicalCanvas, canvas *ebiten.Image) {
	projectBilinear(logicalCanvas, canvas, 0, 0)
}

func projectBilinear(logicalCanvas, targetSubCanvas *ebiten.Image, xMargin, yMargin float32) {
	if projectBilinearShader == nil {
		var err error
		projectBilinearShader, err = ebiten.NewShader(projectBilinearShaderSrc)
		if err != nil { panic(err) }
	}
	projectSetTriangleVertices(logicalCanvas, targetSubCanvas, xMargin, yMargin)
	projectShaderOpts.Images[0] = logicalCanvas
	targetSubCanvas.DrawTrianglesShader(projectShaderVerts, projectShaderIndices, projectBilinearShader, &projectShaderOpts)
}

func projectNearest(logicalCanvas, targetSubCanvas *ebiten.Image) {
	projectSetTriangleVertices(logicalCanvas, targetSubCanvas, 0, 0)
	targetSubCanvas.DrawTriangles(projectShaderVerts, projectShaderIndices, logicalCanvas, nil)
}

func projectSetTriangleVertices(logicalCanvas, targetSubCanvas *ebiten.Image, xMargin, yMargin float32) {
	// set projectShaderVerts
	if projectShaderVerts == nil {
		projectShaderVerts = make([]ebiten.Vertex, 4)
		for i := 0; i < 4; i++ {
			projectShaderVerts[i].ColorR = 1.0
			projectShaderVerts[i].ColorG = 1.0
			projectShaderVerts[i].ColorB = 1.0
			projectShaderVerts[i].ColorA = 1.0
		}
	}

	bounds := targetSubCanvas.Bounds()
	projectShaderVerts[0].DstX = float32(bounds.Min.X) + xMargin // top-left
	projectShaderVerts[0].DstY = float32(bounds.Min.Y) + yMargin // top-left
	projectShaderVerts[1].DstX = float32(bounds.Max.X) - xMargin // top-right
	projectShaderVerts[1].DstY = float32(bounds.Min.Y) + yMargin // top-right
	projectShaderVerts[2].DstX = float32(bounds.Min.X) + xMargin // bottom-left
	projectShaderVerts[2].DstY = float32(bounds.Max.Y) - yMargin // bottom-left
	projectShaderVerts[3].DstX = float32(bounds.Max.X) - xMargin // bottom-right
	projectShaderVerts[3].DstY = float32(bounds.Max.Y) - yMargin // bottom-right

	bounds = logicalCanvas.Bounds()
	projectShaderVerts[0].SrcX = float32(bounds.Min.X) // top-left
	projectShaderVerts[0].SrcY = float32(bounds.Min.Y) // top-left
	projectShaderVerts[1].SrcX = float32(bounds.Max.X) // top-right
	projectShaderVerts[1].SrcY = float32(bounds.Min.Y) // top-right
	projectShaderVerts[2].SrcX = float32(bounds.Min.X) // bottom-left
	projectShaderVerts[2].SrcY = float32(bounds.Max.Y) // bottom-left
	projectShaderVerts[3].SrcX = float32(bounds.Max.X) // bottom-right
	projectShaderVerts[3].SrcY = float32(bounds.Max.Y) // bottom-right
}
