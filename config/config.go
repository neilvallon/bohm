package config

import (
	"encoding/xml"
	"os"
)

type MainConfig struct {
	Samples []Sample `xml:",any"`
}

type Sample struct {
	XMLName xml.Name
	Name    string `xml:"name,attr"`

	Width  int `xml:"width,attr"`
	Height int `xml:"height,attr"`

	N     int `xml:"N,attr"`
	Limit int `xml:"limit,attr"`

	Symmetry int `xml:"symmetry,attr"`
	Ground   int `xml:"ground,attr"`

	Screenshots int `xml:"screenshots,attr"`

	Subset string `xml:"subset,attr"`

	Black bool `xml:"black,attr"`

	Periodic      bool  `xml:"periodic,attr"`
	PeriodicInput *bool `xml:"periodicInput,attr"`
}

func (s *Sample) Set(defaults Defaults) {
	if s.N == 0 {
		s.N = defaults.N
	}

	if s.Width == 0 {
		s.Width = defaults.Width
	}
	if s.Height == 0 {
		s.Height = defaults.Height
	}

	if s.PeriodicInput == nil {
		s.PeriodicInput = &defaults.PeriodicInput
	}

	if s.Symmetry == 0 {
		s.Symmetry = defaults.Symmetry
	}
	if s.Screenshots == 0 {
		s.Screenshots = defaults.Screenshots
	}
}

func Read(name string) (*MainConfig, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg MainConfig
	if err := xml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type Defaults struct {
	N, Width, Height      int
	Symmetry, Screenshots int

	PeriodicInput bool
}
