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

func ipgeobase_generate(wg *sync.WaitGroup, output_dir string) {
	answer := ipgeobase_download()
	if answer != nil {
		print_message("IPGeobase", "Download", "OK")
	}
	archive := ipgeobase_unpack(answer)
	if archive != nil {
		print_message("IPGeobase", "Unpack", "OK")
	}
	cities := ipgeobase_cities(archive)
	if len(cities) > 0 {
		print_message("IPGeobase", "Generate cities", "OK")
	}
	database := ipgeobase_cidr(archive, cities)
	if len(database) > 0 {
		print_message("IPGeobase", "Generate database", "OK")
	}
	ipgeobase_write_map(output_dir, database)
	print_message("IPGeobase", "Write nginx maps", "OK")
	defer wg.Done()
}

func ipgeobase_download() []byte {
	resp, err := http.Get("http://ipgeobase.ru/files/db/Main/geo_files.zip")
	if err != nil {
		print_message("IPGeobase", "Download no answer", "FAIL")
		return nil
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print_message("IPGeobase", "Download bad answer", "FAIL")
		return nil
	}
	return answer
}

func ipgeobase_unpack(response []byte) []*zip.File {
	zip_reader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		print_message("IPGeobase", "Bad zip file", "FAIL")
		return nil
	}
	return zip_reader.File
}

func read_ig_csv(archive []*zip.File, filename string) chan []string {
	yield := make(chan []string)
	go func() {
		for _, file := range archive {
			if file.Name == filename {
				fp, err := file.Open()
				if err != nil {
					print_message("IPGeobase", fmt.Sprintf("Can't open %s", filename), "FAIL")
					yield <- nil
				}
				defer fp.Close()
				utf8, err := charset.NewReader(fp, "text/csv; charset=windows-1251")
				if err != nil {
					print_message("IPGeobase", fmt.Sprintf("%s not in cp1251", filename), "FAIL")
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
						print_message("IPGeobase", fmt.Sprintf("Can't read line from %s", filename), "WARN")
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

func ipgeobase_cities(archive []*zip.File) map[string]City {
	cities := make(map[string]City)
	for record := range read_ig_csv(archive, "cities.txt") {
		if len(record) < 3 {
			print_message("IPGeobase", fmt.Sprintf("cities.txt too short line: %s", record), "FAIL")
			continue
		}
		// Format is:  <city_id>\t<city_name>\t<region>\t<district>\t<lattitude>\t<longitude>
		cid, city, region_name := record[0], record[1], record[2]
		if region, ok := REGIONS[region_name]; ok {
			if cid == "1199" {
				region, _ = REGIONS["Москва"]
			}
			cities[cid] = City{
				Name:   city,
				Reg_ID: region.ID,
				TZ:     region.TZ,
			}
		}
	}
	if len(cities) < 1 {
		print_message("IPGeobase", "Cities db is empty", "FAIL")
	}
	return cities
}

func ipgeobase_cidr(archive []*zip.File, cities map[string]City) map[string]City {
	database := make(map[string]City)
	for record := range read_ig_csv(archive, "cidr_optim.txt") {
		if len(record) < 5 {
			print_message("IPGeobase", fmt.Sprintf("cidr_optim.txt too short line: %s", record), "FAIL")
			continue
		}
		// Format is: <int_start>\t<int_end>\t<ip_range>\t<country_code>\tcity_id
		ip_range, country, cid := remove_space(record[2]), record[3], record[4]
		if city, ok := cities[cid]; country == "RU" && ok {
			database[ip_range] = city
		}
	}
	if len(database) < 1 {
		print_message("IPGeobase", "Database is empty", "FAIL")
	}
	return database
}

func ipgeobase_write_map(output_dir string, database map[string]City) {
	reg := open_map_file(output_dir, "region.txt")
	city := open_map_file(output_dir, "city.txt")
	tz := open_map_file(output_dir, "tz.txt")
	defer reg.Close()
	defer city.Close()
	defer tz.Close()
	ip_ranges := make(IPList, len(database))
	i := 0
	for ip_range := range database {
		ip_ranges[i] = ip_range
		i++
	}
	sort.Sort(ip_ranges)
	for _, ip_range := range ip_ranges {
		info := database[ip_range]
		fmt.Fprintf(city, "%s %s;\n", ip_range, base64.StdEncoding.EncodeToString([]byte(info.Name)))
		fmt.Fprintf(reg, "%s %02d;\n", ip_range, info.Reg_ID)
		fmt.Fprintf(tz, "%s %s;\n", ip_range, info.TZ)
	}

}
