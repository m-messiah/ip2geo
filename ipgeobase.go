package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
)

func (ipgeobase *IPGeobase) Generate() {
	answer, err := ipgeobase.Download()
	if err != nil {
		ipgeobase.ErrorsChan <- Error{err, "IPGeobase", "Download"}
		return
	}
	printMessage("IPGeobase", "Download", "OK")
	err = ipgeobase.Unpack(answer)
	if err != nil {
		ipgeobase.ErrorsChan <- Error{err, "IPGeobase", "Unpack"}
		return
	}
	printMessage("IPGeobase", "Unpack", "OK")
	cities, err := ipgeobase.Cities()
	if err != nil {
		ipgeobase.ErrorsChan <- Error{err, "IPGeobase", "Generate Cities"}
		return
	}
	printMessage("IPGeobase", "Generate cities", "OK")
	err = ipgeobase.Cidr(cities)
	if err != nil {
		ipgeobase.ErrorsChan <- Error{err, "IPGeobase", "Generate db"}
		return
	}
	printMessage("IPGeobase", "Generate database", "OK")
	if err := ipgeobase.WriteMap(); err != nil {
		ipgeobase.ErrorsChan <- Error{err, "IPGeobase", "Write map"}
		return
	}
	printMessage("IPGeobase", "Write nginx maps", "OK")
	ipgeobase.ErrorsChan <- Error{err: nil}
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
	zipReader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		return err
	}
	ipgeobase.archive = zipReader.File
	return nil
}

func (ipgeobase *IPGeobase) Cities() (map[string]City, error) {
	cities := make(map[string]City)
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
			cities[cid] = City{
				Name:  city,
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

func (ipgeobase *IPGeobase) Cidr(cities map[string]City) error {
	database := make(map[string]City)
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
		fmt.Fprintf(city, "%s %s;\n", ipRange, base64.StdEncoding.EncodeToString([]byte(info.Name)))
		fmt.Fprintf(reg, "%s %02d;\n", ipRange, info.RegID)
		fmt.Fprintf(tz, "%s %s;\n", ipRange, info.TZ)
	}
	return nil
}
