package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Username         string
	Password         string
	BaseUrl          string
	Quality          string
	Views            string
	DownloadLocation string
	Token            string
}

var config Config

func ParseConfig(configLocation string) *Config {
	f, err := os.ReadFile(configLocation)
	if err != nil {
		fmt.Println("Could not open config file")
		panic(err)
	}

	if config == (Config{}) {
		fmt.Println("I am here")
		err = json.Unmarshal(f, &config)
		if err != nil {
			fmt.Println("Could not parse the config please validate the json")
			panic(err)
		}
	}

	return &config
}
