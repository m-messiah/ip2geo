package main

import (
	"encoding/binary"
	"fmt"
	"github.com/fatih/color"
	"net"
	"os"
	"path"
	"strings"
	"unicode"
)

func removeSpace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func openMapFile(outputDir, filename string) (*os.File, error) {
	filepath := path.Join(outputDir, filename)
	return os.Create(filepath)
}

func printMessage(module, message, status string) {
	var statusMesage string
	switch status {
	case "OK":
		if LogLevel > 0 {
			return
		}
		statusMesage = color.GreenString(status)
	case "WARN":
		if LogLevel > 1 {
			return
		}
		statusMesage = color.YellowString(status)
	case "FAIL":
		statusMesage = color.RedString(status)
	default:
		if LogLevel > 1 {
			return
		}
		statusMesage = color.BlueString(status)
	}
	fmt.Printf("%-10s | %-60s [%s]\n", module, message, statusMesage)
}

func getIPRange(ipver int, network string) string {
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

func ip2Int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}
