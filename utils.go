package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CreateDirInsideDownloads(dirName string) string {
	log.Printf("Creating downloads directory for %s\n", dirName)

	config := GetConfig()
	err := os.MkdirAll(config.DownloadLocation, 0755)
	if err != nil {
		log.Fatalf("Could not create downloads directory %s with err %v\n", config.DownloadLocation, err)
	}

	// Remove slashes from course name
	dirName = strings.ReplaceAll(dirName, "/", "_")
	dirName = strings.ReplaceAll(dirName, "\\", "_")

	dirPath := filepath.Join(config.DownloadLocation, dirName)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatalf("Could not create directory %s with err %v\n", dirPath, err)
	}

	config.DownloadLocation = dirPath
	log.Printf("Created downloads directory for %s\n", dirName)

	return dirPath
}

func RemoveFile(path string) {
	if err := os.Remove(path); err != nil {
		fmt.Printf("Could not remove %s because %v\n", path, err)
	}
}

// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func removeEmptyLectures(lectures Lectures) Lectures {
	filteredLectures := lectures[:0]
	for _, lecture := range lectures {
		lowercaseTitle := strings.ToLower(lecture.Topic)
		if !(lowercaseTitle == "no class" || lowercaseTitle == "no lecture") {
			filteredLectures = append(filteredLectures, lecture)
		}
	}
	return filteredLectures
}
