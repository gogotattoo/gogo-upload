package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	//"golang.org/x/image/font/basicfont"
	//"crypto/md5"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
	//"io"
)

type BasicImage struct {
	Path   string
	Size   int
	Suffix string
}

type WatermarkImage struct {
	Position    WatermarkPosition
	Size        int
	Path        string
	Transparent float64
}

type CreatImage struct {
	Suffix string
	Path   string
}

type WatermarkPosition int

const (
	TopLeftCorner = iota
	TopRightCorner
	BottomLeftCorner
	BottomRightCorner
	Middle
)

func (b *BasicImage) GetBasicImage(path string) *BasicImage {
	img := &BasicImage{
		Path:   path,
		Size:   100,
		Suffix: "png",
	}

	return img
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{200, 200, 200, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(label)
}

func bestCorner(image image.Image, watermarkBounds image.Rectangle) WatermarkPosition {
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
	return WatermarkPosition(bestIndex)
}

func offsetForBestCorner(resizedBounds, watermarkBounds image.Rectangle, bestCorner WatermarkPosition) image.Point {

	const WATERMARK_PADDING = 20

	offset := image.Pt(WATERMARK_PADDING, WATERMARK_PADDING)
	switch bestCorner {
	case TopRightCorner:
		offset = image.Pt(resizedBounds.Dx()-watermarkBounds.Dx()-WATERMARK_PADDING, WATERMARK_PADDING)
	case BottomLeftCorner:
		offset = image.Pt(WATERMARK_PADDING, resizedBounds.Dy()-watermarkBounds.Dy()-WATERMARK_PADDING)
	case BottomRightCorner:
		offset = image.Pt(resizedBounds.Dx()-watermarkBounds.Dx()-WATERMARK_PADDING,
			resizedBounds.Dy()-watermarkBounds.Dy()-WATERMARK_PADDING)

	}
	return offset
}
func addWatermark(inputPath string, watermark image.Image) {
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
	imgw, _ := os.Create("output/" + strings.Replace(inputPath[2:], "/", "_", -1))
	jpeg.Encode(imgw, m, &jpeg.Options{jpeg.DefaultQuality})
	defer imgw.Close()
}

func makeWatermark(path, onFile string, addLabels bool) *image.RGBA {
	wmb, _ := os.Open(path)
	watermark, _ := png.Decode(wmb)
	defer wmb.Close()
	rect := watermark.Bounds()
	labeledWatermark := image.NewRGBA(rect)

	draw.Draw(labeledWatermark, rect, watermark, rect.Min, draw.Src)
	if addLabels {
		addLabel(labeledWatermark, 50, 70, "/gogo")
		//addLabel(labeledWatermark, 190, 70, "2017/01/12")
		fileInfo, _ := os.Stat(onFile)
		addLabel(labeledWatermark, 190, 70, fileInfo.ModTime().Format("2006/01/02"))
		addLabel(labeledWatermark, 330, 70, "@chushangfeng")
	}
	return labeledWatermark
}
func main() {
	os.Mkdir("output", os.ModePerm)

	c := make(chan error)
	dirPath := os.Args[1]
	watermarkPath := "watermarks/gogo-watermark.png"
	if len(os.Args) > 2 {
		watermarkPath = os.Args[2]
	}
	go func() {
		c <- filepath.Walk(dirPath,
			func(path string, _ os.FileInfo, _ error) error {
				if strings.HasSuffix(strings.ToLower(path), ".jpg") && !strings.HasPrefix(path, "output") {
					fmt.Println(path)
					addWatermark(path, makeWatermark(watermarkPath, path, true))
				}
				return nil
			})
	}()
	<-c

}
