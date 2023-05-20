package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

func JoinChunks(path, title string) string {
	config := GetConfig()

	outfile := filepath.Join(config.DownloadLocation, title)
	outfile = fmt.Sprintf("%s.mkv", outfile)

	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-allowed_extensions", "ts,m3u8", "-i", path, "-c", "copy", outfile)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	return outfile
}

