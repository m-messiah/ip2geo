package main

// Database sortable by network
type Database []GeoItem

func (d Database) Less(i, j int) bool {
	return d[i].NetIP < d[j].NetIP
}

func (d Database) Len() int {
	return len(d)
}

func (d Database) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
