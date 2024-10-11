package pctk

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ViewportEventType is the type of event that happened in the viewport.
type ViewportEventType int

const (
	ViewportEventLeftClick ViewportEventType = iota
	ViewportEventRightClick
	ViewportEventMouseEnter
	ViewportEventMouseLeave
)

// ViewportEvent is an event that happened in the viewport.
type ViewportEvent struct {
	Type ViewportEventType
	Pos  Position
	Item RoomItem
}

// ViewportEventHandler is a function that handles viewport events.
type ViewportEventHandler func(ViewportEvent)

// Viewport is the element of the screen that shows the game world. It is the counterpart of the
// control pane.
type Viewport struct {
	Room *Room

	camera    Camera
	camtarget int
	follow    *Actor
	hover     RoomItem
	dialogs   []Dialog
	handlers  []ViewportEventHandler
}

// Init initializes the viewport with the given camera.
func (v *Viewport) Init(cam Camera) {
	v.camera = cam
}

// SubscribeEventHandler subscribes the given handler to the viewport events.
func (v *Viewport) SubscribeEventHandler(handler ViewportEventHandler) {
	v.handlers = append(v.handlers, handler)
}

// CameraFollowActor makes the camera follow the given actor.
func (v *Viewport) CameraFollowActor(actor *Actor) error {
	if actor.Room != v.Room {
		return fmt.Errorf("Actor %s is not in the room", actor.Name)
	}
	v.follow = actor
	return nil
}

// CameraMoveTo moves the camera to the given position.
func (v *Viewport) CameraMoveTo(pos int) {
	v.camera = v.camera.WithTarget(NewPos(pos, 0))
	v.follow = nil
}

// ProcessFrame processes the frame in the viewport.
func (v *Viewport) ProcessFrame(f *Frame) {
	if v.Room == nil {
		return
	}
	f.WithCamera(&v.camera, func(f *Frame) {
		v.processFrameRoom(f)
		v.processFrameDialogs(f)
		v.processEvents(f)
		v.updateCamera()
		if f.DebugEnabled && f.MouseIn(v.Room.Rect()) {
			v.drawMouseCoords(f.MouseRelativePos())
		}
	})
}

// BeginDialog will prepare the dialog to be shown.
func (a *Viewport) BeginDialog(dialog *Dialog) {
	dialog.Begin()
	if actor := dialog.Actor(); actor != nil {
		a.clearDialogsFrom(actor)
		actor.dialog = dialog
	}
	a.dialogs = append(a.dialogs, *dialog)
}

func (a *Viewport) clearDialogsFrom(actor *Actor) {
	dialogs := make([]Dialog, 0, len(a.dialogs))
	for _, d := range a.dialogs {
		if d.done == nil || d.Actor() != actor {
			dialogs = append(dialogs, d)
		}
	}
	a.dialogs = dialogs
}

func (v *Viewport) processFrameRoom(f *Frame) {
	if v.Room != nil {
		v.Room.Draw(f)
	}
}

func (v *Viewport) processFrameDialogs(f *Frame) {
	dialogs := make([]Dialog, 0, len(v.dialogs))
	for _, d := range v.dialogs {
		d.Draw()
		if !d.Done().IsCompleted() {
			dialogs = append(dialogs, d)
		}
	}
	v.dialogs = dialogs
}

func (v *Viewport) processEvents(f *Frame) {
	if v.Room == nil || !f.MouseIn(v.Room.Rect()) {
		return
	}

	mpos := f.MouseRelativePos()
	item := v.Room.ItemAt(mpos)
	if f.Mouse.LeftClick() {
		v.processEvent(ViewportEvent{
			Type: ViewportEventLeftClick,
			Pos:  mpos,
			Item: item,
		})
	}
	if f.Mouse.RightClick() {
		v.processEvent(ViewportEvent{
			Type: ViewportEventRightClick,
			Pos:  mpos,
			Item: item,
		})
	}
	if item != v.hover {
		if v.hover != nil {
			v.processEvent(ViewportEvent{
				Type: ViewportEventMouseLeave,
				Pos:  mpos,
				Item: v.hover,
			})
		}
		if item != nil {
			v.processEvent(ViewportEvent{
				Type: ViewportEventMouseEnter,
				Pos:  mpos,
				Item: item,
			})
		}
		v.hover = item
	}
}

func (v *Viewport) processEvent(ev ViewportEvent) {
	for _, handler := range v.handlers {
		handler(ev)
	}
}

func (r *Viewport) updateCamera() {
	if r.Room == nil {
		return
	}
	if r.follow != nil {
		r.camtarget = int(r.follow.pos.X) - ScreenWidth/2
	}
	pos := r.camera.Target().X
	if pos != r.camtarget {
		if pos < r.camtarget {
			pos += RoomCameraSpeed
			if pos > r.camtarget {
				pos = r.camtarget
			}
		} else {
			pos -= RoomCameraSpeed
			if pos < r.camtarget {
				pos = r.camtarget
			}
		}
		if pos < 0 {
			pos = 0
		}

		if pos > r.Room.Rect().Size.W-ScreenWidth {
			pos = r.Room.Rect().Size.W - ScreenWidth
		}
	}
	r.camera = r.camera.WithTarget(NewPos(pos, 0))
}

func (m *Viewport) drawMouseCoords(pos Position) {
	cursorText := fmt.Sprintf("(%d,%d)", int32(pos.X), int32(pos.Y))
	textWidth := int(rl.MeasureText(cursorText, 1))
	fontSize := 10
	cursorCoordsX := int32(pos.X - textWidth/2)
	cursorCoordsY := int32(pos.Y + fontSize)

	if pos.X < textWidth {
		cursorCoordsX = int32(pos.X + textWidth/2)
	} else if pos.X > ScreenWidth-textWidth {
		cursorCoordsX = int32(pos.X - textWidth)
	}

	if pos.Y > ScreenHeight-fontSize*2 {
		cursorCoordsY = int32(pos.Y - (fontSize * 2))
	}

	rl.DrawText(cursorText, cursorCoordsX, cursorCoordsY, int32(fontSize), White)
}
