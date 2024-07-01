# ptxt
[![Go Reference](https://pkg.go.dev/badge/tinne26/ptxt.svg)](https://pkg.go.dev/github.com/tinne26/ptxt)

A [ggfnt](https://github.com/tinne26/ggfnt)-compatible pixel font renderer for the [Ebitengine](https://ebitengine.org) game engine.

> [!NOTE]
> This package is still in development. Most basic functionality is already there, but many advanced features are  missing and tests are still very limited. There might also be quite a few bugs, some features left unimplemented without warning and a few more caveats.

## Code example

```Golang
package main

import ( "math" ; "image/color" )
import "github.com/hajimehoshi/ebiten/v2"
import "github.com/tinne26/ptxt"
import "github.com/tinne26/ggfnt-fonts/jammy"

const CanvasWidth, CanvasHeight = 80, 45 // 1/24th of 1920x1080
const WordsPerSec = 2.71828
var Words = []string {
	"PIXEL", "BLOCK", "RETRO", "WORLD", "CLICKY", "GAME",
	"SHARP", "CONTROL", "SIMPLE", "PLAIN", "COLOR", "PALETTE",
}

// ---- Ebitengine's Game interface implementation ----

type Game struct {
	canvas *ebiten.Image // logical canvas
	text *ptxt.Renderer
	wordIndex float64
}

func (*Game) Layout(int, int) (int, int) { panic("F") }
func (*Game) LayoutF(logicWinWidth, logicWinHeight float64) (float64, float64) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	return logicWinWidth*scale, logicWinHeight*scale
}

func (self *Game) Update() error {
	self.wordIndex += WordsPerSec/float64(ebiten.TPS()))
	self.wordIndex  = math.Mod(self.wordIndex, float64(len(Words)))
	return nil
}

func (self *Game) Draw(hiResCanvas *ebiten.Image) {
	// fill background
	self.canvas.Fill(color.RGBA{246, 242, 240, 255})

	// draw text on bottom left corner
	word := Words[int(self.wordIndex)]
	self.text.Draw(self.canvas, word, 6, CanvasHeight - 6)

	// project from logical canvas to high-res (optional ptxt utility)
	ptxt.PixelPerfect.Project(self.canvas, hiResCanvas)
}

// ---- main function ----

func main() {
	// initialize font strand
	strand, err := ptxt.NewStrand(jammy.Font())
	if err != nil { panic(err) }
	
	// create text renderer, set the main properties
	renderer := ptxt.NewRenderer()
	renderer.SetStrand(strand)
	renderer.SetAlign(ptxt.Baseline | ptxt.Left)
	renderer.SetColor(color.RGBA{242, 143, 59, 255})

	// set up Ebitengine and start the game
	ebiten.SetWindowTitle("ptxt-examples/gpu/words")
	canvas := ebiten.NewImage(CanvasWidth, CanvasHeight)
	err = ebiten.RunGame(&Game{ text: renderer, canvas: canvas })
	if err != nil { panic(err) }
}
```
*(You can run the [WASM version](https://tinne26.github.io/ptxt-examples/words) of this example directly in your browser if you want)*

This particular example uses a high resolution layout and then projects from the logical canvas to the screen with `ptxt.PixelPerfect.Project()` (consider also `Proportional`). This model of "render on a logical canvas, project afterwards" is the recommended way to work in most cases.

More examples are available at [tinne26/ptxt-examples](https://github.com/tinne26/ptxt-examples).

## Where to get fonts

You can find a few at [tinne26/ggfnt-fonts](https://github.com/tinne26/ggfnt-fonts).

## How to run on CPU

Font glyphs are always created on the CPU, but by default they are rendered to the target surface with Ebitengine (GPU images). For testing and some image creation processes, though, sometimes it's interesting to do the rendering directly on the CPU. This can be done using the `cputext` tag (e.g. `go run -tags cputext main.go`).

## Bitmap fonts vs vectorial fonts

Before writing **ptxt**, a bitmap font renderer, I wrote [**etxt**](https://github.com/tinne26/etxt), a vectorial font renderer for Ebitengine. The two renderers share many similarities, but **ptxt** is not meant for *scalable UIs*, which allows its model to be slightly simpler; we don't have to worry about text size vs [scaling](https://github.com/tinne26/etxt/blob/main/docs/display-scaling.md), quantization, fractional positioning and others.

While **etxt** [can also be used for pixel fonts](https://github.com/tinne26/etxt/blob/main/docs/pixel-tips.md) in `ttf` or `otf` formats, **ptxt** is a more natural and specialized choice for that. That being said, the documentation on **etxt** is much more representative of general font usage and can be more educative in that respect. The code on **etxt** is also more general and optimized.

One important characteristic of **ptxt** is that it was built to support the [ggfnt](https://github.com/tinne26/ggfnt) format, a custom pixel font format that I created for indie game development. This format comes with its own set of advantages and drawbacks, which are detailed on its own repository.

