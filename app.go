package main

import (
	"flag"
	"os"
)

var LogLevel int = 0

func main() {
	outputDir := flag.String("output", "output", "output directory for files")
	ipgeobase := flag.Bool("ipgeobase", false, "enable ipgeobase generation")
	tor := flag.Bool("tor", false, "enable tor generation")
	ip2proxyFlag := flag.Bool("ip2proxy", false, "enable ip2proxy generation")
	ip2proxyToken := flag.String("ip2proxy-token", "", "Get token here https://lite.ip2location.com/file-download" )
	maxmind := flag.Bool("maxmind", false, "enable maxmind generation")
	maxmindIPVer := flag.Int("ipver", 4, "MaxMind ip version (4 or 6)")
	maxmindLang := flag.String("lang", "ru", "MaxMind city name language")
	maxmindTZNames := flag.Bool("tznames", false, "MaxMind TZ in names format (for example `Europe/Moscow`)")
	maxmindInclude := flag.String("include", "", "MaxMind output filter: only these countries")
	maxmindExclude := flag.String("exclude", "", "MaxMind output filter: except these countries")
	quiet := flag.Bool("q", false, "Be quiet - skip [OK]")
	veryQuiet := flag.Bool("qq", false, "Be very quiet - show only errors")
	flag.Parse()
	if !(*ipgeobase || *tor || *maxmind || *ip2proxyFlag) {
		// By default, generate all maps
		*ipgeobase = true
		*tor = true
		*maxmind = true
		*ip2proxyFlag = *ip2proxyToken != ""
	}
	if *quiet {
		LogLevel = 1
	}
	if *veryQuiet {
		LogLevel = 2
	}
	os.MkdirAll(*outputDir, 0755)
	if LogLevel < 2 {
		printMessage(" ", "Use output directory", *outputDir)
	}
	goroutinesCount := 0
	errorChannel := make(chan Error)
	if *ipgeobase {
		goroutinesCount++
		go ipgeobaseGenerate(*outputDir, errorChannel)
	}

	if *tor {
		goroutinesCount++
		go torGenerate(*outputDir, errorChannel)
	}

	if *maxmind {
		goroutinesCount++
		go maxmindGenerate(*outputDir, *maxmindLang, *maxmindIPVer, *maxmindTZNames, *maxmindInclude, *maxmindExclude, errorChannel)
	}

	if *ip2proxyFlag {
		goroutinesCount++
		o := ip2proxy{
			Token: *ip2proxyToken,
			ErrorsChan: errorChannel,
			OutputDir: *outputDir,
		}
		go o.Get()
	}

	for i := 0; i < goroutinesCount; i++ {
		err := <-errorChannel
		if err.err != nil {
			printMessage(err.Module, err.Action+": "+err.err.Error(), "FAIL")
			os.Exit(1)
		}
	}
	if LogLevel < 1 {
		printMessage(" ", "Generation done", "OK")
	}
}
