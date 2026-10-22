package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"

	"magisterka/quantization/kmeans"
	"os"

	cam "magisterka/camera"
	gui "magisterka/gui"
	constants "magisterka/utils/constants"
	imageUtils "magisterka/utils/image"
	customType "magisterka/utils/types"

	pbr "magisterka/pbr"

	rgui "github.com/gen2brain/raylib-go/raygui"
	"github.com/sqweek/dialog"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	globalPadding := constants.GlobalPadding
	screenWidth := constants.ScreenWidth
	screenHeight := constants.ScreenHeight
	origin := rl.Vector2{X: 0, Y: 0}

	texturesRect := rl.NewRectangle(50, 50, float32(screenWidth)-2*globalPadding, float32(screenHeight)-2*globalPadding)
	rotationArea := rl.NewRectangle(float32(screenWidth)/2-750, float32(screenHeight)/2-500, 1500, 1000)

	palette := make([]color.RGBA, 0)

	textureKeys := []string{
		"diffuse",
		"normal",
		"metallic",
		"roughness",
		"ao",
	}

	texturePaths := map[string]string{
		"diffuse":   "",
		"normal":    "",
		"metallic":  "",
		"roughness": "",
		"ao":        "",
	}

	textures := map[string]rl.Texture2D{
		"diffuse": {},
		"normal":  {},
		"mra":     {},
	}
	mraTextures := customType.TextureSetMRA{}

	cameraParams := customType.Camera{Position: rl.Vector3{X: 10.0, Y: 10.0, Z: 10.0}, Target: rl.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, Up: rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0}, Fov: 45.0}
	camera := cam.InitCam(cameraParams)

	sourceImage := rl.LoadTexture("imageUtils.png")
	quantizedImage := rl.LoadTexture("output.png")
	downscaledImage := rl.LoadTexture("downscaled.png")
	upscaledImage := rl.LoadTexture("upscaled.png")
	images := []rl.Texture2D{sourceImage, quantizedImage, downscaledImage, upscaledImage}

	processing := false
	showTexturesWindow := false

	var kValue float32 = 1.0
	var downscalingValue float32 = 1.0
	var done chan customType.ProcessResult
	rotation := 0.0

	rl.InitWindow(int32(screenWidth), int32(screenHeight), constants.AppTitle)
	defer rl.CloseWindow()
	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	model := rl.LoadModel("./models/cube.obj")

	//PBR
	shader := rl.LoadShader("./pbr/glsl330/pbr.vs", "./pbr/glsl330/pbr.fs")

	l := pbr.Light{}
	l.SetCombineShader(&shader)
	l.Init(0.0, rl.Vector3{X: 1, Y: 1, Z: 1})

	l.NewLight(pbr.LightTypeDirectional, rl.Vector3{}, rl.Vector3{X: 1, Y: -1, Z: -1}, rl.White, 2, &l.Shader)
	l.NewLight(pbr.LightTypeDirectional, rl.Vector3{}, rl.Vector3{X: -1, Y: -0.4, Z: 1}, rl.White, 3, &l.Shader)
	l.NewLight(pbr.LightTypeDirectional, rl.Vector3{}, rl.Vector3{X: 1, Y: -0.2, Z: 1}, rl.White, 2, &l.Shader)
	l.NewLight(pbr.LightTypeDirectional, rl.Vector3{}, rl.Vector3{X: -1, Y: -1, Z: -1}, rl.White, 3, &l.Shader)
	l.NewLight(pbr.LightTypeDirectional, rl.Vector3{}, rl.Vector3{X: 0, Y: -1, Z: 0}, rl.White, 2, &l.Shader)

	p := pbr.PhysicRender{}
	p.SetCombineShader(&shader)
	p.Init()
	p.AmbientColor(rl.Vector3{X: 1, Y: 1, Z: 1}, 0.16)

	p.UseTexAlbedo()
	p.UseTexMRA()
	p.UseTexNormal()
	p.UseTexEmissive()

	configureModelMaterials(&model, shader, rl.White)

	p.AlbedoColorModel(&model, rl.White)
	p.MetallicValue(0.0)
	p.RoughnessValue(0.08)
	p.EmissivePower(0.0)
	p.AoValue(1.0)
	p.NormalValue(0.2)

	p.EmissiveColor(rl.NewColor(255, 162, 0, 255))
	p.TextureMapAlbedo(&model.GetMaterials()[0], rl.LoadTexture("./diffuse_placeholder.png"))
	p.SetTiling(rl.NewVector2(1, 1))
	//modelTexture := rl.LoadTexture("models/textures/diffuse.png")
	//rl.SetMaterialTexture(&model.GetMaterials()[0], rl.MapDiffuse, modelTexture)

	rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {
		rotation++
		mousePosition := rl.GetMousePosition()

		if (rl.IsMouseButtonDown(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mousePosition, rotationArea)) || rl.IsKeyDown(rl.KeyLeftControl) {
			rl.UpdateCamera(&camera, rl.CameraThirdPerson)
		}

		// Keep view position in sync for PBR specular/lighting calculations.
		p.UpadteByCamera(camera.Position)

		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)

		rl.BeginMode3D(camera)
		rl.DrawModel(model, rl.NewVector3(0, 0, 0), 1.0, rl.White)
		rl.DrawGrid(10, 1.0)
		rl.EndMode3D()

		loadModelButton := rgui.Button(rl.NewRectangle(170, 50, 100, 40), "Load Model")
		if loadModelButton {
			path, err := dialog.File().Filter("Wavefront OBJ files", "obj", "glb", "gltf").
				Title("Select model").Load()

			if err != nil || path == "" {
				break
			}
			rl.UnloadModel(model)

			model = rl.LoadModel(path)
			configureModelMaterials(&model, shader, rl.White)
		}

		quantizationBtn := gui.DrawControllButtons(&kValue, &downscalingValue)

		if showTexturesWindow {
			showTexturesWindow = !rgui.WindowBox(texturesRect, "Images")
			gridyfy(images, texturesRect, origin, 2, 100)

		} else {
			showTexturesWindow = rgui.Button(rl.NewRectangle(constants.GlobalPadding, 50, 100, 40), "Show textures")
		}

		for i, key := range textureKeys {
			y := 300 + float32(i*45)
			clicked, path := gui.TextureSelector(key, constants.GlobalPadding, y)

			if clicked {

				switch key {
				case "diffuse":
					{
						if textures[key].ID != 0 {
							rl.UnloadTexture(textures[key])
						}
						textures[key] = rl.LoadTexture(path)
						p.TextureMapAlbedo(&model.GetMaterials()[0], textures["diffuse"])
					}
				case "normal":
					{
						if textures[key].ID != 0 {
							rl.UnloadTexture(textures[key])
						}
						textures[key] = rl.LoadTexture(path)
						p.TextureMapNormal(&model.GetMaterials()[0], textures["normal"])
					}
				case "metallic":
					{
						mraTextures.Metallic = path
						mraTexture, err := imageUtils.ParseToMRA(mraTextures)
						if err == nil {
							textures["mra"] = *mraTexture
							p.TextureMapMetalness(&model.GetMaterials()[0], textures["mra"])
						}
					}
				case "roughness":
					{
						mraTextures.Roughness = path
						mraTexture, err := imageUtils.ParseToMRA(mraTextures)
						if err == nil {
							textures["mra"] = *mraTexture
							p.TextureMapMetalness(&model.GetMaterials()[0], textures["mra"])
						}
					}
				case "ao":
					{
						mraTextures.AO = path
						mraTexture, err := imageUtils.ParseToMRA(mraTextures)
						if err == nil {
							textures["mra"] = *mraTexture
							p.TextureMapMetalness(&model.GetMaterials()[0], textures["mra"])
						}
					}
				default:
					{
					}
				}
				println(texturePaths[key])
			}

			previewTexture := textures[key]
			if key == "metallic" || key == "roughness" || key == "ao" {
				previewTexture = textures["mra"]
			}

			if previewTexture.ID != 0 {
				rl.DrawTexturePro(
					previewTexture,
					parseSourceRec(previewTexture),
					rl.Rectangle{
						X:      constants.GlobalPadding + 90,
						Y:      y,
						Width:  80,
						Height: 40,
					},
					origin,
					0,
					rl.White,
				)
			}
		}

		if quantizationBtn && !processing {

			done = make(chan customType.ProcessResult)
			processing = true

			inputImages := make(map[string]image.Image, len(textures))

			for key, texture := range textures {
				if texture.ID == 0 {
					continue
				}
				inputImages[key] = imageUtils.RaylibTextureToImage(texture)
			}

			rl.UnloadTexture(sourceImage)
			rl.UnloadTexture(quantizedImage)
			rl.UnloadTexture(downscaledImage)
			rl.UnloadTexture(upscaledImage)

			go func() {
				defer close(done)
				for key, texture := range inputImages {
					if key == "diffuse" {
						p, resultImage := processTexture(texture, int(kValue), float64(downscalingValue), key)
						done <- customType.ProcessResult{Key: key, Image: resultImage, Palette: p}
					} else {
						_, resultImage := processTexture(texture, int(kValue), float64(downscalingValue), key)
						done <- customType.ProcessResult{Key: key, Image: resultImage}
					}

				}
			}()
		}

		if processing {
			select {
			case result, open := <-done:

				if !open {
					processing = false
					break
				}

				rayImage := rl.NewImageFromImage(result.Image)
				newTexture := rl.LoadTextureFromImage(rayImage)
				if result.Palette != nil {
					palette = result.Palette
				}
				rl.UnloadImage(rayImage)

				switch result.Key {
				case "diffuse":
					{
						p.TextureMapAlbedo(&model.GetMaterials()[0], newTexture)
					}
				case "normal":
					{
						p.TextureMapNormal(&model.GetMaterials()[0], newTexture)
					}
				case "mra":
					{
						p.TextureMapMetalness(&model.GetMaterials()[0], newTexture)
					}
				default:
					{
					}
				}

				sourceImage = rl.LoadTexture("imageUtils.png")
				quantizedImage = rl.LoadTexture("output.png")
				downscaledImage = rl.LoadTexture("downscaled.png")
				upscaledImage = rl.LoadTexture("upscaled.png")

				images = []rl.Texture2D{
					sourceImage,
					quantizedImage,
					downscaledImage,
					upscaledImage,
				}
			default:
				rl.DrawRectanglePro(rl.NewRectangle(160, 120, 10, 10), rl.NewVector2(5, 5), float32(rotation), rl.LightGray)
			}
		}
		gui.DrawPalette(&palette)
		rl.EndDrawing()
	}
	rl.UnloadTexture(sourceImage)
	rl.UnloadTexture(quantizedImage)
	rl.UnloadTexture(downscaledImage)
	rl.UnloadTexture(upscaledImage)
	rl.UnloadModel(model)
}

