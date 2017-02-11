package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogotattoo/gogo-upload/watermark"
	gia "github.com/ipfs/go-ipfs-api"
)

func myUsage() {
	fmt.Printf("Usage: %s [OPTIONS] directory ...\n", os.Args[0])
	flag.PrintDefaults()
}

var dirPath string

func init() {
	flag.StringVar(&watermark.WatermarkPath, "watermark", "watermarks/gogo.png", "watermark image file")
	flag.StringVar(&watermark.LabelDate, "labelDate", "", "date on the label, default is file's update date")
	flag.BoolVar(&watermark.NeedLabels, "needLabels", false, "set true if you want labels added on watermark")
	flag.StringVar(&watermark.LabelMadeAt, "labelMadeAt", "chushangfeng", "the name of the place it was made at")
	flag.StringVar(&watermark.LabelMadeBy, "labelMadeBy", "gogo", "the name of the artist")

	flag.Usage = myUsage
	if len(os.Args) == 1 {
		myUsage()
		os.Exit(1)
	}
	flag.Parse()
	dirPath = flag.Args()[0]
}

func addWatermarks(dirPath string) {
	c := make(chan error)
	var hashes []string
	go func() {
		c <- filepath.Walk(dirPath,
			func(path string, _ os.FileInfo, _ error) error {
				if strings.HasSuffix(strings.ToLower(path), ".jpg") && !strings.HasPrefix(path, "output") {
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
	os.Mkdir("output", os.ModePerm)
	watermark.WatermarkPath = "/Users/delirium/workspace/gogo.tattoo/gogo-upload/watermarks/gogo-watermark.png"
	addWatermarks(dirPath)
}
