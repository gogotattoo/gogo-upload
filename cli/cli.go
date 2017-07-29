package cli

import (
	"image"
	"image/png"
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
	log.Println(path)
	lp := strings.ToLower(path)

	var outputPath string
	if strings.HasSuffix(lp, ".mp4") || strings.HasSuffix(lp, ".mov") {
		watermark.Only = true
		watermark.NeedLabels = true
		AddWatermarks(path)
		outputPath = path[0:len(path)-4] + ".WATERMARKED.mp4"
		ffmpegCmd := exec.Command("ffmpeg",
			"-i", path,
			"-i", watermark.OutputDir+"/w.png",
			"-strict", "-2",
			"-filter_complex", "overlay=x=(main_w-overlay_w):y=(main_h-overlay_h)",
			outputPath)
		ffmpegCmd.Stdout = os.Stdout
		ffmpegCmd.Stderr = os.Stderr
		ffmpegCmd.Run()

		hash := uploadToIpfs(outputPath)
		videoHashes = append(videoHashes, hash)
		os.Rename(outputPath, outputPath[:len(outputPath)-4]+"_"+hash+".mp4")
		log.Println("Video Hash: ", hash)
		watermark.Only = false
	}
	if (strings.HasSuffix(lp, ".jpeg") || strings.HasSuffix(lp, ".jpg")) &&
		!strings.Contains(path, "._") {
		file, _ := os.Open(watermark.WatermarkPath)
		defer file.Close()
		var w *image.RGBA
		if watermark.V4 {
			w = watermark.MakeWatermarkV4(file, path)
		} else if watermark.V3 {
			w = watermark.MakeWatermarkV3(file, path)
		} else if watermark.V2 {
			w = watermark.MakeWatermarkV2(file, path)
		} else {
			w = watermark.MakeWatermark(file, path)
		}
		outputPath = watermark.AddWatermark(path, w)
		hash := uploadToIpfs(outputPath)
		hashes = append(hashes, hash)
		os.Rename(outputPath, outputPath[:len(outputPath)-4]+"_"+hash+".JPG")
		log.Println("Image Hash: ", hash)
	}
	return nil
}

func uploadToIpfs(path string) string {
	hash, er := sh.AddDir(path)
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
			hash, er = sh.AddDir(path)
			if er != nil {
				log.Println("Error", er)
			}
		}
		log.Println("Ok")
	}
	return hash
}

var hashes []string
var videoHashes []string
var sh *gia.Shell

// AddWatermarks to all the .jpg files in the folder and subfolders of dirPath
func AddWatermarks(dirPath string) []string {
	if watermark.Only {
		file, _ := os.Open(watermark.WatermarkPath)
		defer file.Close()
		var w *image.RGBA
		if watermark.V4 {
			w = watermark.MakeWatermarkV4(file, dirPath)
		} else {
			w = watermark.MakeWatermarkV3(file, dirPath)
		}
		outputPath := watermark.OutputDir + "/w.png"
		imgw, _ := os.Create(outputPath)
		png.Encode(imgw, w)
		defer imgw.Close()
		return nil
	}
	hashes = hashes[:0]
	videoHashes = videoHashes[:0]
	sh = gia.NewShell("localhost:5001")
	c := make(chan error)
	go func() {
		c <- filepath.Walk(dirPath, addWatermarkAndUpload)
	}()
	<-c
	hashesToml := "images_ipfs = ["
	for _, v := range hashes {
		hashesToml += "  \"" + v + "\",\n"
	}
	hashesToml += "]"
	log.Println(hashesToml)
	hashesToml = "videos_ipfs = ["
	for _, v := range videoHashes {
		hashesToml += "  \"" + v + "\",\n"
	}
	hashesToml += "]"
	log.Println(hashesToml)
	if len(videoHashes) != 0 {
		return videoHashes
	}
	return hashes
}
