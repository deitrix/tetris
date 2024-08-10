package main

import (
	_ "embed"
	"log"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/deitrix/tetris/sprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

func RandomPiece() Piece {
	return ClonePiece(pieces[rand.Intn(len(pieces))])
}

func ClonePiece(p Piece) Piece {
	p.Mask = slices.Clone(p.Mask)
	return p
}

var pieces = []Piece{
	{
		Mask: []int{
			0, 0, 0, 0,
			0, 0, 0, 0,
			1, 1, 1, 1,
			0, 0, 0, 0,
		},
		Color:  newColorScale(0.19, 0.78, 0.94, 1),
		Width:  4,
		Height: 4,
	},
	{
		Mask: []int{
			1, 0, 0,
			1, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.35, 0.4, 0.68, 1),
		Width:  3,
		Height: 3,
	},
	{
		Mask: []int{
			0, 0, 1,
			1, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.94, 0.47, 0.13, 1),
		Width:  3,
		Height: 3,
	},
	{
		Mask: []int{
			1, 1,
			1, 1,
		},
		Color:  newColorScale(0.97, 0.83, 0.03, 1),
		Width:  2,
		Height: 2,
	},
	{
		Mask: []int{
			0, 1, 1,
			1, 1, 0,
			0, 0, 0,
		},
		Color:  newColorScale(0.26, 0.71, 0.26, 1),
		Width:  3,
		Height: 3,
	},
	{
		Mask: []int{
			0, 1, 0,
			1, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.68, 0.3, 0.61, 1),
		Width:  3,
		Height: 3,
	},
	{
		Mask: []int{
			1, 1, 0,
			0, 1, 1,
			0, 0, 0,
		},
		Color:  newColorScale(0.94, 0.13, 0.16, 1),
		Width:  3,
		Height: 3,
	},
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

type Cell struct {
	Color ebiten.ColorScale
}

type Game struct {
	// Cells holds the board state, not including the falling piece
	Cells []*Cell
	// FallingPiece is the piece currently being controlled by the player
	FallingPiece Piece
	// Queue is the next 3 pieces that will fall
	Queue [3]Piece
	// HoldPiece is the piece that the player has held for later
	HoldPiece *Piece

	// PauseFallFast is a flag that prevents the player from accidentally fast-falling the next
	// piece after the current piece has landed. This is set to true when the player fast-falls a
	// piece, and it gets cemented into the board. It is reset to false when the player releases the
	// down key.
	PauseFallFast bool
	// DidHoldPiece is a flag that prevents the player from holding a piece more than once per turn.
	DidHoldPiece bool

	// LastAutoFallTime is the time the current piece last advanced downwards
	LastAutoFallTime time.Time
	// AutoFallStep is the time between each automatic fall of the piece. This decreases as the player
	// clears lines.
	AutoFallStep time.Duration
}

func NewGame() *Game {
	g := &Game{
		Cells:            make([]*Cell, 12*21),
		LastAutoFallTime: time.Now(),
		AutoFallStep:     500 * time.Millisecond,
	}
	for i := 0; i < 3; i++ {
		g.Queue[i] = RandomPiece()
	}
	g.placeBorderCells()
	g.nextPiece()
	return g
}

var borderColour = newColorScale(0.5, 0.5, 0.5, 1)

func newColorScale(r, g, b, a float32) ebiten.ColorScale {
	var c ebiten.ColorScale
	c.Scale(r, g, b, a)
	return c
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.cementPiece()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) && !g.DidHoldPiece {
		if g.HoldPiece == nil {
			p := ClonePiece(g.FallingPiece)
			g.HoldPiece = &p
			g.nextPiece()
		} else {
			g.FallingPiece, *g.HoldPiece = *g.HoldPiece, g.FallingPiece
		}
		g.FallingPiece.X = 4
		g.FallingPiece.Y = 0
		g.HoldPiece.X = 0
		g.HoldPiece.Y = 0
		g.DidHoldPiece = true
		return nil
	}

	var isFallingFast bool
	fallStep := g.AutoFallStep
	if !g.PauseFallFast && ebiten.IsKeyPressed(ebiten.KeyDown) {
		fallStep = 25 * time.Millisecond
		isFallingFast = true
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		g.PauseFallFast = false
	}

	if time.Since(g.LastAutoFallTime) > fallStep {
		if g.canMoveDown(g.FallingPiece) {
			g.FallingPiece.Y++
		} else {
			g.cementPiece()
			if isFallingFast {
				g.PauseFallFast = true
			}
		}
		g.LastAutoFallTime = time.Now()
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.KeyPressDuration(ebiten.KeyLeft) > 10) && g.canMoveLeft() {
		g.FallingPiece.X--
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.KeyPressDuration(ebiten.KeyRight) > 10) && g.canMoveRight() {
		g.FallingPiece.X++
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) && g.canRotate() {
		g.rotate()
	}

	return nil
}

func (g *Game) canMoveLeft() bool {
	for i := range g.FallingPiece.Mask {
		if g.FallingPiece.Mask[i] == 0 {
			continue
		}
		width := int(math.Sqrt(float64(len(g.FallingPiece.Mask))))
		x := g.FallingPiece.X + i%width
		y := g.FallingPiece.Y + i/width

		if g.Cells[y*12+x-1] != nil {
			return false
		}
	}
	return true
}

func (g *Game) canMoveRight() bool {
	for i := range g.FallingPiece.Mask {
		if g.FallingPiece.Mask[i] == 0 {
			continue
		}
		width := int(math.Sqrt(float64(len(g.FallingPiece.Mask))))
		x := g.FallingPiece.X + i%width
		y := g.FallingPiece.Y + i/width

		if g.Cells[y*12+x+1] != nil {
			return false
		}
	}
	return true
}

func (g *Game) canMoveDown(p Piece) bool {
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		width := int(math.Sqrt(float64(len(p.Mask))))
		x := p.X + i%width
		y := p.Y + i/width

		if g.Cells[(y+1)*12+x] != nil {
			return false
		}
	}
	return true
}

