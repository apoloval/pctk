package pctk

import (
	"errors"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	DefaultActorSpeakDelay = 500 * time.Millisecond
)

var (
	DefaultActorPosition  = NewPos(160, 90)
	DefaultActorSpeed     = NewPosf(80, 20)
	DefaultActorSize      = NewSize(32, 48)
	DefaultActorDirection = DirRight
	DefaultActorTalkColor = BrigthGrey
	DefaultActorUsePos    = NewPos(ScreenWidth/2, 120)
)

// Actor is an entity that represents a character in the game.
type Actor struct {
	CallRecv  ScriptCallReceiver // The call receiver for the actor
	Costume   ResourceRef        // Reference to the costume of the actor
	Elev      int                // Elevation of the actor
	Name      string             // Name of the actor
	Room      *Room              // The room where the actor is located
	Script    *Script            // The script where this actor is defined.
	Size      Size               // Size of the actor
	TalkColor Color              // Color of the text when the actor talks
	UsePos    Position           // Position where other actors interact with this actor
	UseDir    Direction          // Direction where other actors interact with this actor

	act       *Action
	costume   *Costume
	dialog    *Dialog
	ego       bool
	inventory []*Object
	lookAt    Direction
	pos       Positionf
	speed     Positionf
}

// NewActor creates a new actor with the given ID and name.
func NewActor(name string) *Actor {
	return &Actor{
		TalkColor: DefaultActorTalkColor,
		Size:      DefaultActorSize,
		UsePos:    DefaultActorUsePos,
		UseDir:    DefaultActorDirection,

		Name:  name,
		pos:   DefaultActorPosition.ToPosf(),
		speed: DefaultActorSpeed,
	}
}

// AddToInventory adds an object to the actor's inventory.
func (a *Actor) AddToInventory(obj *Object) {
	a.inventory = append(a.inventory, obj)
	obj.Owner = a
}

// CancelAction cancels the current action of the actor.
func (a *Actor) CancelAction() {
	if a.act != nil {
		a.act.Cancel()
	}
	a.act = nil
}

// Caption returns the name of the actor.
func (a *Actor) Caption() string {
	return a.Name
}

// ItemClass returns the class of the actor.
func (a *Actor) ItemClass() ObjectClass {
	return ObjectClassPerson
}

// DirectionTo returns the direction from the actor to the given position.
func (a *Actor) DirectionTo(pos Position) Direction {
	return a.pos.ToPos().DirectionTo(pos)
}

// Do executes the action in the actor.
func (a *Actor) Do(action *Action) Future {
	if a.act != nil {
		a.act.Cancel()
	}
	a.act = action
	return a.act.Done()
}

// Draw renders the actor in the viewport.
func (a *Actor) Draw() {
	if a.act == nil {
		a.act = Standing(a.lookAt)
	}

	if a.act.RunFrame(a) {
		a.act = nil
	}
}

// Hotspot returns the hotspot of the actor.
func (a *Actor) Hotspot() Rectangle {
	return Rectangle{Pos: a.costumePos(), Size: a.Size}
}

// Inventory returns the inventory of the actor.
func (a *Actor) Inventory() []*Object {
	return a.inventory
}

// IsEgo returns true if the actor is the actor under player's control, false otherwise.
func (a *Actor) IsEgo() bool {
	return a.ego
}

// IsSpeaking returns true if the actor is speaking, false otherwise.
func (a *Actor) IsSpeaking() bool {
	return a.dialog != nil && !a.dialog.Done().IsCompleted()
}

// Load the actor resources.
func (a *Actor) Load(res ResourceLoader) {
	if a.costume == nil && a.Costume != ResourceRefNull {
		a.costume = res.LoadCostume(a.Costume)
	}
}

// Locate the actor in the given room, position and direction.
func (a *Actor) Locate(room *Room, pos Position, dir Direction) {
	a.Room = room
	a.pos = pos.ToPosf()
	a.Do(Standing(dir))
}

