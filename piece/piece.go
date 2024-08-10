package piece

import (
	"math/rand"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	Border = Piece{
		Mask: []int{
			1,
		},
		Color:  newColorScale(0.5, 0.5, 0.5, 1),
		Width:  1,
		Height: 1,
	}

	I = Piece{
		Mask: []int{
			0, 0, 0, 0,
			0, 0, 0, 0,
			1, 1, 1, 1,
			0, 0, 0, 0,
		},
		Color:  newColorScale(0.19, 0.78, 0.94, 1),
		Width:  4,
		Height: 4,
	}

	J = Piece{
		Mask: []int{
			1, 0, 0,
			1, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.35, 0.4, 0.68, 1),
		Width:  3,
		Height: 3,
	}

	L = Piece{
		Mask: []int{
			0, 0, 1,
			1, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.94, 0.47, 0.13, 1),
		Width:  3,
		Height: 3,
	}

	O = Piece{
		Mask: []int{
			1, 1,
			1, 1,
		},
		Color:  newColorScale(0.97, 0.83, 0.03, 1),
		Width:  2,
		Height: 2,
	}

	S = Piece{
		Mask: []int{
			0, 1, 1,
			1, 1, 0,
			0, 0, 0,
		},
		Color:  newColorScale(0.26, 0.71, 0.26, 1),
		Width:  3,
		Height: 3,
	}

	T = Piece{
		Mask: []int{
			0, 1, 0,
			1, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.68, 0.3, 0.61, 1),
		Width:  3,
		Height: 3,
	}

	Z = Piece{
		Mask: []int{
			1, 1, 0,
			0, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.94, 0.13, 0.16, 1),
		Width:  3,
		Height: 3,
	}
)

type Piece struct {
	Mask          []int
	Color         ebiten.ColorScale
	Width, Height int
	X, Y          int
}

// TrimSpace removes empty rows and columns from the piece
func (p Piece) TrimSpace() Piece {
	minX, minY, maxX, maxY := 100, 100, 0, 0
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		x := i % p.Width
		y := i / p.Width
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	newMask := make([]int, (maxX-minX+1)*(maxY-minY+1))
	newWidth := maxX - minX + 1
	for i := range len(newMask) {
		x := i % newWidth
		y := i / newWidth
		ii := (y+minY)*p.Width + x + minX
		newMask[i] = p.Mask[ii]
	}
	return Piece{
		Mask:   newMask,
		Color:  p.Color,
		X:      p.X,
		Y:      p.Y,
		Width:  newWidth,
		Height: maxY - minY + 1,
	}
}

func (p Piece) Clone() Piece {
	p.Mask = slices.Clone(p.Mask)
	return p
}

func (p Piece) Rotate() Piece {
	rp := p.Clone()
	if rp.Width < 3 {
		return rp
	}
	newMask := make([]int, len(rp.Mask))
	for i := range rp.Mask {
		newMask[rotateIndices[len(rp.Mask)][i]] = rp.Mask[i]
	}
	rp.Mask = newMask
	return rp
}

var allPieces = []Piece{I, J, L, O, S, T, Z}

func Rand() Piece {
	return allPieces[rand.Intn(len(allPieces))].Clone()
}

var rotateIndices = map[int][]int{
	9: {
		2, 5, 8,
		1, 4, 7,
		0, 3, 6,
	},
	16: {
		3, 7, 11, 15,
		2, 6, 10, 14,
		1, 5, 9, 13,
		0, 4, 8, 12,
	},
}

func newColorScale(r, g, b, a float32) ebiten.ColorScale {
	var c ebiten.ColorScale
	c.Scale(r, g, b, a)
	return c
}
