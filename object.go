package pctk

import (
	"errors"
	"fmt"
)

// Object represents an object in the game. Objects are defined in the scope of rooms and generated
// by the room scripts.
type Object struct {
	Class   ObjectClass             // The classes the object belongs to as OR-ed bit flags
	Hotspot Rectangle               // The hotspot of the object (for mouse interaction)
	Name    string                  // The name of the object as seen by the player
	Owner   *Actor                  // The actor that owns the object, or nil if not picked up
	Pos     Position                // The position of the object in its room (for rendering)
	Room    *Room                   // The room where the object is declared, and where actions code resides
	Sprites ResourceRef             // The reference to the sprites of the object
	State   *ObjectState            // The current state of the object
	States  map[string]*ObjectState // The states the object can be in
	UseDir  Direction               // The direction the actor when using the object
	UsePos  Position                // The position the actor was when using the object

	callbacks []*ScriptCallback // The callbacks declared in the object
	sprites   *SpriteSheet      // The sprites of the object
}

// NewObject creates a new object.
func NewObject() *Object {
	return &Object{
		States: make(map[string]*ObjectState),
	}
}

// EnableClass enables a class in the object.
func (o *Object) EnableClass(class ObjectClass) {
	o.Class = o.Class.Enable(class)
}

// DisableClass disables a class in the object.
func (o *Object) DisableClass(class ObjectClass) {
	o.Class = o.Class.Disable(class)
}

// Caption returns the name of the object.
func (o *Object) Caption() string {
	return o.Name
}

// DeclareCallback declares a callback in the object.
func (o *Object) DeclareCallback(cb *ScriptCallback) error {
	for _, c := range o.callbacks {
		if c.Name == cb.Name {
			return fmt.Errorf("callback '%s' already declared", cb.Name)
		}
	}
	o.callbacks = append(o.callbacks, cb)
	return nil
}

