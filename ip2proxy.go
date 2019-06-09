package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
)

const (
	IP2PROXY_PRO = iota
	IP2PROXY_LITE
)

type ip2proxyItem struct {
	IPFrom      net.IP
	IPTo        net.IP
	ProxyType   string
	CountryCode string
	Country     string
	Region      string
	City        string
	Company     string
}

type ip2proxy struct {
	Type        int
	items       []*ip2proxyItem
	archive     []*zip.File
	OutputDir   string
	ErrorsChan  chan Error
	Token       string
	Filename    string
	name        string
	csvFilename string
	zipFilename string
	PrintType   bool
}

func (o *ip2proxy) checkErr(err error, message string) bool {
	if err != nil {
		o.ErrorsChan <- Error{err, o.name, message}
		return true
	}
	printMessage(o.name, message, "OK")
	return false
}

func (o *ip2proxy) Get() {
	if o.Type == IP2PROXY_PRO {
		o.name = "ip2proxyPro"
		o.csvFilename = "IP2PROXY-IP-PROXYTYPE-COUNTRY-REGION-CITY-ISP.CSV"
		o.zipFilename = "PX4-IP-PROXYTYPE-COUNTRY-REGION-CITY-ISP"
	} else if o.Type == IP2PROXY_LITE {
		o.name = "ip2proxyLite"
		o.csvFilename = "IP2PROXY-LITE-PX4.CSV"
		o.zipFilename = "PX4LITE"
	} else {
		o.ErrorsChan <- Error{errors.New("Unknown ip2proxy type requested"), o.name, "bad init"}
		return
	}
	fileData, err := o.getZip()
	if o.checkErr(err, "Get ZIP") {
		return
	}
	err = o.unpack(fileData)
	if o.checkErr(err, "Unpack") {
		return
	}
	err = o.Parse(o.csvFilename)
	if o.checkErr(err, "Parse") {
		return
	}
	err = o.Write()
	if o.checkErr(err, "Write Nginx Map") {
		return
	}
	o.ErrorsChan <- Error{err: nil}
}

func (o *ip2proxy) getZip() ([]byte, error) {
	if len(o.Token) > 0 {
		return o.download()
	} else if len(o.Filename) > 0 {
		return ioutil.ReadFile(o.Filename)
	} else {
		return nil, errors.New("Token or Filename must be passed")
	}
}

func (o *ip2proxy) download() ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.ip2location.com/download", nil)
	q := req.URL.Query()
	q.Add("file", o.zipFilename)
	q.Add("token", o.Token)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed with code %d", resp.StatusCode)
	}
	answer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (o *ip2proxy) unpack(response []byte) error {
	file, err := Unpack(response)
	if err == nil {
		o.archive = file
	}
	return err
}

func (o *ip2proxy) Parse(filename string) error {
	var list []*ip2proxyItem
	for record := range readCSVDatabase(o.archive, filename, o.name, ',', false) {
		item, err := o.lineToItem(record)
		if err != nil {
			printMessage(o.name, fmt.Sprintf("Can't parse line from %s with %v", filename, err), "WARN")
			continue
		}
		list = append(list, item)
	}
	o.items = list
	return nil
}

func (o *ip2proxy) lineToItem(line []string) (*ip2proxyItem, error) {
	if len(line) != 8 {
		return nil, fmt.Errorf("Number of field is not 8")
	}
	var (
		ipFromInt, ipToInt int64
		err                error
	)
	if ipFromInt, err = strconv.ParseInt(line[0], 10, 64); err != nil {
		return nil, fmt.Errorf("Can't parse FromIP with: %v", err)
	}
	if ipToInt, err = strconv.ParseInt(line[1], 10, 64); err != nil {
		return nil, fmt.Errorf("Can't parse ToIP with: %v", err)
	}
	return &ip2proxyItem{
		IPFrom:      int2ip(ipFromInt),
		IPTo:        int2ip(ipToInt),
		ProxyType:   line[2],
		CountryCode: line[3],
		Country:     line[4],
		Region:      line[5],
		City:        line[6],
		Company:     line[7],
	}, nil
}

func (o *ip2proxy) Write() error {
	return o.writeNetworks()
}

func (o *ip2proxy) writeNetworks() error {
	file, err := os.Create(path.Join(o.OutputDir, o.name+"_net.txt"))
	if err != nil {
		return err
	}
	defer file.Close()
	var mapValue string
	for _, item := range o.items {
		if o.PrintType {
			mapValue = item.ProxyType
		} else {
			mapValue = "1"
		}
		fmt.Fprintf(file, "%s-%s \"%s\";\n", item.IPFrom, item.IPTo, mapValue)
	}
	return nil
}
