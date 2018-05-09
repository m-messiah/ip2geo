package main

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MaxMind struct {
	database   Database
	archive    []*zip.File
	OutputDir  string
	ErrorsChan chan Error
	lang       string
	ipver      int
	tzNames    bool
	include    string
	exclude    string
}

func (maxmind *MaxMind) Name() string {
	return "MaxMind"
}

func (maxmind *MaxMind) AddError(err Error) {
	maxmind.ErrorsChan <- err
}

func (maxmind *MaxMind) Download() ([]byte, error) {
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

func (maxmind *MaxMind) Unpack(response []byte) error {
	file, err := Unpack(response)
	if err == nil {
		maxmind.archive = file
	}
	return err
}

func (maxmind *MaxMind) lineToItem(record []string, currentTime time.Time) (*string, *GeoItem, error, string) {
	if len(record) < 13 {
		return nil, nil, errors.New("too short line"), "FAIL"
	}
	country := record[4]
	if len(record[10]) < 1 || len(country) < 1 {
		return nil, nil, errors.New("too short country"), ""
	}
	if len(maxmind.include) > 1 && !strings.Contains(maxmind.include, country) {
		return nil, nil, errors.New("country skipped"), ""
	}
	if strings.Contains(maxmind.exclude, country) {
		return nil, nil, errors.New("country excluded"), ""
	}
	tz := record[12]
	if !maxmind.tzNames {
		tz = convertTZToOffset(currentTime, record[12])
	}
	return &record[0], &GeoItem{
		ID:   record[0],
		City: record[10],
		TZ:   tz,
	}, nil, ""
}

func (maxmind *MaxMind) Cities() (map[string]GeoItem, error) {
	locations := make(map[string]GeoItem)
	currentTime := time.Now()
	filename := "GeoLite2-City-Locations-" + maxmind.lang + ".csv"
	for record := range readCSVDatabase(maxmind.archive, filename, "MaxMind", ',', false) {
		key, location, err, severity := maxmind.lineToItem(record, currentTime)
		if err != nil {
			if len(severity) > 0 {
				printMessage("MaxMind", fmt.Sprintf(filename+" %v", err), severity)
			}
			continue
		}
		locations[*key] = *location
	}
	if len(locations) < 1 {
		return nil, errors.New("Locations db is empty")
	}
	return locations, nil
}

func (maxmind *MaxMind) Network(locations map[string]GeoItem) error {
	var database Database
	filename := "GeoLite2-City-Blocks-IPv" + strconv.Itoa(maxmind.ipver) + ".csv"
	for record := range readCSVDatabase(maxmind.archive, filename, "MaxMind", ',', false) {
		if len(record) < 2 {
			printMessage("MaxMind", fmt.Sprintf(filename+" too short line: %s", record), "FAIL")
			continue
		}
		ipRange := getIPRange(maxmind.ipver, record[0])
		netIP := net.ParseIP(strings.Split(ipRange, "-")[0])
		if netIP == nil {
			continue
		}
		geoID := record[1]
		if location, ok := locations[geoID]; ok {
			database = append(database, GeoItem{
				ID:      geoID,
				City:    location.City,
				Network: ipRange,
				TZ:      location.TZ,
				NetIP:   ip2Int(netIP),
			})
		}
	}
	if len(database) < 1 {
		return errors.New("Network db is empty")
	}
	sort.Sort(database)
	maxmind.database = database
	return nil
}

func (maxmind *MaxMind) WriteMap() error {
	city, err := openMapFile(maxmind.OutputDir, "mm_city.txt")
	if err != nil {
		return err
	}
	tz, err := openMapFile(maxmind.OutputDir, "mm_tz.txt")
	if err != nil {
		return err
	}
	defer city.Close()
	defer tz.Close()
	for _, location := range maxmind.database {
		fmt.Fprintf(city, "%s %s;\n", location.Network, base64.StdEncoding.EncodeToString([]byte(location.City)))
		fmt.Fprintf(tz, "%s %s;\n", location.Network, location.TZ)
	}
	return nil
}
