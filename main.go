package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/deitrix/tetris/cell"
	"github.com/deitrix/tetris/piece"
	"github.com/deitrix/tetris/sprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	// cellSize is the size of each cell in pixels
	cellSize = 64
	// boardWidth is the width of the board, not including the walls
	boardWidth = 10
	// boardHeight is the height of the board, not including the floor
	boardHeight = 20
	// floorThickness is the thickness of the floor at the bottom of the board. Must be at least 1.
	floorThickness = 1
	// wallThickness is the thickness of the walls on the left and right sides of the board. Must be
	// at least 1.
	wallThickness = 1
	// queueSize is the number of pieces that are shown in the queue
	queueSize = 3
)

// fallSpeed is the number of frames between each automatic fall of the piece at each level
var fallSpeed = map[int]int{
	0:  53,
	1:  49,
	2:  45,
	3:  41,
	4:  37,
	5:  33,
	6:  28,
	7:  22,
	8:  17,
	9:  11,
	10: 10,
	11: 9,
	12: 8,
	13: 7,
	14: 6,
	16: 5,
	18: 4,
	20: 3,
	22: 2,
	29: 1,
}

var lineScore = map[int]int{
	1: 40,
	2: 100,
	3: 300,
	4: 1200,
}

func getLineScore(level, n int) int {
	return lineScore[n] * (level + 1)
}

func getFallSpeed(level int) int {
	if level > 29 {
		return 1
	}
	if speed, ok := fallSpeed[level]; ok {
		return speed
	}
	return getFallSpeed(level - 1)
}

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
	// OpacityDirection is a flag that indicates whether the opacity of the piece is currently
	// increasing or decreasing.
	OpacityDirection bool
	// TicksSinceFall is the number of ticks since the piece last fell. This is used to determine
	// when the piece should fall automatically.
	TicksSinceFall int
	// TicksSinceMove is the number of ticks since the piece last moved. This is used during the
	// "last chance" period to determine when the piece should be committed.
	TicksSinceMove int
	// LastChanceTicks is the number of ticks since the piece landed. This is used to determine when
	// the piece should be committed during the "last chance" period.
	LastChanceTicks int
	// ScreenWidth is the width of the screen in pixels
	ScreenWidth int
	// ScreenHeight is the height of the screen in pixels
	ScreenHeight int
	// Level is the current level of the game
	Level int
	// Score is the current score of the game
	Score int
	// LinesCleared is the number of lines that have been cleared in the game
	LinesCleared int
	// ShowDebug is a flag that indicates whether debug information should be shown
	ShowDebug bool
}

func NewGame() *Game {
	g := &Game{
		Cells: make([]*Cell, cellCount),
	}
	g.fillQueue()
	g.placeBorderCells()
	g.loadNextPiece()
	return g
}

func (g *Game) Reset() {
	*g = *NewGame()
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.Reset()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.ShowDebug = !g.ShowDebug
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Score += g.earlyCommitScore()
		g.commitPiece()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) && !g.DidHoldPiece {
		g.holdPiece()
		return nil
	}

	var didMove bool
	if (inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.KeyPressDuration(ebiten.KeyLeft) > 10) && g.canMoveLeft() {
		g.FallingPiece.X--
		didMove = true
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.KeyPressDuration(ebiten.KeyRight) > 10) && g.canMoveRight() {
		g.FallingPiece.X++
		didMove = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) && g.canRotate() {
		g.FallingPiece.Rotate()
		didMove = true
	}

	minTicksSinceFall := getFallSpeed(g.Level)
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		minTicksSinceFall = 1
		g.FastFalling = true
		if g.canMoveDown(g.FallingPiece) {
			didMove = true
		}
	} else {
		g.FastFalling = false
	}

	if didMove {
		g.TicksSinceMove = 0
	} else {
		g.TicksSinceMove++
	}
	g.fall(minTicksSinceFall)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawCells(screen)
	g.renderPiece(screen, sprite.Ghost, g.ghostPiece(), 6*cellSize, 0)
	g.renderPiece(screen, sprite.Cell, g.FallingPiece, 6*cellSize, 0)
	g.drawQueue(screen)
	g.drawHeld(screen)
	g.drawScore(screen)
	g.drawDebug(screen)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	g.ScreenWidth = 6*cellSize + boardWidthWithWalls*cellSize + 6*cellSize
	g.ScreenHeight = max(boardHeightWithFloor*cellSize, cellSize+3*queueSize*cellSize)
	return g.ScreenWidth, g.ScreenHeight
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

