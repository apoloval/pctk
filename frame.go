package pctk

import "log"

// Frame represents a frame of the game. It is used to pass information to the application elements
// for rendering and updating their state.
type Frame struct {
	// Num is the frame number. It is useful if some application element needs some randomness or
	// perform some actions every N frames.
	Num uint64

	// DebugEnabled is true if the debug mode is enabled, false otherwise.
	DebugEnabled bool

	// Camera is the current camera this frame is being rendered on.
	Camera *Camera

	// Mouse is the mouse input device.
	Mouse *Mouse
}

// NewFrame creates a new frame.
func NewFrame(mouse *Mouse, debug bool) *Frame {
	return &Frame{
		Mouse:        mouse,
		DebugEnabled: debug,
	}
}

// MouseIn returns true if the mouse cursor is inside the rectangle, false otherwise.
func (f Frame) MouseIn(rect Rectangle) bool {
	return rect.Contains(f.MouseRelativePos())
}

// MouseRelativePos returns the position of the mouse cursor in the frame.
func (f Frame) MouseRelativePos() Position {
	if f.Camera == nil || f.Mouse == nil {
		log.Panicf("frame: camera or mouse not set")
	}
	return f.Mouse.PositionRelative(f.Camera)
}

// WithCamera returns a new frame with the camera set to the given one.
func (f *Frame) WithCamera(c *Camera, act func(*Frame)) {
	if f.Camera != nil {
		log.Panicf("frame: camera already set")
	}

	f.Camera = c
	c.Action(func() {
		act(f)
	})
	f.Camera = nil
}
