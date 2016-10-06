package bohm

import "testing"

const testSeed = 0

func BenchmarkTiledCreate(b *testing.B) {
	summer := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewTiled("Summer", "", 15, 15, false, false)
			_ = m
		}
	}
	circuit := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewTiled("Circuit", "Turnless", 34, 34, true, false)
			_ = m
		}
	}

	b.Run("Summer", summer)
	b.Run("Circuit", circuit)
}

func BenchmarkTiledRun(b *testing.B) {
	s := NewTiled("Summer", "", 15, 15, false, false)
	c := NewTiled("Circuit", "Turnless", 34, 34, true, false)

	summer := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if !s.Run(testSeed, 0) {
				b.Error("CONTRADICTION")
			}
		}
	}
	circuit := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if !c.Run(testSeed, 0) {
				b.Error("CONTRADICTION")
			}
		}
	}

	b.Run("Summer", summer)
	b.Run("Circuit", circuit)
}

func BenchmarkTiledGraphics(b *testing.B) {
	s := NewTiled("Summer", "", 15, 15, false, false)
	if !s.Run(testSeed, 0) {
		b.Error("Summer: CONTRADICTION")
	}

	c := NewTiled("Circuit", "Turnless", 34, 34, true, false)
	if !c.Run(testSeed, 0) {
		b.Error("Circuit: CONTRADICTION")
	}

	summer := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			img, err := s.Graphics()
			if err != nil {
				b.Fatal(err)
			}
			_ = img
		}
	}
	circuit := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			img, err := c.Graphics()
			if err != nil {
				b.Fatal(err)
			}
			_ = img
		}
	}

	b.Run("Summer", summer)
	b.Run("Circuit", circuit)
}

func BenchmarkTiledGraphicsNoCache(b *testing.B) {
	TextureCache = false
	b.Run("Graphics", BenchmarkTiledGraphics)
	TextureCache = true
}

func BenchmarkTiledFull(b *testing.B) {
	summer := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := NewTiled("Summer", "", 15, 15, false, false)

			if !s.Run(testSeed, 0) {
				b.Error("Summer: CONTRADICTION")
			}

			img, err := s.Graphics()
			if err != nil {
				b.Fatal(err)
			}
			_ = img
		}
	}
	circuit := func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := NewTiled("Circuit", "Turnless", 34, 34, true, false)

			if !c.Run(testSeed, 0) {
				b.Error("Circuit: CONTRADICTION")
			}

			img, err := c.Graphics()
			if err != nil {
				b.Fatal(err)
			}
			_ = img
		}
	}

	b.Run("Summer", summer)
	b.Run("Circuit", circuit)
}
