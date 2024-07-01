package ptxt

import "testing"

func TestMeasure(t *testing.T) {
	ensureTestAssetsLoaded()
	if testFont == nil { t.SkipNow() }

	// create strand and renderer
	strand, _ := NewStrand(testFont)
	renderer := NewRenderer()
	renderer.SetStrand(strand)

	// aux variables
	fontLineHeight := testFont.Metrics().LineHeight()
	fontLineHeightWithoutGap := fontLineHeight - int(testFont.Metrics().LineGap())

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

		// consistency tests	
		w1, h1 := renderer.Measure("HEY H")
		w2, h2 := renderer.Measure("HEY HO")
		w3, h3 := renderer.Measure("HEY HOO")
		w4,  _ := renderer.Measure("HEY HO.HEY HO")
		if h1 != fontLineHeightWithoutGap {
			t.Fatalf(
				"expected single line height (%d) to match font line height without gap (%d)",
				h1, fontLineHeightWithoutGap,
			)
		}
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

		// dot height test
		_, dh := renderer.Measure(".")
		if dh != h1 {
			t.Fatalf("expected dot height (%d) to match text line height (%d)", dh, h1)
		}

		// line break and spacing tests
		_, h5 := renderer.Measure("\n")
		if h5 != fontLineHeight {
			t.Fatalf(
				"expected single line break height (%d) to match font line height (%d)",
				h5, fontLineHeight,
			)
		}
		_, h6 := renderer.Measure("\n ")
		if h6 != fontLineHeight + fontLineHeightWithoutGap {
			t.Fatalf(
				"expected line break + content height (%d) to match gapless-line-height*2 + gap (%d)",
				h6, fontLineHeight + fontLineHeightWithoutGap,
			)
		}

		_, hs1 := renderer.Measure("A")
		_, hs2 := renderer.Measure(" ")
		if hs1 != hs2 { t.Fatal("expected same height") }
		_, hs3 := renderer.Measure("A\n\nA")
		_, hs4 := renderer.Measure("    \n\n      ")
		if hs3 != hs4 { t.Fatal("expected same height") }

		// repeat with different scale and ensure consistency
		renderer.SetScale(3)
		s3_w1, s3_h1 := renderer.Measure("HEY H")
		s3_w2, s3_h2 := renderer.Measure("HEY HO")
		s3_w3, s3_h3 := renderer.Measure("HEY HOO")
		s3_w4, _ := renderer.Measure("HEY HO.HEY HO")
		_, s3_dh := renderer.Measure(".")
		_, s3_h5 := renderer.Measure("\n")
		_, s3_h6 := renderer.Measure("\n ")
		_, s3_hs1 := renderer.Measure("A")
		_, s3_hs2 := renderer.Measure(" ")
		_, s3_hs3 := renderer.Measure("A\n\nA")
		_, s3_hs4 := renderer.Measure("    \n\n      ")
		if w1*3 != s3_w1 || h1*3 != s3_h1 { t.Fatal("inconsistent measurings after scaling") }
		if w2*3 != s3_w2 || h2*3 != s3_h2 { t.Fatal("inconsistent measurings after scaling") }
		if w3*3 != s3_w3 || h3*3 != s3_h3 { t.Fatal("inconsistent measurings after scaling") }
		if w4*3 != s3_w4 { t.Fatal("inconsistent measurings after scaling") }
		if dh*3 != s3_dh { t.Fatal("inconsistent measurings after scaling") }
		if h5*3 != s3_h5 { t.Fatal("inconsistent measurings after scaling") }
		if h6*3 != s3_h6 { t.Fatal("inconsistent measurings after scaling") }
		if hs1*3 != s3_hs1 { t.Fatal("inconsistent measurings after scaling") }
		if hs2*3 != s3_hs2 { t.Fatal("inconsistent measurings after scaling") }
		if hs3*3 != s3_hs3 { t.Fatal("inconsistent measurings after scaling") }
		if hs4*3 != s3_hs4 { t.Fatal("inconsistent measurings after scaling") }
		renderer.SetScale(1)

		// extra randoms
		_, hr1 := renderer.Measure("A\nB")
		_, hr2 := renderer.Measure("\nB")
		if hr1 != hr2 { t.Fatal("hr1 != hr2") }
	}

	// test paragraph breaks
	renderer.SetAlign(Baseline | Left)
	_, lbh1_npb := renderer.Measure("HELLO\nWORLD")
	_, lbh2_npb := renderer.Measure("HELLO\n\nWORLD")
	renderer.Advanced().SetParBreakEnabled(true)
	_, lbh1_ypb := renderer.Measure("HELLO\nWORLD")
	_, lbh2_ypb := renderer.Measure("HELLO\n\nWORLD")
	_, lbh3_ypb := renderer.Measure("HELLO\n\n\nWORLD")
	if lbh1_npb != lbh1_ypb {
		t.Fatalf(
			"the height of one line break without paragraph break (%d) should match with par break (%d)",
			lbh1_npb, lbh1_ypb,
		)
	}
	if lbh1_npb != fontLineHeightWithoutGap + fontLineHeight {
		t.Fatal("lbh1_npb != fontLineHeight + fontLineHeightWithoutGap")
	}
	if lbh2_ypb <= lbh1_npb || lbh2_ypb >= lbh2_npb {
		t.Fatalf(
			"expected the height of two line breaks with par break (%d) to be between (%d) and (%d)",
			lbh2_ypb, lbh1_npb, lbh2_npb,
		)
	}
	if lbh2_ypb != fontLineHeightWithoutGap + fontLineHeight + (fontLineHeight >> 1) {
		t.Fatal("lbh2_ypb != fontLineHeightWithoutGap + fontLineHeight + (fontLineHeight >> 1)")
	}
	if lbh3_ypb != lbh2_npb {
		t.Fatalf("lbh3_ypb != lbh2_npb (%d != %d)", lbh3_ypb, lbh2_npb)
	}
}

func TestMeasureWrap(t *testing.T) {
	ensureTestAssetsLoaded()
	if testFont == nil { t.SkipNow() }

	// create strand and renderer
	strand, _ := NewStrand(testFont)
	renderer := NewRenderer()
	renderer.SetStrand(strand)

	helloWidth,  _ := renderer.Measure("HELLO")
	helloWidth2, helloHeight := renderer.Measure("HELLO\nHELLO")
	if helloWidth != helloWidth2 {
		t.Fatalf("expected same widths, got %d and %d", helloWidth, helloWidth2)
	}
	w1, h1 := renderer.MeasureWithWrap("HELLO HELLO", helloWidth)
	if w1 != helloWidth {
		t.Fatalf("expected wrap width = %d, got %d", helloWidth, w1)
	}
	if h1 != helloHeight {
		t.Fatalf("expected wrap height = %d, got %d", helloHeight, h1)
	}
	if len(renderer.run.wrapIndices) == 0 {
		t.Fatal("expected wrap indices to not be empty")
	}

	// TODO: expand with all the relevant cases of single letter, multiple,
	// probably would need before/elide/after too (-), etc. HELLO-HELL,
	// against HELLO-HE, etc.
}
