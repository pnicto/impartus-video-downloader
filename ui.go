package main

import (
	"fmt"
	"log"
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

	// Added the code to check the input whether it lies in the range or not
	for {
	    _, err := fmt.Scanf("%d", &choice)
	    if err != nil {
	        fmt.Println("Invalid input. Please enter a valid number.")
	        continue
	    }
	    if choice < 1 || choice > len(courses) {
	        fmt.Println("Invalid choice. Please enter a valid number within the range.")
	        continue
	    }
	    break
	}
	log.Printf("User chose %d\n", choice)
	log.Printf("Index is %d\n", choice-1)

	CreateDirInsideDownloads(courses[choice-1].SubjectName)

	return choice - 1
}

func ChooseLectures(lectures Lectures) (int, int) {
	log.Println("User entered choose lecture")

	var startIndex int
	var endIndex int

	fmt.Println("Choose a course to download from")
	fmt.Println()
	for i, lecture := range lectures {
		fmt.Printf("%3d LEC %d  %s\n", i+1, lecture.SeqNo, lecture.Topic)
	}
	fmt.Println()

	// Replaced fmt.Println("Enter a range") with below code adding a better example for the range
	fmt.Println("Enter a range (e.g., 1 5 for lectures 1 through 5):")

	// Again added the code to check if the input lies in the range
	for {
	    _, err := fmt.Scanf("%d %d", &startIndex, &endIndex)
	    if err != nil {
	        fmt.Println("Invalid input. Please enter two valid numbers separated by a space.")
	        continue
	    }
	    if startIndex < 1 || endIndex < 1 || startIndex > len(lectures) || endIndex > len(lectures) || startIndex > endIndex {
	        fmt.Println("Invalid range. Please enter a valid range within the lecture indices.")
	        continue
	    }
	    break
	}

	log.Printf("User chose %d %d\n", startIndex, endIndex)
	log.Printf("Indices are %d %d\n", startIndex-1, endIndex-1)

	return startIndex - 1, endIndex - 1
}
