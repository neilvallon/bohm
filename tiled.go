package bohm

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strconv"

	"vallon.me/bohm/config"
)

type textureFn func(x, y int) color.RGBA

type Tiled struct {
	propagator [4][][]bool

	tiles    []textureDef
	tileSize int

	black bool

	//
	FM       Point
	periodic bool

	Model
}

func NewTiled(path, name, subsetName string, width, height int, periodic, black bool) *Tiled {
	tm := &Tiled{
		FM:       Point{width, height},
		periodic: periodic,
		black:    black,
	}

	tm.Model = NewModel(tm)

	tileCfg := config.ReadTileData(filepath.Join(path, name, "data.xml"))
	if tileCfg.Size == 0 {
		tileCfg.Size = 16
	}

	tm.tileSize = tileCfg.Size

	subset := tileCfg.SubsetList(subsetName)

	tm.stationary = tm.stationary[0:0]

	var action [][8]int
	firstOccurrence := make(map[string]int)
	for _, tile := range tileCfg.Tiles {
		tilename := tile.Name
		if len(subset) != 0 && !subset.Contains(tilename) {
			continue
		}

		var a, b func(int) int
		var cardinality int
		switch tile.Symmetry {
		case "L":
			cardinality = 4
			a = func(i int) int { return (i + 1) % 4 }
			b = func(i int) int {
				if i%2 == 0 {
					return i + 1
				}
				return i - 1
			}
		case "T":
			cardinality = 4
			a = func(i int) int { return (i + 1) % 4 }
			b = func(i int) int {
				if i%2 == 0 {
					return i
				}
				return 4 - i
			}
		case "I":
			cardinality = 2
			a = func(i int) int { return 1 - i }
			b = func(i int) int { return i }
		case "\\":
			cardinality = 2
			a = func(i int) int { return 1 - i }
			b = func(i int) int { return 1 - i }
		default:
			cardinality = 1
			a = func(i int) int { return i }
			b = func(i int) int { return i }
		}

		T := len(action)

		firstOccurrence[tilename] = T
		var cmap_ [4][8]int
		cmap := cmap_[:cardinality]
		for t := range cmap {
			cmap[t][0] = T + t
			cmap[t][1] = T + a(t)
			cmap[t][2] = T + a(a(t))
			cmap[t][3] = T + a(a(a(t)))
			cmap[t][4] = T + b(t)
			cmap[t][5] = T + b(a(t))
			cmap[t][6] = T + b(a(a(t)))
			cmap[t][7] = T + b(a(a(a(t))))

			action = append(action, cmap[t])
		}

		if tileCfg.Unique {
			for t := 0; t < cardinality; t++ {
				file := filepath.Join(path, name, fmt.Sprintf("%s %d.png", tile.Name, t))
				tm.tiles = append(tm.tiles, textureDef{name: file, size: tm.tileSize})
			}
		} else {
			file := filepath.Join(path, name, tile.Name+".png")
			for t := 0; t < cardinality; t++ {
				tm.tiles = append(tm.tiles, textureDef{
					name:        file,
					size:        tm.tileSize,
					cardinality: t,
				})
			}
		}

		for t := 0; t < cardinality; t++ {
			weight := tile.Weight
			if weight == 0 {
				weight = 1
			}
			tm.stationary = append(tm.stationary, weight)
		}
	}

	T := len(action)

	for d := range tm.propagator {
		tm.propagator[d] = make([][]bool, T)
		for t := range tm.propagator[d] {
			tm.propagator[d][t] = make([]bool, T)
		}
	}

	tm.wave = make([][][]bool, tm.FM.X)
	tm.changes = make([][]bool, tm.FM.X)

	for x := range tm.wave {
		tm.wave[x] = make([][]bool, tm.FM.Y)
		tm.changes[x] = make([]bool, tm.FM.Y)
		for y := range tm.wave[x] {
			tm.wave[x][y] = make([]bool, T)
		}
	}

	for _, neighbor := range tileCfg.Neighbors {
		split := func(s string) [2]string {
			for i, c := range s {
				if c == ' ' {
					return [2]string{s[:i], s[i+1:]}
				}
			}
			return [2]string{s, ""}
		}

		left := split(neighbor.Left)
		right := split(neighbor.Right)

		if len(subset) != 0 && (!subset.Contains(left[0]) || !subset.Contains(right[0])) {
			continue
		}

		var lInd, rInd int
		if left[1] != "" {
			var err error
			if lInd, err = strconv.Atoi(left[1]); err != nil {
				panic(err)
			}
		}

		if right[1] != "" {
			var err error
			if rInd, err = strconv.Atoi(right[1]); err != nil {
				panic(err)
			}
		}

		L := action[firstOccurrence[left[0]]][lInd]
		D := action[L][1]
		R := action[firstOccurrence[right[0]]][rInd]
		U := action[R][1]

		tm.propagator[0][L][R] = true
		tm.propagator[0][action[L][6]][action[R][6]] = true
		tm.propagator[0][action[R][4]][action[L][4]] = true
		tm.propagator[0][action[R][2]][action[L][2]] = true

		tm.propagator[1][D][U] = true
		tm.propagator[1][action[U][6]][action[D][6]] = true
		tm.propagator[1][action[D][4]][action[U][4]] = true
		tm.propagator[1][action[U][2]][action[D][2]] = true
	}

	for t1 := range tm.propagator[2] {
		for t2 := range tm.propagator[2][t1] {
			tm.propagator[2][t1][t2] = tm.propagator[0][t2][t1]
			tm.propagator[3][t1][t2] = tm.propagator[1][t2][t1]
		}
	}

	return tm
}

