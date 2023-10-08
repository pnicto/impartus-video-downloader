package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const ConfigLocation = "./config.json"

type Config struct {
	Username         string
	Password         string
	BaseUrl          string
	Quality          string
	Views            string
	DownloadLocation string
	Token            string
	TempDirLocation  string
	Threads          int
}

var config Config

func parseConfig(configLocation string) *Config {
	var config Config

	f, err := os.ReadFile(configLocation)
	if err != nil {
		fmt.Println("Could not open config file")
		panic(err)
	}

	err = json.Unmarshal(f, &config)
	if err != nil {
		fmt.Println("Could not parse the config please validate the json")
		panic(err)
	}

	if config.Threads < 1 {
		config.Threads = 10
	}

	return &config
}

func GetConfig() *Config {
	if config == (Config{}) {
		config = *parseConfig(ConfigLocation)
	}

	return &config
}
