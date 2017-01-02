package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
)

type City struct {
	Name   string
	Reg_ID int
	TZ     string
}

func ipgeobase_generate(output_dir string) {
	fmt.Printf("[IPGeobase] Download\t\t\t\t")
	answer := ipgeobase_download()
	if answer != nil {
		color.Green("[OK]")
	}
	fmt.Printf("[IPGeobase] Unpack\t\t\t\t")
	archive := ipgeobase_unpack(answer)
	if archive != nil {
		color.Green("[OK]")
	}
	fmt.Printf("[IPGeobase] Generate cities\t\t\t")
	cities := ipgeobase_cities(archive)
	if len(cities) > 0 {
		color.Green("[OK]")
	}
	fmt.Printf("[IPGeobase] Generate database\t\t\t")
	database := ipgeobase_cidr(archive, cities)
	if len(database) > 0 {
		color.Green("[OK]")
	}
	fmt.Printf("[IPGeobase] Nginx maps\t\t\t\t")
	ipgeobase_write_map(output_dir, database)
	color.Green("[OK]")
}

func ipgeobase_download() []byte {
	resp, err := http.Get("http://ipgeobase.ru/files/db/Main/geo_files.zip")
	if err != nil {
		color.Red("[FAIL]\nUrl no answer: %s", err.Error())
		return nil
	}
	defer resp.Body.Close()
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		color.Red("[FAIL]\nCan't read answer: %s", err.Error())
		return nil
	}
	return answer
}

func ipgeobase_unpack(response []byte) []*zip.File {
	zip_reader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		color.Red("[FAIL]\nBad zip file: %s", err.Error())
		return nil
	}
	return zip_reader.File
}

func read_csv(archive []*zip.File, filename string) chan []string {
	yield := make(chan []string)
	go func() {
		for _, file := range archive {
			if file.Name == filename {
				fp, err := file.Open()
				if err != nil {
					color.Red("[FAIL]\nCan't open %s: %s", filename, err.Error())
					yield <- nil
				}
				defer fp.Close()
				utf8, err := charset.NewReader(fp, "text/csv; charset=windows-1251")
				if err != nil {
					color.Red("[FAIL]\n%s not in cp1251!: %s", filename, err.Error())
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
						color.Yellow("[WARN]\ncan't read line from %s: %s", filename, err.Error())
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
	for record := range read_csv(archive, "cities.txt") {
		if len(record) < 3 {
			color.Red("[FAIL]\ncities.txt too short line: %s", record)
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
		color.Red("[FAIL]\nCities db is empty")
	}
	return cities
}

func ipgeobase_cidr(archive []*zip.File, cities map[string]City) map[string]City {
	database := make(map[string]City)
	for record := range read_csv(archive, "cidr_optim.txt") {
		if len(record) < 5 {
			color.Red("[FAIL]\ncidr_optim.txt too short line: %s", record)
			continue
		}
		// Format is: <int_start>\t<int_end>\t<ip_range>\t<country_code>\tcity_id
		ip_range, country, cid := remove_space(record[2]), record[3], record[4]
		if city, ok := cities[cid]; country == "RU" && ok {
			database[ip_range] = city
		}
	}
	if len(database) < 1 {
		color.Red("[FAIL]\nDatabase is empty")
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
	ip_ranges := make([]string, len(database))
	i := 0
	for ip_range := range database {
		ip_ranges[i] = ip_range
		i++
	}
	sort.Strings(ip_ranges)
	for _, ip_range := range ip_ranges {
		info := database[ip_range]
		fmt.Fprintf(city, "%s %s;\n", ip_range, base64.StdEncoding.EncodeToString([]byte(info.Name)))
		fmt.Fprintf(reg, "%s %02d;\n", ip_range, info.Reg_ID)
		fmt.Fprintf(tz, "%s %s;\n", ip_range, info.TZ)
	}

}
