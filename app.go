package main

import (
	"flag"
	"github.com/jinzhu/configor"
	"os"
)

type ip2ProxyConfig struct {
	Enabled  bool `default:"false"`
	Token    string
	Filename string
}

// Config - all configuration for tool defined here
var Config = struct {
	LogLevel  int    `default:"0"`
	OutputDir string `default:"output"`
	TOR       struct {
		Enabled bool `default:"false"`
	}
	IP2Proxy struct {
		Lite      ip2ProxyConfig
		Pro       ip2ProxyConfig
		PrintType bool `default:"false"`
	}
	MaxMind struct {
		Enabled   bool   `default:"false"`
		IPVer     int    `default:"4"`
		Lang      string `default:"ru"`
		TZNames   bool   `default:"false"`
		Include   string
		Exclude   string
		NoBase64  bool `default:"false"`
		NoCountry bool `default:"false"`
	}
	IPGeobase struct {
		Enabled bool `default:"false"`
	}
}{}

func main() {
	configFile := flag.String("c", "", "Read config from file")
	quiet := flag.Bool("q", false, "Be quiet - skip [OK]")
	veryQuiet := flag.Bool("qq", false, "Be very quiet - show only errors")
	version := flag.Bool("version", false, "Print version information and exit")

	flag.StringVar(&Config.OutputDir, "output", "output", "output directory for files")
	flag.BoolVar(&Config.IPGeobase.Enabled, "ipgeobase", false, "enable ipgeobase generation")
	flag.BoolVar(&Config.TOR.Enabled, "tor", false, "enable tor generation")
	flag.BoolVar(&Config.IP2Proxy.Lite.Enabled, "ip2proxy", false, "enable ip2proxy PX4-LITE generation")
	flag.BoolVar(&Config.IP2Proxy.Pro.Enabled, "ip2proxy-pro", false, "enable ip2proxy PX4 generation")
	flag.StringVar(&Config.IP2Proxy.Lite.Token, "ip2proxy-token", "", "Get token here https://lite.ip2location.com/file-download")
	flag.StringVar(&Config.IP2Proxy.Pro.Token, "ip2proxy-pro-token", "", "ip2proxy download token")
	flag.StringVar(&Config.IP2Proxy.Lite.Filename, "ip2proxy-lite-filename", "", "Filename of already downloaded ip2proxy-lite db")
	flag.StringVar(&Config.IP2Proxy.Pro.Filename, "ip2proxy-pro-filename", "", "Filename of already downloaded ip2proxy db")
	flag.BoolVar(&Config.IP2Proxy.PrintType, "ip2proxy-print-type", false, "Print proxy type in map, instead of `1`")
	flag.BoolVar(&Config.MaxMind.Enabled, "maxmind", false, "enable maxmind generation")
	flag.IntVar(&Config.MaxMind.IPVer, "ipver", 4, "MaxMind ip version (4 or 6)")
	flag.StringVar(&Config.MaxMind.Lang, "lang", "ru", "MaxMind city name language")
	flag.BoolVar(&Config.MaxMind.TZNames, "tznames", false, "MaxMind TZ in names format (for example `Europe/Moscow`)")
	flag.StringVar(&Config.MaxMind.Include, "include", "", "MaxMind output filter: only these countries")
	flag.StringVar(&Config.MaxMind.Exclude, "exclude", "", "MaxMind output filter: except these countries")
	flag.BoolVar(&Config.MaxMind.NoBase64, "nobase64", false, "MaxMind Cities as-is (without base64 encode). DO NOT USE IT IF YOU NOT SURE ABOUT MaxMind encoding")
	flag.BoolVar(&Config.MaxMind.NoCountry, "nocountry", false, "do not add maxmind country maps")
	flag.Parse()
	if *version {
		printMessage("ip2geo", "version "+VERSION, "OK")
		return
	}
	configor.Load(&Config, *configFile)
	if *quiet {
		Config.LogLevel = 1
	}
	if *veryQuiet {
		Config.LogLevel = 2
	}
	if !(Config.IPGeobase.Enabled || Config.TOR.Enabled || Config.MaxMind.Enabled || Config.IP2Proxy.Lite.Enabled || Config.IP2Proxy.Pro.Enabled) {
		// By default, generate all maps
		Config.IPGeobase.Enabled = true
		Config.TOR.Enabled = true
		Config.MaxMind.Enabled = true
		Config.IP2Proxy.Lite.Enabled = Config.IP2Proxy.Lite.Token != "" || Config.IP2Proxy.Lite.Filename != ""
		Config.IP2Proxy.Pro.Enabled = Config.IP2Proxy.Pro.Token != "" || Config.IP2Proxy.Pro.Filename != ""
	}

	os.MkdirAll(Config.OutputDir, 0755)
	if Config.LogLevel < 2 {
		printMessage(" ", "Use output directory", Config.OutputDir)
	}
	goroutinesCount := 0
	errorChannel := make(chan Error)
	if Config.IPGeobase.Enabled {
		goroutinesCount++
		i := IPGeobase{
			OutputDir:  Config.OutputDir,
			ErrorsChan: errorChannel,
		}
		go Generate(&i)
	}

	if Config.TOR.Enabled {
		goroutinesCount++
		t := Tor{
			OutputDir:  Config.OutputDir,
			ErrorsChan: errorChannel,
		}
		go t.Generate()
	}

	if Config.MaxMind.Enabled {
		goroutinesCount++
		m := MaxMind{
			OutputDir:  Config.OutputDir,
			ErrorsChan: errorChannel,
			lang:       Config.MaxMind.Lang,
			ipver:      Config.MaxMind.IPVer,
			tzNames:    Config.MaxMind.TZNames,
			include:    Config.MaxMind.Include,
			exclude:    Config.MaxMind.Exclude,
			noBase64:   Config.MaxMind.NoBase64,
			noCountry:  Config.MaxMind.NoCountry,
		}
		go Generate(&m)
	}

	for t, ip2ProxyType := range [2]ip2ProxyConfig{Config.IP2Proxy.Lite, Config.IP2Proxy.Pro} {
		if ip2ProxyType.Enabled {
			goroutinesCount++
			var name string
			if t == 0 {
				name = "ip2proxyLite"
			} else {
				name = "ip2proxyPro"

			}
			o := ip2proxy{
				Name:       name,
				Token:      ip2ProxyType.Token,
				Filename:   ip2ProxyType.Filename,
				ErrorsChan: errorChannel,
				OutputDir:  Config.OutputDir,
				PrintType:  Config.IP2Proxy.PrintType,
			}
			go o.Get()
		}
	}

	for i := 0; i < goroutinesCount; i++ {
		err := <-errorChannel
		if err.err != nil {
			printMessage(err.Module, err.Action+": "+err.err.Error(), "FAIL")
			os.Exit(1)
		}
	}
	if Config.LogLevel < 1 {
		printMessage(" ", "Generation done", "OK")
	}
}
