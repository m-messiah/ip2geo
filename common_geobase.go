package main

type geoItem struct {
	Name        string
	RegID       int
	ID          string
	City        string
	Network     string
	TZ          string
	Country     string
	CountryCode string
}

// GeoBase interface for downloadable and convertible geo database
type GeoBase interface {
	download() ([]byte, error)
	unpack([]byte) error
	citiesDB() (map[string]geoItem, error)
	writeMap(map[string]geoItem) error
	name() string
	addError(Error)
}

// Generate GeoBase (download, unpack, parse and write in nginx map format)
func Generate(geobase GeoBase) {
	answer, err := geobase.download()
	if err != nil {
		geobase.addError(Error{err, geobase.name(), "Download"})
		return
	}
	printMessage(geobase.name(), "Download", "OK")
	err = geobase.unpack(answer)
	if err != nil {
		geobase.addError(Error{err, geobase.name(), "Unpack"})
		return
	}
	printMessage(geobase.name(), "Unpack", "OK")
	cities, err := geobase.citiesDB()
	if err != nil {
		geobase.addError(Error{err, geobase.name(), "Generate Cities"})
		return
	}
	printMessage(geobase.name(), "Generate cities", "OK")
	if err := geobase.writeMap(cities); err != nil {
		geobase.addError(Error{err, geobase.name(), "Write map"})
		return
	}
	printMessage(geobase.name(), "Write nginx maps", "OK")
	geobase.addError(Error{err: nil})
}