// FindCallback finds a callback in the object by name.
func (o *Object) FindCallback(name string) *ScriptCallback {
	for _, c := range o.callbacks {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// GetScriptField returns the state of the object with the given name.
func (o *Object) GetScriptField(name string) *ScriptEntityValue {
	if state, ok := o.States[name]; ok {
		return &ScriptEntityValue{Type: ScriptEntityState, UserData: state}
	}
	return nil
}

// ItemClass returns the class of the object.
func (o *Object) ItemClass() ObjectClass {
	return o.Class
}

// CurrentState returns the current state of the object.
func (o *Object) CurrentState() *ObjectState {
	return o.State
}

// Draw renders the object in the viewport.
func (o *Object) Draw(f *Frame) {
	if !o.IsVisible() {
		return
	}
	if st := o.CurrentState(); st != nil && st.Anim != nil {
		pos := o.Pos.Sub(NewPos(o.sprites.frameSize.W/2, o.sprites.frameSize.H))
		st.Anim.Draw(o.sprites, pos)
	}
}

// IsVisible returns true if the object is visible in the room, false otherwise.
func (o *Object) IsVisible() bool {
	return o.Owner == nil && !o.Class.Is(ObjectClassUntouchable)
}

// Load the object resources.
func (o *Object) Load(res ResourceLoader) {
	if o.sprites == nil && o.Sprites != ResourceRefNull {
		o.sprites = res.LoadSpriteSheet(o.Sprites)
	}
}

// ItemOwner returns the actor that owns the object, or nil if not picked up.
func (o *Object) ItemOwner() *Actor {
	if o == nil {
		return nil
	}
	return o.Owner
}

// ItemPosition returns the position of the object.
func (o *Object) ItemPosition() Position {
	return o.Pos
}

// ItemUsePosition returns the position where actors interact with the object.
func (o *Object) ItemUsePosition() (Position, Direction) {
	return o.UsePos, o.UseDir
}

// ObjectState represents a state of an object.
type ObjectState struct {
	Anim   *Animation // The animation while in this state.
	Object *Object    // Object related to the state.
}

// ObjectClass represents a class of objects. Classes are aimed to be used as bit flags that can be
// OR-ed together. As this type is backed by a uint64, there can be up to 64 different classes.
// There are two kind of classes: the built-in classes and the custom classes.
type ObjectClass uint64

const (
	// ObjectClassPerson is a built-in class that represents objects that are persons.
	ObjectClassPerson ObjectClass = 1 << 0

	// ObjectClassUntouchable is a built-in class that represents objects the player cannot interact
	// with and will not be visible.
	ObjectClassUntouchable ObjectClass = 1 << 1

	// ObjectClassPickable is a built-in class that represents objects that can be picked up by the
	// player.
	ObjectClassPickable ObjectClass = 1 << 2

	// ObjectClassOpenable is a built-in class that represents objects that can be opened by the
	// player.
	ObjectClassOpenable ObjectClass = 1 << 3

	// ObjectClassCloseable is a built-in class that represents objects that can be closed by the
	// player.
	ObjectClassCloseable ObjectClass = 1 << 4

	// ObjectClassApplicable is a built-in class that represents objects that can be applied to
	// other objects. This is what determines that "use" verb requires an object to be applied to.
	ObjectClassApplicable ObjectClass = 1 << 5
)

// WithObjectClasses returns a new object class with the given classes.
func WithObjectClasses(head ObjectClass, tail ...ObjectClass) ObjectClass {
	for _, class := range tail {
		head |= class
	}
	return head
}

// Enable enables a class in the object class.
func (c ObjectClass) Enable(class ObjectClass) ObjectClass {
	return c | class
}

// Disable disables a class in the object class.
func (c ObjectClass) Disable(class ObjectClass) ObjectClass {
	return c &^ class
}

// Is returns true if the class is enabled in the object classes.
func (c ObjectClass) Is(class ObjectClass) bool {
	return c&class != 0
}

// IsOneOf returns true if some class of c is also present in other
func (c ObjectClass) IsOneOf(head ObjectClass, tail ...ObjectClass) bool {
	return c&WithObjectClasses(head, tail...) != 0
}

// IsAllOf returns true if all classes of c are also present in other
func (c ObjectClass) IsAllOf(head ObjectClass, tail ...ObjectClass) bool {
	return c&WithObjectClasses(head, tail...) == c
}

// IsNoneOf returns true if none of the classes of c are present in other
func (c ObjectClass) IsNoneOf(head ObjectClass, tail ...ObjectClass) bool {
	return c&WithObjectClasses(head, tail...) == 0
}

// ObjectDefaults represents the script call receiver for default actions.
type ObjectDefaults struct {
	callbacks []*ScriptCallback
}

// CallFunction calls a default function in the script with the given arguments.
func (o *ObjectDefaults) CallFunction(function string, args []ScriptEntityValue) Future {
	cb := o.FindCallback(function)
	if cb == nil {
		return AlreadyFailed(fmt.Errorf("default callback '%s' not found", function))
	}

	return cb.Invoke(args)
}

// DeclareCallback declares a callback in the object defaults.
func (o *ObjectDefaults) DeclareCallback(cb *ScriptCallback) error {
	for _, c := range o.callbacks {
		if c.Name == cb.Name {
			return fmt.Errorf("callback '%s' already declared", cb.Name)
		}
	}
	o.callbacks = append(o.callbacks, cb)
	return nil
}

// FindCallback finds a callback in the object defaults by name.
func (o *ObjectDefaults) FindCallback(name string) *ScriptCallback {
	for _, c := range o.callbacks {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// SetObjectDefaults sets the default actions for objects. This method can be called only once. If
// called again, it will return an error.
func (a *App) SetObjectDefaults(def *ObjectDefaults) error {
	if a.defaults != nil {
		return errors.New("Defaults already set")
	}
	a.defaults = def
	return nil
}
