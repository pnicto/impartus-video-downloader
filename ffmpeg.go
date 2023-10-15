package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func JoinChunks(path, title string) string {
	config := GetConfig()

	outfile := filepath.Join(config.DownloadLocation, title)
	outfile = fmt.Sprintf("%s.mkv", outfile)

	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-allowed_extensions", "ts,m3u8", "-threads", fmt.Sprint(config.Threads), "-i", path, "-c", "copy", outfile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}

	return outfile
}

func JoinViews(leftFile, rightFile, title string) {
	config := GetConfig()
	outfile := fmt.Sprintf("%s BOTH.mkv", leftFile[:len(leftFile)-9])

	cmd := exec.Command("ffmpeg", "-y", "-threads", fmt.Sprint(config.Threads), "-hide_banner", "-i", rightFile, "-i", leftFile, "-map", "0", "-map", "1", "-c", "copy", outfile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}

	fmt.Println(outfile)
}

func JoinChunksFromM3U8(f string, title string) string {
	config := GetConfig()
	outfile := filepath.Join(config.DownloadLocation, title)

	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-i", f, "-c", "copy", outfile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	fmt.Println("created outfile at", outfile)
	return outfile
}
