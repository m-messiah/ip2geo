package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
)

func torGenerate(wg *sync.WaitGroup, outputDir string) {
	torLists := make(chan map[string]bool, 2)
	go torBlutmagieDownload(torLists)
	go torTorProjectDownload(torLists)
	torlist := torMerge(torLists)
	if torlist != nil {
		printMessage("TOR", "Merge", "OK")
	}
	torWriteMap(outputDir, torlist)
	printMessage("TOR", "Write nginx maps", "OK")
	defer wg.Done()
}

func torBlutmagieDownload(ch chan map[string]bool) {
	resp, err := http.Get("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
	if err != nil {
		printMessage("TOR", "Blutmagie Download", "FAIL")
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

func torWriteMap(outputDir string, torlist IPList) {
	tor := openMapFile(outputDir, "tor.txt")
	defer tor.Close()
	for _, ip := range torlist {
		fmt.Fprintf(tor, "%s-%s 1;\n", ip, ip)
	}

}
