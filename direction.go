package ptxt

import "strconv"

// Determines the main direction of the text. See [Renderer.SetDirection]().
type Direction uint8

const (
	Horizontal Direction = iota // left to right
	Vertical // [UNIMPLEMENTED] vertical, lines going LTR. font needs vert layout
	Sideways // sideways, glyph tops on the left side, bottom to top
	SidewaysRight // sideways, glyph tops on the right side, top to bottom
)

// Returns a textual representation of the direction.
func (self Direction) String() string {
	switch self {
	case Horizontal: return "Horizontal"
	case Vertical: return "Vertical"
	case Sideways: return "Sideways"
	case SidewaysRight: return "SidewaysRight"
	default:
		return "DirectionInvalid#" + strconv.Itoa(int(self))
	}
}
