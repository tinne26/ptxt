package ptxt

import "io"
import "os"
import _ "unsafe"

import "github.com/tinne26/ptxt/core"
import "github.com/tinne26/ptxt/strand"

import "github.com/tinne26/ggfnt"

// Handy type alias for font strands.
//type Strand = strand.Strand

// Tries to parse a font from the given source and creates
// a default strand for it. Accepted types are [*ggfnt.Font],
// [io.Reader], []byte and string (as a filepath).
//
// For the specific case of a non-nil [*ggfnt.Font], the returned
// error is always guaranteed to be nil.
//
// The returned strand is always non-nil if error is nil.
func NewStrand(source any) (*strand.Strand, error) {
	switch typedSource := source.(type) {
	case ggfnt.Font:
		panic("[ptxt.NewStrand] please use *ggfnt.Font, not ggfnt.Font")
	case *ggfnt.Font:
		return strand.New(typedSource), nil
	case io.Reader:
		return newStrandFromReader(typedSource)
	case []byte:
		return newStrandFromReader(&byteSliceReader{ data: typedSource })
	case string:
		file, err := os.Open(typedSource)
		if err != nil { return nil, err }
		strand, err := newStrandFromReader(file)
		if err != nil {
			_ = file.Close()
			return strand, err
		}
		return strand, file.Close()
	default:
		return nil, errMsg("invalid ggfnt font source type")
	}
}

// Tries to parse a font from the given reader and creates
// a default strand for it.
func newStrandFromReader(fontReader io.Reader) (*strand.Strand, error) {
	font, err := ggfnt.Parse(fontReader)
	if err != nil { return nil, err }
	return strand.New(font), nil
}

//go:linkname lnkBeginPass github.com/tinne26/ptxt/strand.(*StrandMapping).beginPass
func lnkBeginPass(*strand.StrandMapping, strand.GlyphPickerPass) error

//go:linkname lnkFinishPass github.com/tinne26/ptxt/strand.(*StrandMapping).finishPass
func lnkFinishPass(*strand.StrandMapping, strand.GlyphPickerPass)

//go:linkname lnkAppendCodePoint github.com/tinne26/ptxt/strand.(*StrandMapping).appendCodePoint
func lnkAppendCodePoint(*strand.StrandMapping, rune, []ggfnt.GlyphIndex) []ggfnt.GlyphIndex

//go:linkname lnkFinishMapping github.com/tinne26/ptxt/strand.(*StrandMapping).finishMapping
func lnkFinishMapping(*strand.StrandMapping, []ggfnt.GlyphIndex) []ggfnt.GlyphIndex

//go:linkname lnkSetBlendMode github.com/tinne26/ptxt/strand.(*Strand).setBlendMode
func lnkSetBlendMode(*strand.Strand, core.BlendMode)

//go:linkname lnkDrawHorzMask github.com/tinne26/ptxt/strand.(*Strand).drawHorzMask
func lnkDrawHorzMask(*strand.Strand, core.Target, core.GlyphMask, int, int, int, [4]float32)

//go:linkname lnkDrawSidewaysMask github.com/tinne26/ptxt/strand.(*Strand).drawSidewaysMask
func lnkDrawSidewaysMask(*strand.Strand, core.Target, core.GlyphMask, int, int, int, [4]float32)

//go:linkname lnkDrawSidewaysRightMask github.com/tinne26/ptxt/strand.(*Strand).drawSidewaysRightMask
func lnkDrawSidewaysRightMask(*strand.Strand, core.Target, core.GlyphMask, int, int, int, [4]float32)

func strandFullHorzInterspacing(fontStrand *strand.Strand) int {
	horzInterspacing := fontStrand.Font().Metrics().HorzInterspacing()
	return int(horzInterspacing) + int(fontStrand.HorzInterspacingShift())
}

func strandFullVertLineHeight(fontStrand *strand.Strand) int {
	lineHeight := fontStrand.Font().Metrics().LineHeight()
	return lineHeight + int(fontStrand.VertInterspacingShift())
}
