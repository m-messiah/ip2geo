package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func maxmindGenerate(wg *sync.WaitGroup, outputDir, lang string, ipver int, tzNames bool, include, exclude string) {
	answer := maxmindDownload()
	if answer != nil {
		printMessage("MaxMind", "Download", "OK")
	}
	archive := maxmindUnpack(answer)
	if archive != nil {
		printMessage("MaxMind", "Unpack", "OK")
	}
	cities := maxmindCities(archive, lang, tzNames, include, exclude)
	if len(cities) > 0 {
		printMessage("MaxMind", "Generate cities", "OK")
	}
	database := maxmindNetwork(archive, ipver, cities)
	if len(database) > 0 {
		printMessage("MaxMind", "Generate database", "OK")
	}
	maxmindWriteMap(outputDir, database)
	printMessage("MaxMind", "Write nginx maps", "OK")
	defer wg.Done()
}

func maxmindDownload() []byte {
	resp, err := http.Get("http://geolite.maxmind.com/download/geoip/database/GeoLite2-City-CSV.zip")
	if err != nil {
		printMessage("MaxMind", "Download no answer", "FAIL")
		return nil
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		printMessage("MaxMind", "Download bad answer", "FAIL")
		return nil
	}
	return answer
}

func maxmindUnpack(response []byte) []*zip.File {
	zipReader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		printMessage("MaxMind", "Bad zip file", "FAIL")
		return nil
	}
	return zipReader.File
}

func readMaxMindCSV(archive []*zip.File, filename string) chan []string {
	yield := make(chan []string)
	go func() {
		for _, file := range archive {
			if strings.Contains(file.Name, filename) {
				fp, err := file.Open()
				if err != nil {
					printMessage("MaxMind", fmt.Sprintf("Can't open %s", filename), "FAIL")
					yield <- nil
				}
				defer fp.Close()
				r := csv.NewReader(fp)
				r.LazyQuotes = true
				for {
					record, err := r.Read()
					// Stop at EOF.
					if err == io.EOF {
						break
					}
					if err != nil {
						printMessage("MaxMind", fmt.Sprintf("Can't read line from %s", filename), "WARN")
						continue
					}
					yield <- record
				}
			}
		}
		close(yield)
	}()
	return yield
}

func maxmindCities(archive []*zip.File, language string, tznames bool, include, exclude string) map[string]Location {
	locations := make(map[string]Location)
	currentTime := time.Now()
	for record := range readMaxMindCSV(archive, "GeoLite2-City-Locations-"+language+".csv") {
		if len(record) < 13 {
			printMessage("MaxMind", fmt.Sprintf("GeoLite2-City-Locations-"+language+".csv"+" too short line: %s", record), "FAIL")
			continue
		}
		country := record[4]
		if len(record[10]) < 1 || len(country) < 1 {
			continue
		}
		if len(include) < 1 || strings.Contains(include, country) {
			if !strings.Contains(exclude, country) {
				tz := record[12]
				if !tznames {
					tz = convertTZToOffset(currentTime, record[12])
				}
				locations[record[0]] = Location{
					ID:   record[0],
					City: record[10],
					TZ:   tz,
				}
			}
		}
	}
	if len(locations) < 1 {
		printMessage("MaxMind", "Locations db is empty", "FAIL")
	}
	return locations
}

func maxmindNetwork(archive []*zip.File, ipver int, locations map[string]Location) Database {
	var database Database
	for record := range readMaxMindCSV(archive, "GeoLite2-City-Blocks-IPv"+strconv.Itoa(ipver)+".csv") {
		if len(record) < 2 {
			printMessage("MaxMind", fmt.Sprintf("GeoLite2-City-Blocks-IPv"+strconv.Itoa(ipver)+".csv"+" too short line: %s", record), "FAIL")
			continue
		}
		ipRange := getIPRange(ipver, record[0])
		netIP := net.ParseIP(strings.Split(ipRange, "-")[0])
		if netIP == nil {
			continue
		}
		geoID := record[1]
		if location, ok := locations[geoID]; ok {
			database = append(database, Location{
				ID:      geoID,
				City:    location.City,
				Network: ipRange,
				TZ:      location.TZ,
				NetIP:   ip2Int(netIP),
			})
		}
	}
	if len(database) < 1 {
		printMessage("MaxMind", "Network db is empty", "FAIL")
	}
	sort.Sort(database)
	return database
}

func convertTZToOffset(t time.Time, tz string) string {
	location, err := time.LoadLocation(tz)
	if err != nil {
		return ""
	}
	_, offset := t.In(location).Zone()
	return fmt.Sprintf("UTC%+d", offset/3600)
}

func maxmindWriteMap(outputDir string, database Database) {
	city := openMapFile(outputDir, "mm_city.txt")
	tz := openMapFile(outputDir, "mm_tz.txt")
	defer city.Close()
	defer tz.Close()
	for _, location := range database {
		fmt.Fprintf(city, "%s %s;\n", location.Network, base64.StdEncoding.EncodeToString([]byte(location.City)))
		fmt.Fprintf(tz, "%s %s;\n", location.Network, location.TZ)
	}
}
