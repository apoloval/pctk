package pctk

import (
	"errors"
	"fmt"
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
	callbacks  []*ScriptCallback  // The callbacks declared in the room
	objects    map[string]*Object // The objects declared in the room
	wbmatrix   *WalkBoxMatrix     // The wbmatrix defines the walkable areas within the room and their adjacency.
}

// NewRoom creates a new room ready to be used.
func NewRoom() *Room {
	return &Room{
		objects: make(map[string]*Object),
	}
}

// Rect returns the rectangle of the room.
func (r *Room) Rect() Rectangle {
	if r.background == nil {
		return Rectangle{}
	}
	return NewRect(0, 0, int(r.background.Width()), int(r.background.Height()))
}

// DeclareCallback declares a callback in the room.
func (r *Room) DeclareCallback(cb *ScriptCallback) error {
	for _, c := range r.callbacks {
		if c.Name == cb.Name {
			return fmt.Errorf("callback '%s' already declared", cb.Name)
		}
	}
	r.callbacks = append(r.callbacks, cb)
	return nil
}

// DeclareObject declares an object in the room.
func (r *Room) DeclareObject(tag string, obj *Object) {
	if _, found := r.objects[tag]; found {
		log.Fatalf("Object already declared: %s", tag)
	}
	obj.Room = r
	r.objects[tag] = obj
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

// FindCallback returns the callback with the given name, or nil if not found.
func (r *Room) FindCallback(name string) *ScriptCallback {
	for _, cb := range r.callbacks {
		if cb.Name == name {
			return cb
		}
	}
	return nil
}

// GetScriptField returns the script field with the given name, or nil if not found.
func (r *Room) GetScriptField(name string) *ScriptEntityValue {
	for tag, obj := range r.objects {
		if tag == name {
			return &ScriptEntityValue{
				Type:     ScriptEntityObject,
				UserData: obj,
			}
		}
	}
	for _, wb := range r.wbmatrix.walkBoxes {
		if wb.walkBoxID == name {
			return &ScriptEntityValue{
				Type:     ScriptEntityWalkBox,
				UserData: wb,
			}
		}
	}
	return nil
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

// RoomItem is an item from a room that can be represented in the viewport.
type RoomItem interface {
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
	for _, r := range a.rooms {
		if r == room {
			room.Load(a.res)
			return a.viewport.ActivateRoom(room)
		}
	}
	return AlreadyFailed(errors.New("Room not declared"))
}
