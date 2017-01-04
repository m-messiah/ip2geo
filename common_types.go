package main

import (
	"bytes"
	"net"
	"strings"
)

// Sortable by network
type IPList []string

func (d IPList) Less(i, j int) bool {
	ip_i := net.ParseIP(strings.Split(d[i], "-")[0])
	ip_j := net.ParseIP(strings.Split(d[j], "-")[0])
	return bytes.Compare(ip_i, ip_j) < 0
}

func (d IPList) Len() int {
	return len(d)
}

func (d IPList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
