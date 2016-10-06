package bohm

import (
	"image"
	"math"
	"math/rand"
)

type Point struct {
	X, Y int
}

type Model struct {
	wave       [][][]bool
	changes    [][]bool
	stationary []float64

	distribution []float64

	random *rand.Rand

	logProb []float64

	ModelDep
}

func NewModel(dep ModelDep) Model {
	return Model{
		ModelDep: dep,
	}
}

func (m *Model) Run(seed int64, limit int) bool {
	T := len(m.stationary)

	if cap(m.logProb) == 0 {
		m.logProb = make([]float64, 0, T)
	} else {
		m.logProb = m.logProb[0:0]
	}

	for _, s := range m.stationary {
		m.logProb = append(m.logProb, math.Log(s))
	}

	logT := math.Log(float64(T))

	m.ModelDep.Clear()

	m.random = rand.New(rand.NewSource(seed))

	for l := 0; l < limit || limit == 0; l++ {
		if result := m.Observe(logT, m.logProb); result != observeNil {
			return result == observeTrue
		}
		for m.ModelDep.Propagate() {
		}
	}

	return true
}

func (m *Model) Clear() {
	for x, col := range m.wave {
		for y, row := range col {
			for t := range row {
				row[t] = true
			}
			m.changes[x][y] = false
		}
	}
}

type observeState int

const (
	observeNil observeState = iota
	observeTrue
	observeFalse
)

func (m *Model) Observe(logT float64, logProb []float64) observeState {
	var min float64 = 1E+3

	var sum, mainSum, logSum, noise, entropy float64
	argminx, argminy := -1, -1
	var amount, T int

	for x := range m.wave {
		for y := range m.wave[x] {
			T = len(m.wave[x])

			if m.ModelDep.OnBoundary(x, y) {
				continue
			}

			amount = 0
			sum = 0

			for t, on := range m.wave[x][y] {
				if on {
					amount++
					sum += m.stationary[t]
				}
			}

			if sum == 0 {
				return observeFalse
			}

			noise = 1E-6 * m.random.Float64()

			if amount == 1 {
				entropy = 0
			} else if amount == T {
				entropy = logT
			} else {
				mainSum = 0
				logSum = math.Log(sum)
				for t, on := range m.wave[x][y] {
					if on {
						mainSum += m.stationary[t] * logProb[t]
					}
				}
				entropy = logSum - mainSum/sum
			}

			if entropy > 0 && entropy+noise < min {
				min = entropy + noise
				argminx = x
				argminy = y
			}
		}
	}

	if argminx == -1 && argminy == -1 {
		return observeTrue
	}

	if cap(m.logProb) == 0 {
		m.distribution = make([]float64, 0, T)
	} else {
		m.distribution = m.distribution[0:0]
	}

	for t, on := range m.wave[argminx][argminy] {
		if on {
			m.distribution = append(m.distribution, m.stationary[t])
		} else {
			m.distribution = append(m.distribution, 0)
		}
	}

	r := randIndex(m.distribution, m.random.Float64())

	for t := range m.wave[argminx][argminy] {
		m.wave[argminx][argminy][t] = t == r
	}
	m.changes[argminx][argminy] = true

	return observeNil
}

func randIndex(a []float64, r float64) int {
	var sum float64
	for _, n := range a {
		sum += n
	}

	if sum == 0 {
		return int(r * float64(len(a)))
	}

	r *= sum

	var x float64
	for i, n := range a {
		x += n
		if r <= x {
			return i
		}
	}

	return 0
}

type ModelDep interface {
	Clear()
	Graphics() (image.Image, error)
	OnBoundary(x, y int) bool
	Propagate() bool
}
