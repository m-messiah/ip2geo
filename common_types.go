package main

import (
	"bytes"
	"net"
	"strings"
)

// IPList sortable by network
type IPList []string

func (d IPList) Less(i, j int) bool {
	ipI := net.ParseIP(strings.Split(d[i], "-")[0])
	ipJ := net.ParseIP(strings.Split(d[j], "-")[0])
	return bytes.Compare(ipI, ipJ) < 0
}

func (d IPList) Len() int {
	return len(d)
}

func (d IPList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Error of GeoBase handle
type Error struct {
	err    error
	Module string
	Action string
}
