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

	for {
		_, err := fmt.Scanf("%d\n", &choice)
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

func ChooseLectures(lectures Lectures) (int, int, bool) {
	log.Println("User entered choose lecture")

	var startIndex int
	var endIndex int

	for left, right := 0, len(lectures)-1; left < right; left, right = left+1, right-1 {
		lectures[left], lectures[right] = lectures[right], lectures[left]
	}

	fmt.Println()
	fmt.Println("Choose the lecture range you want to download")
	for i, lecture := range lectures {
		fmt.Printf("%3d) LEC %3d %s\n", i+1, lecture.SeqNo, lecture.Topic)
	}
	fmt.Println()

	fmt.Println("Enter a range (e.g., 1 5 for lectures 1 through 5):")

	for {
		_, err := fmt.Scanf("%d %d\n", &startIndex, &endIndex)
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

	var skipEmptyLectures string

	fmt.Println("Skip lectures with titles like 'No class' or 'No lecture'? [Y/n]")
	fmt.Scanf("%s\n", &skipEmptyLectures)
	skipEmptyLectures = strings.ToLower(skipEmptyLectures)

	for skipEmptyLectures != "y" && skipEmptyLectures != "n" && skipEmptyLectures != "" {
		fmt.Println("Please enter a valid choice: [Y/n]")
		fmt.Scanf("%s\n", &skipEmptyLectures)
	}

	if skipEmptyLectures == "n" {
		log.Printf("User chose not to skip empty lectures\n")
	} else {
		log.Printf("User chose to skip empty lectures\n")
	}

	return startIndex - 1, endIndex - 1, (skipEmptyLectures != "n")
}
