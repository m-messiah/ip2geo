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

func tor_generate(wg *sync.WaitGroup, output_dir string) {
	torlists := make(chan map[string]bool, 2)
	go tor_blutmagie_download(torlists)
	go tor_torproject_download(torlists)
	torlist := tor_merge(torlists)
	if torlist != nil {
		print_message("TOR", "Merge", "OK")
	}
	tor_write_map(output_dir, torlist)
	print_message("TOR", "Write nginx maps", "OK")
	defer wg.Done()
}

func tor_blutmagie_download(ch chan map[string]bool) {
	resp, err := http.Get("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
	if err != nil {
		print_message("TOR", "Blutmagie Download", "FAIL")
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
			print_message("TOR", "can't read line from blutmagie", "WARN")
			continue
		}
		if len(line) < 1 {
			continue
		}
		torlist[strings.TrimSpace(line)] = true
	}
	print_message("TOR", "Blutmagie Download", "OK")
	ch <- torlist
}

func tor_torproject_download(ch chan map[string]bool) {
	resp, err := http.Get("https://check.torproject.org/exit-addresses")
	if err != nil {
		print_message("TOR", "Torproject Download", "FAIL")
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
			print_message("TOR", "Can't read line from torproject", "WARN")
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
	print_message("TOR", "Torproject Download", "OK")
	ch <- torproject
}

func tor_merge(ch chan map[string]bool) IPList {
	a := <-ch
	for k, v := range <-ch {
		a[k] = v
	}
	ip_list := make(IPList, len(a))
	i := 0
	for ip := range a {
		ip_list[i] = ip
		i++
	}
	sort.Sort(ip_list)
	return ip_list
}

func tor_write_map(output_dir string, torlist IPList) {
	tor := open_map_file(output_dir, "tor.txt")
	defer tor.Close()
	for _, ip := range torlist {
		fmt.Fprintf(tor, "%s-%s 1;\n", ip, ip)
	}

}
