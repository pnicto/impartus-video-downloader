package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func JoinViews(leftFile, rightFile, name string) {
	title := fmt.Sprintf("%s BOTH.mkv", name)
	outfile := filepath.Join(config.DownloadLocation, title)

	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-i", leftFile, "-i", rightFile, "-map", "0", "-map", "1", "-c", "copy", outfile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	// fmt.Println("created outfile at", outfile)
	err = os.Remove(leftFile)
	if err != nil {
		fmt.Println("Could not remove", leftFile)
	}
	// fmt.Println("Removed", leftFile)
	err = os.Remove(rightFile)
	if err != nil {
		fmt.Println("Could not remove", rightFile)
	}
	// fmt.Println("Removed", rightFile)
}

func JoinChunksFromM3U8(f string, title string) string {
	config := GetConfig()
	outfile := filepath.Join(config.DownloadLocation, title)

	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-i", f, "-c", "copy", outfile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	// fmt.Println("created outfile at", outfile)
	return outfile
}
