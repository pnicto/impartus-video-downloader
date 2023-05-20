package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const CONFIG_LOCATION = "./config.json"

type Config struct {
	Username         string
	Password         string
	BaseUrl          string
	Quality          string
	Views            string
	DownloadLocation string
	Token            string
	TempDirLocation  string
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

	return &config
}

func GetConfig() *Config {
	if config == (Config{}) {
		config = *parseConfig(CONFIG_LOCATION)
	}

	return &config
}
