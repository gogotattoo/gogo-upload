package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/jteeuwen/go-pkg-optarg"

	"github.com/gogotattoo/gogo-upload/watermark"
	gia "github.com/ipfs/go-ipfs-api"
)

//
// func myUsage() {
// 	fmt.Printf("Usage: %s [OPTIONS] directory ...\n", os.Args[0])
// 	flag.PrintDefaults()
// }

var dirPath string

func init() {
	flag.Header("General")
	flag.Add("i", "inputDir", "directory with newly made tattoo photos", ".")
	flag.Add("o", "outputDir", "directory with newly made tattoo photos", "~/output")
	flag.Header("Watermark")
	flag.Add("w", "watermark", "watermark image file", "watermarks/gogo.png")
	flag.Add("l", "needLabels", "set true if you want labels added on watermark", false)
	flag.Add("d", "labelDate", "date on the label, default is file's update date", "")
	flag.Add("a", "labelMadeAt", "the name of the place it was made at", "chushangfeng")
	flag.Add("b", "labelMadeBy", "the name of the artist", "gogo")

	for opt := range flag.Parse() {
		switch opt.Name {
		case "watermark":
			watermark.WatermarkPath = opt.String()
		case "labelDate":
			watermark.LabelDate = opt.String()
		case "needLabels":
			watermark.NeedLabels = opt.Bool()
		case "labelMadeAt":
			watermark.LabelMadeAt = opt.String()
		case "labelMadeBy":
			watermark.LabelMadeBy = opt.String()
		case "inputDir":
			inputDir = opt.String()
		case "outputDir":
			watermark.OutputDir = opt.String()
		}
	}

	//dirPath = flag.Remainder
}

var inputDir string

func addWatermarks(dirPath string) {
	c := make(chan error)
	var hashes []string
	go func() {
		c <- filepath.Walk(inputDir,
			func(path string, _ os.FileInfo, _ error) error {
				if strings.HasSuffix(strings.ToLower(path), ".jpg") && !strings.Contains(path, "._") {
					fmt.Println(path)
					outputPath := watermark.AddWatermark(path, watermark.MakeWatermark(watermark.WatermarkPath, path))
					hash, _ := gia.NewShell("localhost:5001").AddDir(outputPath)
					hashes = append(hashes, hash)
					fmt.Println("Hash: ", hash)
				}
				return nil
			})
	}()

	<-c
	hashes_toml := "["
	for _, v := range hashes {
		hashes_toml += "  \"" + v + "\",\n"
	}
	hashes_toml += "]"
	fmt.Println(hashes_toml)
}

func main() {
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}
	os.Mkdir(watermark.OutputDir, os.ModePerm)
	addWatermarks(dirPath)
}
