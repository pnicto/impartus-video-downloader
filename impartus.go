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
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
)

type (
	LoginResponse struct {
		Message  string `json:"message"`
		Token    string `json:"token"`
		UserType int    `json:"userType"`
		Success  bool   `json:"success"`
	}

	Courses []Course
	Course  struct {
		Institute            string `json:"institute"`
		SubjectName          string `json:"subjectName"`
		SessionName          string `json:"sessionName"`
		ProfessorName        string `json:"professorName"`
		Department           string `json:"department"`
		Coverpic             string `json:"coverpic"`
		SessionID            int    `json:"sessionId"`
		ProfessorID          int    `json:"professorId"`
		DepartmentID         int    `json:"departmentId"`
		InstituteID          int    `json:"instituteId"`
		SubjectID            int    `json:"subjectId"`
		VideoCount           int    `json:"videoCount"`
		FlippedLecturesCount int    `json:"flippedLecturesCount"`
	}

	Lectures []Lecture
	Lecture  struct {
		SubjectDescription  any    `json:"subjectDescription"`
		SessionName         string `json:"sessionName"`
		ClassroomName       string `json:"classroomName"`
		FilePath2           string `json:"filePath2"`
		FilePath            string `json:"filePath"`
		EndTime             string `json:"endTime"`
		Topic               string `json:"topic"`
		StartTime           string `json:"startTime"`
		Coverpic            string `json:"coverpic"`
		SubjectCode         string `json:"subjectCode"`
		ProfessorImageURL   string `json:"professorImageUrl"`
		ProfessorName       string `json:"professorName"`
		Institute           string `json:"institute"`
		SubjectName         string `json:"subjectName"`
		Department          string `json:"department"`
		VideoID             int    `json:"videoId"`
		TapNToggle          int    `json:"tapNToggle"`
		Trending            int    `json:"trending"`
		SeqNo               int    `json:"seqNo"`
		DepartmentID        int    `json:"departmentId"`
		ProfessorID         int    `json:"professorId"`
		InstituteID         int    `json:"instituteId"`
		Ttid                int    `json:"ttid"`
		Selfenroll          int    `json:"selfenroll"`
		SubjectID           int    `json:"subjectId"`
		ActualDuration      int    `json:"actualDuration"`
		ClassroomID         int    `json:"classroomId"`
		Type                int    `json:"type"`
		Status              int    `json:"status"`
		SlideCount          int    `json:"slideCount"`
		Noaudio             int    `json:"noaudio"`
		Views               int    `json:"views"`
		DocumentCount       int    `json:"documentCount"`
		LessonPlanAvailable int    `json:"lessonPlanAvailable"`
		SessionID           int    `json:"sessionId"`
		LastPosition        int    `json:"lastPosition"`
	}

	StreamInfo struct {
		Quality string
		URL     string
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

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Could not create request %v", err)
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Referer", "https://bitshyd.impartus.com/login/")
	req.Header.Add("User-Agent", GetRandomUserAgent())

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Login failed with error %v", err)
	}
	defer response.Body.Close()

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
	resp, _ := GetClientAuthorized(url, config.Token)
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
	resp, _ := GetClientAuthorized(url, config.Token)
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

// TODO: Refine decryptChunk
func decryptChunk(filePath string, key []byte) string {
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
	return outPath
}

func getStreamInfos(lecture Lecture) []StreamInfo {
	config := GetConfig()
	var streamInfos []StreamInfo
	uri := fmt.Sprintf("%s/fetchvideo?ttid=%d&token=%s&type=index.m3u8", config.BaseUrl, lecture.Ttid, config.Token)

	resp, err := GetClientAuthorized(uri, config.Token)
	if err != nil {
		log.Println("Could not get stream infos", err)
		return []StreamInfo{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing response body")
		}
	}(resp.Body)

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading in response body")
	}

	lines := strings.Split(string(res), "\n")

	pattern := `\d*x\d*`
	re := regexp.MustCompile(pattern)

	for _, line := range lines {
		if strings.HasPrefix(line, "http") || strings.HasPrefix(line, "https") {
			match := re.FindStringSubmatch(line)
			if len(match) > 0 {
				resolution := strings.Split(match[0], "x")
				streamInfos = append(streamInfos, StreamInfo{Quality: resolution[1], URL: line})
			}
		}
	}

	return streamInfos
}

func getStreamUrl(streamInfos []StreamInfo) string {
	config := GetConfig()
	var streamUrl string
	for _, streamInfo := range streamInfos {
		if (streamInfo.Quality == "450" || streamInfo.Quality == "480") && strings.HasPrefix(config.Quality, "4") {
			streamUrl = streamInfo.URL
			break
		} else if streamInfo.Quality == config.Quality {
			streamUrl = streamInfo.URL
			break
		}
	}
	return streamUrl
}