func processTexture(img image.Image, k int, downscalingValue float64, label string) ([]color.RGBA, *image.RGBA) {

	quantized, palette := kmeans.Quantize(img, k, 20)
	downscaled := imageUtils.DownscaleNN(quantized, downscalingValue)
	upscaled := imageUtils.UpscaleNN(downscaled, downscalingValue)

	out, _ := os.Create((fmt.Sprintf("./models/textures/processed/output_%s.png", label)))
	out2, _ := os.Create((fmt.Sprintf("./models/textures/processed/downscaled%s.png", label)))
	out3, _ := os.Create((fmt.Sprintf("./models/textures/processed/upscaled_%s.png", label)))
	defer out.Close()
	defer out2.Close()
	defer out3.Close()

	png.Encode(out, quantized)
	png.Encode(out2, downscaled)
	png.Encode(out3, upscaled)

	//finalTexture := rl.LoadTextureFromImage(rl.NewImageFromImage(upscaled))

	return palette, upscaled
}

func parseSourceRec(texture rl.Texture2D) rl.Rectangle {
	return rl.Rectangle{X: 0, Y: 0, Width: float32(texture.Width), Height: float32(texture.Height)}
}

func configureModelMaterials(model *rl.Model, shader rl.Shader, color rl.Color) {
	materials := model.GetMaterials()
	for i := range materials {
		materials[i].Shader = shader
		materials[i].GetMap(rl.MapAlbedo).Color = color
	}
}

