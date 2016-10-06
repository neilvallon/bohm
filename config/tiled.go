package config

import (
	"encoding/xml"
	"os"
)

func ReadTileData(name string) *tileSet {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := xml.NewDecoder(f)

	var ts tileSet
	if err := dec.Decode(&ts); err != nil {
		panic(err)
	}

	return &ts
}

type tile struct {
	Name     string  `xml:"name,attr"`
	Symmetry string  `xml:"symmetry,attr"`
	Weight   float64 `xml:"weight,attr"`
}

type tileSet struct {
	Size      int    `xml:"size,attr"`
	Unique    bool   `xml:"unique,attr"`
	Tiles     []tile `xml:"tiles>tile"`
	Neighbors []struct {
		Left  string `xml:"left,attr"`
		Right string `xml:"right,attr"`
	} `xml:"neighbors>neighbor"`
	Subsets []struct {
		Name  string `xml:"name,attr"`
		Tiles []tile `xml:"tile"`
	} `xml:"subsets>subset"`
}

func (ts tileSet) SubsetList(subset string) (set setList) {
	if subset == "" {
		return
	}

	for _, ss := range ts.Subsets {
		if ss.Name == subset {
			for _, t := range ss.Tiles {
				set = append(set, t.Name)
			}
			break
		}
	}
	return
}

type setList []string

func (l setList) Contains(name string) bool {
	for _, s := range l {
		if s == name {
			return true
		}
	}
	return false
}
