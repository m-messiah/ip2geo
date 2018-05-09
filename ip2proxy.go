package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
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
	items      []*ip2proxyItem
	archive    []*zip.File
	OutputDir  string
	ErrorsChan chan Error
	Token      string
}

func (o *ip2proxy) Get() {
	answer, err := o.download()
	if err != nil {
		o.ErrorsChan <- Error{err, "ip2proxy", "Download"}
		return
	}
	printMessage("ip2proxy", "Download", "OK")
	if err := o.unpack(answer); err != nil {
		o.ErrorsChan <- Error{err, "ip2proxy", "Unpack"}
		return
	}
	printMessage("ip2proxy", "Unpack", "OK")
	if err := o.Parse("IP2PROXY-LITE-PX4.CSV"); err != nil {
		o.ErrorsChan <- Error{err, "ip2proxy", "Parse"}
		return
	}
	printMessage("ip2proxy", "Parse", "OK")
	if err := o.Write(); err != nil {
		o.ErrorsChan <- Error{err, "ip2proxy", "Write Nginx Map"}
		return
	}
	printMessage("ip2proxy", "Write Nginx Map", "OK")
	o.ErrorsChan <- Error{err: nil}
}

func (o *ip2proxy) download() ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.ip2location.com/download", nil)
	q := req.URL.Query()
	q.Add("file", "PX4LITE")
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
	for record := range readCSVDatabase(o.archive, filename, "ip2proxy", ',', false) {
		item, err := o.lineToItem(record)
		if err != nil {
			printMessage("ip2proxy", fmt.Sprintf("Can't parse line from %s with %v", filename, err), "WARN")
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
	file, err := os.Create(path.Join(o.OutputDir, "ip2proxy_net.txt"))
	if err != nil {
		return err
	}
	defer file.Close()
	for _, item := range o.items {
		fmt.Fprintf(file, "%s-%s 1;\n", item.IPFrom, item.IPTo)
	}
	return nil
}
