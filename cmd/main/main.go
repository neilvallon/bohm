package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"image/png"

	"vallon.me/bohm"
	"vallon.me/bohm/config"
	"vallon.me/shortening"
)

var (
	overlappingDefaults = config.Defaults{
		N:             2,
		Width:         48,
		Height:        48,
		PeriodicInput: true,
		Symmetry:      8,
		Screenshots:   2,
	}
	tiledDefaults = config.Defaults{
		Width:       10,
		Height:      10,
		Screenshots: 2,
	}
)

var random *rand.Rand

func init() {
	flag.StringVar(&inputFile, "i", "", "xml file containing texture jobs")
	flag.StringVar(&textureDir, "t", "", "directory containing texture images")
	flag.StringVar(&outputDir, "o", "./out", "output directory")

	flag.Parse()

	if inputFile == "" || textureDir == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	seed := time.Now().UnixNano()
	random = rand.New(rand.NewSource(seed))
	log.Printf("SEED: %q\n", shortening.Encode(uint64(seed)))
}

var inputFile, textureDir, outputDir string

func main() {
	cfg, err := config.Read(inputFile)
	if err != nil {
		panic(err)
	}

	count := 1
	for _, s := range cfg.Samples {
		name := s.Name
		log.Println(name)

		var m interface {
			Run(seed int64, limit int) bool
			Graphics() (image.Image, error)
		}

		switch s.XMLName.Local {
		case "overlapping":
			s.Set(overlappingDefaults)
			m = bohm.NewOverlapping(textureDir, s.Name, s.N,
				s.Width, s.Height, *s.PeriodicInput, s.Periodic, s.Symmetry, s.Ground)

		case "simpletiled":
			s.Set(tiledDefaults)
			m = bohm.NewTiled(textureDir, s.Name, s.Subset, s.Width, s.Height, s.Periodic, s.Black)

		default:
			log.Println(s.XMLName.Local, "not implemented")
			continue
		}

		for i := 0; i < s.Screenshots; i++ {
			for k := 0; k < 10; k++ {
				seed := random.Int63()
				ident := fmt.Sprintf("%d %s+%s %d", count, s.Name, shortening.Encode(uint64(seed)), i)
				if m.Run(seed, s.Limit) {
					log.Printf("[%s]\tDONE\n", ident)

					img, err := m.Graphics()
					if err != nil {
						panic(err)
					}

					saveImage(filepath.Join(outputDir, ident+".png"), img)
					break
				} else {
					log.Printf("[%s]\tCONTRADICTION\n", ident)
				}
			}
		}
		count++
	}
}

func saveImage(name string, img image.Image) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
