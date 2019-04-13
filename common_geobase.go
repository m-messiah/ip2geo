package main

type GeoItem struct {
	Name        string
	RegID       int
	ID          string
	City        string
	Network     string
	TZ          string
	Country     string
	CountryCode string
}

type GeoBase interface {
	Download() ([]byte, error)
	Unpack([]byte) error
	Cities() (map[string]GeoItem, error)
	WriteMap(map[string]GeoItem) error
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
	if err := geobase.WriteMap(cities); err != nil {
		geobase.AddError(Error{err, geobase.Name(), "Write map"})
		return
	}
	printMessage(geobase.Name(), "Write nginx maps", "OK")
	geobase.AddError(Error{err: nil})
}
