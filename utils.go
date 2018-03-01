package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatih/color"
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

// Convert uint to net.IP
func int2ip(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

// Convert net.IP to int64
func ip2int(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}
