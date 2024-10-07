package pctk

import (
	"errors"
	"log"
	"slices"
)

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

// DeclareObject declares an object in the room.
func (r *Room) DeclareObject(obj *Object) {
	obj.Room = r
	r.objects = append(r.objects, obj)
	// TODO testing purpose
	box0 := NewWalkBox("walkbox0", [4]Positionf{{5, 140}, {320, 140}, {270, 110}, {80, 110}}, 1)
	box1 := NewWalkBox("walkbox1", [4]Positionf{{80, 110}, {270, 110}, {175, 100}, {145, 100}}, 0.95)
	box2 := NewWalkBox("walkbox2", [4]Positionf{{145, 100}, {175, 100}, {175, 90}, {145, 90}}, 0.8)
	box3 := NewWalkBox("walkbox3", [4]Positionf{{145, 90}, {175, 90}, {175, 80}, {145, 80}}, 0.6)
	box4 := NewWalkBox("walkbox4", [4]Positionf{{155, 80}, {165, 80}, {165, 75}, {155, 75}}, 0.3)

	r.wbmatrix = NewWalkBoxMatrix([]*WalkBox{box0, box1, box2, box3, box4})

}

// DeclareWalkBoxMatrix declares walk box matrix for the room.
func (r *Room) DeclareWalkBoxMatrix(walkboxes []*WalkBox) {
	r.wbmatrix = NewWalkBoxMatrix(walkboxes)
}

// Draw renders the room in the viewport.
func (r *Room) Draw() {
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
		item.Draw()
	}

	// TODO if Debug Mode
	r.wbmatrix.Draw()
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
	CallReceiver() ScriptCallReceiver
	Caption() string
	Draw()
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
			if a.room != nil {
				job = RecoverWithValue(
					a.room.script.CallMethod(a.room.callrecv, "exit", nil),
					func(err error) any {
						log.Printf("Failed to call room exit function: %v", err)
						return nil
					},
				)
			}
			a.room = room
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

func (a *App) drawSceneViewport() {
	if a.room != nil {
		a.room.Draw()
	}
}
