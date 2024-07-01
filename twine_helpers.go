package ptxt

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Related to [RendererTwine.RegisterFunc]().
type TwineEffectArgs struct {
	Renderer *Renderer
	Payload []byte
	OX, OY int
	StartIndex int
	EndIndex int
	MinWidth int
	StartWrap bool // whether the effect is re-starting after a line break
	EndWrap bool // whether the effect splits at a line break
	// DrawPass (not really, I still have to figure out the model,
	//           sometimes we only apply the func at measure time?
	//           Maybe I can tell based on the func type if it's
	//           only measuring, only drawing, or both)
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// ...
func (self *TwineEffectArgs) AssertPayloadLen(numBytes int) {
	panic("unimplemented")
}

// NOTICE: TWINES ARE UNIMPLEMENTED
// 
// Related to [Twine.PushPadder](). Most of the time, you
// create padders directly with TwinePadder{ PrePad: 16 } or
// similar, defining only the required fields.
type TwinePadder struct {
	PrePad       uint16 // padding before the block
	PostPad      uint16 // padding after the block
	MinWidth     uint16 // can't be disconnected from PrePad or LineStartPad
	LineStartPad uint16 // padding after block break (should be <= PrePad)
	LineBreakPad uint16 // padding on block break (should be <= PostPad)
	UnitsScaled  bool // if true, units are considered to be already scaled
}