func (g *Game) earlyCommitScore() int {
	p := g.FallingPiece
	for i := 0; ; i++ {
		if !g.canMoveDown(p) {
			return i * 2
		}
		p.Y++
	}
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
	g.LastChanceTicks = 0
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

func (g *Game) fall(minTicksSinceFall int) {
	if g.canMoveDown(g.FallingPiece) {
		g.FallingPiece.Opacity = 255
		if g.TicksSinceFall >= minTicksSinceFall {
			g.FallingPiece.Y++
			if g.FastFalling {
				g.Score++
			}
			g.TicksSinceFall = 0
			g.TicksSinceMove = 0
		} else {
			g.TicksSinceFall++
		}
	} else if g.TicksSinceMove >= 30 || g.LastChanceTicks >= 120 {
		g.commitPiece()
		g.TicksSinceFall = 0
	} else {
		g.LastChanceTicks++
		if g.OpacityDirection {
			g.FallingPiece.Opacity += 8
			if g.FallingPiece.Opacity >= 255 {
				g.FallingPiece.Opacity = 255
				g.OpacityDirection = false
			}
		} else {
			g.FallingPiece.Opacity -= 8
			if g.FallingPiece.Opacity <= 128 {
				g.FallingPiece.Opacity = 128
				g.OpacityDirection = true
			}
		}
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
}

func (g *Game) clearLines() {
	lines := 0
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
			lines++
		}
	}
	if lines > 0 {
		g.Score += getLineScore(g.Level, lines)
		g.LinesCleared += lines
		g.Level = min(g.LinesCleared/10, 29)
	}
}

func (g *Game) removeRow(row int) {
	for y := row; y > 0; y-- {
		for x := wallThickness; x < boardWidthWithWalls-wallThickness; x++ {
			g.Cells[y*boardWidthWithWalls+x] = g.Cells[(y-1)*boardWidthWithWalls+x]
		}
	}
}

func (g *Game) drawCells(screen *ebiten.Image) {
	for x := 0; x < boardWidthWithWalls; x++ {
		for y := 0; y < boardHeightWithFloor; y++ {
			i := y*boardWidthWithWalls + x
			if g.Cells[i] == nil {
				continue
			}
			drawCell(screen, sprite.Cell, 6*cellSize+x*cellSize, y*cellSize, cellSize, cellSize, g.Cells[i].Tint, 255)
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

func (g *Game) drawScore(screen *ebiten.Image) {
	drawText(screen, sprite.Roboto, "Score", 48, 24, g.ScreenHeight-168, color.White)
	drawText(screen, sprite.Roboto, fmt.Sprintf("%d", g.Score), 48, 192, g.ScreenHeight-168, color.White)
	drawText(screen, sprite.Roboto, "Level", 48, 24, g.ScreenHeight-96, color.White)
	drawText(screen, sprite.Roboto, fmt.Sprintf("%d", g.Level+1), 48, 192, g.ScreenHeight-96, color.White)
	drawText(screen, sprite.Roboto, "Lines", 48, 24, g.ScreenHeight-24, color.White)
	drawText(screen, sprite.Roboto, fmt.Sprintf("%d", g.LinesCleared), 48, 192, g.ScreenHeight-24, color.White)
}

func (g *Game) drawDebug(screen *ebiten.Image) {
	if !g.ShowDebug {
		return
	}
	drawText(screen, sprite.Roboto, strings.Join([]string{
		fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()),
		fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()),
		fmt.Sprintf("Fall Speed: %d", getFallSpeed(g.Level)),
		fmt.Sprintf("Ticks Since Fall: %d", g.TicksSinceFall),
		fmt.Sprintf("Ticks Since Move: %d", g.TicksSinceMove),
		fmt.Sprintf("Last Chance Ticks: %d", g.LastChanceTicks),
		fmt.Sprintf("Fast Falling: %t", g.FastFalling),
		fmt.Sprintf("Did Hold Piece: %t", g.DidHoldPiece),
		fmt.Sprintf("Early-commit Score: %d", g.earlyCommitScore()),
	}, "\n"), 32, 24, 256, color.White)
}

var fontFaceCache = make(map[*opentype.Font]map[float64]font.Face)

func drawText(img *ebiten.Image, f *opentype.Font, t string, size float64, x, y int, c color.Color) {
	if _, ok := fontFaceCache[f]; !ok {
		fontFaceCache[f] = make(map[float64]font.Face)
	}
	if _, ok := fontFaceCache[f][size]; !ok {
		var err error
		fontFaceCache[f][size], err = opentype.NewFace(f, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingNone,
		})
		if err != nil {
			log.Fatalf("failed to create face: %v", err)
		}
	}
	text.Draw(img, t, fontFaceCache[f][size], x, y, c)
}

func (g *Game) renderPiece(screen, sprite *ebiten.Image, p piece.Piece, xoff, yoff int) {
	for i := range p.Mask {
		if p.Mask[i] == 0 {
			continue
		}
		x := i % p.Width
		y := i / p.Width
		drawCell(screen, sprite, (p.X+x)*cellSize+xoff, (p.Y+y)*cellSize+yoff, cellSize, cellSize, p.Tint, uint8(p.Opacity))
	}
}

func drawCell(screen *ebiten.Image, img *ebiten.Image, x, y, width, height int, tint cell.Tint, opacity uint8) {
	var op ebiten.DrawImageOptions
	op.ColorScale.ScaleWithColor(tint.NRGBA())
	op.ColorScale.ScaleAlpha(float32(opacity) / 255)
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