func (g *Game) rotatedPiece() Piece {
	p := g.FallingPiece
	newMask := make([]int, len(p.Mask))
	for i := range p.Mask {
		newMask[rotateIndices[len(p.Mask)][i]] = p.Mask[i]
	}
	p.Mask = newMask
	return p
}

func (g *Game) ghostPiece() Piece {
	p := g.FallingPiece
	for {
		if !g.canMoveDown(p) {
			break
		}
		p.Y++
	}
	return p
}

func (g *Game) canRotate() bool {
	if len(g.FallingPiece.Mask) == 4 {
		return false
	}
	p := g.rotatedPiece()
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		width := int(math.Sqrt(float64(len(p.Mask))))
		x := p.X + i%width
		y := p.Y + i/width
		if g.Cells[y*12+x] != nil {
			return false
		}
	}
	return true
}

func (g *Game) rotate() {
	g.FallingPiece = g.rotatedPiece()
}

// cementPiece cements the falling piece into the board
func (g *Game) cementPiece() {
	for g.canMoveDown(g.FallingPiece) {
		g.FallingPiece.Y++
	}
	for i := range g.FallingPiece.Mask {
		if g.FallingPiece.Mask[i] == 0 {
			continue
		}
		width := int(math.Sqrt(float64(len(g.FallingPiece.Mask))))
		x := g.FallingPiece.X + i%width
		y := g.FallingPiece.Y + i/width
		g.Cells[y*12+x] = &Cell{
			Color: g.FallingPiece.Color,
		}
	}
	g.nextPiece()
	g.clearLines()
	g.DidHoldPiece = false
}

func (g *Game) nextPiece() {
	g.FallingPiece = g.Queue[0]
	for i := 0; i < 2; i++ {
		g.Queue[i] = g.Queue[i+1]
	}
	g.Queue[2] = RandomPiece()
	g.FallingPiece.X = 4
	g.FallingPiece.Y = 0
	g.LastAutoFallTime = time.Now()
}

func (g *Game) clearLines() {
	for y := 0; y < 20; y++ {
		full := true
		for x := 1; x < 11; x++ {
			if g.Cells[y*12+x] == nil {
				full = false
				break
			}
		}
		if full {
			g.removeRow(y)
		}
	}
}

func (g *Game) removeRow(row int) {
	for y := row; y > 0; y-- {
		for x := 1; x < 11; x++ {
			g.Cells[y*12+x] = g.Cells[(y-1)*12+x]
		}
	}
	if g.AutoFallStep > 25*time.Millisecond {
		g.AutoFallStep -= 5 * time.Millisecond
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	for x := 0; x < 12; x++ {
		for y := 0; y < 21; y++ {
			i := y*12 + x
			if g.Cells[i] == nil {
				continue
			}
			drawImage(screen, sprite.Cell, 400+x*64, y*64, 64, 64, &ebiten.DrawImageOptions{
				ColorScale: g.Cells[i].Color,
			})
		}
	}
	g.renderPiece(screen, sprite.Ghost, g.ghostPiece(), 400, 0)
	g.renderPiece(screen, sprite.Cell, g.FallingPiece, 400, 0)
	for i, p := range g.Queue {
		p = p.TrimSpace()
		xoff := 1400 - p.Width*64/2
		yoff := 200 + i*200 - p.Height*64/2
		g.renderPiece(screen, sprite.Cell, p.TrimSpace(), xoff, yoff)
	}
	if p := g.HoldPiece; p != nil {
		xoff := 200 - p.Width*64/2
		yoff := 200 - p.Height*64/2
		g.renderPiece(screen, sprite.Cell, *p, xoff, yoff)
	}
}

func (g *Game) renderPiece(screen, sprite *ebiten.Image, p Piece, xoff, yoff int) {
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		x := i % p.Width
		y := i / p.Width
		drawImage(screen, sprite, (p.X+x)*64+xoff, (p.Y+y)*64+yoff, 64, 64, &ebiten.DrawImageOptions{
			ColorScale: p.Color,
		})
	}
}

func drawImage(screen *ebiten.Image, img *ebiten.Image, x, y, width, height int, op *ebiten.DrawImageOptions) {
	if op == nil {
		op = &ebiten.DrawImageOptions{}
	}
	op.GeoM.Scale(float64(width)/float64(img.Bounds().Dx()), float64(height)/float64(img.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return 12*64 + 800, 21 * 64
}

func (g *Game) placeBorderCells() {
	for x := 0; x < 12; x++ {
		for y := 0; y < 21; y++ {
			if x == 0 || x == 11 || y == 20 {
				i := y*12 + x
				g.Cells[i] = &Cell{
					Color: borderColour,
				}
			}
		}
	}
}

func main() {
	log.SetFlags(0)
	if err := sprite.Load(); err != nil {
		log.Fatalf("failed to load sprites: %v", err)
	}

	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(1920, 1080)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatalf("failed to run game: %v", err)
	}
}
