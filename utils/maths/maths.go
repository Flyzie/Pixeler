package maths

import (
	"magisterka/utils/types"
	"math"
)

func Distance(centroid, pixel types.Pixel) float64 {
	dr := centroid.R - pixel.R
	dg := centroid.G - pixel.G
	db := centroid.B - pixel.B
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func Clamp(v float64) float64 {
	return math.Max(0, math.Min(255, v))
}
