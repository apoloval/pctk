package pctk

import (
	"bytes"
	"log"

	"github.com/Shopify/go-lua"
)

func (s *Script) luaInit(app *App) {
	if s.lua == nil {
		s.lua = NewLuaInterpreter()
		lua.BaseOpen(s.lua.State)
		s.lua.DeclareActorType(app, s)
		s.lua.DeclareColorType()
		s.lua.DeclareControlType(app)
		s.lua.DeclareUtilityFunctions(app)
		s.lua.DeclareDirectionType()
		s.lua.DeclareFutureType()
		s.lua.DeclareMusicType(app)
		s.lua.DeclareObjectDefaultsType(app, s)
		s.lua.DeclareObjectType()
		s.lua.DeclarePositionType()
		s.lua.DeclareRectType()
		s.lua.DeclareRoomType(app, s)
		s.lua.DeclareSentenceChoiceType(app)
		s.lua.DeclareSizeType()
		s.lua.DeclareSoundType(app)
		s.lua.DeclareWalkBoxType(app)

		s.lua.DeclareExportFunction(func(exp ScriptNamedEntityValue) {
			s.exports = append(s.exports, exp)
		})
		s.lua.DeclareImportFunction(func(script ResourceRef, handler ScriptEntityHandler) {
			other := app.LoadScript(script)
			for _, exp := range other.Exports() {
				handler(exp)
			}
		})
	}
}

func (s *Script) luaRun(app *App) {
	if s.lua == nil {
		log.Panic("Script not initialized")
	}
	input := bytes.NewReader(s.Code)
	if err := s.lua.Load(input, "="+s.ref.String(), ""); err != nil {
		log.Fatalf("Error loading script '%s': %s", s.ref, err.Error())
	}
	if err := s.lua.ProtectedCall(0, lua.MultipleReturns, 0); err != nil {
		log.Fatalf("Error running script '%s': %s", s.ref, err.Error())
	}
}

func (s *Script) luaCallFunction(
	recv ScriptCallReceiver,
	function string,
	args []ScriptEntityValue,
) Future {
	prom := NewPromise()
	go func() {
		if s.lua == nil {
			log.Panic("Script not initialized")
		}

		err := s.lua.CallFunction(recv, function, args)
		if err != nil {
			prom.CompleteWithError(err)
			return
		}
		prom.Complete()
	}()
	return prom
}

func (s *Script) luaCallMethod(
	recv ScriptCallReceiver,
	method string,
	args []ScriptEntityValue,
) Future {
	prom := NewPromise()
	go func() {
		if s.lua == nil {
			log.Panic("Script not initialized")
		}

		err := s.lua.CallMethod(recv, method, args)
		if err != nil {
			prom.CompleteWithError(err)
			return
		}
		prom.Complete()
	}()
	return prom
}