func (tm *Tiled) Propagate() bool {
	var change bool
	for x2 := range tm.changes {
		for y2 := range tm.changes[x2] {
			for d := range tm.propagator {
				x1 := x2
				y1 := y2

				switch d {
				case 0:
					if x2 == 0 {
						if !tm.periodic {
							continue
						}
						x1 = tm.FM.X - 1
					} else {
						x1 = x2 - 1
					}
				case 1:
					if y2 == tm.FM.Y-1 {
						if !tm.periodic {
							continue
						}
						y1 = 0
					} else {
						y1 = y2 + 1
					}
				case 2:
					if x2 == tm.FM.X-1 {
						if !tm.periodic {
							continue
						}
						x1 = 0
					} else {
						x1 = x2 + 1
					}
				default:
					if y2 == 0 {
						if !tm.periodic {
							continue
						}
						y1 = tm.FM.Y - 1
					} else {
						y1 = y2 - 1
					}
				}

				if !tm.changes[x1][y1] {
					continue
				}

				for t2, on := range tm.wave[x2][y2] {
					if on {
						b := false
						for t1, on := range tm.wave[x1][y1] {
							b = on && tm.propagator[d][t1][t2]
							if b {
								break
							}
						}
						if !b {
							tm.wave[x2][y2][t2] = false
							tm.changes[x2][y2] = true
							change = true
						}
					}
				}
			}
		}
	}
	return change
}

func (Tiled) OnBoundary(_, _ int) bool { return false }

func (tm *Tiled) Graphics() (image.Image, error) {
	result := image.NewRGBA(image.Rect(0, 0, tm.FM.X*tm.tileSize, tm.FM.Y*tm.tileSize))

	tileBuf := make([]float64, tm.tileSize*tm.tileSize*4)
	for x, col := range tm.wave {
		for y, row := range col {
			var amount int
			var lambda float64
			for t, on := range row {
				if on {
					amount++
					lambda += tm.stationary[t]
				}
			}

			for i := 0; i < len(tileBuf); i += 4 {
				tileBuf[i] = 0x00
				tileBuf[i+1] = 0x00
				tileBuf[i+2] = 0x00
				tileBuf[i+3] = 0x00
			}

			if !tm.black || amount != len(row) {
				for t, on := range row {
					if !on {
						continue
					}
					weight := tm.stationary[t] / lambda

					texture, err := tm.tiles[t].Open()
					if err != nil {
						return nil, err
					}

					tile := texture.CarTile(tm.tiles[t].cardinality)
					for p := 0; p < len(tileBuf); p += 4 {
						tileBuf[p] += float64(tile[p]) * weight
						tileBuf[p+1] += float64(tile[p+1]) * weight
						tileBuf[p+2] += float64(tile[p+2]) * weight
						tileBuf[p+3] += float64(tile[p+3]) * weight
					}
				}
			}

			for yt := 0; yt < tm.tileSize; yt++ {
				row := tileBuf[(yt*tm.tileSize)*4:]
				for xt := 0; xt < tm.tileSize; xt++ {
					cell := row[xt*4:]
					result.SetRGBA(x*tm.tileSize+xt, y*tm.tileSize+yt, color.RGBA{
						R: uint8(cell[0]),
						G: uint8(cell[1]),
						B: uint8(cell[2]),
						A: uint8(cell[3]),
					})
				}
			}
		}
	}
	return result, nil
}
