package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatalln("Please add ffmpeg to your path")
	}

	// Logging
	logFile, err := os.OpenFile("run.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Could not start logs")
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	LoginAndSetToken()
	courses := GetCourses()

	courseIndex := ChooseCourse(courses)
	lectures := GetLectures(courses[courseIndex])

	startLectureIndex, endLectureIndex, skipEmptyLectures := ChooseLectures(lectures)
	var chosenLectures Lectures
	if skipEmptyLectures {
		chosenLectures = removeEmptyLectures(lectures[startLectureIndex : endLectureIndex+1])
	} else {
		chosenLectures = lectures[startLectureIndex : endLectureIndex+1]
	}

	downloadedPlaylists := DownloadPlaylist(GetPlaylist(chosenLectures))
	metadataFiles := CreateTempM3U8Files(downloadedPlaylists)
	config := GetConfig()

	for _, file := range metadataFiles {
		var left, right string
		if file.FirstViewFile != "" && config.Views != "left" {
			left = JoinChunksFromM3U8(file.FirstViewFile, fmt.Sprintf("LEC %03d %s RIGHT VIEW.mp4", file.Playlist.SeqNo, file.Playlist.Title))
		}

		if file.SecondViewFile != "" && config.Views != "right" {
			right = JoinChunksFromM3U8(file.SecondViewFile, fmt.Sprintf("LEC %03d %s LEFT VIEW.mp4", file.Playlist.SeqNo, file.Playlist.Title))
		}

		if left != "" && right != "" && config.Views == "both" {
			JoinViews(left, right, fmt.Sprintf("LEC %03d %s", file.Playlist.SeqNo, file.Playlist.Title))
		}
	}
}
