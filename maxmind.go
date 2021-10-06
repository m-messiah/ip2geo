package main

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// MaxMind - GeoBase compatible generator for geolite.maxmind.com
type MaxMind struct {
	archive    []*zip.File
	OutputDir  string
	ErrorsChan chan Error
	maxMindConfig
}

func (maxmind *MaxMind) name() string {
	return "MaxMind"
}

func (maxmind *MaxMind) addError(err Error) {
	maxmind.ErrorsChan <- err
}

func (maxmind *MaxMind) download() ([]byte, error) {
	// If used filename, no download
	if len(maxmind.Filename) > 0 {
		return ioutil.ReadFile(maxmind.Filename)
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City-CSV&suffix=zip", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("license_key", maxmind.LicenseKey)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
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

func (maxmind *MaxMind) unpack(response []byte) error {
	file, err := Unpack(response)
	if err == nil {
		maxmind.archive = file
	}
	return err
}

func (maxmind *MaxMind) lineToItem(record []string, currentTime time.Time) (*string, *geoItem, string, error) {
	if len(record) < 13 {
		return nil, nil, "FAIL", errors.New("too short line")
	}
	countryCode := record[4]
	if len(countryCode) < 1 || len(record[5]) < 1 {
		return nil, nil, "", errors.New("too short country")
	}
	if len(maxmind.Include) > 1 && !strings.Contains(maxmind.Include, countryCode) {
		return nil, nil, "", errors.New("country skipped")
	}
	if strings.Contains(maxmind.Exclude, countryCode) {
		return nil, nil, "", errors.New("country Excluded")
	}
	tz := record[12]
	if !maxmind.TZNames {
		tz = convertTZToOffset(currentTime, record[12])
	}
	if len(record[10]) < 1 {
		return nil, nil, "", errors.New("too short city name")
	}
	return &record[0], &geoItem{
		ID:          record[0],
		City:        record[10],
		TZ:          tz,
		CountryCode: record[4],
		Country:     record[5],
	}, "", nil
}

func (maxmind *MaxMind) citiesDB() (map[string]geoItem, error) {
	locations := make(map[string]geoItem)
	currentTime := time.Now()
	filename := "GeoLite2-City-Locations-" + maxmind.Lang + ".csv"
	for record := range readCSVDatabase(maxmind.archive, filename, "MaxMind", ',', false) {
		key, location, severity, err := maxmind.lineToItem(record, currentTime)
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

func (maxmind *MaxMind) parseNetwork(locations map[string]geoItem) <-chan geoItem {
	database := make(chan geoItem)
	go func() {
		var ipRange string
		var geoID string
		filename := "GeoLite2-City-Blocks-IPv" + strconv.Itoa(maxmind.IPVer) + ".csv"
		for record := range readCSVDatabase(maxmind.archive, filename, "MaxMind", ',', false) {
			if len(record) < 2 {
				printMessage("MaxMind", fmt.Sprintf(filename+" too short line: %s", record), "FAIL")
				continue
			}
			ipRange = getIPRange(maxmind.IPVer, record[0])
			if ipRange == "" {
				continue
			}
			geoID = record[1]
			if location, ok := locations[geoID]; ok {
				location.Network = ipRange
				database <- location
			}
		}
		close(database)
	}()
	return database
}

func (maxmind *MaxMind) writeMap(locations map[string]geoItem) error {
	city, err := openMapFile(maxmind.OutputDir, "mm_city.txt")
	if err != nil {
		return err
	}
	tz, err := openMapFile(maxmind.OutputDir, "mm_tz.txt")
	if err != nil {
		return err
	}
	var country *os.File
	var countryCode *os.File
	if !maxmind.NoCountry {
		country, err = openMapFile(maxmind.OutputDir, "mm_country.txt")
		if err != nil {
			return err
		}
		countryCode, err = openMapFile(maxmind.OutputDir, "mm_country_code.txt")
		if err != nil {
			return err
		}
		defer country.Close()
		defer countryCode.Close()
	}
	defer city.Close()
	defer tz.Close()

	for location := range maxmind.parseNetwork(locations) {
		var cityName string
		var countryName string
		if maxmind.NoBase64 {
			cityName = "\"" + strings.Replace(location.City, "\"", "\\\"", -1) + "\""
			countryName = "\"" + strings.Replace(location.Country, "\"", "\\\"", -1) + "\""
		} else {
			cityName = base64.StdEncoding.EncodeToString([]byte(location.City))
			countryName = base64.StdEncoding.EncodeToString([]byte(location.Country))
		}

		fmt.Fprintf(city, "%s %s;\n", location.Network, cityName)
		fmt.Fprintf(tz, "%s %s;\n", location.Network, location.TZ)
		if !maxmind.NoCountry {
			fmt.Fprintf(country, "%s %s;\n", location.Network, countryName)
			fmt.Fprintf(countryCode, "%s %s;\n", location.Network, location.CountryCode)
		}
	}
	return nil
}
