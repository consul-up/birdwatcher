package main

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed birds.json
var birdData []byte

//go:embed canaries.json
var canaryData []byte

type birdRawData struct {
	Title     string `json:"title"`
	Thumbnail struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
	ExtractHTML string `json:"extract_html"`
}

// birds returns a list of birds parsed from birds.json.
func birds() []birdRawData {
	var birdList []birdRawData
	if err := json.Unmarshal(birdData, &birdList); err != nil {
		log.Fatalf("unable to parse birds.json: %s\n", err)
	}
	return birdList
}

// canaries returns a list of canaries parsed from canaries.json.
func canaries() []birdRawData {
	var canaryList []birdRawData
	if err := json.Unmarshal(canaryData, &canaryList); err != nil {
		log.Fatalf("unable to parse canaries.json: %s\n", err)
	}
	return canaryList
}
