package main

import (
	"flag"
	"os"
)

var logLevel int

func main() {
	outputDir := flag.String("output", "output", "output directory for files")
	ipgeobase := flag.Bool("ipgeobase", false, "enable ipgeobase generation")
	tor := flag.Bool("tor", false, "enable tor generation")
	ip2proxyLiteFlag := flag.Bool("ip2proxy", false, "enable ip2proxy PX4-LITE generation")
	ip2proxyFlag := flag.Bool("ip2proxy-pro", false, "enable ip2proxy PX4 generation")
	ip2proxyLiteToken := flag.String("ip2proxy-token", "", "Get token here https://lite.ip2location.com/file-download")
	ip2proxyToken := flag.String("ip2proxy-pro-token", "", "ip2proxy download token")
	ip2proxyLiteFilename := flag.String("ip2proxy-lite-filename", "", "Filename of already downloaded ip2proxy-lite db")
	ip2proxyFilename := flag.String("ip2proxy-pro-filename", "", "Filename of already downloaded ip2proxy db")
	ip2proxyPrintType := flag.Bool("ip2proxy-print-type", false, "Print proxy type in map, instead of `1`")
	maxmind := flag.Bool("maxmind", false, "enable maxmind generation")
	maxmindIPVer := flag.Int("ipver", 4, "MaxMind ip version (4 or 6)")
	maxmindLang := flag.String("lang", "ru", "MaxMind city name language")
	maxmindTZNames := flag.Bool("tznames", false, "MaxMind TZ in names format (for example `Europe/Moscow`)")
	maxmindInclude := flag.String("include", "", "MaxMind output filter: only these countries")
	maxmindExclude := flag.String("exclude", "", "MaxMind output filter: except these countries")
	maxmindNoBase64 := flag.Bool("nobase64", false, "MaxMind Cities as-is (without base64 encode). DO NOT USE IT IF YOU NOT SURE ABOUT MaxMind encoding")
	maxmindNoCountry := flag.Bool("nocountry", false, "do not add maxmind country maps")
	quiet := flag.Bool("q", false, "Be quiet - skip [OK]")
	veryQuiet := flag.Bool("qq", false, "Be very quiet - show only errors")
	version := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()
	if *version {
		printMessage("ip2geo", "version "+VERSION, "OK")
		return
	}
	if !(*ipgeobase || *tor || *maxmind || *ip2proxyLiteFlag || *ip2proxyFlag) {
		// By default, generate all maps
		*ipgeobase = true
		*tor = true
		*maxmind = true
		*ip2proxyLiteFlag = *ip2proxyLiteToken != "" || *ip2proxyLiteFilename != ""
		*ip2proxyFlag = *ip2proxyToken != "" || *ip2proxyFilename != ""
	}
	if *quiet {
		logLevel = 1
	}
	if *veryQuiet {
		logLevel = 2
	}
	os.MkdirAll(*outputDir, 0755)
	if logLevel < 2 {
		printMessage(" ", "Use output directory", *outputDir)
	}
	goroutinesCount := 0
	errorChannel := make(chan Error)
	if *ipgeobase {
		goroutinesCount++
		i := IPGeobase{
			OutputDir:  *outputDir,
			ErrorsChan: errorChannel,
		}
		go Generate(&i)
	}

	if *tor {
		goroutinesCount++
		t := Tor{
			OutputDir:  *outputDir,
			ErrorsChan: errorChannel,
		}
		go t.Generate()
	}

	if *maxmind {
		goroutinesCount++
		m := MaxMind{
			OutputDir:  *outputDir,
			ErrorsChan: errorChannel,
			lang:       *maxmindLang,
			ipver:      *maxmindIPVer,
			tzNames:    *maxmindTZNames,
			include:    *maxmindInclude,
			exclude:    *maxmindExclude,
			noBase64:   *maxmindNoBase64,
			noCountry:  *maxmindNoCountry,
		}
		go Generate(&m)
	}

	if *ip2proxyLiteFlag {
		goroutinesCount++
		o := ip2proxy{
			Name:       "ip2proxyLite",
			Token:      *ip2proxyLiteToken,
			Filename:   *ip2proxyLiteFilename,
			ErrorsChan: errorChannel,
			OutputDir:  *outputDir,
			PrintType:  *ip2proxyPrintType,
		}
		go o.Get()
	}

	if *ip2proxyFlag {
		goroutinesCount++
		o := ip2proxy{
			Name:       "ip2proxyPro",
			Token:      *ip2proxyToken,
			Filename:   *ip2proxyFilename,
			ErrorsChan: errorChannel,
			OutputDir:  *outputDir,
			PrintType:  *ip2proxyPrintType,
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
	if logLevel < 1 {
		printMessage(" ", "Generation done", "OK")
	}
}
