package ptxt

import "io"
import "cmp"

const brokenCode = "broken code"
const preViolation = "precondition violation"

const maxInt32 = 0x7FFF_FFFF

type errMsg string
func (self errMsg) Error() string { return string(self) }

func setBufferSize[T any](buffer []T, size int) []T {
	if cap(buffer) >= size {
		return buffer[ : size]
	} else {
		return make([]T, size)
	}
}

func clamp[T cmp.Ordered](x, a, b T) T {
	if x <= a { return a }
	if x >= b { return b }
	return x
}

// implements io.Reader for []byte
type byteSliceReader struct { data []byte ; index int }
func (self *byteSliceReader) Read(buffer []byte) (int, error) {
	maxRead := len(self.data) - self.index
	if maxRead <= 0 { return 0, io.EOF }
	if len(buffer) == 0 { return 0, nil }
	if len(buffer) >= maxRead {
		copy(buffer, self.data[self.index : self.index + maxRead])
		self.index += maxRead
		return maxRead, io.EOF
	} else {
		copy(buffer, self.data[self.index : self.index + len(buffer)])
		self.index += len(buffer)
		return len(buffer), nil
	}
}
