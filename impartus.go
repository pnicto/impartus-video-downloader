package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
