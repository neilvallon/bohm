package bohm

import (
	"image"
	"image/color"
	_ "image/png"
	"os"
)

type Overlapping struct {
	propagator [][][][]int
	N          int

	patterns [][]byte
	colors   []color.Color
	ground   int

	//
	T        int
	FM       Point
	periodic bool

	Model
}

func NewOverlapping(name string, N, width, height int, periodicInput, periodicOutput bool, symmetry, ground int) *Overlapping {
	om := &Overlapping{
		N:        N,
		periodic: periodicOutput,
		ground:   ground,
		FM:       Point{X: width, Y: height},
	}

	om.Model = NewModel(om)

	bitmap, err := openBMP("samples/" + name + ".png")
	if err != nil {
		panic(err)
	}

	var SMX, SMY int
	{
		rect := bitmap.Bounds()
		SMX = rect.Max.X
		SMY = rect.Max.Y
	}
	// byte[,] sample = new byte[SMX, SMY];
	sample := make([][]byte, SMX)
	for x := range sample {
		sample[x] = make([]byte, SMY)
	}

	// var colors []color.Color
	for y := 0; y < SMY; y++ {
		for x := 0; x < SMX; x++ {
			color := bitmap.At(x, y)

			var i int
			for _, c := range om.colors {
				if c == color {
					break
				}
				i++
			}

			if i == len(om.colors) {
				om.colors = append(om.colors, color)
			}
			sample[x][y] = byte(i)
		}
	}

	C := len(om.colors)
	W := power(C, N*N)

	pattern := func(f func(int, int) byte) []byte {
		result := make([]byte, N*N)
		for y := 0; y < N; y++ {
			for x := 0; x < N; x++ {
				result[x+y*N] = f(x, y)
			}
		}
		return result
	}

	patternFromSample := func(x, y int) []byte {
		return pattern(func(dx, dy int) byte {
			return sample[(x+dx)%SMX][(y+dy)%SMY]
		})
	}

	rotate := func(p []byte) []byte {
		return pattern(func(x, y int) byte {
			return p[N-1-y+x*N]
		})
	}

	reflect := func(p []byte) []byte {
		return pattern(func(x, y int) byte {
			return p[N-1-x+y*N]
		})
	}

	index := func(p []byte) int {
		result := 0
		power := 1
		for i := 0; i < len(p); i++ {
			result += int(p[len(p)-1-i]) * power
			power *= C
		}
		return result
	}

	patternFromIndex := func(ind int) []byte {
		residue := ind
		power := W
		result := make([]byte, N*N)
		for i := 0; i < len(result); i++ {
			power /= C

			count := 0
			for residue >= power {
				residue -= power
				count++
			}

			result[i] = byte(count)
		}

		return result
	}

	// Dictionary<int, int> weights = new Dictionary<int, int>();
	weights := make(map[int]int)
	var ordering []int
	for y := 0; (periodicInput && y < SMY) || (!periodicInput && y < SMY-N+1); y++ {
		for x := 0; (periodicInput && x < SMX) || (!periodicInput && x < SMX-N+1); x++ {
			var ps [8][]byte

			ps[0] = patternFromSample(x, y)
			ps[1] = reflect(ps[0])
			ps[2] = rotate(ps[0])
			ps[3] = reflect(ps[2])
			ps[4] = rotate(ps[2])
			ps[5] = reflect(ps[4])
			ps[6] = rotate(ps[4])
			ps[7] = reflect(ps[6])

			for k := 0; k < symmetry; k++ {
				ind := index(ps[k])
				if _, ok := weights[ind]; ok {
					weights[ind]++
				} else {
					weights[ind] = 1
					ordering = append(ordering, ind)
				}
			}
		}
	}

	om.T = len(weights)
	om.ground = (om.ground + om.T) % om.T

	om.patterns = make([][]byte, om.T)
	om.stationary = make([]float64, om.T)
	om.propagator = make([][][][]int, om.T)

	for i, w := range ordering {
		om.patterns[i] = patternFromIndex(w)
		om.stationary[i] = float64(weights[w])
	}

	om.wave = make([][][]bool, om.FM.X)
	om.changes = make([][]bool, om.FM.X)
	for x := 0; x < om.FM.X; x++ {
		om.wave[x] = make([][]bool, om.FM.Y)
		om.changes[x] = make([]bool, om.FM.Y)
		for y := 0; y < om.FM.Y; y++ {
			om.wave[x][y] = make([]bool, om.T)
			om.changes[x][y] = false
			for t := 0; t < om.T; t++ {
				om.wave[x][y][t] = true
			}
		}
	}

	// Func<byte[], byte[], int, int, bool> agrees = (p1, p2, dx, dy) =>
	agrees := func(p1, p2 []byte, dx, dy int) bool {
		xmin := dx
		xmax := N
		if dx < 0 {
			xmin = 0
			xmax += dx
		}

		ymin := dy
		ymax := N
		if dy < 0 {
			ymin = 0
			ymax += dy
		}

		for y := ymin; y < ymax; y++ {
			for x := xmin; x < xmax; x++ {
				if p1[x+N*y] != p2[x-dx+N*(y-dy)] {
					return false
				}
			}
		}
		return true
	}

	for t := 0; t < om.T; t++ {
		om.propagator[t] = make([][][]int, 2*N-1)
		for x := 0; x < 2*N-1; x++ {
			om.propagator[t][x] = make([][]int, 2*N-1)
			for y := 0; y < 2*N-1; y++ {
				var list []int
				for t2 := 0; t2 < om.T; t2++ {
					if agrees(om.patterns[t], om.patterns[t2], x-N+1, y-N+1) {
						list = append(list, t2)
					}
				}
				om.propagator[t][x][y] = make([]int, len(list))
				for c := 0; c < len(list); c++ {
					om.propagator[t][x][y][c] = list[c]
				}
			}
		}
	}

	return om
}

