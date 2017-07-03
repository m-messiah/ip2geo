package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func maxmindGenerate(outputDir, lang string, ipver int, tzNames bool, include, exclude string, errors_chan chan Error) {
	answer, err := maxmindDownload()
	if err != nil {
		errors_chan <- Error{err, "MaxMind", "Download"}
		return
	} else {
		printMessage("MaxMind", "Download", "OK")
	}
	archive, err := maxmindUnpack(answer)
	if err != nil {
		errors_chan <- Error{err, "MaxMind", "Unpack"}
		return
	} else {
		printMessage("MaxMind", "Unpack", "OK")
	}
	cities, err := maxmindCities(archive, lang, tzNames, include, exclude)
	if err != nil {
		errors_chan <- Error{err, "MaxMind", "Generate Cities"}
		return
	} else {
		printMessage("MaxMind", "Generate cities", "OK")
	}
	database, err := maxmindNetwork(archive, ipver, cities)
	if err != nil {
		errors_chan <- Error{err, "MaxMind", "Generate db"}
		return
	} else {
		printMessage("MaxMind", "Generate db", "OK")
	}
	if err := maxmindWriteMap(outputDir, database); err != nil {
		errors_chan <- Error{err, "MaxMind", "Write nginx maps"}
		return
	} else {
		printMessage("MaxMind", "Write nginx maps", "OK")
	}
	errors_chan <- Error{err: nil}
}

func maxmindDownload() ([]byte, error) {
	resp, err := http.Get("http://geolite.maxmind.com/download/geoip/database/GeoLite2-City-CSV.zip")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func maxmindUnpack(response []byte) ([]*zip.File, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		return nil, err
	}
	return zipReader.File, nil
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

func maxmindCities(archive []*zip.File, language string, tznames bool, include, exclude string) (map[string]Location, error) {
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
		return nil, errors.New("Locations db is empty")
	}
	return locations, nil
}

func maxmindNetwork(archive []*zip.File, ipver int, locations map[string]Location) (Database, error) {
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
		return nil, errors.New("Network db is empty")
	}
	sort.Sort(database)
	return database, nil
}

func convertTZToOffset(t time.Time, tz string) string {
	location, err := time.LoadLocation(tz)
	if err != nil {
		return ""
	}
	_, offset := t.In(location).Zone()
	return fmt.Sprintf("UTC%+d", offset/3600)
}

func maxmindWriteMap(outputDir string, database Database) error {
	city, err := openMapFile(outputDir, "mm_city.txt")
	if err != nil {
		return err
	}
	tz, err := openMapFile(outputDir, "mm_tz.txt")
	if err != nil {
		return err
	}
	defer city.Close()
	defer tz.Close()
	for _, location := range database {
		fmt.Fprintf(city, "%s %s;\n", location.Network, base64.StdEncoding.EncodeToString([]byte(location.City)))
		fmt.Fprintf(tz, "%s %s;\n", location.Network, location.TZ)
	}
	return nil
}
