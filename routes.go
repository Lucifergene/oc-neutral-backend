package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Config struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	UpdatedAt   string `json:"updatedAt"`
	ConfigURL   string `json:"configURL"`
	User        string `json:"user"`
}

func downloadConfig(user string, configURL string, configName string) error {
	fmt.Println("Downloading config from: ", configURL)
	fmt.Print("\n")

	// Get the data
	resp, err := http.Get(configURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	path := "uploads/" + user
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	out, err := os.Create(path + "/" + configName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var config Config
	json.NewDecoder(r.Body).Decode(&config)
	fmt.Println(config)
	json.NewEncoder(w).Encode(config)

	err := downloadConfig(config.User, config.ConfigURL, config.Name)
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded: " + config.ConfigURL)

}
