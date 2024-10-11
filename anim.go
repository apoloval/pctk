package pctk

import (
	"io"
	"time"
)

const (
	DefaultAnimationDelay = 100 * time.Millisecond
	DefaultScale          = 1.0
)

// Animation represents a sequence of images that can be played.
type Animation struct {
	frames []animationFrame
	flip   bool

	currentFrame int
	lastFrame    time.Time
}

// NewAnimation creates a new animation.
func NewAnimation() *Animation {
	return &Animation{}
}

// AddFrames adds a sequence of frames to the animation. The frames are located at the row-th row
// of the sprite sheet. The delay is the time to wait before moving to the next frame. The sequence
// is the columns of the sheet row that will be repeated in the animation in the order they are
// given.
func (a *Animation) AddFrames(delay time.Duration, row int, sequence ...int) *Animation {
	for _, col := range sequence {
		a.frames = append(a.frames, animationFrame{uint(col), uint(row), delay})
	}
	return a
}

// Flip sets the flip flag for the animation.
func (a *Animation) Flip(flip bool) *Animation {
	a.flip = flip
	return a
}

// BinaryEncode encodes the animation to a binary format. The format is as follows:
// - byte: the flip flag.
// - uint32: the number of frames.
// - for each frame:
//   - byte: the sprite column.
//   - byte: the sprite row.
//   - uint64: the delay.
func (a *Animation) BinaryEncode(w io.Writer) (n int, err error) {
	n, err = BinaryEncode(w, a.flip, uint32(len(a.frames)))
	for _, frame := range a.frames {
		nn, err := BinaryEncode(w, byte(frame.col), byte(frame.row), uint64(frame.delay))
		n += nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// BinaryDecode decodes the animation from a binary format. See BinaryEncode for the format.
func (a *Animation) BinaryDecode(r io.Reader) error {
	var count uint32
	if err := BinaryDecode(r, &a.flip, &count); err != nil {
		return err
	}
	a.frames = make([]animationFrame, count)
	for i := uint32(0); i < count; i++ {
		var col, row byte
		var delay uint64
		if err := BinaryDecode(r, &col, &row, &delay); err != nil {
			return err
		}
		a.frames[i] = animationFrame{
			col:   uint(col),
			row:   uint(row),
			delay: time.Duration(delay),
		}
	}
	return nil
}

// Draw renders the animation in the viewport.
func (a *Animation) Draw(sprites *SpriteSheet, pos Position, scale float32) {
	if a == nil {
		return
	}
	if a.frames[a.currentFrame].delay < time.Since(a.lastFrame) {
		a.lastFrame = time.Now()
		a.currentFrame++
		if a.currentFrame >= len(a.frames) {
			a.currentFrame = 0
		}
	}

	sprites.DrawSprite(
		a.frames[a.currentFrame].col,
		a.frames[a.currentFrame].row,
		pos,
		scale,
		a.flip,
	)
}

type animationFrame struct {
	col, row uint
	delay    time.Duration
}
