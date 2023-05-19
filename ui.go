package main

import "fmt"

func ChooseCourse(courses Courses) int {
	var choice int

	fmt.Println("Choose a course to download from")
	fmt.Println()
	for i, course := range courses {
		fmt.Printf("%3d %s\n", i+1, course.SubjectName)
	}
	fmt.Println()
	fmt.Println("Enter a number")
	// TODO: Check input is within range and a number
	fmt.Scanf("%d", &choice)

	CreateDirInsideDownloads(courses[choice-1].SubjectName)

	return choice - 1
}

func ChooseLectures(lectures Lectures) (int, int) {
	var startIndex int
	var endIndex int

	fmt.Println("Choose a course to download from")
	fmt.Println()
	for i, lecture := range lectures {
		fmt.Printf("%3d LEC %d  %s\n", i+1, lecture.SeqNo, lecture.Topic)
	}
	fmt.Println()

	// TODO: Add better range examples here
	fmt.Println("Enter a range")
	// TODO: Check input is within range and a number
	fmt.Scanf("%d %d", &startIndex, &endIndex)

	return startIndex - 1, endIndex - 1
}
