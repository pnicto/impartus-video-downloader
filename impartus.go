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
)

func LoginAndSetToken(config *Config) {
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
