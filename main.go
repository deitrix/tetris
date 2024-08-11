package main

import (
	_ "embed"
	"log"
	"time"

	"github.com/deitrix/tetris/cell"
	"github.com/deitrix/tetris/piece"
	"github.com/deitrix/tetris/sprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	// cellSize is the size of each cell in pixels
	cellSize = 64
	// boardWidth is the width of the board, not including the walls
	boardWidth = 20
	// boardHeight is the height of the board, not including the floor
	boardHeight = 50
	// floorThickness is the thickness of the floor at the bottom of the board. Must be at least 1.
	floorThickness = 5
	// wallThickness is the thickness of the walls on the left and right sides of the board. Must be
	// at least 1.
	wallThickness = 2
	// queueSize is the number of pieces that are shown in the queue
	queueSize = 10
	// minAutoFallStep is the minimum time between each automatic fall of the piece
	minAutoFallStep = 25 * time.Millisecond
	// initialAutoFallStep is the time between each automatic fall of the piece at the start of the
	// game
	initialAutoFallStep = 500 * time.Millisecond
	// autoFallStepDecrement is the amount of time that the auto-fall step decreases by each time a
	// line is cleared
	autoFallStepDecrement = 5 * time.Millisecond
)

const (
	boardWidthWithWalls  = boardWidth + wallThickness*2
	boardHeightWithFloor = boardHeight + floorThickness
	cellCount            = boardWidthWithWalls * boardHeightWithFloor
)

type Cell struct {
	Tint cell.Tint
}

type Game struct {
	// Cells holds the board state, not including the falling piece
	Cells []*Cell
	// Queue is the next 3 pieces that will fall
	Queue [queueSize]piece.Piece

	// DidHoldPiece is a flag that prevents the player from holding a piece more than once per turn.
	DidHoldPiece bool
	// HoldPiece is the piece that the player has held for later
	HoldPiece *piece.Piece

	// FallingPiece is the piece currently being controlled by the player
	FallingPiece piece.Piece
	// FastFalling is a flag that indicates whether the player is currently fast-falling the piece
	FastFalling bool
	// FastFallingPaused is a flag that prevents the player from accidentally fast-falling the next
	// piece after the current piece has landed. This is set to true when the player fast-falls a
	// piece, and it gets committed into the board. It is reset to false when the player releases
	// the down key.
	FastFallingPaused bool

	// LastAutoFallTime is the time the current piece last advanced downwards
	LastAutoFallTime time.Time
	// AutoFallStep is the time between each automatic fall of the piece. This decreases as the player
	// clears lines.
	AutoFallStep time.Duration
}

func NewGame() *Game {
	g := &Game{
		Cells:        make([]*Cell, cellCount),
		AutoFallStep: initialAutoFallStep,
	}
	g.fillQueue()
	g.placeBorderCells()
	g.loadNextPiece()
	return g
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.commitPiece()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) && !g.DidHoldPiece {
		g.holdPiece()
		return nil
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.KeyPressDuration(ebiten.KeyLeft) > 10) && g.canMoveLeft() {
		g.FallingPiece.X--
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.KeyPressDuration(ebiten.KeyRight) > 10) && g.canMoveRight() {
		g.FallingPiece.X++
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) && g.canRotate() {
		g.FallingPiece.Rotate()
	}

	fallStep := g.AutoFallStep
	if !g.FastFallingPaused && ebiten.IsKeyPressed(ebiten.KeyDown) {
		fallStep = minAutoFallStep // fast-fall speed
		g.FastFalling = true
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		g.FastFallingPaused = false
	}

	g.fall(fallStep)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawCells(screen)
	g.renderPiece(screen, sprite.Ghost, g.ghostPiece(), 6*cellSize, 0)
	g.renderPiece(screen, sprite.Cell, g.FallingPiece, 6*cellSize, 0)
	g.drawQueue(screen)
	g.drawHeld(screen)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	screenWidth = 6*cellSize + boardWidthWithWalls*cellSize + 6*cellSize
	screenHeight = max(boardHeightWithFloor*cellSize, cellSize+3*queueSize*cellSize)
	return
}

func (g *Game) canMoveLeft() bool {
	for i := range g.FallingPiece.Mask {
		if g.FallingPiece.Mask[i] == 0 {
			continue
		}
		x := g.FallingPiece.X + i%g.FallingPiece.Width
		y := g.FallingPiece.Y + i/g.FallingPiece.Width
		if g.Cells[y*boardWidthWithWalls+x-1] != nil {
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
		x := g.FallingPiece.X + i%g.FallingPiece.Width
		y := g.FallingPiece.Y + i/g.FallingPiece.Width
		if g.Cells[y*boardWidthWithWalls+x+1] != nil {
			return false
		}
	}
	return true
}

func (g *Game) canMoveDown(p piece.Piece) bool {
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		x := p.X + i%p.Width
		y := p.Y + i/p.Width
		if g.Cells[(y+1)*boardWidthWithWalls+x] != nil {
			return false
		}
	}
	return true
}

func (g *Game) ghostPiece() piece.Piece {
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
	p := g.FallingPiece.Clone()
	p.Rotate()
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		x := p.X + i%p.Width
		y := p.Y + i/p.Width
		if g.Cells[y*boardWidthWithWalls+x] != nil {
			return false
		}
	}
	return true
}

