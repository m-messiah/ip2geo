package main

import "archive/zip"

// Location MaxMind main structure
type Location struct {
	ID      string
	City    string
	Network string
	TZ      string
	NetIP   uint32
}

// Database sortable by network
type Database []Location

func (d Database) Less(i, j int) bool {
	return d[i].NetIP < d[j].NetIP
}

func (d Database) Len() int {
	return len(d)
}

func (d Database) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

type MaxMind struct {
	database   Database
	archive    []*zip.File
	OutputDir  string
	ErrorsChan chan Error
	lang       string
	ipver      int
	tzNames    bool
	include    string
	exclude    string
}
