package image

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"magisterka/utils/types"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, format, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	fmt.Println("Loaded format: ", format)
	return img, nil
}

func SaveImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func DownscaleNN(img image.Image, factor float64) *image.RGBA {
	bounds := img.Bounds()
	height := int(float64(bounds.Dy()) / factor)
	width := int(float64(bounds.Dx()) / factor)

	if width <= 0 {
		width = 1
	}

	if height <= 0 {
		height = 1
	}

	out := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		sx := int((float64(x) + 0.5) * factor)
		for y := 0; y < height; y++ {
			sy := int((float64(y) + 0.5) * factor)

			out.Set(x, y, img.At(sx, sy))
		}
	}

	return out
}

func UpscaleNN(img image.Image, factor float64) *image.RGBA {
	bounds := img.Bounds()
	height := int(float64(bounds.Dy()) * factor)
	width := int(float64(bounds.Dx()) * factor)

	if width <= 0 {
		width = 1
	}

	if height <= 0 {
		height = 1
	}

	out := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		sx := int((float64(x) + 0.5) / factor)
		for y := 0; y < height; y++ {
			sy := int((float64(y) + 0.5) / factor)

			out.Set(x, y, img.At(sx, sy))
		}
	}

	return out
}

func ParseToMRA(textures types.TextureSetMRA) (*rl.Texture2D, error) {

	if textures.Metallic == "" && textures.Roughness == "" && textures.AO == "" {
		return nil, errors.New("You need to select at least one texture")
	}

	var metallic, roughness, ao *rl.Image
	var pixelsMetallic, pixelsRoughness, pixelsAO []color.RGBA

	if textures.Metallic != "" {
		metallic = rl.LoadImage(textures.Metallic)
		pixelsMetallic = rl.LoadImageColors(metallic)
		defer rl.UnloadImage(metallic)
		defer rl.UnloadImageColors(pixelsMetallic)
	}

	if textures.Roughness != "" {
		roughness = rl.LoadImage(textures.Roughness)
		pixelsRoughness = rl.LoadImageColors(roughness)
		defer rl.UnloadImage(roughness)
		defer rl.UnloadImageColors(pixelsRoughness)
	}

	if textures.AO != "" {
		ao = rl.LoadImage(textures.AO)
		pixelsAO = rl.LoadImageColors(ao)
		defer rl.UnloadImage(ao)
		defer rl.UnloadImageColors(pixelsAO)
	}

	var width, height int32

	switch {
	case metallic != nil:
		width = metallic.Width
		height = metallic.Height
	case roughness != nil:
		width = roughness.Width
		height = roughness.Height
	case ao != nil:
		width = ao.Width
		height = ao.Height
	}

	for name, texture := range map[string]*rl.Image{
		"metallic":  metallic,
		"roughness": roughness,
		"ao":        ao,
	} {
		if texture != nil && (texture.Width != width || texture.Height != height) {
			return nil, fmt.Errorf("%s dimensions don't match", name)
		}
	}

	imgMRA := rl.GenImageColor(int(width), int(height), rl.Blank)

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			i := y*int(width) + x

			r, g, b := uint8(0), uint8(0), uint8(0)

			if textures.Metallic != "" {
				r = pixelsMetallic[i].R
			}

			if textures.Roughness != "" {
				g = pixelsRoughness[i].R
			}

			if textures.AO != "" {
				b = pixelsAO[i].R
			}

			color := rl.Color{R: r, G: g, B: b, A: 255}

			rl.ImageDrawPixel(imgMRA, int32(x), int32(y), color)
		}

	}

	ok := rl.ExportImage(*imgMRA, "MRA.png")
	if !ok {
		return nil, fmt.Errorf("failed to export MRA.png")
	} else {
		fmt.Println("Image exported")
	}

	texture := rl.LoadTextureFromImage(imgMRA)
	rl.UnloadImage(imgMRA)

	return &texture, nil

}

func RaylibTextureToImage(texture rl.Texture2D) image.Image {
	rlImage := rl.LoadImageFromTexture(texture)
	defer rl.UnloadImage(rlImage)

	pixels := rl.LoadImageColors(rlImage)
	defer rl.UnloadImageColors(pixels)

	img := image.NewRGBA(image.Rect(0, 0, int(rlImage.Width), int(rlImage.Height)))
	for y := 0; y < int(rlImage.Height); y++ {
		for x := 0; x < int(rlImage.Width); x++ {
			img.SetRGBA(x, y, pixels[y*int(rlImage.Width)+x])
		}
	}

	return img
}
