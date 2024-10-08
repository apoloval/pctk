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
	callrecv   ScriptCallReceiver // The call receiver for the room
	campos     int                // The camera position
	camtarget  int                // The camera target position
	camfollow  *Actor             // The actor to follow with the camera
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

// CameraFollowActor makes the camera follow the given actor.
func (r *Room) CameraFollowActor(actor *Actor) error {
	if actor.Room != r {
		return fmt.Errorf("Actor %s is not in the room", actor.Name)
	}
	r.camfollow = actor
	return nil
}

// CameraMoveTo moves the camera to the given position.
func (r *Room) CameraMoveTo(pos int) {
	r.camtarget = pos
	r.camfollow = nil
}

// CameraPos returns the current camera position.
func (r *Room) CameraPos() int {
	return r.campos
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
func (r *Room) Draw(debugEnabled bool) {
	r.background.Draw(NewPos(-r.campos, 0), White)
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

	if debugEnabled && r.wbmatrix != nil {
		r.wbmatrix.Draw(r.campos)
	}

	r.updateCamera()
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
	pos.X += r.campos

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

func (r *Room) updateCamera() {
	if r.camfollow != nil {
		r.camtarget = int(r.camfollow.pos.X) - ScreenWidth/2
	}
	if r.campos != r.camtarget {
		if r.campos < r.camtarget {
			r.campos += RoomCameraSpeed
			if r.campos > r.camtarget {
				r.campos = r.camtarget
			}
		} else {
			r.campos -= RoomCameraSpeed
			if r.campos < r.camtarget {
				r.campos = r.camtarget
			}
		}
		if r.campos < 0 {
			r.campos = 0
		}
		if r.campos > int(r.background.Width())-ScreenWidth {
			r.campos = int(r.background.Width()) - ScreenWidth
		}
	}
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
		a.room.Draw(a.debugEnabled)
	}
}
