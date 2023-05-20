package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CreateDirInsideDownloads(dirName string) string {
	config := GetConfig()
	err := os.MkdirAll(config.DownloadLocation, 0755)
	if err != nil {
		fmt.Printf("Could not create downloads directory %s with err %v\n", config.DownloadLocation, err)
		panic(err)
	}

	// Remove slashes from course name
	dirName = strings.ReplaceAll(dirName, "/", "_")
	dirName = strings.ReplaceAll(dirName, "\\", "_")

	dirPath := filepath.Join(config.DownloadLocation, dirName)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		fmt.Printf("Could not create directory %s with err %v\n", dirPath, err)
		panic(err)
	}

	config.DownloadLocation = dirPath

	return dirPath
}

func RemoveFile(path string) {
	if err := os.Remove(path); err != nil {
		fmt.Printf("Could not remove %s because %v", path, err)
	}
}
