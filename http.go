package main

import (
	"fmt"
	"log"
	"net/http"
)

func GetClientAuthorized(url string, token string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create http request for GET %s", url)
		panic(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed with error %v for GET %s", err, url)
	}

	return response
}
