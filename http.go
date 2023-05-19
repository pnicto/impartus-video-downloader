package main

import (
	"fmt"
	"net/http"
)

var client *http.Client

func GetClientAuthorized(url string, token string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Failed to create http request for GET %s", url)
		panic(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if client != (&http.Client{}) {
		client = &http.Client{}
	}

	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed with error %v for GET %s", err, url)
	}

	return response
}
