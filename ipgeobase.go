package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
)

func ipgeobaseGenerate(outputDir string, errors_chan chan Error) {
	answer, err := ipgeobaseDownload()
	if err != nil {
		errors_chan <- Error{err, "IPGeobase", "Download"}
		return
	} else {
		printMessage("IPGeobase", "Download", "OK")
	}
	archive, err := ipgeobaseUnpack(answer)
	if err != nil {
		errors_chan <- Error{err, "IPGeobase", "Unpack"}
		return
	} else {
		printMessage("IPGeobase", "Unpack", "OK")
	}
	cities, err := ipgeobaseCities(archive)
	if err != nil {
		errors_chan <- Error{err, "IPGeobase", "Generate Cities"}
		return
	} else {
		printMessage("IPGeobase", "Generate cities", "OK")
	}
	database, err := ipgeobaseCidr(archive, cities)
	if err != nil {
		errors_chan <- Error{err, "IPGeobase", "Generate db"}
		return
	} else {
		printMessage("IPGeobase", "Generate database", "OK")
	}
	if err := ipgeobaseWriteMap(outputDir, database); err != nil {
		errors_chan <- Error{err, "IPGeobase", "Write map"}
		return
	} else {
		printMessage("IPGeobase", "Write nginx maps", "OK")
	}
	errors_chan <- Error{err: nil}
}

func ipgeobaseDownload() ([]byte, error) {
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

func ipgeobaseUnpack(response []byte) ([]*zip.File, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		return nil, err
	}
	return zipReader.File, nil
}

func readIgCSV(archive []*zip.File, filename string) chan []string {
	yield := make(chan []string)
	go func() {
		for _, file := range archive {
			if file.Name == filename {
				fp, err := file.Open()
				if err != nil {
					printMessage("IPGeobase", fmt.Sprintf("Can't open %s", filename), "FAIL")
					yield <- nil
				}
				defer fp.Close()
				utf8, err := charset.NewReader(fp, "text/csv; charset=windows-1251")
				if err != nil {
					printMessage("IPGeobase", fmt.Sprintf("%s not in cp1251", filename), "FAIL")
					yield <- nil
				}
				r := csv.NewReader(utf8)
				r.Comma, r.LazyQuotes = '\t', true
				for {
					record, err := r.Read()
					// Stop at EOF.
					if err == io.EOF {
						break
					}
					if err != nil {
						printMessage("IPGeobase", fmt.Sprintf("Can't read line from %s", filename), "WARN")
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

func ipgeobaseCities(archive []*zip.File) (map[string]City, error) {
	cities := make(map[string]City)
	for record := range readIgCSV(archive, "cities.txt") {
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

func ipgeobaseCidr(archive []*zip.File, cities map[string]City) (map[string]City, error) {
	database := make(map[string]City)
	for record := range readIgCSV(archive, "cidr_optim.txt") {
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
		return nil, errors.New("Database is empty")
	}
	return database, nil
}

func ipgeobaseWriteMap(outputDir string, database map[string]City) error {
	reg, err := openMapFile(outputDir, "region.txt")
	if err != nil {
		return err
	}
	city, err := openMapFile(outputDir, "city.txt")
	if err != nil {
		return err
	}
	tz, err := openMapFile(outputDir, "tz.txt")
	if err != nil {
		return err
	}
	defer reg.Close()
	defer city.Close()
	defer tz.Close()
	ipRanges := make(IPList, len(database))
	i := 0
	for ipRange := range database {
		ipRanges[i] = ipRange
		i++
	}
	sort.Sort(ipRanges)
	for _, ipRange := range ipRanges {
		info := database[ipRange]
		fmt.Fprintf(city, "%s %s;\n", ipRange, base64.StdEncoding.EncodeToString([]byte(info.Name)))
		fmt.Fprintf(reg, "%s %02d;\n", ipRange, info.RegID)
		fmt.Fprintf(tz, "%s %s;\n", ipRange, info.TZ)
	}
	return nil
}
