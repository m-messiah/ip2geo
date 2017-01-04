package main

import (
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"path"
	"strings"
	"unicode"
)

func remove_space(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func open_map_file(output_dir, filename string) *os.File {
	filepath := path.Join(output_dir, filename)
	f, err := os.Create(filepath)
	if err != nil {
		color.Red("[FAIL]\nCan't open %s: %s", filepath, err.Error())
		return nil
	}
	return f
}

func print_message(module, message, status string) {
	var status_mesage string
	switch status {
	case "OK":
		status_mesage = color.GreenString(status)
	case "FAIL":
		status_mesage = color.RedString(status)
	case "WARN":
		status_mesage = color.YellowString(status)
	default:
		status_mesage = color.BlueString(status)
	}
	fmt.Printf("%-10s | %-30s[%s]\n", module, message, status_mesage)
}

func get_ip_range(ipver int, network string) string {
	if ipver == 4 {
		_, ipnet, err := net.ParseCIDR(network)
		if err != nil {
			return ""
		}
		ipb := make(net.IP, net.IPv4len)
		copy(ipb, ipnet.IP)
		for i, v := range ipb {
			ipb[i] = v | ^ipnet.Mask[i]
		}
		return fmt.Sprintf("%s-%s", ipnet.IP, ipb)
	}
	return network
}
