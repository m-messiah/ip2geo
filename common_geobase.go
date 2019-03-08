package main

type GeoItem struct {
	Name        string
	RegID       int
	ID          string
	City        string
	Network     string
	TZ          string
	NetIP       uint32
	Country     string
	CountryCode string
}

type GeoBase interface {
	Download() ([]byte, error)
	Unpack([]byte) error
	Cities() (map[string]GeoItem, error)
	Network(map[string]GeoItem) error
	WriteMap() error
	Name() string
	AddError(Error)
}

func Generate(geobase GeoBase) {
	answer, err := geobase.Download()
	if err != nil {
		geobase.AddError(Error{err, geobase.Name(), "Download"})
		return
	}
	printMessage(geobase.Name(), "Download", "OK")
	err = geobase.Unpack(answer)
	if err != nil {
		geobase.AddError(Error{err, geobase.Name(), "Unpack"})
		return
	}
	printMessage(geobase.Name(), "Unpack", "OK")
	cities, err := geobase.Cities()
	if err != nil {
		geobase.AddError(Error{err, geobase.Name(), "Generate Cities"})
		return
	}
	printMessage(geobase.Name(), "Generate cities", "OK")
	err = geobase.Network(cities)
	if err != nil {
		geobase.AddError(Error{err, geobase.Name(), "Generate db"})
		return
	}
	printMessage(geobase.Name(), "Generate database", "OK")
	if err := geobase.WriteMap(); err != nil {
		geobase.AddError(Error{err, geobase.Name(), "Write map"})
		return
	}
	printMessage(geobase.Name(), "Write nginx maps", "OK")
	geobase.AddError(Error{err: nil})
}
