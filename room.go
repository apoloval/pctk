package pctk

import (
	"errors"
	"log"
	"slices"
)

// RoomCameraSpeed is the speed at which the camera moves in the room.
const RoomCameraSpeed = 2

// Room represents a room in the game.
type Room struct {
	Background ResourceRef // The reference to the background image

	actors     []*Actor           // The actors in the room
	background *Image             // The background image of the room
	callrecv   ScriptCallReceiver // The call receiver for the room
	objects    []*Object          // The objects declared in the room
	script     *Script            // The script where this room is defined. Used to call the room functions.
	wbmatrix   *WalkBoxMatrix     // The wbmatrix defines the walkable areas within the room and their adjacency.
}

// NewRoom creates a new room with the given background image.
func NewRoom(bg *Image) *Room {
	if bg != nil && (bg.Width() < ScreenWidth || bg.Height() < ViewportHeight) {
		log.Fatal("Background image is too small")
	}
	return &Room{
		background: bg,
	}
}

// Rect returns the rectangle of the room.
func (r *Room) Rect() Rectangle {
	if r.background == nil {
		return Rectangle{}
	}
	return NewRect(0, 0, int(r.background.Width()), int(r.background.Height()))
}

// DeclareObject declares an object in the room.
func (r *Room) DeclareObject(obj *Object) {
	obj.Room = r
	r.objects = append(r.objects, obj)
}

// DeclareWalkBoxMatrix declares walk box matrix for the room.
func (r *Room) DeclareWalkBoxMatrix(walkboxes []*WalkBox) {
	r.wbmatrix = NewWalkBoxMatrix(walkboxes)
}

// Draw renders the room in the viewport.
func (r *Room) Draw(frame *Frame) {
	r.background.Draw(NewPos(0, 0), White)
	items := make([]RoomItem, 0, len(r.actors)+len(r.objects))
	for _, actor := range r.actors {
		items = append(items, actor)
	}
	for _, obj := range r.objects {
		items = append(items, obj)
	}
	slices.SortFunc(items, func(a, b RoomItem) int {
		return a.ItemPosition().Y - b.ItemPosition().Y
	})
	for _, item := range items {
		item.Draw(frame)
	}

	if frame.DebugEnabled && r.wbmatrix != nil {
		r.wbmatrix.Draw()
	}
}

// Load the room resources.
func (r *Room) Load(res ResourceLoader) {
	if r.background == nil {
		r.background = res.LoadImage(r.Background)
		if r.background == nil {
			log.Fatalf("Background image not found: %s", r.Background)
		}
	}
	for _, obj := range r.objects {
		obj.Load(res)
	}
}

// ItemAt returns the item at the given position in the room.
func (r *Room) ItemAt(pos Position) RoomItem {
	if r == nil {
		return nil
	}
	for _, actor := range r.actors {
		if !actor.IsEgo() && actor.Hotspot().Contains(pos) {
			return actor
		}
	}
	for _, obj := range r.objects {
		if obj.IsVisible() && obj.Hotspot.Contains(pos) {
			return obj
		}
	}
	return nil
}

// ObjectByID returns the object with the given ID, or nil if not found.
func (r *Room) ObjectByID(id string) *Object {
	for _, obj := range r.objects {
		if obj.Name == id {
			return obj
		}
	}
	return nil
}

// PutActor puts an actor in the room.
func (r *Room) PutActor(actor *Actor) {
	actor.Room = r
	for _, act := range r.actors {
		if act == actor {
			return
		}
	}
	r.actors = append(r.actors, actor)
}

// GetScaleAtPosition returns the scale at the specified position within the room.
func (r *Room) GetScaleAtPosition(p Position) float32 {
	w, _ := r.wbmatrix.walkBoxAt(p.ToPosf())
	if w != nil {
		return w.Scale()
	}
	return DefaultScale
}

// RoomItem is an item from a room that can be represented in the viewport.
type RoomItem interface {
	CallReceiver() ScriptCallReceiver
	Caption() string
	Draw(*Frame)
	ItemClass() ObjectClass
	ItemOwner() *Actor
	ItemPosition() Position
	ItemUsePosition() (Position, Direction)
}

// DeclareRoom declares a new room in the application.
func (a *App) DeclareRoom(room *Room) error {
	for _, r := range a.rooms {
		if r == room {
			return errors.New("Room already declared")
		}
	}
	a.rooms = append(a.rooms, room)
	return nil
}

// StartRoom starts the given room in the application.
func (a *App) StartRoom(room *Room) Future {
	var job Future
	for _, r := range a.rooms {
		if r == room {
			if prev := a.viewport.Room; prev != nil {
				job = RecoverWithValue(
					prev.script.CallMethod(prev.callrecv, "exit", nil),
					func(err error) any {
						log.Printf("Failed to call room exit function: %v", err)
						return nil
					},
				)
			}
			a.viewport.Room = room
			room.Load(a.res)
			job = Continue(job, func(a any) Future {
				return RecoverWithValue(
					room.script.CallMethod(room.callrecv, "enter", nil),
					func(err error) any {
						log.Printf("Failed to call room enter function: %v", err)
						return nil
					},
				)
			})
			return job
		}
	}
	prom := NewPromise()
	prom.CompleteWithErrorf("Room not declared")
	return prom
}
