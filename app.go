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
	maxmind := flag.Bool("maxmind", false, "enable maxmind generation")
	maxmindIPVer := flag.Int("ipver", 4, "MaxMind ip version (4 or 6)")
	maxmindLang := flag.String("lang", "ru", "MaxMind city name language")
	maxmindTZNames := flag.Bool("tznames", false, "MaxMind TZ in names format (for example `Europe/Moscow`)")
	maxmindInclude := flag.String("include", "", "MaxMind output filter: only these countries")
	maxmindExclude := flag.String("exclude", "", "MaxMind output filter: except these countries")
	quiet := flag.Bool("q", false, "Be quiet - skip [OK]")
	very_quiet := flag.Bool("qq", false, "Be very quiet - show only errors")
	flag.Parse()
	if !(*ipgeobase || *tor || *maxmind) {
		// By default, generate all maps
		*ipgeobase = true
		*tor = true
		*maxmind = true
	}
	if *quiet {
		LogLevel = 1
	}
	if *very_quiet {
		LogLevel = 2
	}
	os.MkdirAll(*outputDir, 0755)
	if LogLevel < 2 {
		printMessage(" ", "Use output directory", *outputDir)
	}
	goroutines_count := 0
	error_channel := make(chan Error)
	if *ipgeobase {
		goroutines_count++
		go ipgeobaseGenerate(*outputDir, error_channel)
	}

	if *tor {
		goroutines_count++
		go torGenerate(*outputDir, error_channel)
	}

	if *maxmind {
		goroutines_count++
		go maxmindGenerate(*outputDir, *maxmindLang, *maxmindIPVer, *maxmindTZNames, *maxmindInclude, *maxmindExclude, error_channel)
	}

	for i := 0; i < goroutines_count; i++ {
		err := <-error_channel
		if err.err != nil {
			printMessage(err.Module, err.Action+": "+err.err.Error(), "FAIL")
			os.Exit(1)
		}
	}
	if LogLevel < 1 {
		printMessage(" ", "Generation done", "OK")
	}
}
