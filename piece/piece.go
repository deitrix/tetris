package piece

import (
	"math/rand"
	"slices"

	"github.com/deitrix/tetris/cell"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	I = Piece{
		Mask: []int{
			0, 0, 0, 0,
			0, 0, 0, 0,
			1, 1, 1, 1,
			0, 0, 0, 0,
		},
		Tint:   cell.Cyan,
		Width:  4,
		Height: 4,
	}

	J = Piece{
		Mask: []int{
			1, 0, 0,
			1, 1, 1,
			0, 0, 0,
		},
		Tint:   cell.Blue,
		Width:  3,
		Height: 3,
	}

	L = Piece{
		Mask: []int{
			0, 0, 1,
			1, 1, 1,
			0, 0, 0,
		},
		Tint:   cell.Orange,
		Width:  3,
		Height: 3,
	}

	O = Piece{
		Mask: []int{
			1, 1,
			1, 1,
		},
		Tint:   cell.Yellow,
		Width:  2,
		Height: 2,
	}

	S = Piece{
		Mask: []int{
			0, 1, 1,
			1, 1, 0,
			0, 0, 0,
		},
		Tint:   cell.Green,
		Width:  3,
		Height: 3,
	}

	T = Piece{
		Mask: []int{
			0, 1, 0,
			1, 1, 1,
			0, 0, 0,
		},
		Tint:   cell.Purple,
		Width:  3,
		Height: 3,
	}

	Z = Piece{
		Mask: []int{
			1, 1, 0,
			0, 1, 1,
			0, 0, 0,
		},
		Tint:   cell.Red,
		Width:  3,
		Height: 3,
	}
)

type Piece struct {
	Mask          []int
	Tint          cell.Tint
	Opacity       int
	Width, Height int
	X, Y          int
	Orientation   int
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
		Mask:    newMask,
		Tint:    p.Tint,
		Opacity: p.Opacity,
		X:       p.X,
		Y:       p.Y,
		Width:   newWidth,
		Height:  maxY - minY + 1,
	}
}

func (p Piece) Clone() Piece {
	p.Mask = slices.Clone(p.Mask)
	return p
}

func (p *Piece) Rotate() {
	if p.Width < 3 {
		return
	}
	newMask := make([]int, len(p.Mask))
	for i := range p.Mask {
		newMask[rotateIndices[len(p.Mask)][i]] = p.Mask[i]
	}
	p.Mask = newMask
	p.Orientation = (p.Orientation + 1) % 4
}

func (p *Piece) ResetRotation() {
	for p.Orientation != 0 {
		p.Rotate()
	}
}

var allPieces = []Piece{I, J, L, O, S, T, Z}

func Rand() Piece {
	p := allPieces[rand.Intn(len(allPieces))].Clone()
	p.Opacity = 255
	return p
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
