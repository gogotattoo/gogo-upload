package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogotattoo/gogo-upload/watermark"
	gia "github.com/ipfs/go-ipfs-api"
)

func addWatermarkAndUpload(path string, fi os.FileInfo, err error) error {
	if strings.HasSuffix(strings.ToLower(path), ".jpg") && !strings.Contains(path, "._") {
		fmt.Println(path)
		outputPath := watermark.AddWatermark(path, watermark.MakeWatermark(watermark.WatermarkPath, path))
		hash, _ := sh.AddDir(outputPath)
		hashes = append(hashes, hash)
		fmt.Println("Hash: ", hash)
	}
	return nil
}

var hashes []string
var sh *gia.Shell

func AddWatermarks(dirPath string) {
	sh = gia.NewShell("localhost:5001")
	c := make(chan error)
	go func() {
		c <- filepath.Walk(dirPath, addWatermarkAndUpload)
	}()
	<-c

	hashes_toml := "["
	for _, v := range hashes {
		hashes_toml += "  \"" + v + "\",\n"
	}
	hashes_toml += "]"
	fmt.Println(hashes_toml)
}
