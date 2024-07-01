//go:build cputext
package ptxt

import "testing"

func TestMeasureMask(t *testing.T) {
	ensureTestAssetsLoaded()
	if testFont == nil { t.SkipNow() }

	// create strand and renderer
	strand, _ := NewStrand(testFont)
	renderer := NewRenderer()
	renderer.SetStrand(strand)
	renderer.Advanced().SetBoundingMode(MaskBounding)

	// loop and test basic measuring for different aligns
	for _, align := range []Align{ (Baseline | Left), (Bottom | Right), Center } {
		// TODO: add scale tests, see that it's the exact same results but multiplied

		// configure renderer with current params
		renderer.SetAlign(align)

		// check zero measure
		zw, zh := renderer.Measure("")
		if zw != 0 || zh != 0 {
			t.Fatal("expected zero with and height")
		}

		zw, zh = renderer.Measure("\n")
		if zw != 0 || zh != 0 {
			t.Fatal("expected zero with and height")
		}

		zw, zh = renderer.Measure(" \n ")
		if zw != 0 || zh != 0 {
			t.Fatal("expected zero with and height")
		}

		zw, zh = renderer.Measure("\n\n \n")
		if zw != 0 || zh != 0 {
			t.Fatal("expected zero with and height")
		}

		// consistency tests	
		w1, h1 := renderer.Measure("HEY H")
		w2, h2 := renderer.Measure("HEY HO")
		w3, h3 := renderer.Measure("HEY HOO")
		w4,  _ := renderer.Measure("HEY HO.HEY HO")
		w5, h5 := renderer.Measure("HEY HO.HEY HO \n")
		if w3 >= w1*2 {
			t.Fatalf("expected w3 < w1*2, but got w3 = %d, w1 = %d", w3, w1)
		}
		if w1 >= w2 {
			t.Fatalf("expected w1 < w2, but got w2 = %d, w1 = %d", w2, w1)
		}
		if w3 <= w2 {
			t.Fatalf("expected w3 > w2, but got w3 = %d, w2 = %d", w3, w2)
		}
		if h1 != h2 || h2 != h3 {
			t.Fatalf("inconsistent heights (%d, %d, %d)", h1, h2, h3)
		}
		if w4 <= w2*2 {
			t.Fatalf("expected w4 > w2*2, but got w4 = %d, w2 = %d", w4, w2)
		}

		// consistency check
		if w5 != w4 || h5 != h1 {
			t.Fatalf("expected w5, h5 = w4, h1 (%d, %d vs %d, %d)", w5, h5, w4, h1)
		}

		// dot height test
		_, dh := renderer.Measure(".")
		if dh == h1 {
			t.Fatalf("expected dot height (%d) to be different from h1 (%d)", dh, h1)
		}

		// random measure test
		_, _ = renderer.Measure("A")
		wk1, hk1 := renderer.Measure("A\nA")
		_, _ = wk1, hk1
		wk2, hk2 := renderer.Measure("A\n\nA")
		if wk1 != wk2 {
			t.Fatalf("expected wk1 == wk2, got %d != %d", wk1, wk2)
		}
		if hk1 == hk2 {
			t.Fatalf("expected hk1 != hk2, got %d", hk1)
		}

		// ensure that bottoms respect glyph elevation
		_, eh1 := renderer.Measure("-")
		_, oy1 := renderer.Advanced().LastBoundsOffset()
		_, eh2 := renderer.Measure(".")
		_, oy2 := renderer.Advanced().LastBoundsOffset()
		if oy1 == oy2 {
			t.Fatalf("expected oy1 != oy2, got %d", oy1)
		}
		if oy1 + eh1 == oy2 + eh2 {
			t.Fatalf("expected oy1 + eh1 != oy2 + eh2, got (%d + %d) == (%d + %d)", oy1, eh1, oy2, eh2)
		}
	}

	// scale tests
	sw1, sh1 := renderer.Measure("A_")
	renderer.SetScale(4)
	sw4, sh4 := renderer.Measure("A_")
	if sw1*4 != sw4 || sh1*4 != sh4 {
		t.Fatalf("expected x4 scaling to result in x4 bounds, but got (%d, %d) vs (%d, %d)", sw1, sh1, sw4, sh4)
	}
}
