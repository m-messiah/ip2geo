package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
)

func ipgeobaseGenerate(wg *sync.WaitGroup, outputDir string) {
	answer := ipgeobaseDownload()
	if answer != nil {
		printMessage("IPGeobase", "Download", "OK")
	}
	archive := ipgeobaseUnpack(answer)
	if archive != nil {
		printMessage("IPGeobase", "Unpack", "OK")
	}
	cities := ipgeobaseCities(archive)
	if len(cities) > 0 {
		printMessage("IPGeobase", "Generate cities", "OK")
	}
	database := ipgeobaseCidr(archive, cities)
	if len(database) > 0 {
		printMessage("IPGeobase", "Generate database", "OK")
	}
	ipgeobaseWriteMap(outputDir, database)
	printMessage("IPGeobase", "Write nginx maps", "OK")
	defer wg.Done()
}

func ipgeobaseDownload() []byte {
	resp, err := http.Get("http://ipgeobase.ru/files/db/Main/geo_files.zip")
	if err != nil {
		printMessage("IPGeobase", "Download no answer", "FAIL")
		return nil
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		printMessage("IPGeobase", "Download bad answer", "FAIL")
		return nil
	}
	return answer
}

func ipgeobaseUnpack(response []byte) []*zip.File {
	zipReader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		printMessage("IPGeobase", "Bad zip file", "FAIL")
		return nil
	}
	return zipReader.File
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

func ipgeobaseCities(archive []*zip.File) map[string]City {
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
		printMessage("IPGeobase", "Cities db is empty", "FAIL")
	}
	return cities
}

func ipgeobaseCidr(archive []*zip.File, cities map[string]City) map[string]City {
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
		printMessage("IPGeobase", "Database is empty", "FAIL")
	}
	return database
}

func ipgeobaseWriteMap(outputDir string, database map[string]City) {
	reg := openMapFile(outputDir, "region.txt")
	city := openMapFile(outputDir, "city.txt")
	tz := openMapFile(outputDir, "tz.txt")
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

}
