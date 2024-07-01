// ptxt is a package for pixel art font rendering designed to be used with
// Ebitengine, a 2D game engine made by Hajime Hoshi for Golang.
//
// To get started, you should initialize a font [*strand.Strand] and a [*Renderer]:
//   strand, err := ptxt.NewStrand(font) // font can be *ggfnt.Font, []byte, filename...
//   if err != nil { panic(err) }
//   
//   text := ptxt.NewRenderer() // create the text renderer
//   text.SetStrand(strand) // link the font strand to the renderer
//   text.SetAlign(ptxt.Center)
//   text.SetColor(color.RGBA{192, 0, 255, 255})
//
// As shown above, you can adjust the basic renderer properties with functions like
// [Renderer.SetColor](), [Renderer.SetScale](), [Renderer.SetAlign]() and many others.
//
// Once you have everything configured to your liking, drawing is quite straightforward:
//    text.Draw(canvas, "TEXT IS ME", x, y)
//
// To learn more, make sure to read the docs and check the [examples]!
//
// [examples]: https://github.com/tinne26/ptxt-examples
package ptxt
