package ptxt

// Aligns tell a [Renderer] how to interpret the coordinates
// that [Renderer.Draw]() and other operations receive.
//
// More concretely: given some text, we have a text box or bounding
// rectangle that contains it. The text align specifies which part of
// that bounding box has to be aligned to the given coordinates. For
// example: drawing "POOP" at (0, 0) with a centered align means that
// the center of the text box will be aligned to the (0, 0) coordinate.
// We should see the bottom half of "OP" on the top-left corner of our
// rendering [core.Target].
//
// Notice that align names are based on the [Horizontal] text [Direction].
// A top left align applied to [Sideways] text will make the text start
// on the bottom left of the bounding rectangle.
//
// See [Renderer.SetAlign]() for further explanations.
//
// As a final note, notice that there are some important differences
// for aligns when it comes to bitmap vs vectorial fonts:
//  - The error in text centering can be very visible. This is party
//    because very low resolutions scaled up will make errors more
//    obvious, and partly because the vertical center of many small
//    fonts can't even be accurate in the first place.
//  - Many bitmap fonts don't have lowercase letters, so the
//    lowercase height will be zero and [Midline] won't be usable.
type Align uint8

// Returns the vertical component of the align. If the align
// is valid and [Align.HasVertComponent](), the result can only
// be [Top], [CapLine], [Midline], [VertCenter], [Baseline],
// [LastBaseline] or [Bottom].
func (self Align) Vert() Align { return alignVertBits & self }

// Returns the horizontal component of the align. If the
// align is valid and [Align.HasHorzComponent]() is true,
// the result can only be [Left], [HorzCenter] or [Right].
func (self Align) Horz() Align { return alignHorzBits & self }

// Returns whether the vertical component of the align is set.
func (self Align) HasVertComponent() bool { return alignVertBits & self != 0 }

// Returns whether the horizontal component of the align is set.
func (self Align) HasHorzComponent() bool { return alignHorzBits & self != 0 }

// Returns the result of overriding the current align with
// the non-empty components of the new align. If both
// components are defined for the new align, the result
// will be the new align itself. If only one component
// is defined, only that component will be overwritten.
// If the new align is completely empty, the value of
// the current align will be returned unmodified.
func (self Align) Adjusted(align Align) Align {
	horz := align.Horz()
	vert := align.Vert()
	if horz != 0 {
		if vert != 0 { return align }
		return horz | self.Vert()
	} else if vert != 0 {
		return self.Horz() | vert
	} else {
		return self
	}
}

// Returns a value between 'left' and 'right' based on the current horizontal align:
//  - [Left]: the function returns 'left'.
//  - [Right]: the function returns 'right'.
//  - Otherwise: the function returns the middle point between 'left' and 'right'.
func (self Align) GetHorzAnchor(left, right int) int {
	switch self.Horz() {
	case Left  : return left
	case Right : return right
	default: // assume horz center even when undefined
		return (left + right) >> 1
	}
}

// Returns a textual representation of the align. Some examples:
//   (Top | Right).String() == "(Top | Right)"
//   (Right | Top).String() == "(Top | Right)"
//   Center.String() == "(VertCenter | HorzCenter)"
//   (Baseline | Left).String() == "(Baseline | Left)"
//   HorzCenter.String() == "(HorzCenter)"
//   Bottom.String() == "(Bottom)"
func (self Align) String() string {
	if self == 0 { return "(ZeroAlign)" }
	if self.Vert() == 0 { return "(" + self.horzString() + ")" }
	if self.Horz() == 0 { return "(" + self.vertString() + ")" }
	return "(" + self.vertString() + " | " + self.horzString() + ")"
}

func (self Align) vertString() string {
	switch self.Vert() {
	case Top: return "Top"
	case CapLine: return "CapLine"
	case Midline: return "Midline"
	case VertCenter: return "VertCenter"
	case Baseline: return "Baseline"
	case Bottom: return "Bottom"
	case LastBaseline: return "LastBaseline"
	default:
		return "VertUnknown"
	}
}

func (self Align) horzString() string {
	switch self.Horz() {
	case Left: return "Left"
	case HorzCenter: return "HorzCenter"
	case Right: return "Right"
	default:
		return "HorzUnknown"
	}
}

// Aligns have a vertical and a horizontal component. To set
// both components at once, you can use a bitwise OR:
//   Renderer.SetAlign(ptxt.Left | ptxt.Bottom)
// To retrieve or compare the individual components, avoid
// bitwise operations and use [Align.Vert]() and [Align.Horz]()
// instead.
const (
	// Horizontal aligns
	Left       Align = 0b0010_0000
	HorzCenter Align = 0b0100_0000
	Right      Align = 0b1000_0000

	// Vertical aligns
	Top          Align = 0b0000_0001 // top of font's ascent
	CapLine      Align = 0b0000_0011 // top of font's cap height (rarely used)
	Midline      Align = 0b0000_0010 // top of xheight (rarely used)
	VertCenter   Align = 0b0000_1001 // middle of line height
	Baseline     Align = 0b0000_0100 // aligned to baseline
	Bottom       Align = 0b0000_1000 // bottom of font's descent
	LastBaseline Align = 0b0000_1100 // last Baseline (for multiline text)

	// Full aligns
	Center Align = HorzCenter | VertCenter
	
	alignVertBits Align = 0b0000_1111 // bit mask
	alignHorzBits Align = 0b1111_0000 // bit mask
)
// Internal note: the fact that the combinations of Bottom | Top,
// Baseline | Bottom and Midline | Bottom are valid is intentional
// and I intend to preserve it like that, despite the fact that it
// is ugly that Top | Baseline and Top | Midline don't result in
// Baseline and Midline respectively too. Perfect combinations are
// not possible, so this is almost only a cute detail.
