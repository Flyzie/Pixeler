package gui

import (
	"image/color"
	"magisterka/utils/constants"
	"strconv"

	rgui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/sqweek/dialog"
)

func DrawControllButtons(kValue *float32, downscalingValue *float32) bool {
	quantizationBtn := rgui.Button(rl.NewRectangle(50, 100, 100, 40), "Process textures")
	rgui.Slider(rl.NewRectangle(50, 160, 300, 10), "0", strconv.Itoa(int(*kValue)), kValue, 1, 50)
	rgui.Slider(rl.NewRectangle(50, 180, 300, 10), "0", strconv.FormatFloat(float64(*downscalingValue), 'f', -1, 32), downscalingValue, 1, 20)

	return quantizationBtn
}

func DrawPalette(palette *[]color.RGBA) {
	for i, paletteColor := range *palette {
		rlColor := rl.NewColor(paletteColor.R, paletteColor.G, paletteColor.B, paletteColor.A)
		rl.DrawRectangle(constants.ScreenWidth-int32(25*(i+1)), constants.ScreenHeight-25, 25, 25, rlColor)
	}
}

func TextureSelector(name string, x, y float32) (bool, string) {
	clicked := rgui.Button(rl.NewRectangle(x, y, 80, 40), name)

	if !clicked {
		return false, ""
	}

	path, err := dialog.File().Filter("Image files", "png", "jpg", "jpeg").Title("Select texture").Load()

	if err != nil || path == "" {
		return false, ""
	}

	return clicked, path
}
