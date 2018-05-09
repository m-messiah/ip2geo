package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type Tor struct {
	OutputDir  string
	ErrorsChan chan Error
	list       IPList
	temp_lists chan map[string]bool
}

func (tor *Tor) Generate() {
	tor.temp_lists = make(chan map[string]bool, 2)
	go tor.BlutmagieDownload()
	go tor.TorProjectDownload()
	if err := tor.Merge(); err != nil {
		tor.ErrorsChan <- Error{err, "TOR", "Merge"}
		return
	}
	printMessage("TOR", "Merge", "OK")
	if err := tor.WriteMap(); err != nil {
		tor.ErrorsChan <- Error{err, "TOR", "nginx"}
		return
	}
	printMessage("TOR", "Write nginx maps", "OK")
	tor.ErrorsChan <- Error{err: nil}
}

func (tor *Tor) BlutmagieDownload() {
	resp, err := http.Get("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
	if err != nil {
		printMessage("TOR", "Blutmagie Download", "FAIL")
		tor.temp_lists <- nil
		return
	}
	defer resp.Body.Close()
	torlist := make(map[string]bool)
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		if err != nil {
			printMessage("TOR", "can't read line from blutmagie", "WARN")
			continue
		}
		if len(line) < 1 {
			continue
		}
		torlist[strings.TrimSpace(line)] = true
	}
	printMessage("TOR", "Blutmagie Download", "OK")
	tor.temp_lists <- torlist
}

func (tor *Tor) TorProjectDownload() {
	resp, err := http.Get("https://check.torproject.org/exit-addresses")
	if err != nil {
		printMessage("TOR", "Torproject Download", "FAIL")
		tor.temp_lists <- nil
		return
	}
	defer resp.Body.Close()
	torproject := make(map[string]bool)
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			printMessage("TOR", "Can't read line from torproject", "WARN")
			continue
		}
		if len(line) < 1 {
			continue
		}
		if !strings.Contains(line, "ExitAddress") {
			continue
		}
		fields := strings.Fields(line)
		torproject[fields[1]] = true
	}
	printMessage("TOR", "Torproject Download", "OK")
	tor.temp_lists <- torproject
}

func (tor *Tor) Merge() error {
	result := make(map[string]bool)
	for i := 0; i < 2; i++ {
		m := <-tor.temp_lists
		if m == nil {
			continue
		}
		for k, v := range m {
			result[k] = v
		}
	}
	ipList := make(IPList, len(result))
	i := 0
	for ip := range result {
		ipList[i] = ip
		i++
	}
	sort.Sort(ipList)
	tor.list = ipList
	if len(tor.list) > 0 {
		return nil
	}
	return errors.New("torlist empty")
}

func (tor *Tor) WriteMap() error {
	tor_file, err := openMapFile(tor.OutputDir, "tor.txt")
	if err != nil {
		return err
	}
	defer tor_file.Close()
	for _, ip := range tor.list {
		fmt.Fprintf(tor_file, "%s-%s 1;\n", ip, ip)
	}
	return nil

}
