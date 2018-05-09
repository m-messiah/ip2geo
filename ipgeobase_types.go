package main

import "archive/zip"

// City ipgeobase main structure
type City struct {
	Name  string
	RegID int
	TZ    string
}

type IPGeobase struct {
	OutputDir  string
	ErrorsChan chan Error
	database   map[string]City
	archive    []*zip.File
}
