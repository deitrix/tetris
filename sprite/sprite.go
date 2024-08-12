package sprite

import (
	"fmt"

	"github.com/deitrix/tetris/res"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var Cell, Ghost *ebiten.Image

var spriteMap = map[string]**ebiten.Image{
	"cell":  &Cell,
	"ghost": &Ghost,
}

func Load() (err error) {
	for name, img := range spriteMap {
		*img, _, err = ebitenutil.NewImageFromFileSystem(res.FS, fmt.Sprintf("sprite/%s.png", name))
		if err != nil {
			return fmt.Errorf("loading %s: %w", name, err)
		}
	}
	robotoBS, err := res.FS.ReadFile("roboto.ttf")
	if err != nil {
		return fmt.Errorf("loading roboto.ttf: %w", err)
	}
	robotoFont, err := opentype.Parse(robotoBS)
	if err != nil {
		return fmt.Errorf("parsing roboto.ttf: %w", err)
	}
	Roboto, err = opentype.NewFace(robotoFont, &opentype.FaceOptions{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return fmt.Errorf("parsing roboto.ttf: %w", err)
	}
	return nil
}
