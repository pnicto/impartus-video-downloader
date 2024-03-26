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
	defer outFile.Close()
	if err != nil {
		fmt.Printf("Could not download chunk %d %v", chunk, err)
	}

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

func DownloadPlaylist(playlists []ParsedPlaylist) []DownloadedPlaylist {
	config := GetConfig()
	var downloaded []DownloadedPlaylist

	err := os.MkdirAll(config.TempDirLocation, 0755)
	if err != nil {
		log.Fatalln("Could not create temp dir")
	}

	for _, playlist := range playlists {
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
					fmt.Println()
					fmt.Println("Chunk", i, "download failed")
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
					fmt.Println()
					fmt.Println("Chunk", i, "download failed")
					continue
				}
				chunkPath := decryptChunk(f, decryptionKey)
				downloadedPlaylist.SecondViewChunks = append(downloadedPlaylist.SecondViewChunks, chunkPath)
				bar.Add(1)
			}
		}

		downloadedPlaylist.Playlist = playlist
		downloaded = append(downloaded, downloadedPlaylist)
	}
	return downloaded
}

type M3U8File struct {
	FirstViewFile  string
	SecondViewFile string
	Playlist       ParsedPlaylist
}

func CreateTempM3U8Files(downloadedPlaylists []DownloadedPlaylist) []M3U8File {
	var m3u8Files []M3U8File

	config := GetConfig()
	for _, playlist := range downloadedPlaylists {
		var m3u8File M3U8File

		if len(playlist.FirstViewChunks) > 0 {
			firstView, err := os.Create(fmt.Sprintf("%s/%d_first.m3u8", config.TempDirLocation, playlist.Playlist.Id))
			defer firstView.Close()

			if err != nil {
				fmt.Printf("Could not create temp m3u8 file for ttid %d with error %v", playlist.Playlist.Id, err)
			}

			firstView.WriteString(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=NONE`)

			for _, chunk := range playlist.FirstViewChunks {
				firstView.WriteString("#EXTINF:1\n")
				firstView.WriteString("../" + chunk + "\n")
			}

			firstView.WriteString("#EXT-X-ENDLIST")

			m3u8File.FirstViewFile = firstView.Name()
		}

		if len(playlist.SecondViewChunks) > 0 {
			secondView, err := os.Create(fmt.Sprintf("%s/%d_second.m3u8", config.TempDirLocation, playlist.Playlist.Id))
			if err != nil {
				fmt.Printf("Could not create temp m3u8 file for ttid %d with error %v", playlist.Playlist.Id, err)
			}
			defer secondView.Close()

			secondView.WriteString(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:11
#EXT-X-KEY:METHOD=NONE`)

			for _, chunk := range playlist.SecondViewChunks {
				secondView.WriteString("#EXTINF:1\n")
				secondView.WriteString("../" + chunk + "\n")
			}

			secondView.WriteString("#EXT-X-ENDLIST")

			m3u8File.SecondViewFile = secondView.Name()
		}

		m3u8File.Playlist = playlist.Playlist
		m3u8Files = append(m3u8Files, m3u8File)
	}
	return m3u8Files
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
