package pctk

import (
	"io"
	"log"
	"sync"

	"github.com/google/uuid"
)

// ScriptEntityType is the type of a user data entity in a script.
type ScriptEntityType string

const (
	// ScriptEntityAny is the type of an entity that can be any type.
	ScriptEntityAny ScriptEntityType = "any"

	// ScriptEntityActor is the type of an Actor entity.
	ScriptEntityActor ScriptEntityType = "actor"

	// ScriptEntityAnimation is the type of an Animation entity.
	ScriptEntityAnimation ScriptEntityType = "animation"

	// ScriptEntityClass is the type of an ObjectClass entity.
	ScriptEntityClass ScriptEntityType = "class"

	// ScriptEntityColor is the type of a Color entity.
	ScriptEntityColor ScriptEntityType = "color"

	// ScriptEntityControl is the type of a Control entity.
	ScriptEntityControl ScriptEntityType = "control"

	// ScriptEntityDir is the type of a Direction entity.
	ScriptEntityDir ScriptEntityType = "direction"

	// ScriptEntityFuture is the type of a Future entity.
	ScriptEntityFuture ScriptEntityType = "future"

	// ScriptEntityMusic is the type of a Music entity.
	ScriptEntityMusic ScriptEntityType = "music"

	// ScriptEntityObject is the type of an Object entity.
	ScriptEntityObject ScriptEntityType = "object"

	// ScriptEntityObjectDefaults is the type of an ObjectDefaults entity.
	ScriptEntityObjectDefaults ScriptEntityType = "defaults"

	// ScriptEntityPos is the type of a Position entity.
	ScriptEntityPos ScriptEntityType = "position"

	// ScriptEntityRect is the type of a Rectangle entity.
	ScriptEntityRect ScriptEntityType = "rect"

	// ScriptEntityRef is the type of a ResourceRef entity.
	ScriptEntityRef ScriptEntityType = "ref"

	// ScriptEntityRoom is the type of a Room entity.
	ScriptEntityRoom ScriptEntityType = "room"

	// ScriptEntitySize is the type of a Size entity.
	ScriptEntitySize ScriptEntityType = "size"

	// ScriptEntitySentenceChoice is the type of a SentenceChoice entity.
	ScriptEntitySentenceChoice ScriptEntityType = "choice"

	// ScriptEntitySound is the type of a Sound entity.
	ScriptEntitySound ScriptEntityType = "sound"

	// ScriptEntityState is the type of an ObjectState entity.
	ScriptEntityState ScriptEntityType = "state"

	// ScriptEntityWalkBox is the type of a Walkbox entity.
	ScriptEntityWalkBox ScriptEntityType = "walkbox"
)

// RegistryName returns the name of the entity type in the Lua registry.
func (t ScriptEntityType) RegistryName() string {
	return "pctk." + string(t)
}

// String returns the string representation of the entity type.
func (t ScriptEntityType) String() string {
	return string(t)
}

// ScriptEntityValue is a value from a script that is visible by other scripts.
type ScriptEntityValue struct {
	// EntityType is the type of the script entity value.
	Type ScriptEntityType

	// UserData is user data of the script entity value.
	UserData any
}

// ScriptNamedEntityValue is a name and value from a script that is visible by other scripts.
type ScriptNamedEntityValue struct {
	ScriptEntityValue

	// Name is the name of the script entity value.
	Name string
}

// ScriptEntityHandler is a function to handle a script entity.
type ScriptEntityHandler func(exp ScriptNamedEntityValue)

// ScriptImportHandler is a function that can be called from Lua using the interpreter to import
// entity values from other scripts. The handler will be called with each exported entity from the
// script.
type ScriptImportHandler func(script ResourceRef, handler ScriptEntityHandler)

// ScriptCallbackReceiver is an interface to receive script callbacks.
type ScriptCallbackReceiver interface {
	DeclareCallback(cb *ScriptCallback) error
	FindCallback(name string) *ScriptCallback
}

// ScriptCustomGetter is an interface to get custom values from a script. User values that implement
// this interface can expose other user values to the script through getters.
type ScriptCustomGetter interface {
	GetScriptField(name string) *ScriptEntityValue
}

// ScriptCallbackID is the identifier of a callback function in a script.
type ScriptCallbackID string

// NewScriptCallbackID creates a new script callback identifier.
func NewScriptCallbackID() ScriptCallbackID {
	return ScriptCallbackID(uuid.New().String())
}

// String returns the string representation of the script callback identifier.
func (id ScriptCallbackID) String() string {
	return string(id)
}

// ScriptCallback represents a callback function in a script.
type ScriptCallback struct {
	ID     ScriptCallbackID // The identifier used to find it in the callbacks table
	Name   string           // The name of the callback
	Script *Script          // The script where the callback is declared
}

// Invoke calls the callback with the given arguments.
func (c ScriptCallback) Invoke(args []ScriptEntityValue) Future {
	return c.Script.CallMethod(c.ID, args)
}

// ScriptLanguage represents the language of a script.
type ScriptLanguage byte

const (
	// ScriptUndefined is an undefined script language.
	ScriptUndefined ScriptLanguage = iota

	// ScriptLua is the Lua script language.
	ScriptLua
)

// Script represents a script.
type Script struct {
	Language ScriptLanguage
	Code     []byte

	mutex   sync.Mutex
	ref     ResourceRef
	exports []ScriptNamedEntityValue

	lua *LuaInterpreter
}

// NewScript creates a new script.
func NewScript(lang ScriptLanguage, code []byte) *Script {
	return &Script{
		Language: lang,
		Code:     code,
	}
}

// CallMethod calls a method for a call receiver with the given arguments.
func (s *Script) CallMethod(cb ScriptCallbackID, args []ScriptEntityValue) Future {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.Language {
	case ScriptLua:
		return s.luaCallMethod(cb, args)
	default:
		log.Panicf("Unknown script language: %0x", s.Language)
		return nil
	}
}

// BinaryDecode decodes the script from a binary stream. The format is:
//   - byte: the script language.
//   - uint32: the length of the script code.
//   - []byte: the script code.
func (s *Script) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, s.Language, uint32(len(s.Code)), s.Code)
}

// BinaryDecode decodes the script from a binary stream. See Script.BinaryEncode for the format.
func (s *Script) BinaryDecode(r io.Reader) error {
	var lang ScriptLanguage
	var length uint32
	if err := BinaryDecode(r, &lang, &length); err != nil {
		return err
	}

	code := make([]byte, length)
	if err := BinaryDecode(r, &code); err != nil {
		return err
	}

	s.Language = ScriptLanguage(lang)
	s.Code = code
	return nil
}

// Exports returns the exported entities of the script.
func (s *Script) Exports() []ScriptNamedEntityValue {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.exports
}

// Run the script. This will evaluate the code in the script, running the declarations (if any) and
// preparing the code to receive calls.
func (s *Script) Run(app *App) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.Language {
	case ScriptLua:
		s.luaInit(app)
		s.luaRun()
	default:
		log.Panicf("Unknown script language: %0x", s.Language)
	}
}

// LoadScript loads a script from the resources. If the script is already loaded, it will return the
// loaded script. Otherwise, it will load the script, run it, and return it.
func (a *App) LoadScript(ref ResourceRef) *Script {
	script, ok := a.scripts[ref]
	if !ok {
		script = a.res.LoadScript(ref)
		if script == nil {
			log.Panicf("Script not found: %s", ref.String())
		}
		a.scripts[ref] = script
		script.Run(a)
	}
	return script
}
