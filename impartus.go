package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type (
	LoginResponse struct {
		Success  bool   `json:"success"`
		Message  string `json:"message"`
		UserType int    `json:"userType"`
		Token    string `json:"token"`
	}

	Courses []Course
	Course  struct {
		SubjectID            int    `json:"subjectId"`
		SubjectName          string `json:"subjectName"`
		SessionID            int    `json:"sessionId"`
		SessionName          string `json:"sessionName"`
		ProfessorID          int    `json:"professorId"`
		ProfessorName        string `json:"professorName"`
		DepartmentID         int    `json:"departmentId"`
		Department           string `json:"department"`
		InstituteID          int    `json:"instituteId"`
		Institute            string `json:"institute"`
		Coverpic             string `json:"coverpic"`
		VideoCount           int    `json:"videoCount"`
		FlippedLecturesCount int    `json:"flippedLecturesCount"`
	}

	Lectures []Lecture
	Lecture  struct {
		Type                int    `json:"type"`
		Ttid                int    `json:"ttid"`
		SeqNo               int    `json:"seqNo"`
		Status              int    `json:"status"`
		VideoID             int    `json:"videoId"`
		SubjectID           int    `json:"subjectId"`
		SubjectName         string `json:"subjectName"`
		Selfenroll          int    `json:"selfenroll"`
		Coverpic            string `json:"coverpic"`
		SubjectCode         string `json:"subjectCode"`
		SubjectDescription  any    `json:"subjectDescription"`
		InstituteID         int    `json:"instituteId"`
		Institute           string `json:"institute"`
		DepartmentID        int    `json:"departmentId"`
		Department          string `json:"department"`
		ClassroomID         int    `json:"classroomId"`
		ClassroomName       string `json:"classroomName"`
		SessionID           int    `json:"sessionId"`
		SessionName         string `json:"sessionName"`
		Topic               string `json:"topic"`
		ProfessorID         int    `json:"professorId"`
		ProfessorName       string `json:"professorName"`
		ProfessorImageURL   string `json:"professorImageUrl"`
		StartTime           string `json:"startTime"`
		EndTime             string `json:"endTime"`
		ActualDuration      int    `json:"actualDuration"`
		TapNToggle          int    `json:"tapNToggle"`
		FilePath            string `json:"filePath"`
		FilePath2           string `json:"filePath2"`
		SlideCount          int    `json:"slideCount"`
		Noaudio             int    `json:"noaudio"`
		Views               int    `json:"views"`
		DocumentCount       int    `json:"documentCount"`
		LessonPlanAvailable int    `json:"lessonPlanAvailable"`
		Trending            int    `json:"trending"`
		LastPosition        int    `json:"lastPosition"`
	}
)

func LoginAndSetToken() {
	log.Println("Attempt to login")

	config := GetConfig()
	url := fmt.Sprintf("%s/auth/signin", config.BaseUrl)

	requestBody, err := json.Marshal(map[string]string{"username": config.Username, "password": config.Password})
	if err != nil {
		log.Fatalf("Could not marshal login body %v", err)
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Login failed with error %v", err)
	}

	if response.StatusCode == http.StatusUnauthorized {
		fmt.Println("Wrong credentials please retry")
		log.Fatalln("Wrong credentials please retry")
	}

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalln("Error in reading the response body when login failed")
		}

		log.Println(url, string(body))
		log.Fatalln("Something went wrong please create a new issue on github")
	}

	var loginResponse LoginResponse
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	if err != nil {
		log.Fatalf("Could not decode login body %v", err)
	}

	config.Token = loginResponse.Token
	log.Printf("Token set with length %d\n", len(config.Token))
}

func GetCourses() Courses {
	log.Println("Getting courses")

	var courses Courses
	config := GetConfig()

	url := fmt.Sprintf("%s/subjects", config.BaseUrl)
	resp := GetClientAuthorized(url, config.Token)
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&courses)
	if err != nil {
		log.Fatalf("Could not decode response %v", err)
	}

	log.Printf("Fetched %d courses\n", len(courses))
	log.Println(courses)

	return courses
}

func GetLectures(course Course) Lectures {
	log.Println("Getting lectures")

	var lectures Lectures
	config := GetConfig()

	url := fmt.Sprintf("%s/subjects/%d/lectures/%d", config.BaseUrl, course.SubjectID, course.SessionID)
	resp := GetClientAuthorized(url, config.Token)
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&lectures)
	if err != nil {
		log.Fatalf("Could not decode response %v", err)
	}

	log.Printf("Fetched %d lectures\n", len(lectures))
	log.Println(lectures)

	return lectures
}

