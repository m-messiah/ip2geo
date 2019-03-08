package main

import "strings"

// Database sortable by network
type Database []GeoItem

func (d Database) Less(i, j int) bool {
	return strings.Compare(d[i].Network, d[j].Network) < 0
}

func (d Database) Len() int {
	return len(d)
}

func (d Database) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
