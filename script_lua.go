package pctk

import (
	"bytes"
	"log"

	"github.com/Shopify/go-lua"
)

func (s *Script) luaInit(app *App) {
	if s.lua == nil {
		s.lua = NewLuaInterpreter(app, s)
		lua.BaseOpen(s.lua.State)
		s.lua.DeclareActorType()
		s.lua.DeclareColorType()
		s.lua.DeclareControlType()
		s.lua.DeclareUtilityFunctions()
		s.lua.DeclareDirectionType()
		s.lua.DeclareFutureType()
		s.lua.DeclareMusicType()
		s.lua.DeclareObjectDefaultsType()
		s.lua.DeclareObjectType()
		s.lua.DeclarePositionType()
		s.lua.DeclareRectType()
		s.lua.DeclareRoomType()
		s.lua.DeclareSentenceChoiceType()
		s.lua.DeclareSizeType()
		s.lua.DeclareSoundType()
		s.lua.DeclareWalkBoxType()

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

func (s *Script) luaRun() {
	if s.lua == nil {
		log.Panic("Script not initialized")
	}
	if err := s.lua.Execute(s.ref.String(), bytes.NewReader(s.Code)); err != nil {
		log.Fatalf("Error running script '%s': %s", s.ref.String(), err.Error())
	}
}

func (s *Script) luaCallMethod(cb ScriptCallbackID, args []ScriptEntityValue) Future {
	prom := NewPromise()
	go func() {
		if s.lua == nil {
			log.Panic("Script not initialized")
		}

		err := s.lua.CallMethod(cb, args)
		if err != nil {
			prom.CompleteWithError(err)
			return
		}
		prom.Complete()
	}()
	return prom
}
