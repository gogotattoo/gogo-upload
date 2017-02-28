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
	if strings.HasSuffix(strings.ToLower(path), ".jpg") && !strings.Contains(path, "._") {
		log.Println(path)
		file, _ := os.Open(watermark.WatermarkPath)
		defer file.Close()
		outputPath := watermark.AddWatermark(path, watermark.MakeWatermark(file, path))
		hash, _ := sh.AddDir(outputPath)
		hashes = append(hashes, hash)
		log.Println("Hash: ", hash)
	}
	return nil
}

var hashes []string
var sh *gia.Shell

// AddWatermarks to all the .jpg files in the folder and subfolders of dirPath
func AddWatermarks(dirPath string) {
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
}
