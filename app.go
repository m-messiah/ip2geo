package main

import (
	"flag"
	"os"
	"sync"
)

func main() {
	outputDir := flag.String("output", "output", "output directory for files")
	ipgeobase := flag.Bool("ipgeobase", false, "enable ipgeobase generation")
	tor := flag.Bool("tor", false, "enable tor generation")
	maxmind := flag.Bool("maxmind", false, "enable maxmind generation")
	maxmindLang := flag.String("lang", "ru", "MaxMind city name language")
	maxmindIPVer := flag.Int("ipver", 4, "MaxMind ip version (4 or 6)")
	maxmindInclude := flag.String("include", "", "MaxMind output filter: only these countries")
	maxmindExclude := flag.String("exclude", "", "MaxMind output filter: except these countries")
	flag.Parse()
	if !(*ipgeobase || *tor || *maxmind) {
		// By default, generate all maps
		*ipgeobase = true
		*tor = true
		*maxmind = true
	}
	os.MkdirAll(*outputDir, 0755)
	printMessage(" ", "Use output directory", *outputDir)
	var wg sync.WaitGroup
	if *ipgeobase {
		wg.Add(1)
		go ipgeobaseGenerate(&wg, *outputDir)
	}

	if *tor {
		wg.Add(1)
		go torGenerate(&wg, *outputDir)
	}

	if *maxmind {
		wg.Add(1)
		go maxmindGenerate(&wg, *outputDir, *maxmindLang, *maxmindIPVer, *maxmindInclude, *maxmindExclude)
	}

	wg.Wait()
}
