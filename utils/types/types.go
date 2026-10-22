package types

import (
	"image"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Pixel struct {
	R float64
	G float64
	B float64
}

type Camera struct {
	Position rl.Vector3
	Target   rl.Vector3
	Up       rl.Vector3
	Fov      float32
}

type TextureSetMRA struct {
	Metallic  string
	Roughness string
	AO        string
}

type ProcessResult struct {
	Key     string
	Image   *image.RGBA
	Palette []color.RGBA
}
