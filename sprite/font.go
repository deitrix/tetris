package sprite

import (
	"fmt"

	"github.com/deitrix/tetris/res"
	"golang.org/x/image/font/opentype"
)

var (
	Roboto    *opentype.Font
	Monospace *opentype.Font
)

var fontMap = map[string]**opentype.Font{
	"roboto":    &Roboto,
	"monospace": &Monospace,
}

func loadFonts() (err error) {
	for name, f := range fontMap {
		bs, err := res.FS.ReadFile(fmt.Sprintf("font/%s.ttf", name))
		if err != nil {
			return fmt.Errorf("loading roboto.ttf: %w", err)
		}
		*f, err = opentype.Parse(bs)
		if err != nil {
			return fmt.Errorf("parsing roboto.ttf: %w", err)
		}
	}
	return nil
}
