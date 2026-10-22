package kmeans

import (
	"image"
	"image/color"
	"magisterka/utils/maths"
	"magisterka/utils/types"
	"math/rand"
)

func Quantize(img image.Image, k int, maxIterations int) (*image.RGBA, []color.RGBA) {
	bounds := img.Bounds()
	height := bounds.Dy()
	width := bounds.Dx()

	pixels := make([]types.Pixel, 0, width*height)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()

			pixels = append(pixels, types.Pixel{R: float64(r >> 8), G: float64(g >> 8), B: float64(b >> 8)})
		}
	}

	centroids := make([]types.Pixel, k)

	for i := 0; i < k; i++ {
		centroids[i] = pixels[rand.Intn(len(pixels))]
	}

	pixelAssignments := make([]int, len(pixels))

	for iter := 0; iter < maxIterations; iter++ {
		for i, p := range pixels {
			bestCluster := 0

			bestDistance := maths.Distance(centroids[0], p)

			for c := 1; c < k; c++ {
				d := maths.Distance(centroids[c], p)

				if d < bestDistance {
					bestDistance = d
					bestCluster = c
				}
			}

			pixelAssignments[i] = bestCluster
		}

		centroidSums := make([]types.Pixel, k)
		centroidCount := make([]int, k)

		for i, cluster := range pixelAssignments {
			centroidSums[cluster].R += pixels[i].R
			centroidSums[cluster].G += pixels[i].G
			centroidSums[cluster].B += pixels[i].B
			centroidCount[cluster]++
		}

		moved := false

		for c := 0; c < k; c++ {
			if centroidCount[c] == 0 {
				continue
			}

			newCentroid := types.Pixel{
				R: centroidSums[c].R / float64(centroidCount[c]),
				G: centroidSums[c].G / float64(centroidCount[c]),
				B: centroidSums[c].B / float64(centroidCount[c]),
			}

			if maths.Distance(centroids[c], newCentroid) > 1.0 {
				moved = true
			}

			centroids[c] = newCentroid
		}

		if !moved {
			break
		}
	}

	out := image.NewRGBA(bounds)

	pixelIndex := 0

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {

			cluster := pixelAssignments[pixelIndex]

			c := centroids[cluster]

			out.Set(x, y, color.RGBA{
				R: uint8(maths.Clamp((c.R))),
				G: uint8(maths.Clamp((c.G))),
				B: uint8(maths.Clamp((c.B))),
				A: 255,
			})

			pixelIndex++
		}
	}

	palette := make([]color.RGBA, k)

	for i, c := range centroids {
		palette[i] = color.RGBA{
			R: uint8(maths.Clamp(c.R)),
			G: uint8(maths.Clamp(c.G)),
			B: uint8(maths.Clamp(c.B)),
			A: 255,
		}
	}

	return out, palette
}