// commitPiece commits the currently falling piece into the board, such that it can no longer be
// moved. It also loads the next piece into the falling piece, and clears any lines that have been
// filled.
func (g *Game) commitPiece() {
	for g.canMoveDown(g.FallingPiece) {
		g.FallingPiece.Y++
	}
	for i := range g.FallingPiece.Mask {
		if g.FallingPiece.Mask[i] == 0 {
			continue
		}
		x := g.FallingPiece.X + i%g.FallingPiece.Width
		y := g.FallingPiece.Y + i/g.FallingPiece.Width
		g.Cells[y*boardWidthWithWalls+x] = &Cell{
			Tint: g.FallingPiece.Tint,
		}
	}
	g.loadNextPiece()
	g.clearLines()
	g.DidHoldPiece = false
}

func (g *Game) holdPiece() {
	if g.HoldPiece == nil {
		p := g.FallingPiece.Clone()
		g.HoldPiece = &p
		g.loadNextPiece()
	} else {
		g.FallingPiece, *g.HoldPiece = *g.HoldPiece, g.FallingPiece
	}
	g.FallingPiece.ResetRotation()
	g.FallingPiece.X = boardWidthWithWalls/2 - g.FallingPiece.Width/2
	g.FallingPiece.Y = 0
	g.HoldPiece.ResetRotation()
	g.HoldPiece.X = 0
	g.HoldPiece.Y = 0
	g.DidHoldPiece = true
}

func (g *Game) fall(step time.Duration) {
	if time.Since(g.LastAutoFallTime) > step {
		if g.canMoveDown(g.FallingPiece) {
			g.FallingPiece.Y++
		} else {
			g.commitPiece()
			if g.FastFalling {
				g.FastFallingPaused = true
			}
		}
		g.LastAutoFallTime = time.Now()
	}
}

func (g *Game) loadNextPiece() {
	g.FallingPiece = g.Queue[0]
	for i := 0; i < queueSize-1; i++ {
		g.Queue[i] = g.Queue[i+1]
	}
	g.Queue[queueSize-1] = piece.Rand()
	g.FallingPiece.X = boardWidthWithWalls/2 - g.FallingPiece.Width/2
	g.FallingPiece.Y = 0
	g.LastAutoFallTime = time.Now()
}

func (g *Game) clearLines() {
	for y := 0; y < boardHeight; y++ {
		full := true
		for x := wallThickness; x < boardWidthWithWalls-wallThickness; x++ {
			if g.Cells[y*boardWidthWithWalls+x] == nil {
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
		for x := wallThickness; x < boardWidthWithWalls-wallThickness; x++ {
			g.Cells[y*boardWidthWithWalls+x] = g.Cells[(y-1)*boardWidthWithWalls+x]
		}
	}
	if g.AutoFallStep > minAutoFallStep {
		g.AutoFallStep -= autoFallStepDecrement
	}
}

func (g *Game) drawCells(screen *ebiten.Image) {
	for x := 0; x < boardWidthWithWalls; x++ {
		for y := 0; y < boardHeightWithFloor; y++ {
			i := y*boardWidthWithWalls + x
			if g.Cells[i] == nil {
				continue
			}
			drawCell(screen, sprite.Cell, 6*cellSize+x*cellSize, y*cellSize, cellSize, cellSize, g.Cells[i].Tint)
		}
	}
}

func (g *Game) drawQueue(screen *ebiten.Image) {
	for i, p := range g.Queue {
		p = p.TrimSpace()
		xoff := (6+boardWidthWithWalls+3)*cellSize - p.Width*cellSize/2
		yoff := 2*cellSize + i*(3*cellSize) - p.Height*cellSize/2
		g.renderPiece(screen, sprite.Cell, p.TrimSpace(), xoff, yoff)
	}
}

func (g *Game) drawHeld(screen *ebiten.Image) {
	if p := g.HoldPiece; p != nil {
		p := p.TrimSpace()
		xoff := 3*cellSize - p.Width*cellSize/2
		yoff := 2*cellSize - p.Height*cellSize/2
		g.renderPiece(screen, sprite.Cell, p, xoff, yoff)
	}
}

func (g *Game) renderPiece(screen, sprite *ebiten.Image, p piece.Piece, xoff, yoff int) {
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		x := i % p.Width
		y := i / p.Width
		drawCell(screen, sprite, (p.X+x)*cellSize+xoff, (p.Y+y)*cellSize+yoff, cellSize, cellSize, p.Tint)
	}
}

func drawCell(screen *ebiten.Image, img *ebiten.Image, x, y, width, height int, tint cell.Tint) {
	var op ebiten.DrawImageOptions
	op.ColorScale.ScaleWithColor(tint.NRGBA())
	op.GeoM.Scale(float64(width)/float64(img.Bounds().Dx()), float64(height)/float64(img.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, &op)
}

func (g *Game) fillQueue() {
	for i := 0; i < queueSize; i++ {
		g.Queue[i] = piece.Rand()
	}
}

func (g *Game) placeBorderCells() {
	for x := 0; x < boardWidthWithWalls; x++ {
		for y := 0; y < boardHeightWithFloor; y++ {
			if x < wallThickness || x >= boardWidthWithWalls-wallThickness || y >= boardHeightWithFloor-floorThickness {
				i := y*boardWidthWithWalls + x
				g.Cells[i] = &Cell{Tint: cell.Wall}
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
