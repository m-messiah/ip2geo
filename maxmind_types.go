package main

type Location struct {
	ID      string
	City    string
	Network string
	TZ      string
	NetIP   uint32
}

// Sortable by network
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
