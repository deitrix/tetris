package sprite

import (
	"fmt"

	"github.com/deitrix/tetris/res"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	return nil
}
