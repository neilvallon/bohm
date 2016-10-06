package bohm

import (
	"image"
	"os"
)

type textureDef struct {
	name              string
	size, cardinality int
}

func (def textureDef) Open() (*texture, error) {
	if t, ok := textureCache[def.name]; ok {
		return t, nil
	}

	f, err := os.Open(def.name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	t := &texture{cardCache: [4][]uint8{img.(*image.RGBA).Pix}}
	t.cardCache[0] = img.(*image.RGBA).Pix
	t.Size = def.size

	textureCache[def.name] = t
	return t, nil
}

type texture struct {
	Size      int
	cardCache [4][]uint8
}

func (t *texture) CarTile(cardinality int) []byte {
	if t.cardCache[cardinality] == nil {
		t.buildCache(cardinality)
	}
	return t.cardCache[cardinality]
}

func (t *texture) buildCache(cardinality int) {
	switch cardinality {
	case 1:
		t.cardCache[1] = make([]uint8, len(t.cardCache[0]))
		for y := 0; y < t.Size; y++ {
			row := y * t.Size
			for x := 0; x < t.Size; x++ {
				to := t.cardCache[1][(row+x)*4:]
				from := t.cardCache[0][(x*t.Size+(t.Size-1-y))*4:]
				to[0], to[1], to[2], to[3] = from[0], from[1], from[2], from[3]
			}
		}

	case 2:
		t.cardCache[2] = make([]uint8, len(t.cardCache[0]))
		for y := 0; y < t.Size; y++ {
			row := y * t.Size
			for x := 0; x < t.Size; x++ {
				to := t.cardCache[2][(row+x)*4:]
				from := t.cardCache[0][((t.Size-1-y)*t.Size+(t.Size-1-x))*4:]
				to[0], to[1], to[2], to[3] = from[0], from[1], from[2], from[3]
			}
		}

	case 3:
		t.cardCache[3] = make([]uint8, len(t.cardCache[0]))
		for y := 0; y < t.Size; y++ {
			row := y * t.Size
			for x := 0; x < t.Size; x++ {
				to := t.cardCache[3][(row+x)*4:]
				from := t.cardCache[0][4*(t.Size*(t.Size-1-x)+y):]
				to[0], to[1], to[2], to[3] = from[0], from[1], from[2], from[3]
			}
		}
	}
}

var (
	TextureCache = true
	textureCache = make(map[string]*texture)
)

func _OpenTexture(name string) *texture {
	if t, ok := textureCache[name]; ok && TextureCache {
		return t
	}

	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	t := &texture{}
	t.cardCache[0] = img.(*image.RGBA).Pix

	textureCache[name] = t
	return t
}