func getDecryptionKey(encryptionKey []byte) []byte {
	encryptionKey = encryptionKey[2:]
	for i, j := 0, len(encryptionKey)-1; i < j; i, j = i+1, j-1 {
		encryptionKey[i], encryptionKey[j] = encryptionKey[j], encryptionKey[i]
	}

	return encryptionKey
}

func getResolution(quality string) string {
	var resolution string

	switch quality {
	case "720":
		resolution = "1280x720"
	case "450":
		resolution = "800x450"
	case "144":
		resolution = "256x144"
	}

	return resolution
}

func createTempM3U8File(ttid int, view string) (*os.File, string) {
	config := GetConfig()

	err := os.MkdirAll(config.TempDirLocation, 0755)
	if err != nil {
		fmt.Printf("Could not create temp directory %s with err %v\n", config.TempDirLocation, err)
		panic(err)
	}

	tempM3U8File := filepath.Join(config.TempDirLocation, fmt.Sprintf("%d_%s.m3u8", ttid, view))

	f, err := os.Create(tempM3U8File)
	if err != nil {
		fmt.Printf("Could not create temp m3u8 file for ttid %d with error %v", ttid, err)
	}

	return f, tempM3U8File
}

func getM3U8(ttid int) string {
	config := GetConfig()
	resolution := getResolution(config.Quality)
	url := fmt.Sprintf("%s/fetchvideo?tag=LC&inm3u8=http%%3A%%2F%%2F172.16.3.45%%2F%%2Fdownload1%%2F%d_hls%%2F%s_27%%2F%s_27.m3u8", config.BaseUrl, ttid, resolution, resolution)

	resp := GetClientAuthorized(url, config.Token)
	defer resp.Body.Close()

	m3u8Data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Could not read m3u8 data %v", err)
		panic(err)
	}

	return string(m3u8Data)
}

func downloadChunk(ttid int, resolution string, view string, chunk int) string {
	config := GetConfig()
	chunkUrl := fmt.Sprintf("%s/fetchvideo?ts=http%%3A%%2F%%2F172.16.3.45%%2F%%2Fdownload1%%2F%d_hls%%2F%s_27%%2F%s_27%s_%04d_hls_0.ts", config.BaseUrl, ttid, resolution, resolution, view, chunk)

	resp := GetClientAuthorized(chunkUrl, config.Token)
	defer resp.Body.Close()

	outFilepath := filepath.Join(config.TempDirLocation, fmt.Sprintf("%d_%04d_%s.ts.temp", ttid, chunk, view))
	outFile, err := os.Create(outFilepath)
	if err != nil {
		fmt.Printf("Could not download chunk %d %d %v", ttid, chunk, err)
	}

	outFileContent, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Could not read chunk %d %d %v", ttid, chunk, err)
	}

	_, err = outFile.Write(outFileContent)
	if err != nil {
		fmt.Printf("Could not write chunk %d %d %v", ttid, chunk, err)
	}

	return outFilepath
}

// TODO: Refine decryptChunk
func decryptChunk(filePath string, key []byte) {
	outPath := filePath[:len(filePath)-5]

	infile, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Could not open chunk %s", filePath)
	}

	length := 16 - (len(infile) % 16)
	infile = append(infile, bytes.Repeat([]byte{byte(length)}, length)...)

	iv := bytes.Repeat([]byte{0}, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(infile))
	mode.CryptBlocks(plainText, infile)

	err = os.WriteFile(outPath, plainText, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func writeM3U8FileConditionally(leftFile, rightFile *os.File, leftContent, rightContent string) {
	config := GetConfig()

	switch config.Views {
	case "left":
		_, err := leftFile.WriteString(leftContent)
		if err != nil {
			log.Fatalf("Could not write to left m3u8 file")
		}
	case "right":
		_, err := rightFile.WriteString(rightContent)
		if err != nil {
			log.Fatalf("Could not write to left m3u8 file")
		}
	case "both":
		_, err := leftFile.WriteString(leftContent)
		if err != nil {
			log.Fatalf("Could not write to left m3u8 file")
		}
		_, err = rightFile.WriteString(rightContent)
		if err != nil {
			log.Fatalf("Could not write to left m3u8 file")
		}
	}
}

func downloadChunkConditonally(ttid int, resolution string, chunk int, decryptionKey []byte) {
	config := GetConfig()

	switch config.Views {
	case "left":
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			chunkPath := downloadChunk(ttid, resolution, "v3", chunk)
			decryptChunk(chunkPath, decryptionKey)
			RemoveFile(chunkPath)
		}()
		wg.Wait()

	case "right":
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			chunkPath := downloadChunk(ttid, resolution, "v1", chunk)
			decryptChunk(chunkPath, decryptionKey)
			RemoveFile(chunkPath)
		}()
		wg.Wait()

	case "both":
		var wg sync.WaitGroup

		wg.Add(2)

		go func() {
			defer wg.Done()
			chunkPathLeft := downloadChunk(ttid, resolution, "v3", chunk)
			decryptChunk(chunkPathLeft, decryptionKey)
			RemoveFile(chunkPathLeft)
		}()

		go func() {
			defer wg.Done()
			chunkPathRight := downloadChunk(ttid, resolution, "v1", chunk)
			decryptChunk(chunkPathRight, decryptionKey)
			RemoveFile(chunkPathRight)
		}()

		wg.Wait()
	}
}