func parseDestRec(texture rl.Texture2D, col, row int32, scale, offset float32) rl.Rectangle {
	return rl.Rectangle{X: float32(float32(col*texture.Width)*scale) + offset, Y: float32(float32(row*texture.Height)*scale) + offset, Width: float32(texture.Width) * scale, Height: float32(texture.Height) * scale}
}

func parseToMaxDestRec(texture rl.Texture2D, col, row int32, maxWidth, maxHeight, offset float32) rl.Rectangle {
	scaleX := maxWidth / float32(texture.Width)
	scaleY := maxHeight / float32(texture.Height)

	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	return rl.Rectangle{X: float32(float32(col*texture.Width)*scale) + offset, Y: float32(float32(row*texture.Height)*scale) + offset, Width: float32(texture.Width) * scale, Height: float32(texture.Height) * scale}
}

func gridyfy(textures []rl.Texture2D, container rl.Rectangle, origin rl.Vector2, cols int32, offset float32) {
	row := 0
	col := 0

	for i, texture := range textures {
		if i%int(cols) == 0 && i != 0 {
			row++
			col = 0
		}
		rl.DrawTexturePro(texture, parseSourceRec(texture), parseToMaxDestRec(texture, int32(col), int32(row), (container.Width/float32(cols)), (container.Height/float32(cols)), offset), origin, 0, rl.White)
		col++
	}
}
