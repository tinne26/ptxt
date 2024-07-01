package strand

import "image/color"
import "fmt"
import "strconv"

const brokenCode = "broken code"

const nonPremultRGBA = "non-premultiplied RGBA"

func isPremultiplied(rgba color.RGBA) bool {
	return (rgba.A >= rgba.R) && (rgba.A >= rgba.G) && (rgba.A >= rgba.B)
}

func itoaRune(r rune) string {
	return strconv.FormatInt(int64(r), 10)
}

func runeToUnicodeCode(r rune) string {
	return fmt.Sprintf("\\u%04X", int64(r))
}
