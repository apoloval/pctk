package pctk

import (
	"time"
)

const (
	// LettersPerSecond is the number of letters that an adult human could read easily per second.
	LettersPerSecond = 10
)

var (
	// DefaultDialogColor is the default color of a dialog.
	DefaultDialogColor = Magenta

	// DefaultDialogPosition is the default position of a dialog.
	DefaultDialogPosition = Position{X: 160, Y: 20}
)

// Dialog is a dialog that will be shown in the screen.
type Dialog struct {
	actor  *Actor
	bounds Rectangle
	text   string
	pos    Position
	color  Color
	speed  float32
	done   *Promise
}

// NewDialog creates a new dialog with the given properties.
func NewDialog(actor *Actor, text string, pos Position, color Color, speed float32) *Dialog {
	if color == Blank {
		color = DefaultDialogColor
	}
	if speed == 0 {
		speed = 1
	}
	return &Dialog{
		actor: actor,
		text:  text,
		pos:   pos,
		color: color,
		speed: speed,
	}
}

// SetBounds sets the bounds of the dialog.
func (d *Dialog) SetBounds(bounds Rectangle) {
	d.bounds = bounds
}

// Actor returns the actor that is speaking the dialog, or nil if it comes from a external voice.
func (d *Dialog) Actor() *Actor {
	return d.actor
}

// Begin the dialog. This will set the timer to complete the dialog.
func (d *Dialog) Begin() {
	duration := time.Duration(len(d.text)/LettersPerSecond) * time.Second
	if duration < 2*time.Second {
		duration = 2 * time.Second
	}
	duration /= time.Duration(d.speed)

	d.done = NewPromise()
	time.AfterFunc(duration, func() {
		d.done.Complete()
	})
}

// Done will return a future that will be completed when the dialog is done. If the dialog
// is not beginned, it will return nil.
func (d *Dialog) Done() Future {
	if d.done == nil {
		return nil
	}
	return d.done
}

// Draw will draw the dialog in the screen. It returns true if the dialog is completed.
func (d *Dialog) Draw() {
	if d.done != nil && d.done.IsCompleted() {
		return
	}
	DrawDialogText(d.text, d.pos, d.bounds, d.color)
}
