package pctk

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Mouse is a custom mouse cursor.
type Mouse struct {
	Enabled bool

	col Color
	tx  rl.Texture2D
}

// NewMouseCursor creates a new mouse cursor.
func NewMouseCursor() *Mouse {
	return &Mouse{
		col: rl.NewColor(0xAA, 0xAA, 0xAA, 0xFF),
		tx: rl.LoadTextureFromImage(
			rl.NewImage(mouseCursorData(), 15, 15, 1, rl.UncompressedR8g8b8a8),
		),
	}
}

// Draw renders the mouse cursor at the current position.
func (m *Mouse) Draw(frame *Frame) {
	pos := m.PositionRelative(frame.Camera)
	if m.Enabled && rl.IsCursorOnScreen() {
		rl.DrawTexture(m.tx, int32(pos.X-7), int32(pos.Y-7), m.col)
		m.col.R = max(0xAA, m.col.R+6)
		m.col.G = max(0xAA, m.col.G+6)
		m.col.B = max(0xAA, m.col.B+6)
	}
}

// LeftClick returns true if the left mouse button is pressed.
func (m *Mouse) LeftClick() bool {
	return rl.IsMouseButtonPressed(rl.MouseLeftButton)
}

// RightClick returns true if the right mouse button is pressed.
func (m *Mouse) RightClick() bool {
	return rl.IsMouseButtonPressed(rl.MouseRightButton)
}

// OnScreen returns true if the mouse is on the screen.
func (m *Mouse) OnScreen() bool {
	return rl.IsCursorOnScreen()
}

// PositionAbsolute returns the absolute current mouse position (in terms of screen native
// resolution).
func (m *Mouse) PositionAbsolute() Position {
	if !m.Enabled {
		return Position{-1, -1}
	}
	return positionFromRaylib(rl.GetMousePosition())
}

// PositionRelative returns the current mouse position relative to the camera.
func (m *Mouse) PositionRelative(cam *Camera) Position {
	return cam.ScreenToWorldPosition(m.PositionAbsolute())
}

func mouseCursorData() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 15, 15))
	for i := 0; i <= 5; i++ {
		img.Set(i, 7, White)
		img.Set(7, i, White)
	}
	for i := 9; i <= 15; i++ {
		img.Set(i, 7, White)
		img.Set(7, i, White)
	}
	return img.Pix
}
