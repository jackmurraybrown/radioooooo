package analytics

// ✮⋆‧° geo lookup — resolves IP to country code, then forgets the IP
// https://dev.maxmind.com/geoip/geolite2-free-geolocation-data

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoResolver struct {
	db *geoip2.Reader
}

func NewGeoResolver(dbPath string) (*GeoResolver, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("geoip: open %s: %w", dbPath, err)
	}
	return &GeoResolver{db: db}, nil
}

// resolves an IP to a two-letter country code. returns "XX" if unknown.
func (g *GeoResolver) Country(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "XX"
	}
	record, err := g.db.Country(parsed)
	if err != nil || record.Country.IsoCode == "" {
		return "XX"
	}
	return record.Country.IsoCode
}

func (g *GeoResolver) Close() error {
	return g.db.Close()
}
