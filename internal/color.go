package internal

import "image/color"

func RGBAToFloat32(rgba color.RGBA) [4]float32 {
	return [4]float32{
		float32(rgba.R)/255,
		float32(rgba.G)/255,
		float32(rgba.B)/255,
		float32(rgba.A)/255,
	}
}
