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
