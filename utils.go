package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	"golang.org/x/net/html/charset"
)

func removeSpace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func getFileFromZip(archive []*zip.File, filename string) (*zip.File, error) {
	for _, file := range archive {
		if strings.Contains(file.Name, filename) {
			return file, nil
		}
	}
	return nil, errors.New("file not found")
}

func openMapFile(outputDir, filename string) (*os.File, error) {
	filepath := path.Join(outputDir, filename)
	return os.Create(filepath)
}

func printMessage(module, message, status string) {
	var statusMesage string
	switch status {
	case "OK":
		if logLevel > 0 {
			return
		}
		statusMesage = color.GreenString(status)
	case "WARN":
		if logLevel > 1 {
			return
		}
		statusMesage = color.YellowString(status)
	case "FAIL":
		statusMesage = color.RedString(status)
	default:
		if logLevel > 1 {
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
	if strings.Contains(network, ":") && strings.Contains(network, "/") {
		return network
	}
	return ""
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

func readCSVDatabase(archive []*zip.File, filename string, dbType string, comma rune, windowsEncoding bool) chan []string {
	yield := make(chan []string)
	go func() {
		defer close(yield)
		file, err := getFileFromZip(archive, filename)
		if err != nil {
			printMessage(dbType, fmt.Sprintf("%s %s", filename, err.Error()), "FAIL")
			return
		}
		fp, err := file.Open()
		if err != nil {
			printMessage(dbType, fmt.Sprintf("Can't open %s", filename), "FAIL")
			yield <- nil
		}
		defer fp.Close()
		var r *csv.Reader
		if windowsEncoding {
			utf8, err := charset.NewReader(fp, "text/csv; charset=windows-1251")
			if err != nil {
				printMessage(dbType, fmt.Sprintf("%s not in cp1251", filename), "FAIL")
				yield <- nil
			}
			r = csv.NewReader(utf8)
		} else {
			r = csv.NewReader(fp)
		}
		r.Comma, r.LazyQuotes = comma, true
		for {
			record, err := r.Read()
			// Stop at EOF.
			if err == io.EOF {
				break
			}
			if err != nil {
				printMessage(dbType, fmt.Sprintf("Can't read line from %s", filename), "WARN")
				continue
			}
			yield <- record
		}

	}()
	return yield
}

func convertTZToOffset(t time.Time, tz string) string {
	location, err := time.LoadLocation(tz)
	if err != nil {
		return ""
	}
	_, offset := t.In(location).Zone()
	return fmt.Sprintf("UTC%+d", offset/3600)
}

// Unpack zip file and return slice of zip.File`s
func Unpack(response []byte) ([]*zip.File, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	var file []*zip.File
	if err == nil {
		file = zipReader.File
	}
	return file, err
}
