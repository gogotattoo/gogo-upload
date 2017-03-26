package cli

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
		hash, er := sh.AddDir(outputPath)
		if er != nil {
			log.Println("Error", er)
		}
		if len(hash) == 0 {
			log.Println("ipfs api is not running... attempting to launch it")
			daemonCmd := exec.Command("ipfs", "daemon")
			daemonCmd.Stdout = os.Stdout
			daemonCmd.Stderr = os.Stderr
			go daemonCmd.Start()
			for len(hash) == 0 {
				time.Sleep(time.Second * 20)
				log.Print("Trying after 20 seconds... ")
				sh = gia.NewShell("localhost:5001")
				hash, er = sh.AddDir(outputPath)
				if er != nil {
					log.Println("Error", er)
				}
				log.Println("Hash: ", hash)
			}
			log.Println("Ok")
		}
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
