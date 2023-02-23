package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type ip2proxyItem struct {
	IPFrom      net.IP
	IPTo        net.IP
	ProxyType   string
	CountryCode string
	Country     string
	Region      string
	City        string
	ISP         string
}

type ip2proxy struct {
	archive     []*zip.File
	OutputDir   string
	ErrorsChan  chan Error
	Token       string
	Filename    string
	Name        string
	csvFilename string
	zipFilename string
	PrintType   bool
}

func (o *ip2proxy) checkErr(err error, message string) bool {
	if err != nil {
		o.ErrorsChan <- Error{err, o.Name, message}
		return true
	}
	printMessage(o.Name, message, "OK")
	return false
}

func (o *ip2proxy) Get() {
	if o.Name == "ip2proxyPro" {
		o.csvFilename = "IP2PROXY-IP-PROXYTYPE-COUNTRY-REGION-CITY-ISP.CSV"
		o.zipFilename = "PX4"
	} else if o.Name == "ip2proxyLite" {
		o.csvFilename = "IP2PROXY-LITE-PX4.CSV"
		o.zipFilename = "PX4LITE"
	} else {
		o.ErrorsChan <- Error{errors.New("Unknown ip2proxy type requested"), o.Name, "bad init"}
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
		return os.ReadFile(o.Filename)
	} else {
		return nil, errors.New("Token or Filename must be passed")
	}
}

func (o *ip2proxy) download() ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.ip2location.com/download", nil)
	if err != nil {
		return nil, err
	}
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
	answer, err := io.ReadAll(resp.Body)
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

func (o *ip2proxy) Parse(filename string) <-chan *ip2proxyItem {
	database := make(chan *ip2proxyItem)
	go func() {
		for record := range readCSVDatabase(o.archive, filename, o.Name, ',', false) {
			item, err := o.lineToItem(record)
			if err != nil {
				printMessage(o.Name, fmt.Sprintf("Can't parse line from %s with %v", filename, err), "WARN")
				continue
			}
			database <- item
		}
		close(database)
	}()
	return database
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
		ISP:         line[7],
	}, nil
}

func (o *ip2proxy) Write() error {
	netFile, err := os.Create(path.Join(o.OutputDir, o.Name+"_net.txt"))
	if err != nil {
		return err
	}
	defer netFile.Close()
	ispFile, err := os.Create(path.Join(o.OutputDir, o.Name+"_isp.txt"))
	if err != nil {
		return err
	}
	defer ispFile.Close()
	var mapValue string
	for item := range o.Parse(o.csvFilename) {
		if o.PrintType {
			mapValue = item.ProxyType
		} else {
			mapValue = "1"
		}
		fmt.Fprintf(netFile, "%s-%s \"%s\";\n", item.IPFrom, item.IPTo, mapValue)
		fmt.Fprintf(ispFile, "%s-%s \"%s\";\n", item.IPFrom, item.IPTo, strings.Replace(item.ISP, "\"", "\\\"", -1))
	}
	return nil
}