// ItemOwner returns the actor that owns the actor in its inventory. Typically nil unless you manage to
// model that actors can be picked up (as if they were dogs or monkeys).
func (a *Actor) ItemOwner() *Actor {
	return nil
}

// ItemPosition returns the position of the actor.
func (a *Actor) ItemPosition() Position {
	return a.pos.ToPos()
}

// CallReceiver returns the call receiver for the actor.
func (a *Actor) CallReceiver() ScriptCallReceiver {
	return a.CallRecv
}

// ItemUsePosition returns the position where actors interact with the actor.
func (a *Actor) ItemUsePosition() (Position, Direction) {
	return a.UsePos, a.UseDir
}

func (a *Actor) costumePos() Position {
	pos := a.pos.ToPos().Sub(NewPos((a.Size.W / 2), a.Size.H-a.Elev))
	if a.Room != nil {
		pos.X -= a.Room.campos
	}
	return pos
}

func (a *Actor) dialogPos() Position {
	return a.costumePos().Above(a.Size.H)
}

// Action is an action that an actor is performing.
type Action struct {
	prom *Promise
	f    func(*Actor, *Promise)
}

// Standing creates a new action that makes an actor stand in the given direction.
func Standing(dir Direction) *Action {
	return &Action{
		prom: NewPromise(),
		f: func(a *Actor, done *Promise) {
			a.lookAt = dir
			costume := CostumeIdle(dir)
			if a.IsSpeaking() {
				costume = CostumeSpeak(dir)
			}
			if cos := a.costume; cos != nil {
				cos.draw(costume, a.costumePos())
			}
		},
	}
}

// WalkingTo creates a new action that makes an actor walk to a given position.
func WalkingTo(pos Position) *Action {
	return &Action{
		prom: NewPromise(),
		f: func(a *Actor, done *Promise) {
			if cos := a.costume; cos != nil {
				cos.draw(CostumeWalk(a.lookAt), a.costumePos())
			}

			if a.pos.ToPos() == pos {
				done.Complete()
				return
			}

			a.lookAt = a.pos.ToPos().DirectionTo(pos)
			a.pos = a.pos.Move(pos.ToPosf(), a.speed.Scale(rl.GetFrameTime()))
		},
	}
}

// SpeakingTo creates a new action that makes an actor speak to a dialog.
func SpeakingTo(dialog Future) *Action {
	return &Action{
		prom: NewPromise(),
		f: func(a *Actor, done *Promise) {
			if cos := a.costume; cos != nil {
				cos.draw(CostumeSpeak(a.lookAt), a.costumePos())
			}
			if dialog.IsCompleted() {
				done.Complete()
			}
		},
	}
}

// Cancel cancels the action.
func (a *Action) Cancel() {
	a.prom.Break()
}

// Done returns a future that will be completed when the action is done.
func (a *Action) Done() Future {
	return a.prom
}

// RunFrame runs a frame of the action.
func (a *Action) RunFrame(actor *Actor) (completed bool) {
	a.f(actor, a.prom)
	return a.prom.IsCompleted()
}

// DeclareActor declares a new actor in the app.
func (a *App) DeclareActor(actor *Actor) error {
	for _, obj := range a.actors {
		if obj == actor {
			return errors.New("actor already declared")
		}
	}
	a.actors = append(a.actors, actor)
	return nil
}

// SelectEgo sets actor as the ego.
func (a *App) SelectEgo(actor *Actor) {
	if a.ego != nil {
		a.ego.ego = false
	}
	a.ego = actor
	if a.ego != nil {
		a.ego.ego = true
	}
}

// ActorShow shows the actor in the active room
func (a *App) ActorShow(actor *Actor, pos Position, lookAt Direction) error {
	if a.room == nil {
		return errors.New("no active room to show actor")
	}

	actor.Load(a.res)
	a.room.PutActor(actor)
	actor.Locate(a.room, pos, lookAt)
	return nil
}
