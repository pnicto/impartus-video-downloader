package main

import (
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

	startLectureIndex, endLectureIndex := ChooseLectures(lectures)
	GetMetadata(lectures[startLectureIndex : endLectureIndex+1])
}
