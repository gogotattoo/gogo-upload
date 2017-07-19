package watermark

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	//"golang.org/x/image/font/basicfont"
	//"crypto/md5"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
	//"io"
)

// BasicImage TODO: This should be used
type BasicImage struct {
	Path   string
	Size   int
	Suffix string
}

// WMImage represents an image with a watermark
type WMImage struct {
	Position    WMPosition
	Size        int
	Path        string
	Transparent float64
}

// WMPosition defines where the watermark will be placed
type WMPosition int

const (
	// TopLeftCorner puts watermark in the *top left* corner of the final image
	TopLeftCorner = iota
	// TopRightCorner places watermark in the *top right* corner of the final image
	TopRightCorner
	// BottomLeftCorner places watermark in the *bottom left* corner of the final image
	BottomLeftCorner
	// BottomRightCorner places watermark in the *bottom right* corner of the final image
	BottomRightCorner
	// Middle places watermark in the *middle* of the final image
	Middle
)

func (b *BasicImage) getBasicImage(path string) *BasicImage {
	img := &BasicImage{
		Path:   path,
		Size:   100,
		Suffix: "png",
	}

	return img
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{200, 200, 200, 255}
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(label)
}

func bestCorner(image image.Image, watermarkBounds image.Rectangle) WMPosition {
	bounds := image.Bounds()
	// Testirng the top left corner
	var rgb [4]float32
	for x := 0; x < watermarkBounds.Dx(); x++ {
		for y := 0; y < watermarkBounds.Dy(); y++ {
			r, g, b, a := image.At(x, y).RGBA()
			rgb[0] += 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b) + float32(a)
		}
	}
	// Top Right
	for x := bounds.Dx() - watermarkBounds.Dx(); x < bounds.Dx(); x++ {
		for y := 0; y < watermarkBounds.Dy(); y++ {
			r, g, b, a := image.At(x, y).RGBA()
			rgb[1] += 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b) + float32(a)
		}
	}
	// Bottom Left
	for x := 0; x < watermarkBounds.Dx(); x++ {
		for y := bounds.Dy() - watermarkBounds.Dy(); y < bounds.Dy(); y++ {
			r, g, b, a := image.At(x, y).RGBA()
			rgb[2] += 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b) + float32(a)
		}
	}
	// Bottom Right
	for x := bounds.Dx() - watermarkBounds.Dx(); x < bounds.Dx(); x++ {
		for y := bounds.Dy() - watermarkBounds.Dy(); y < bounds.Dy(); y++ {
			r, g, b, a := image.At(x, y).RGBA()
			rgb[3] += 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b) + float32(a)
		}
	}
	fmt.Println(rgb)
	bestIndex := 0
	var prevBest float32
	for index, value := range rgb {
		if value > prevBest {
			prevBest = value
			bestIndex = index
		}
		fmt.Println(value, index)
	}
	return WMPosition(bestIndex)
}

func offsetForBestCorner(resizedBounds, watermarkBounds image.Rectangle, bestCorner WMPosition) image.Point {

	const WMPadding = 20

	offset := image.Pt(WMPadding, WMPadding)
	switch bestCorner {
	case TopRightCorner:
		offset = image.Pt(resizedBounds.Dx()-watermarkBounds.Dx()-WMPadding, WMPadding)
	case BottomLeftCorner:
		offset = image.Pt(WMPadding, resizedBounds.Dy()-watermarkBounds.Dy()-WMPadding)
	case BottomRightCorner:
		offset = image.Pt(resizedBounds.Dx()-watermarkBounds.Dx()-WMPadding,
			resizedBounds.Dy()-watermarkBounds.Dy()-WMPadding)

	}
	return offset
}

// AddWatermark adds a given image.Image watermark to the file provided(path),
// returns the path to the output file
// TODO: improve output file name format
// suggestions:
//		- 2006_01_02_artist_tattooname_shopname_ipfshash.jpg
func AddWatermark(inputPath string, watermark image.Image) string {
	imgb, _ := os.Open(inputPath)
	img, err := jpeg.Decode(imgb)
	if err != nil {
		fmt.Print(err)
	}
	defer imgb.Close()

	resizedImage := resize.Resize(2048, 0, img, resize.Lanczos3)
	resizedBounds := resizedImage.Bounds()
	watermarkBounds := watermark.Bounds()

	// Let's try to find the brightest corner for our watermark
	bestCorner := bestCorner(resizedImage, watermarkBounds)
	offset := offsetForBestCorner(resizedBounds, watermarkBounds, bestCorner)
	fmt.Println("BestCorner: ", bestCorner)
	//    wb := watermark.Bounds()
	// fmt.Println("Original image: ", img.Bounds().Dx(), " x ", img.Bounds().Dy())
	// fmt.Println(rect.Dx(), " x ", rect.Dy())
	// fmt.Println(wb.Dx(), " x ", wb.Dy())

	m := image.NewRGBA(resizedBounds)
	draw.Draw(m, resizedBounds, resizedImage, image.ZP, draw.Src)
	draw.Draw(m, watermarkBounds.Add(offset), watermark, image.ZP, draw.Over)

	// h := md5.New()
	// io.WriteString(h, inputPath)
	// imgw, _ := os.Create("output/" + fmt.Sprintf("%x", h.Sum(nil)) + ".jpg")
	outputPath := OutputDir + "/" + strings.Replace(inputPath[:], "/", "_", -1)
	imgw, _ := os.Create(outputPath)
	jpeg.Encode(imgw, m, &jpeg.Options{Quality: jpeg.DefaultQuality})
	defer imgw.Close()
	return outputPath
}

