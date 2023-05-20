package main

import (
	"log"
	"os/exec"
)

func main() {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatalf("Please add ffmpeg to your path")
	}

	LoginAndSetToken()
	courses := GetCourses()

	courseIndex := ChooseCourse(courses)
	lectures := GetLectures(courses[courseIndex])

	startLectureIndex, endLectureIndex := ChooseLectures(lectures)
	GetMetadata(lectures[startLectureIndex : endLectureIndex+1])
}