func (om *Overlapping) OnBoundary(x, y int) bool {
	return !om.periodic && (x+om.N > om.FM.X || y+om.N > om.FM.Y)
}

func (om *Overlapping) Propagate() bool {
	change := false
	var b bool
	var x2, y2, sx, sy int
	var allowed []bool

	for x1 := 0; x1 < om.FM.X; x1++ {
		for y1 := 0; y1 < om.FM.Y; y1++ {
			if om.changes[x1][y1] {
				om.changes[x1][y1] = false
				for dx := -om.N + 1; dx < om.N; dx++ {
					for dy := -om.N + 1; dy < om.N; dy++ {
						x2 = x1 + dx
						y2 = y1 + dy

						sx = x2
						if sx < 0 {
							sx += om.FM.X
						} else if sx >= om.FM.X {
							sx -= om.FM.X
						}

						sy = y2
						if sy < 0 {
							sy += om.FM.Y
						} else if sy >= om.FM.Y {
							sy -= om.FM.Y
						}

						if !om.periodic && (sx+om.N > om.FM.X || sy+om.N > om.FM.Y) {
							continue
						}
						allowed = om.wave[sx][sy]

						for t2 := 0; t2 < om.T; t2++ {
							if !allowed[t2] {
								continue
							}
							b = false
							prop := om.propagator[t2][om.N-1-dx][om.N-1-dy]
							for i1 := 0; i1 < len(prop) && !b; i1++ {
								b = om.wave[x1][y1][prop[i1]]
							}
							if !b {
								om.changes[sx][sy] = true
								change = true
								allowed[t2] = false
							}
						}
					}
				}
			}
		}
	}
	return change
}

func (om *Overlapping) Graphics() (image.Image, error) {
	result := image.NewRGBA(image.Rect(0, 0, om.FM.X, om.FM.Y))
	for y := 0; y < om.FM.Y; y++ {
		for x := 0; x < om.FM.X; x++ {
			var contributors []byte
			for dy := 0; dy < om.N; dy++ {
				for dx := 0; dx < om.N; dx++ {
					sx := x - dx
					if sx < 0 {
						sx += om.FM.X
					}

					sy := y - dy
					if sy < 0 {
						sy += om.FM.Y
					}

					if om.OnBoundary(sx, sy) {
						continue
					}

					for t, on := range om.wave[sx][sy] {
						if on {
							contributors = append(contributors, om.patterns[t][dx+dy*om.N])
						}
					}
				}
			}

			var r, g, b, a uint32
			for _, c := range contributors {
				r_, g_, b_, a_ := om.colors[c].RGBA()
				r += r_
				g += g_
				b += b_
				a += a_
			}

			lambda := 1.0 / float64(len(contributors))
			result.SetRGBA(x, y, color.RGBA{
				R: uint8(lambda * float64(r)),
				G: uint8(lambda * float64(g)),
				B: uint8(lambda * float64(b)),
				A: uint8(lambda * float64(a)),
			})
		}
	}
	return result, nil
}

func (om *Overlapping) Clear() {
	om.Model.Clear()

	if om.ground != 0 {
		for x := 0; x < om.FM.X; x++ {
			for t := 0; t < om.T; t++ {
				if t != om.ground {
					om.wave[x][om.FM.Y-1][t] = false
				}
			}
			om.changes[x][om.FM.Y-1] = true

			for y := 0; y < om.FM.Y-1; y++ {
				om.wave[x][y][om.ground] = false
				om.changes[x][y] = true
			}

			for om.Propagate() {
			}
		}
	}
}

func power(a, n int) int {
	product := 1
	for i := 0; i < n; i++ {
		product *= a
	}
	return product
}

func openBMP(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)

	return img, err
}
