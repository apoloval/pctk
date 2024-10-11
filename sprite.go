package pctk

import (
	"io"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SpriteSheet represents a collection of sprites arranged in a grid-shaped sheet.
type SpriteSheet struct {
	raw       *rl.Image
	tex       rl.Texture2D
	frameSize Size
}

// LoadSpriteSheetFromFile loads a sprite sheet from a image file.
func LoadSpriteSheetFromFile(path string, frameSize Size) *SpriteSheet {
	return &SpriteSheet{
		raw:       rl.LoadImage(path),
		frameSize: frameSize,
	}
}

// Release releases the resources used by the sprite sheet.
func (s *SpriteSheet) Release() {
	rl.UnloadTexture(s.tex)
}

// DrawSprite draws a sprite from the sprite sheet at the given position.
func (s *SpriteSheet) DrawSprite(col, row uint, pos Position, scale float32, flip bool) {
	src := Rectangle{
		Pos: Position{
			int(s.frameSize.W) * int(col),
			int(s.frameSize.H) * int(row),
		},
		Size: s.frameSize,
	}
	dest := Rectangle{
		Pos: pos,
		Size: Size{
			int(float32(s.frameSize.W) /** scale*/),
			int(float32(s.frameSize.H) /** scale*/),
		},
	}

	// TODO: calculate origin based on scale
	origin := Position{
		X: 0,
		Y: 0,
	}

	if flip {
		src.Size = src.Size.FlipH()
	}

	rl.DrawTexturePro(
		s.texture(),
		src.toRaylib(),
		dest.toRaylib(),
		origin.toRaylib(),
		0.0,
		rl.White,
	)
}

// BinaryEncode encodes the sprite sheet to a binary format. The encoded format is:
// - uint16: the width of each sprite.
// - uint16: the height of each sprite.
// - uint32: the length of the image bytes.
// - []byte: the image bytes in PNG format.
func (s *SpriteSheet) BinaryEncode(w io.Writer) (int, error) {
	bytes := rl.ExportImageToMemory(*s.raw, ".png")
	return BinaryEncode(w, uint16(s.frameSize.W), uint16(s.frameSize.H), uint32(len(bytes)), bytes)
}

// BinaryDecode decodes the sprite sheet from a binary format. See SpriteSheet.BinaryEncode for the
// format.
func (s *SpriteSheet) BinaryDecode(r io.Reader) error {
	var w, h uint16
	var size uint32
	if err := BinaryDecode(r, &w, &h, &size); err != nil {
		return err
	}
	bytes := make([]byte, size)
	if err := BinaryDecode(r, bytes); err != nil {
		return err
	}
	s.frameSize = Size{int(w), int(h)}
	s.raw = rl.LoadImageFromMemory(".png", bytes, int32(len(bytes)))
	return nil
}

func (s *SpriteSheet) texture() rl.Texture2D {
	if !rl.IsTextureReady(s.tex) {
		s.tex = rl.LoadTextureFromImage(s.raw)
	}
	return s.tex
}
