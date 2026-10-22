package camera

import (
	"magisterka/utils/types"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func InitCam(camParams types.Camera) rl.Camera3D {
	perspective := rl.CameraProjection(0)

	camera := rl.Camera3D{
		Position:   camParams.Position,
		Target:     camParams.Target,
		Up:         camParams.Up,
		Fovy:       camParams.Fov,
		Projection: perspective,
	}

	return camera
}
