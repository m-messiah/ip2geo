package main

import (
	"bytes"
	"net"
)

type Location struct {
	ID      string
	City    string
	Network string
	TZ      string
}

// Sortable by network
type Database []Location

func (d Database) Less(i, j int) bool {
	_, ipnet_i, _ := net.ParseCIDR(d[i].Network)
	_, ipnet_j, _ := net.ParseCIDR(d[j].Network)
	return bytes.Compare(ipnet_i.IP, ipnet_j.IP) < 0
}

func (d Database) Len() int {
	return len(d)
}

func (d Database) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
