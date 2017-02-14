package watermark

import "testing"

func TestMakeWatermark(t *testing.T) {
	path := "../input.jpg"
	WatermarkPath = "../watermarks/gogo-watermark.png"
	MakeWatermark(WatermarkPath, path)
}

func TestAddWatermark(t *testing.T) {
	path := "../input.jpg"
	WatermarkPath = "../watermarks/gogo-watermark.png"
	watermark := MakeWatermark(WatermarkPath, path)
	AddWatermark(path, watermark)
}
