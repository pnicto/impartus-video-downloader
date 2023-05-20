package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	var loginResponse LoginResponse
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	if err != nil {
		log.Fatalf("Could not decode login body %v", err)
	}

	config.Token = loginResponse.Token
}

func GetCourses() Courses {
	var courses Courses
	config := GetConfig()

	url := fmt.Sprintf("%s/subjects", config.BaseUrl)
	resp := GetClientAuthorized(url, config.Token)
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&courses)
	if err != nil {
		log.Fatalf("Could not decode response %v", err)
	}

	return courses
}

func GetLectures(course Course) Lectures {
	var lectures Lectures
	config := GetConfig()

	url := fmt.Sprintf("%s/subjects/%d/lectures/%d", config.BaseUrl, course.SubjectID, course.SessionID)
	resp := GetClientAuthorized(url, config.Token)
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&lectures)
	if err != nil {
		log.Fatalf("Could not decode response %v", err)
	}

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

func createTempM3U8File(ttid int) (*os.File, string) {
	config := GetConfig()
	fmt.Println(config)

	err := os.MkdirAll(config.TempDirLocation, 0755)
	if err != nil {
		fmt.Printf("Could not create temp directory %s with err %v\n", config.TempDirLocation, err)
		panic(err)
	}

	tempM3U8File := filepath.Join(config.TempDirLocation, fmt.Sprintf("%d.m3u8", ttid))

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

	outFilepath := filepath.Join(config.TempDirLocation, fmt.Sprintf("%04d_%s.ts.temp", chunk, view))
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

	err = ioutil.WriteFile(outPath, plainText, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func GetMedata(lectures Lectures) map[string]string {
	m3u8Filepaths := make(map[string]string)

	chunksCount := 0
	config := GetConfig()

	for _, lecture := range lectures {
		keyUrl := fmt.Sprintf("%s/fetchvideo/getVideoKey?ttid=%d&keyid=0", config.BaseUrl, lecture.Ttid)
		view := config.Views
		resolution := getResolution(config.Quality)

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

		m3u8File, m3u8Filepath := createTempM3U8File(lecture.Ttid)
		defer m3u8File.Close()
		m3u8Filepaths[m3u8Filepath] = fmt.Sprintf("LEC %03d %s", lecture.SeqNo, lecture.Topic)

		// TODO: Handle for different views
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "#EXT-X-DISCONTINUITY") {
				m3u8File.WriteString("#EXT-X-ENDLIST\n")
				break
			} else if strings.HasPrefix(line, "#EXT-X-KEY") {
				m3u8File.WriteString("#EXT-X-KEY:METHOD=NONE")
				continue
			} else if strings.HasPrefix(line, "#") || line == "" {
				m3u8File.WriteString(line + "\n")
				continue
			} else {
				m3u8File.WriteString(fmt.Sprintf("%04d_v1.ts\n", chunksCount))
				chunksCount++
			}
		}

		for i := 0; i < chunksCount; i++ {
			switch view {
			case "both":
				// TODO: Need to think of a way to handle this
				// downloadChunk(lecture.Ttid, resolution, "v1", i)
				// downloadChunk(lecture.Ttid, resolution, "v3", i)
			case "right":
				chunkPath := downloadChunk(lecture.Ttid, resolution, "v1", i)
				decryptChunk(chunkPath, decryptionKey)
			case "left":
				downloadChunk(lecture.Ttid, resolution, "v3", i)
			}
		}
	}
	return m3u8Filepaths
}
