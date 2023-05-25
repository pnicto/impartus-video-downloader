package main

import (
	"fmt"
	"log"
	"strings"
)

func ChooseCourse(courses Courses) int {
	log.Println("User entered choose course")

	var choice int

	fmt.Println("Choose a course to download from")
	fmt.Println()
	for i, course := range courses {
		fmt.Printf("%3d %s\n", i+1, course.SubjectName)
	}
	fmt.Println()
	fmt.Println("Enter a number")
	// TODO: Check input is within range and a number
	fmt.Scanf("%d\n", &choice)
	log.Printf("User chose %d\n", choice)
	log.Printf("Index is %d\n", choice-1)

	CreateDirInsideDownloads(courses[choice-1].SubjectName)

	return choice - 1
}

func ChooseLectures(lectures Lectures) (int, int, bool) {
	log.Println("User entered choose lecture")

	var startIndex int
	var endIndex int

	fmt.Println("Choose the lecture range you want to download")
	fmt.Println()
	for i, lecture := range lectures {
		fmt.Printf("%3d LEC %d  %s\n", i+1, lecture.SeqNo, lecture.Topic)
	}
	fmt.Println()

	// TODO: Add better range examples here
	fmt.Println("Enter a range")
	// TODO: Check input is within range and a number
	fmt.Scanf("%d %d\n", &startIndex, &endIndex)

	log.Printf("User chose %d %d\n", startIndex, endIndex)
	log.Printf("Indices are %d %d\n", startIndex-1, endIndex-1)

	var skipEmptyLectures string

	fmt.Println("Skip lectures with titles like 'No class' or 'No lecture'? [Y/n]")
	fmt.Scanf("%s\n", &skipEmptyLectures)
	skipEmptyLectures = strings.ToLower(skipEmptyLectures)

	for skipEmptyLectures != "y" && skipEmptyLectures != "n" && skipEmptyLectures != "" {
		fmt.Println("Please enter a valid choice: [Y/n]")
		fmt.Scanf("%s\n", &skipEmptyLectures)
	}

	if skipEmptyLectures == "" {
		fmt.Println("Skipping empty lectures by default")
	}

	if skipEmptyLectures == "n" {
		log.Printf("User chose not to skip empty lectures\n")
	} else {
		log.Printf("User chose to skip empty lectures\n")
	}

	return startIndex - 1, endIndex - 1, (skipEmptyLectures != "n")
}
