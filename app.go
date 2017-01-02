package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
)

func main() {
	output_dir := flag.String("output", "output", "output directory for files")
	ipgeobase := flag.Bool("ipgeobase", false, "enable ipgeobase generation")
	tor := flag.Bool("tor", false, "enable tor generation")
	maxmind := flag.Bool("maxmind", false, "enable maxmind generation")
	// maxmind_lang := flag.String("lang", "ru", "MaxMind city name language")
	// maxmind_ipver := flag.Int("ipver", 4, "MaxMind ip version (4 or 6)")
	// maxmind_include := flag.String("include", "", "MaxMind output filter: only these countries")
	// maxmind_exclude := flag.String("exclude", "", "MaxMind output filter: except these countries")
	flag.Parse()
	if !(*ipgeobase || *tor || *maxmind) {
		// By default, generate all maps
		*ipgeobase = true
		*tor = true
		*maxmind = true
	}
	os.MkdirAll(*output_dir, 0755)
	fmt.Printf("Use %s as output directory\n", color.BlueString("%s", *output_dir))
	if *ipgeobase {
		ipgeobase_generate(*output_dir)
	}
}
