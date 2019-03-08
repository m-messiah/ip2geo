package main

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
)

type IPGeobase struct {
	OutputDir  string
	ErrorsChan chan Error
	database   map[string]GeoItem
	archive    []*zip.File
}

func (ipgeobase *IPGeobase) Name() string {
	return "IPGeobase"
}

func (ipgeobase *IPGeobase) AddError(err Error) {
	ipgeobase.ErrorsChan <- err
}

func (ipgeobase *IPGeobase) Download() ([]byte, error) {
	resp, err := http.Get("http://ipgeobase.ru/files/db/Main/geo_files.zip")
	if err != nil {
		printMessage("IPGeobase", "Download no answer", "FAIL")
		return nil, err
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		printMessage("IPGeobase", "Download bad answer", "FAIL")
		return nil, err
	}
	return answer, nil
}

func (ipgeobase *IPGeobase) Unpack(response []byte) error {
	file, err := Unpack(response)
	if err == nil {
		ipgeobase.archive = file
	}
	return err
}

func (ipgeobase *IPGeobase) Cities() (map[string]GeoItem, error) {
	cities := make(map[string]GeoItem)
	for record := range readCSVDatabase(ipgeobase.archive, "cities.txt", "IPGeobase", '\t', true) {
		if len(record) < 3 {
			printMessage("IPGeobase", fmt.Sprintf("cities.txt too short line: %s", record), "FAIL")
			continue
		}
		// Format is:  <city_id>\t<city_name>\t<region>\t<district>\t<lattitude>\t<longitude>
		cid, city, regionName := record[0], record[1], record[2]
		if region, ok := REGIONS[regionName]; ok {
			if cid == "1199" {
				region, _ = REGIONS["Москва"]
			}
			cities[cid] = GeoItem{
				City:  city,
				RegID: region.ID,
				TZ:    region.TZ,
			}
		}
	}
	if len(cities) < 1 {
		return nil, errors.New("Cities db is empty")
	}
	return cities, nil
}

func (ipgeobase *IPGeobase) Network(cities map[string]GeoItem) error {
	database := make(map[string]GeoItem)
	for record := range readCSVDatabase(ipgeobase.archive, "cidr_optim.txt", "IPGeobase", '\t', true) {
		if len(record) < 5 {
			printMessage("IPGeobase", fmt.Sprintf("cidr_optim.txt too short line: %s", record), "FAIL")
			continue
		}
		// Format is: <int_start>\t<int_end>\t<ip_range>\t<country_code>\tcity_id
		ipRange, country, cid := removeSpace(record[2]), record[3], record[4]
		if city, ok := cities[cid]; country == "RU" && ok {
			database[ipRange] = city
		}
	}
	if len(database) < 1 {
		return errors.New("Database is empty")
	}
	ipgeobase.database = database
	return nil
}

func (ipgeobase *IPGeobase) WriteMap() error {
	reg, err := openMapFile(ipgeobase.OutputDir, "region.txt")
	if err != nil {
		return err
	}
	city, err := openMapFile(ipgeobase.OutputDir, "city.txt")
	if err != nil {
		return err
	}
	tz, err := openMapFile(ipgeobase.OutputDir, "tz.txt")
	if err != nil {
		return err
	}
	defer reg.Close()
	defer city.Close()
	defer tz.Close()
	ipRanges := make(IPList, len(ipgeobase.database))
	i := 0
	for ipRange := range ipgeobase.database {
		ipRanges[i] = ipRange
		i++
	}
	sort.Sort(ipRanges)
	for _, ipRange := range ipRanges {
		info := ipgeobase.database[ipRange]
		fmt.Fprintf(city, "%s %s;\n", ipRange, base64.StdEncoding.EncodeToString([]byte(info.City)))
		fmt.Fprintf(reg, "%s %02d;\n", ipRange, info.RegID)
		fmt.Fprintf(tz, "%s %s;\n", ipRange, info.TZ)
	}
	return nil
}
