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
)

func maxmind_generate(wg *sync.WaitGroup, output_dir, lang string, ipver int, include, exclude string) {
	answer := maxmind_download()
	if answer != nil {
		print_message("MaxMind", "Download", "OK")
	}
	archive := maxmind_unpack(answer)
	if archive != nil {
		print_message("MaxMind", "Unpack", "OK")
	}
	cities := maxmind_cities(archive, lang, include, exclude)
	if len(cities) > 0 {
		print_message("MaxMind", "Generate cities", "OK")
	}
	database := maxmind_network(archive, ipver, cities)
	if len(database) > 0 {
		print_message("MaxMind", "Generate database", "OK")
	}
	maxmind_write_map(output_dir, database)
	print_message("MaxMind", "Write nginx maps", "OK")
	defer wg.Done()
}

func maxmind_download() []byte {
	resp, err := http.Get("http://geolite.maxmind.com/download/geoip/database/GeoLite2-City-CSV.zip")
	if err != nil {
		print_message("MaxMind", "Download no answer", "FAIL")
		return nil
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print_message("MaxMind", "Download bad answer", "FAIL")
		return nil
	}
	return answer
}

func maxmind_unpack(response []byte) []*zip.File {
	zip_reader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		print_message("MaxMind", "Bad zip file", "FAIL")
		return nil
	}
	return zip_reader.File
}

func read_mm_csv(archive []*zip.File, filename string) chan []string {
	yield := make(chan []string)
	go func() {
		for _, file := range archive {
			if strings.Contains(file.Name, filename) {
				fp, err := file.Open()
				if err != nil {
					print_message("MaxMind", fmt.Sprintf("Can't open %s", filename), "FAIL")
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
						print_message("MaxMind", fmt.Sprintf("Can't read line from %s", filename), "WARN")
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

func maxmind_cities(archive []*zip.File, language, include, exclude string) map[string]Location {
	locations := make(map[string]Location)
	for record := range read_mm_csv(archive, "GeoLite2-City-Locations-"+language+".csv") {
		if len(record) < 13 {
			print_message("MaxMind", fmt.Sprintf("GeoLite2-City-Locations-"+language+".csv"+" too short line: %s", record), "FAIL")
			continue
		}
		country := record[4]
		if len(record[10]) < 1 || len(country) < 1 {
			continue
		}
		if len(include) < 1 || strings.Contains(include, country) {
			if !strings.Contains(exclude, country) {
				locations[record[0]] = Location{
					ID:   record[0],
					City: record[10],
					TZ:   record[12],
				}
			}
		}
	}
	if len(locations) < 1 {
		print_message("MaxMind", "Locations db is empty", "FAIL")
	}
	return locations
}

func maxmind_network(archive []*zip.File, ipver int, locations map[string]Location) Database {
	var database Database
	for record := range read_mm_csv(archive, "GeoLite2-City-Blocks-IPv"+strconv.Itoa(ipver)+".csv") {
		if len(record) < 2 {
			print_message("MaxMind", fmt.Sprintf("GeoLite2-City-Blocks-IPv"+strconv.Itoa(ipver)+".csv"+" too short line: %s", record), "FAIL")
			continue
		}
		ip_range := get_ip_range(ipver, record[0])
		net_ip := net.ParseIP(strings.Split(ip_range, "-")[0])
		if net_ip == nil {
			continue
		}
		geo_id := record[1]
		if location, ok := locations[geo_id]; ok {
			database = append(database, Location{
				ID:      geo_id,
				City:    location.City,
				Network: ip_range,
				TZ:      location.TZ,
				NetIP:   ip2int(net_ip),
			})
		}
	}
	if len(database) < 1 {
		print_message("MaxMind", "Network db is empty", "FAIL")
	}
	sort.Sort(database)
	return database
}

func maxmind_write_map(output_dir string, database Database) {
	city := open_map_file(output_dir, "mm_city.txt")
	// TODO: Convert TZ in delta format
	// tz := open_map_file(output_dir, "mm_tz.txt")
	defer city.Close()
	// defer tz.Close()
	for _, location := range database {
		fmt.Fprintf(city, "%s %s;\n", location.Network, base64.StdEncoding.EncodeToString([]byte(location.City)))
		// fmt.Fprintf(tz, "%s %s;\n", ip_range, location.TZ)
	}

}
