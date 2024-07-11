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

// Basically, [][4]float32 would be the ideal way to work with colors,
// but shaders need it passed as a flat []float32 slice, so we are
// using this type as glue
type rgbaSlice struct {
	data []float32
}

func newRGBASlice(size int) rgbaSlice {
	return rgbaSlice{ data: make([]float32, (size << 2)) }
}

func (self *rgbaSlice) Len() int {
	return len(self.data) >> 2
}

func (self *rgbaSlice) SliceAt(index int) []float32 {
	index <<= 2
	return self.data[index : index + 4]
}

func (self *rgbaSlice) At(index int) [4]float32 {
	index <<= 2
	return [4]float32(self.data[index : index + 4])
}

func (self *rgbaSlice) Set(index int, rgba [4]float32) {
	index <<= 2
	self.data[index + 0] = rgba[0]
	self.data[index + 1] = rgba[1]
	self.data[index + 2] = rgba[2]
	self.data[index + 3] = rgba[3]
}

