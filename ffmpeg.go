package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

func JoinChunks(path, topic string) {
	config := GetConfig()

	outfile := (filepath.Join(config.DownloadLocation, topic))
	fmt.Println(outfile)

	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-allowed_extensions", "ts,m3u8", "-i", path, "-c", "copy", fmt.Sprintf("%s.mkv", outfile))
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