func joinChunksConditionally(leftFilePath, rightFilePath, titleLeft, titleRight string) {
	config := GetConfig()

	switch config.Views {
	case "left":
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			JoinChunks(leftFilePath, titleLeft)
		}()
		wg.Wait()

	case "right":
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			JoinChunks(rightFilePath, titleRight)
		}()
		wg.Wait()

	case "both":
		var wg sync.WaitGroup
		leftChan := make(chan string)
		rightChan := make(chan string)

		go func() {
			leftChan <- JoinChunks(leftFilePath, titleLeft)
		}()
		go func() {
			rightChan <- JoinChunks(rightFilePath, titleRight)
		}()

		leftOutFile := <-leftChan
		rightOutFile := <-rightChan

		title := titleLeft[:len(titleLeft)-9]

		wg.Add(1)
		go func() {
			defer wg.Done()
			JoinViews(leftOutFile, rightOutFile, title)
		}()
		wg.Wait()

		RemoveFile(leftOutFile)
		RemoveFile(rightOutFile)

		wg.Wait()
	}
}

func GetMetadata(lectures Lectures) {
	config := GetConfig()
	resolution := getResolution(config.Quality)

	var wg sync.WaitGroup

	for _, lecture := range lectures {
		wg.Add(1)

		chunksCount := 0

		keyUrl := fmt.Sprintf("%s/fetchvideo/getVideoKey?ttid=%d&keyid=0", config.BaseUrl, lecture.Ttid)
		resp := GetClientAuthorized(keyUrl, config.Token)
		defer resp.Body.Close()

		keyUrlContent, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Could not get keyUrlContent %v", err)
			panic(err)
		}

		decryptionKey := getDecryptionKey(keyUrlContent)

		m3u8Data := getM3U8(lecture.Ttid)
		scanner := bufio.NewScanner(strings.NewReader(m3u8Data))

		m3u8FileRight, m3u8FilepathRight := createTempM3U8File(lecture.Ttid, "v1")
		defer m3u8FileRight.Close()

		m3u8FileLeft, m3u8FilepathLeft := createTempM3U8File(lecture.Ttid, "v3")
		defer m3u8FileLeft.Close()

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "#EXT-X-DISCONTINUITY") {
				writeM3U8FileConditionally(m3u8FileLeft, m3u8FileRight, "#EXT-X-ENDLIST\n", "#EXT-X-ENDLIST\n")
				break
			} else if strings.HasPrefix(line, "#EXT-X-KEY") {
				writeM3U8FileConditionally(m3u8FileLeft, m3u8FileRight, "#EXT-X-KEY:METHOD=NONE\n", "#EXT-X-KEY:METHOD=NONE\n")
				continue
			} else if strings.HasPrefix(line, "#") || line == "" {
				writeM3U8FileConditionally(m3u8FileLeft, m3u8FileRight, line+"\n", line+"\n")
				continue
			} else {
				writeM3U8FileConditionally(m3u8FileLeft, m3u8FileRight, fmt.Sprintf("%d_%04d_%s.ts\n", lecture.Ttid, chunksCount, "v3"), fmt.Sprintf("%d_%04d_%s.ts\n", lecture.Ttid, chunksCount, "v1"))
				chunksCount++
			}
		}

		var chunkWg sync.WaitGroup

		for i := 0; i < chunksCount; i++ {
			chunkWg.Add(1)

			go func(i int) {
				defer chunkWg.Done()
				downloadChunkConditonally(lecture.Ttid, resolution, i, decryptionKey)
			}(i)
		}

		chunkWg.Wait()
		fmt.Println("Entering ffmpeg")

		leftTitle := fmt.Sprintf("LEC %03d %s LEFT", lecture.SeqNo, lecture.Topic)
		rightTitle := fmt.Sprintf("LEC %03d %s RIGHT", lecture.SeqNo, lecture.Topic)

		go func() {
			defer wg.Done()
			joinChunksConditionally(m3u8FilepathLeft, m3u8FilepathRight, leftTitle, rightTitle)
		}()
	}
	wg.Wait()
}
