package cli

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogotattoo/gogo-upload/watermark"
	gia "github.com/ipfs/go-ipfs-api"
)

func addWatermarkAndUpload(path string, fi os.FileInfo, err error) error {
	lp := strings.ToLower(path)
	if (strings.HasSuffix(lp, ".jpeg") || strings.HasSuffix(lp, ".jpg")) &&
		!strings.Contains(path, "._") {
		log.Println(path)
		file, _ := os.Open(watermark.WatermarkPath)
		defer file.Close()
		var outputPath string
		if watermark.V2 {
			outputPath = watermark.AddWatermark(path, watermark.MakeWatermarkV2(file, path))
		} else {
			outputPath = watermark.AddWatermark(path, watermark.MakeWatermark(file, path))
		}
		hash, _ := sh.AddDir(outputPath)
		hashes = append(hashes, hash)
		os.Rename(outputPath, outputPath[:len(outputPath)-4]+"_"+hash+".JPG")
		log.Println("Hash: ", hash)
	}
	return nil
}

var hashes []string
var sh *gia.Shell

// AddWatermarks to all the .jpg files in the folder and subfolders of dirPath
func AddWatermarks(dirPath string) []string {
	hashes = hashes[:0]
	sh = gia.NewShell("localhost:5001")
	c := make(chan error)
	go func() {
		c <- filepath.Walk(dirPath, addWatermarkAndUpload)
	}()
	<-c
	hashesToml := "["
	for _, v := range hashes {
		hashesToml += "  \"" + v + "\",\n"
	}
	hashesToml += "]"
	log.Println(hashesToml)
	return hashes
}
