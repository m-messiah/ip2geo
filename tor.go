package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
)

func tor_generate(wg *sync.WaitGroup, output_dir string) {
	fmt.Printf("[TOR] Blutmagie Download\t\t\t")
	torlists := make(chan map[string]bool, 2)
	go tor_blutmagie_download(torlists)
	fmt.Printf("[TOR] Torproject Download\t\t\t")
	go tor_torproject_download(torlists)
	fmt.Printf("[TOR] Merge\t\t\t\t\t")
	torlist := tor_merge(torlists)
	if torlist != nil {
		color.Green("[OK]")
	}
	fmt.Printf("[TOR] Nginx maps\t\t\t\t")
	tor_write_map(output_dir, torlist)
	color.Green("[OK]")
	defer wg.Done()
}

func tor_blutmagie_download(ch chan map[string]bool) {
	resp, err := http.Get("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
	if err != nil {
		color.Red("[FAIL]\nUrl no answer: %s", err.Error())
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
			color.Yellow("[WARN]\ncan't read line from blutmagie: %s", err.Error())
			continue
		}
		if len(line) < 1 {
			continue
		}
		torlist[strings.TrimSpace(line)] = true
	}
	ch <- torlist
}

func tor_torproject_download(ch chan map[string]bool) {
	resp, err := http.Get("https://check.torproject.org/exit-addresses")
	if err != nil {
		color.Red("[FAIL]\nUrl no answer: %s", err.Error())
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
			color.Yellow("[WARN]\ncan't read line from torproject: %s", err.Error())
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
	ch <- torproject
}

func tor_merge(ch chan map[string]bool) []string {
	a := <-ch
	for k, v := range <-ch {
		a[k] = v
	}
	ip_list := make([]string, len(a))
	i := 0
	for ip := range a {
		ip_list[i] = ip
		i++
	}
	sort.Strings(ip_list)
	return ip_list
}

func tor_write_map(output_dir string, torlist []string) {
	tor := open_map_file(output_dir, "tor.txt")
	defer tor.Close()
	for _, ip := range torlist {
		fmt.Fprintf(tor, "%s-%s 1;\n", ip, ip)
	}

}
