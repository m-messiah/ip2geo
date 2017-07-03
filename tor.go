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

func torGenerate(outputDir string, errors_chan chan Error) {
	torLists := make(chan map[string]bool, 2)
	go torBlutmagieDownload(torLists)
	go torTorProjectDownload(torLists)
	torlist := torMerge(torLists)
	if torlist != nil && len(torlist) > 0 {
		printMessage("TOR", "Merge", "OK")
	} else {
		errors_chan <- Error{errors.New("torlist empty"), "TOR", "Merge"}
		return
	}
	if err := torWriteMap(outputDir, torlist); err != nil {
		errors_chan <- Error{err, "TOR", "nginx"}
		return
	} else {
		printMessage("TOR", "Write nginx maps", "OK")
	}
	errors_chan <- Error{err: nil}
}

func torBlutmagieDownload(ch chan map[string]bool) {
	resp, err := http.Get("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
	if err != nil {
		printMessage("TOR", "Blutmagie Download", "FAIL")
		ch <- nil
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
	ch <- torlist
}

func torTorProjectDownload(ch chan map[string]bool) {
	resp, err := http.Get("https://check.torproject.org/exit-addresses")
	if err != nil {
		printMessage("TOR", "Torproject Download", "FAIL")
		ch <- nil
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
	ch <- torproject
}

func torMerge(ch chan map[string]bool) IPList {
	a := <-ch
	for k, v := range <-ch {
		a[k] = v
	}
	ipList := make(IPList, len(a))
	i := 0
	for ip := range a {
		ipList[i] = ip
		i++
	}
	sort.Sort(ipList)
	return ipList
}

func torWriteMap(outputDir string, torlist IPList) error {
	tor, err := openMapFile(outputDir, "tor.txt")
	if err != nil {
		return err
	}
	defer tor.Close()
	for _, ip := range torlist {
		fmt.Fprintf(tor, "%s-%s 1;\n", ip, ip)
	}
	return nil

}
