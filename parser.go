package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type ParsedPlaylist struct {
	KeyURL           string
	Title            string
	FirstViewURLs    []string
	SecondViewURLs   []string
	Id               int
	SeqNo            int
	HasMultipleViews bool
}

func PlaylistParser(scanner *bufio.Scanner, id int, title string, seqNo int) ParsedPlaylist {
	var parsedOutput ParsedPlaylist
	var isFirstView = true
	var firstViewUrls []string
	var secondViewUrls []string

	parsedOutput.Id = id
	parsedOutput.Title = title
	parsedOutput.SeqNo = seqNo

	for scanner.Scan() {
		l := scanner.Text()
		if parsedOutput.KeyURL == "" && strings.HasPrefix(l, "#EXT-X-KEY") {
			pattern := `URI="([^"]+)"`
			re := regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(l)
			if len(match) == 2 {
				url := match[1]
				parsedOutput.KeyURL = url
			} else {
				fmt.Println("URL not found in the line.")
			}
		} else if strings.HasPrefix(l, "#EXT-X-DISCONTINUITY") {
			isFirstView = false
		} else if !strings.HasPrefix(l, "#EXT") {
			if isFirstView {
				firstViewUrls = append(firstViewUrls, l)
			} else {
				secondViewUrls = append(secondViewUrls, l)
			}
		}
	}

	if isFirstView {
		parsedOutput.HasMultipleViews = false
		parsedOutput.FirstViewURLs = firstViewUrls
	} else {
		parsedOutput.HasMultipleViews = true
		parsedOutput.FirstViewURLs = firstViewUrls
		parsedOutput.SecondViewURLs = secondViewUrls
	}

	return parsedOutput
}
