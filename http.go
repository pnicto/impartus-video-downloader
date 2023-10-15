package main

import (
	"fmt"
	"log"
	"net/http"
)

func GetClientAuthorized(url string, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create http request for GET %s", url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		return response, fmt.Errorf("Request failed with error %v for GET %s\n", err, url)
	}

	return response, nil
}
