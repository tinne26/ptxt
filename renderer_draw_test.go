//go:build cputext
package ptxt

import "testing"

import "image"

func TestDrawWrap(t *testing.T) {
	if testFont == nil { t.SkipNow() }

	// create strand and renderer
	strand, _ := NewStrand(testFont)
	renderer := NewRenderer()
	renderer.SetStrand(strand)
	renderer.SetAlign(Top | Left)

	w, h := renderer.Measure("HELLO\nHELLO")
	target1 := image.NewRGBA(image.Rect(0, 0, w, h))
	target2 := image.NewRGBA(image.Rect(0, 0, w, h))
	renderer.Draw(target1, "HELLO\nHELLO", 0, 0)
	if len(renderer.run.wrapIndices) != 0 {
		t.Fatal("expected wrap indices to be empty")
	}
	renderer.DrawWithWrap(target2, "HELLO HELLO", 0, 0, w)
	if len(renderer.run.wrapIndices) == 0 {
		t.Fatal("expected wrap indices to not be empty")
	}
	if !equalSlices(target1.Pix, target2.Pix) {
		outFilename1 := "testfail_draw_wrap_target1.png"
		outFilename2 := "testfail_draw_wrap_target2.png"
		exportAsPNG(outFilename1, target1)
		exportAsPNG(outFilename2, target2)
		t.Fatalf("wrap draw not matching manual break draw, exported to '%s', '%s'", outFilename1, outFilename2)
	}
}
