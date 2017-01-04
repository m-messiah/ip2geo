package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"net/http"
	"sort"
	"strings"
)

func tor_generate(output_dir string) {
	fmt.Printf("[TOR] Blutmagie Download\t\t\t")
	torlist := tor_blutmagie_download()
	if torlist != nil {
		color.Green("[OK]")
	}
	fmt.Printf("[TOR] Torproject Download\t\t\t")
	torproject := tor_torproject_download()
	if torproject != nil {
		color.Green("[OK]")
	}
	fmt.Printf("[TOR] Merge\t\t\t\t\t")
	tor_merge(torlist, torproject)
	if torlist != nil {
		color.Green("[OK]")
	}
	fmt.Printf("[TOR] Nginx maps\t\t\t\t")
	tor_write_map(output_dir, torlist)
	color.Green("[OK]")
}

func tor_blutmagie_download() map[string]bool {
	resp, err := http.Get("https://torstatus.blutmagie.de/ip_list_exit.php/Tor_ip_list_EXIT.csv")
	if err != nil {
		color.Red("[FAIL]\nUrl no answer: %s", err.Error())
		return nil
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
	return torlist
}

func tor_torproject_download() map[string]bool {
	resp, err := http.Get("https://check.torproject.org/exit-addresses")
	if err != nil {
		color.Red("[FAIL]\nUrl no answer: %s", err.Error())
		return nil
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
	return torproject
}

func tor_merge(a, b map[string]bool) {
	for k, v := range b {
		a[k] = v
	}
}

func tor_write_map(output_dir string, torlist map[string]bool) {
	tor := open_map_file(output_dir, "tor.txt")
	defer tor.Close()
	ip_list := make([]string, len(torlist))
	i := 0
	for ip := range torlist {
		ip_list[i] = ip
		i++
	}
	sort.Strings(ip_list)
	for _, ip := range ip_list {
		fmt.Fprintf(tor, "%s-%s 1;\n", ip, ip)
	}

}
