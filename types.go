package traefikgeoip2

import (
	"fmt"
	"net"
	"strconv"

	"github.com/IncSW/geoip2"
)

// Unknown constant for undefined/unresolvable data.
const Unknown = "XX"

// DefaultDBPath default GeoIP2 database path.
const DefaultDBPath = "GeoLite2-Country.mmdb"

const (
	// CountryHeader carries the ISO 3166-1 alpha-2 country code.
	CountryHeader = "X-GeoIP2-Country"
	// RegionHeader carries the ISO 3166-2 subdivision code.
	RegionHeader = "X-GeoIP2-Region"
	// CityHeader carries the English city name.
	CityHeader = "X-GeoIP2-City"
	// PostalCodeHeader carries the postal/ZIP code.
	// Only populated when using a GeoLite2-City (or GeoIP2-City) database.
	// Empty string when the DB has no postal data for the IP.
	PostalCodeHeader = "X-GeoIP2-PostalCode"
	// IPAddressHeader carries the resolved client IP used for the lookup.
	IPAddressHeader = "X-GeoIP2-IPAddress"
	// PrivateIPHeader is "true" when the client IP is a private/LAN/loopback address
	// and no real GeoIP lookup was performed.
	PrivateIPHeader = "X-GeoIP2-IsPrivate"
	// LatitudeHeader carries the latitude from the GeoIP2 Location data.
	// Only populated when using a City database.
	LatitudeHeader = "X-GeoIP2-Latitude"
	// LongitudeHeader carries the longitude from the GeoIP2 Location data.
	// Only populated when using a City database.
	LongitudeHeader = "X-GeoIP2-Longitude"
)

// GeoIPResult holds the resolved geo data for a single IP lookup.
type GeoIPResult struct {
	country    string
	region     string
	city       string
	postalCode string
	latitude   string
	longitude  string
}

// LookupGeoIP2 is the function signature for a GeoIP database lookup.
type LookupGeoIP2 func(ip net.IP) (*GeoIPResult, error)

// CreateCityDBLookup returns a LookupGeoIP2 backed by a City database.
// Populates country, region, city and postal code (where available).
func CreateCityDBLookup(rdr *geoip2.CityReader) LookupGeoIP2 {
	return func(ip net.IP) (*GeoIPResult, error) {
		rec, err := rdr.Lookup(ip)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		retval := GeoIPResult{
			country:    rec.Country.ISOCode,
			region:     Unknown,
			city:       Unknown,
			postalCode: Unknown,
			latitude:   Unknown,
			longitude:  Unknown,
		}
		if city, ok := rec.City.Names["en"]; ok {
			retval.city = city
		}
		if rec.Subdivisions != nil {
			retval.region = rec.Subdivisions[0].ISOCode
		}
		if rec.Postal.Code != "" {
			retval.postalCode = rec.Postal.Code
		}
		if rec.Location.Latitude != 0 || rec.Location.Longitude != 0 {
			retval.latitude = strconv.FormatFloat(rec.Location.Latitude, 'f', -1, 64)
			retval.longitude = strconv.FormatFloat(rec.Location.Longitude, 'f', -1, 64)
		}
		return &retval, nil
	}
}

// CreateCountryDBLookup returns a LookupGeoIP2 backed by a Country database.
// Country DBs do not contain city or postal data; those fields are set to Unknown.
func CreateCountryDBLookup(rdr *geoip2.CountryReader) LookupGeoIP2 {
	return func(ip net.IP) (*GeoIPResult, error) {
		rec, err := rdr.Lookup(ip)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		return &GeoIPResult{
			country:    rec.Country.ISOCode,
			region:     Unknown,
			city:       Unknown,
			postalCode: Unknown,
			latitude:   Unknown,
			longitude:  Unknown,
		}, nil
	}
}