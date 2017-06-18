package main

import (
	"os"

	flag "github.com/jteeuwen/go-pkg-optarg"

	"github.com/gogotattoo/gogo-upload/cli"
	"github.com/gogotattoo/gogo-upload/watermark"
)

const (
	defaultInputDir                 = "."
	defaultOutputDir                = defaultInputDir
	defaultAuthor                   = "gogo"
	defaultPlace                    = "chushangfeng"
	defaultWatermarkWithLabelPath   = "watermarks/gogo-watermark.png"
	defaultWatermarkWithLabelV2Path = "watermarks/gogo-label-v2.png"
	defaultWatermarkPath            = "watermarks/gogo.png"
)

var inputDir string

func init() {
	flag.Header("General")
	flag.Add("i", "inputDir", "directory with newly made tattoo photos", defaultInputDir)
	flag.Add("o", "outputDir", "directory with newly made tattoo photos", "[same with input]")
	flag.Header("Watermark")
	flag.Add("w", "watermark", "watermark image file", defaultWatermarkPath)
	flag.Add("l", "needLabels", "set true if you want labels added on watermark", false)
	flag.Add("d", "labelDate", "date on the label, default is file's update date", "[file date]")
	flag.Add("a", "labelMadeAt", "the name of the place it was made at", defaultPlace)
	flag.Add("b", "labelMadeBy", "the name of the artist", defaultAuthor)
	flag.Add("v2", "watermarkVersion2", "the second generation of our watermarks", true)
	flag.Add("v3", "watermarkVersion3", "the third awesome generation of our watermarks", true)

	// Default values
	watermark.WatermarkPath = defaultWatermarkPath
	watermark.LabelMadeAt = defaultPlace
	watermark.LabelMadeBy = defaultAuthor
	inputDir = defaultInputDir

	watermarkFileOverriten := false
	for opt := range flag.Parse() {
		switch opt.Name {
		case "watermark":
			watermarkFileOverriten = true
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
		case "watermarkVersion2":
			watermark.V2 = true
			watermark.V3 = false
		case "watermarkVersion3":
			watermark.V2 = false
			watermark.V3 = true
		}
	}

	if len(watermark.OutputDir) == 0 {
		watermark.OutputDir = inputDir
	}

	if watermark.NeedLabels && !watermarkFileOverriten {
		if watermark.V2 {
			watermark.WatermarkPath = defaultWatermarkWithLabelV2Path
		} else {
			watermark.WatermarkPath = defaultWatermarkWithLabelPath
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}
	os.Mkdir(watermark.OutputDir, os.ModePerm)
	cli.AddWatermarks(inputDir)
}
