package main

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// IPGeobase - GeoBase compatible generator for ipgeobase.ru
type IPGeobase struct {
	OutputDir  string
	ErrorsChan chan Error
	archive    []*zip.File
}

func (ipgeobase *IPGeobase) name() string {
	return "IPGeobase"
}

func (ipgeobase *IPGeobase) addError(err Error) {
	ipgeobase.ErrorsChan <- err
}

func (ipgeobase *IPGeobase) download() ([]byte, error) {
	resp, err := http.Get("http://ipgeobase.ru/files/db/Main/geo_files.zip")
	if err != nil {
		printMessage("IPGeobase", "Download no answer", "FAIL")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	answer, err := io.ReadAll(resp.Body)
	if err != nil {
		printMessage("IPGeobase", "Download bad answer", "FAIL")
		return nil, err
	}
	return answer, nil
}

func (ipgeobase *IPGeobase) unpack(response []byte) error {
	file, err := Unpack(response)
	if err == nil {
		ipgeobase.archive = file
	}
	return err
}

func (ipgeobase *IPGeobase) citiesDB() (map[string]geoItem, error) {
	cities := make(map[string]geoItem)
	for record := range readCSVDatabase(ipgeobase.archive, "cities.txt", "IPGeobase", '\t', true) {
		if len(record) < 3 {
			printMessage("IPGeobase", fmt.Sprintf("cities.txt too short line: %s", record), "FAIL")
			continue
		}
		// Format is:  <city_id>\t<city_name>\t<region>\t<district>\t<lattitude>\t<longitude>
		cid, city, regionName := record[0], record[1], record[2]
		if region, ok := REGIONS[regionName]; ok {
			if cid == "1199" {
				region = REGIONS["Москва"]
			}
			cities[cid] = geoItem{
				City:  city,
				RegID: region.ID,
				TZ:    region.TZ,
			}
		}
	}
	if len(cities) < 1 {
		return nil, errors.New("cities db is empty")
	}
	return cities, nil
}

func (ipgeobase *IPGeobase) parseNetwork(cities map[string]geoItem) <-chan geoItem {
	database := make(chan geoItem)
	go func() {
		for record := range readCSVDatabase(ipgeobase.archive, "cidr_optim.txt", "IPGeobase", '\t', true) {
			if len(record) < 5 {
				printMessage("IPGeobase", fmt.Sprintf("cidr_optim.txt too short line: %s", record), "FAIL")
				continue
			}
			// Format is: <int_start>\t<int_end>\t<ip_range>\t<country_code>\tcity_id
			ipRange, country, cid := removeSpace(record[2]), record[3], record[4]
			if info, ok := cities[cid]; country == "RU" && ok {
				info.Network = ipRange
				database <- info
			}
		}
		close(database)
	}()
	return database
}

func (ipgeobase *IPGeobase) writeMap(cities map[string]geoItem) (retErr error) {
	reg, err := openMapFile(ipgeobase.OutputDir, "region.txt")
	if err != nil {
		return err
	}
	defer func() {
		if cerr := reg.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()
	city, err := openMapFile(ipgeobase.OutputDir, "city.txt")
	if err != nil {
		return err
	}
	defer func() {
		if cerr := city.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()
	tz, err := openMapFile(ipgeobase.OutputDir, "tz.txt")
	if err != nil {
		return err
	}
	defer func() {
		if cerr := tz.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	for info := range ipgeobase.parseNetwork(cities) {
		if _, err := fmt.Fprintf(city, "%s %s;\n", info.Network, base64.StdEncoding.EncodeToString([]byte(info.City))); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(reg, "%s %02d;\n", info.Network, info.RegID); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(tz, "%s %s;\n", info.Network, info.TZ); err != nil {
			return err
		}
	}
	return nil
}