// MakeWatermarkV3 creates an *image.RGBA with given watermark
// Parameter NeedLabels defines if date, artist and place should be added to the
// label
func MakeWatermarkV3(wmReader io.Reader, onFile string) (labeledWatermark *image.RGBA) {
	wmk, _ := png.Decode(wmReader)
	rect := wmk.Bounds()
	labeledWatermark = image.NewRGBA(rect)

	draw.Draw(labeledWatermark, rect, wmk, rect.Min, draw.Src)
	if !NeedLabels {
		return labeledWatermark
	}
	addLabel(labeledWatermark, 70, 90, "/"+LabelMadeBy)
	//addLabel(labeledWatermark, 190, 70, "2017/01/16")
	fileInfo, _ := os.Stat(onFile)
	date := fileInfo.ModTime().Format("2006/01/02")
	if len(LabelDate) != 0 {
		date = LabelDate
	}
	addLabel(labeledWatermark, 220, 90, date)
	if len(LabelMadeAt) != 0 {
		addLabel(labeledWatermark, 380, 90, "@"+LabelMadeAt)
	}
	fmt.Println("Date:", date)
	return
}

// MakeWatermarkV2 creates an *image.RGBA with given watermark
// Parameter NeedLabels defines if date, artist and place should be added to the
// label
func MakeWatermarkV2(wmReader io.Reader, onFile string) (labeledWatermark *image.RGBA) {
	wmk, _ := png.Decode(wmReader)
	rect := wmk.Bounds()

	resizedWatermark := resize.Resize(uint(wmk.Bounds().Max.X/2), 0, wmk, resize.Lanczos3)
	resizedBounds := resizedWatermark.Bounds()
	labeledWatermark = image.NewRGBA(resizedBounds)

	draw.Draw(labeledWatermark, rect, resizedWatermark, rect.Min, draw.Src)
	if !NeedLabels {
		return labeledWatermark
	}
	leftOffset := 85
	topOffset := 90
	distOffset := 200
	addLabel(labeledWatermark, leftOffset, topOffset, "/"+LabelMadeBy)
	//addLabel(labeledWatermark, 190, 70, "2017/01/16")
	fileInfo, _ := os.Stat(onFile)
	date := fileInfo.ModTime().Format("2006/01/02")
	if len(LabelDate) != 0 {
		date = LabelDate
	}
	addLabel(labeledWatermark, leftOffset+distOffset, topOffset, date)
	if len(LabelMadeAt) != 0 {
		addLabel(labeledWatermark, leftOffset+2*distOffset, topOffset, "@"+LabelMadeAt)
	}
	fmt.Println("Date:", date)
	return
}

// MakeWatermark creates an *image.RGBA with given watermark
// Parameter NeedLabels defines if date, artist and place should be added to the
// label
func MakeWatermark(wmReader io.Reader, onFile string) (labeledWatermark *image.RGBA) {
	wmk, _ := png.Decode(wmReader)
	rect := wmk.Bounds()
	labeledWatermark = image.NewRGBA(rect)

	draw.Draw(labeledWatermark, rect, wmk, rect.Min, draw.Src)
	if !NeedLabels {
		return labeledWatermark
	}
	addLabel(labeledWatermark, 50, 70, "/"+LabelMadeBy)
	//addLabel(labeledWatermark, 190, 70, "2017/01/16")
	fileInfo, _ := os.Stat(onFile)
	date := fileInfo.ModTime().Format("2006/01/02")
	if len(LabelDate) != 0 {
		date = LabelDate
	}
	addLabel(labeledWatermark, 190, 70, date)
	if len(LabelMadeAt) != 0 {
		addLabel(labeledWatermark, 330, 70, "@"+LabelMadeAt)
	}
	fmt.Println("Date:", date)
	return
}

var (
	// OutputDir global output directory parameter
	OutputDir string
	// WatermarkPath defines the path to the watermark png image, it should be of the
	// proper format: check examples provided for our masters in /watermars foulder
	WatermarkPath string
	// NeedLabels defines if date, artist name and shop name should be added on the watermark
	NeedLabels bool
	// LabelDate format: 2006/01/02
	LabelDate string
	// LabelMadeAt the name of the place it was made at
	LabelMadeAt string
	// LabelMadeBy default: gogo
	LabelMadeBy string
	// V2 - set true to use the second generation of watermarks
	V2 bool
	// V3 - set true to use the third generation of watermarks
	V3 bool
	// Only to generate watermark only
	Only bool
)
