package pctk

import (
	"errors"
	"io"
	"log"
	"sync"

	"github.com/google/uuid"
)

// ErrEntityNotFound is the error returned when a function or method is unknown in a script.
var ErrScriptFunctionUnknown = errors.New("script function unknown")

// ScriptEntityType is the type of a user data entity in a script.
type ScriptEntityType string

const (
	// ScriptEntityActor is the type of an Actor entity.
	ScriptEntityActor ScriptEntityType = "actor"

	// ScriptEntityAnimation is the type of an Animation entity.
	ScriptEntityAnimation ScriptEntityType = "animation"

	// ScriptEntityClass is the type of an ObjectClass entity.
	ScriptEntityClass ScriptEntityType = "class"

	// ScriptEntityColor is the type of a Color entity.
	ScriptEntityColor ScriptEntityType = "color"

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

// ScriptCallReceiver is the identifier of a script element that can receive function calls.
type ScriptCallReceiver string

// NewScriptCallReceiver creates a new script instance identifier.
func NewScriptCallReceiver() ScriptCallReceiver {
	return ScriptCallReceiver(uuid.New().String())
}

// String returns the string representation of the script instance identifier.
func (id ScriptCallReceiver) String() string {
	return string(id)
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

// CallFunction calls a function in the script with the given arguments.
func (s *Script) CallFunction(
	recv ScriptCallReceiver,
	function string,
	args []ScriptEntityValue,
) Future {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.Language {
	case ScriptLua:
		return s.luaCallFunction(recv, function, args)
	default:
		log.Panicf("Unknown script language: %0x", s.Language)
		return nil
	}
}

// CallMethod calls a method for a call receiver with the given arguments.
func (s *Script) CallMethod(
	recv ScriptCallReceiver,
	method string,
	args []ScriptEntityValue,
) Future {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.Language {
	case ScriptLua:
		return s.luaCallMethod(recv, method, args)
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
		s.luaRun(app)
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
			log.Panicf("Script not found: %s", ref)
		}
		script.Run(a)
	}
	return script
}
