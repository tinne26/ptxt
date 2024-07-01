package ptxt

import "github.com/tinne26/ptxt/strand"

// Font [*strand.Strand] index used with [RendererStrands] to
// manage, select and switch between font strands.
type StrandIndex uint8

// This type exists only for documentation and structuring purposes,
// acting as a [gateway] to fetch, add and modify renderer font strands.
//
// Working with multiple strands can be helpful if you are using multiple
// fonts... but it only becomes truly crucial once you start trying
// to render single blocks of text with multiple fonts and formats
// at once (see [Twine]).
//
// In general, this type is used through method chaining:
//   index := renderer.Strands().Add(fontStrand)
//
// [gateway]: https://pkg.go.dev/github.com/tinne26/ptxt#Renderer
type RendererStrands Renderer

// Returns the font strand stored at the given index.
// Panics if the index is not valid.
func (self *RendererStrands) Get(index StrandIndex) *strand.Strand {
	return self.strands[index]
}

// Adds a new font strand to the renderer and returns its
// index, which can be used to fetch it, select it or replace
// it later.
//
// Only up to 255 strands can be simultaneously stored in the
// renderer.
func (self *RendererStrands) Add(fontStrand *strand.Strand) StrandIndex {
	if len(self.strands) >= 255 { panic("can't exceed 255 strands") }
	if fontStrand == nil { panic("nil strand") }
	
	if self.strands[0] == nil {
		self.strands[0] = fontStrand
		return 1 // base case on initialization
	} else {
		index := StrandIndex(len(self.strands))
		self.strands = append(self.strands, fontStrand)
		return index
	}
}

// Sets the renderer's active font strand.
func (self *RendererStrands) Select(index StrandIndex) {
	if int(index) >= len(self.strands) { panic("strand index out of bounds") }
	self.strandIndex = index
}

// Returns the index of the currently selected font strand.
func (self *RendererStrands) Index() StrandIndex {
	return self.strandIndex
}

// Returns the number of font strands currently stored in the renderer.
func (self *RendererStrands) Count() int {
	return len(self.strands)
}

// Replaces the strand at the given index with the new strand.
// Passing an invalid index or a nil strand will make the
// method panic.
func (self *RendererStrands) Replace(index StrandIndex, fontStrand *strand.Strand) {
	if int(index) >= len(self.strands) { panic("strand index out of bounds") }
	if fontStrand == nil { panic("nil strand") }
	self.strands[index] = fontStrand
}

// Removes all font strands from the renderer's memory.
func (self *RendererStrands) ClearAll() {
	self.strands = self.strands[ : 1]
	self.strands[0] = nil
	self.strandIndex = 0
}

// Probably won't implement removal at all.
//func (self *RendererStrands) Remove(StrandIndex) {}
