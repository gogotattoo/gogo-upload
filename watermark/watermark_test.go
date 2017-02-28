package watermark

import (
	"io"
	"os"
	"testing"
)

const (
	inputPath     = "../input.jpg"
	watermarkPath = "../watermarks/gogo.png"
)

var (
	fileReader io.Reader
)

func init() {
}

func TestMakeWatermark(t *testing.T) {

	fileReader, err := os.Open(watermarkPath)
	defer fileReader.Close()
	if err != nil {
		panic("Error: cannot open " + inputPath)
	}
	MakeWatermark(fileReader, inputPath)
}

func TestAddWatermark(t *testing.T) {
	file, _ := os.Open(watermarkPath)
	defer file.Close()
	watermark := MakeWatermark(file, inputPath)
	AddWatermark(inputPath, watermark)
}

func TestAddWatermarkWithLabels(t *testing.T) {
	file, _ := os.Open(watermarkPath)
	defer file.Close()
	NeedLabels = true
	watermark := MakeWatermark(file, inputPath)
	AddWatermark(inputPath, watermark)
}

func TestAddWatermarkWithLabelsAndGivenDate(t *testing.T) {
	file, _ := os.Open(watermarkPath)
	defer file.Close()
	NeedLabels = true
	LabelDate = "2017/02/25"
	watermark := MakeWatermark(file, inputPath)
	AddWatermark(inputPath, watermark)
}