func GetPlaylist(lectures []Lecture) []ParsedPlaylist {
	var parsedPlaylists []ParsedPlaylist

	for _, lecture := range lectures {
		streamInfos := getStreamInfos(lecture)
		streamUrl := getStreamUrl(streamInfos)
		resp, err := GetClientAuthorized(streamUrl, GetConfig().Token)
		if err != nil {
			fmt.Println("Could not get stream url", err)
			continue
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		parsedPlaylists = append(parsedPlaylists, PlaylistParser(scanner, lecture.Ttid, lecture.Topic, lecture.SeqNo))
	}

	return parsedPlaylists
}

func downloadUrl(url string, id int, chunk int, view string) (string, error) {
	resp, err := GetClientAuthorized(url, config.Token)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Could not close response body %v", err)
		}
	}(resp.Body)

	outFilepath := filepath.Join(config.TempDirLocation, fmt.Sprintf("%d_%s_%04d.ts.temp", id, view, chunk))
	outFile, err := os.Create(outFilepath)
	if err != nil {
		fmt.Printf("Could not download chunk %d %v", chunk, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Printf("Could not write chunk %d %v", chunk, err)
	}
	outFile.Sync()

	return outFilepath, nil
}

type DownloadedPlaylist struct {
	FirstViewChunks  []string
	SecondViewChunks []string
	Playlist         ParsedPlaylist
}

func DownloadPlaylist(playlist ParsedPlaylist) DownloadedPlaylist {
	config := GetConfig()
	var downloadedPlaylist DownloadedPlaylist

	resp, _ := GetClientAuthorized(playlist.KeyURL, config.Token)
	defer resp.Body.Close()
	keyUrlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Could not get keyUrlContent %v", err)
		panic(err)
	}

	decryptionKey := getDecryptionKey(keyUrlContent)

	if len(playlist.FirstViewURLs) > 0 && config.Views != "left" {
		bar := progressbar.NewOptions64(-1, progressbar.OptionSetDescription(fmt.Sprintf("Lec %03d downloading right view chunks", playlist.SeqNo)))
		for i, url := range playlist.FirstViewURLs {
			f, err := downloadUrl(url, playlist.Id, i, "first")
			if err != nil {
				// fmt.Println()
				// fmt.Println("[WARNING] Chunk", i, "download failed")
				continue
			}
			chunkPath := decryptChunk(f, decryptionKey)
			downloadedPlaylist.FirstViewChunks = append(downloadedPlaylist.FirstViewChunks, chunkPath)
			bar.Add(1)
		}
	}

	if len(playlist.SecondViewURLs) > 0 && config.Views != "right" {
		bar := progressbar.NewOptions64(-1, progressbar.OptionSetDescription(fmt.Sprintf("Lec %03d downloading left view chunks", playlist.SeqNo)))
		for i, url := range playlist.SecondViewURLs {
			f, err := downloadUrl(url, playlist.Id, i, "second")
			if err != nil {
				// fmt.Println()
				// fmt.Println("[WARNING] Chunk", i, "download failed")
				continue
			}
			chunkPath := decryptChunk(f, decryptionKey)
			downloadedPlaylist.SecondViewChunks = append(downloadedPlaylist.SecondViewChunks, chunkPath)
			bar.Add(1)
		}
	}

	downloadedPlaylist.Playlist = playlist
	return downloadedPlaylist
}

type M3U8File struct {
	FirstViewFile  string
	SecondViewFile string
	Playlist       ParsedPlaylist
}

func CreateTempM3U8File(downloadedPlaylist DownloadedPlaylist) M3U8File {
	config := GetConfig()
	var m3u8File M3U8File

	if len(downloadedPlaylist.FirstViewChunks) > 0 {
		firstView, err := os.Create(fmt.Sprintf("%s/%d_first.m3u8", config.TempDirLocation, downloadedPlaylist.Playlist.Id))
		if err != nil {
			fmt.Printf("Could not create temp m3u8 file for ttid %d with error %v", downloadedPlaylist.Playlist.Id, err)
		}
		defer firstView.Close()

		firstView.WriteString(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=NONE`)

		for _, chunk := range downloadedPlaylist.FirstViewChunks {
			firstView.WriteString("#EXTINF:1\n")
			firstView.WriteString("../" + chunk + "\n")
		}

		firstView.WriteString("#EXT-X-ENDLIST")

		m3u8File.FirstViewFile = firstView.Name()
	}

	if len(downloadedPlaylist.SecondViewChunks) > 0 {
		secondView, err := os.Create(fmt.Sprintf("%s/%d_second.m3u8", config.TempDirLocation, downloadedPlaylist.Playlist.Id))
		if err != nil {
			fmt.Printf("Could not create temp m3u8 file for ttid %d with error %v", downloadedPlaylist.Playlist.Id, err)
		}
		defer secondView.Close()

		secondView.WriteString(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=NONE`)

		for _, chunk := range downloadedPlaylist.SecondViewChunks {
			secondView.WriteString("#EXTINF:1\n")
			secondView.WriteString("../" + chunk + "\n")
		}

		secondView.WriteString("#EXT-X-ENDLIST")

		m3u8File.SecondViewFile = secondView.Name()
	}

	m3u8File.Playlist = downloadedPlaylist.Playlist

	return m3u8File
}

func DownloadLectureSlides(lecture Lecture) {
	config := GetConfig()
	path := fmt.Sprintf("./slides/%s/L%03d %s.pdf", lecture.SubjectName, lecture.SeqNo, lecture.Topic)
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("Could not create file", path, "with error", err)
	}
	url := fmt.Sprintf("%s/videos/%d/auto-generated-pdf", config.BaseUrl, lecture.VideoID)
	resp, _ := GetClientAuthorized(url, config.Token)
	defer resp.Body.Close()
	num, _ := io.Copy(f, resp.Body)
	fmt.Printf("Downloaded %d bytes\n", num)
}
