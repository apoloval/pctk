package pctk

import (
	"fmt"
	"time"

	"github.com/Shopify/go-lua"
)

// LuaFunction is a function that can be called from Lua using the interpreter.
type LuaFunction func(*LuaInterpreter) int

// LuaInterpreter is a wrapper around the lua.State that provides convenience methods for pctk.
type LuaInterpreter struct {
	*lua.State
}

// NewLuaInterpreter creates a new LuaInterpreter.
func NewLuaInterpreter() *LuaInterpreter {
	return &LuaInterpreter{lua.NewState()}
}

// WrapInterpreter wraps a lua.State into a LuaInterpreter.
func WrapInterpreter(l *lua.State) *LuaInterpreter {
	return &LuaInterpreter{l}
}

// CallFunction calls a function the entity instance previously registed using RegisterInstance.
func (l *LuaInterpreter) CallFunction(
	recv ScriptCallReceiver,
	method string,
	args []ScriptEntityValue,
) error {
	lua.NewMetaTable(l.State, "pctk.callreceivers")
	l.Field(-1, recv.String())
	if l.IsNil(-1) {
		l.Pop(2)
		return fmt.Errorf("call receiver '%s' not found during function '%s' call", recv, method)
	}

	l.PushString(method)
	l.RawGet(-2)
	if !l.IsFunction(-1) {
		l.Pop(3)
		return fmt.Errorf("function '%s' not found in call receiver '%s': %w",
			method, recv, ErrScriptFunctionUnknown)
	}

	for _, arg := range args {
		l.PushEntity(arg.Type, arg.UserData)
	}
	err := l.ProtectedCall(len(args), 0, 0)
	l.Pop(2)
	return err
}

// CallFunction calls a function the entity instance previously registed using RegisterInstance.
func (l *LuaInterpreter) CallMethod(
	recv ScriptCallReceiver,
	method string,
	args []ScriptEntityValue,
) error {
	lua.NewMetaTable(l.State, "pctk.callreceivers")
	l.Field(-1, recv.String())
	if l.IsNil(-1) {
		l.Pop(2)
		return fmt.Errorf("call receiver '%s' not found during method '%s' call: %w",
			recv, method, ErrScriptFunctionUnknown)
	}

	l.PushString(method)
	l.RawGet(-2)
	if !l.IsFunction(-1) {
		l.Pop(3)
		return fmt.Errorf("method '%s' not found in call receiver '%s': %w",
			method, recv, ErrScriptFunctionUnknown)
	}
	l.PushValue(-2) // Push the call receiver as first argument
	for _, arg := range args {
		l.PushEntity(arg.Type, arg.UserData)
	}
	err := l.ProtectedCall(len(args)+1, 0, 0)
	l.Pop(2)
	return err
}

// CheckEntity checks if the entity in the given index is of the given type, and returns it.
func (l *LuaInterpreter) CheckEntity(index int, typ ScriptEntityType) any {
	index = l.AbsIndex(index)
	if !l.IsTable(index) {
		lua.ArgumentError(l.State, index, fmt.Sprintf("expected entity %s", typ))
	}
	if !l.MetaTable(index) {
		lua.ArgumentError(l.State, index, fmt.Sprintf("expected entity %s", typ))
	}

	lua.MetaTableNamed(l.State, typ.RegistryName())
	match := l.RawEqual(-1, -2)
	l.Pop(2)
	if !match {
		lua.ArgumentError(l.State, index, fmt.Sprintf("expected entity %s", typ))
	}

	l.Field(index, "__userdata")
	obj := l.ToUserData(-1)
	l.Pop(1)

	return obj
}

// CheckFieldInteger checks if the field of the table at index is an integer, and returns it.
func (l *LuaInterpreter) CheckFieldInteger(index int, name string) (val int) {
	l.WithField(index, name, func() { val = lua.CheckInteger(l.State, -1) })
	return
}

// CheckFieldEntity checks if the field of the table at index is an entity of the given type, and
// returns it.
func (l *LuaInterpreter) CheckFieldEntity(index int, name string, typ ScriptEntityType) (val any) {
	l.WithField(index, name, func() { val = l.CheckEntity(-1, typ) })
	return
}

// CheckFieldString checks if the field of the table at index is a string, and returns it.
func (l *LuaInterpreter) CheckFieldString(index int, name string) (val string) {
	l.WithField(index, name, func() { val = lua.CheckString(l.State, -1) })
	return
}

// DeclareActorType declares the type of an Actor in the Lua interpreter. The handle function is
// called to declare the actor instances by the caller.
func (l *LuaInterpreter) DeclareActorType(app *App, script *Script) {
	l.DeclareColorType()
	l.DeclareDirectionType()
	l.DeclareFutureType()
	l.DeclarePositionType()
	l.DeclareSizeType()
	l.DeclareReferenceType()

	if l.DeclareEntityType(ScriptEntityActor) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityActor, "actor",
		func(l *LuaInterpreter) int {
			actor := NewActor(l.CheckFieldString(1, "name"))
			actor.Script = script
			l.WithOptionalField(1, "costume", func() {
				actor.Costume = l.CheckEntity(-1, ScriptEntityRef).(ResourceRef)
			})
			l.WithOptionalField(1, "size", func() {
				actor.Size = l.CheckEntity(-1, ScriptEntitySize).(Size)
			})
			l.WithOptionalField(1, "talkcolor", func() {
				actor.TalkColor = l.CheckEntity(-1, ScriptEntityColor).(Color)
			})
			l.WithOptionalField(1, "usepos", func() {
				actor.UsePos = l.CheckEntity(-1, ScriptEntityPos).(Position)
			})
			l.WithOptionalField(1, "usedir", func() {
				actor.UseDir = l.CheckEntity(-1, ScriptEntityDir).(Direction)
			})

			err := app.DeclareActor(actor)
			if err != nil {
				lua.Errorf(l.State, "error declaring actor: %s", err)
			}

			l.PushEntity(ScriptEntityActor, actor)
			actor.CallRecv = l.RegisterCallReceiver(-1)
			return 1
		},
	)
	l.DeclareEntityGetter(ScriptEntityActor, "name", func(l *LuaInterpreter) int {
		actor := l.CheckEntity(1, ScriptEntityActor).(*Actor)
		l.PushString(actor.Caption())
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityActor, "size", func(l *LuaInterpreter) int {
		actor := l.CheckEntity(1, ScriptEntityActor).(*Actor)
		l.PushEntity(ScriptEntitySize, actor.Size)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityActor, "talkcolor", func(l *LuaInterpreter) int {
		actor := l.CheckEntity(1, ScriptEntityActor).(*Actor)
		l.PushEntity(ScriptEntityColor, actor.TalkColor)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityActor, "usepos", func(l *LuaInterpreter) int {
		actor := l.CheckEntity(1, ScriptEntityActor).(*Actor)
		l.PushEntity(ScriptEntityPos, actor.UsePos)
		return 1
	})
	l.DeclareEntityMethod(ScriptEntityActor, "lookdir", func(l *LuaInterpreter) int {
		var cmd ActorStand
		cmd.Actor = l.CheckEntity(1, ScriptEntityActor).(*Actor)
		cmd.Direction = l.CheckEntity(2, ScriptEntityDir).(Direction)
		app.RunCommand(cmd).Wait()
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityActor, "say", func(l *LuaInterpreter) int {
		var cmd ActorSpeak
		cmd.Actor = l.CheckEntity(1, ScriptEntityActor).(*Actor)
		cmd.Text = lua.CheckString(l.State, 2)
		l.WithOptionalField(3, "color", func() {
			cmd.Color = l.CheckEntity(-1, ScriptEntityColor).(Color)
		})
		done := app.RunCommand(cmd)
		l.PushEntity(ScriptEntityFuture, done)
		return 1
	})
	l.DeclareEntityMethod(ScriptEntityActor, "select", func(l *LuaInterpreter) int {
		var cmd ActorSelectEgo
		cmd.Actor = l.CheckEntity(1, ScriptEntityActor).(*Actor)
		app.RunCommand(cmd).Wait()
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityActor, "show", func(l *LuaInterpreter) int {
		var cmd ActorShow
		cmd.Actor = l.CheckEntity(1, ScriptEntityActor).(*Actor)
		l.WithOptionalField(2, "pos", func() {
			cmd.Position = l.CheckEntity(-1, ScriptEntityPos).(Position)
		})
		l.WithOptionalField(2, "lookat", func() {
			cmd.LookAt = l.CheckEntity(-1, ScriptEntityDir).(Direction)
		})
		app.RunCommand(cmd).Wait()
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityActor, "toinventory", func(l *LuaInterpreter) int {
		var cmd ActorAddToInventory
		cmd.Actor = l.CheckEntity(1, ScriptEntityActor).(*Actor)
		cmd.Object = l.CheckEntity(2, ScriptEntityObject).(*Object)
		app.RunCommand(cmd).Wait()
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityActor, "walkto", func(l *LuaInterpreter) int {
		var cmd ActorWalkToPosition
		cmd.Actor = l.CheckEntity(1, ScriptEntityActor).(*Actor)
		cmd.Position = l.CheckEntity(2, ScriptEntityPos).(Position)
		done := app.RunCommand(cmd)
		l.PushEntity(ScriptEntityFuture, done)
		return 1
	})
	l.DeclareEntityEqual(ScriptEntityActor, func(l *LuaInterpreter) int {
		if l.EntityTypeOf(1) != ScriptEntityActor || l.EntityTypeOf(2) != ScriptEntityActor {
			l.PushBoolean(false)
			return 1
		}
		a := l.CheckEntity(1, ScriptEntityActor).(*Actor)
		b := l.CheckEntity(2, ScriptEntityActor).(*Actor)
		l.PushBoolean(a == b)
		return 1
	})
	l.DeclareEntityToString(ScriptEntityActor, func(l *LuaInterpreter) int {
		actor := l.CheckEntity(1, ScriptEntityActor).(*Actor)
		l.PushString(fmt.Sprintf("actor:(name=%s)", actor.Caption()))
		return 1
	})
}

// DeclareAnimType declares the type of an Animation in the Lua interpreter.
func (l *LuaInterpreter) DeclareAnimType() {
	if l.DeclareEntityType(ScriptEntityAnimation) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityAnimation, "fixedanim", func(l *LuaInterpreter) int {
		anim := NewAnimation()
		anim.AddFrames(
			1*time.Second,
			l.CheckFieldInteger(1, "row"),
			l.CheckFieldInteger(1, "col"),
		)
		l.PushEntity(ScriptEntityAnimation, anim)
		return 1
	})
}

// DeclareClassType declares the type of an ObjectClass in the Lua interpreter.
func (l *LuaInterpreter) DeclareClassType() {
	if l.DeclareEntityType(ScriptEntityClass) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityClass, "class",
		func(l *LuaInterpreter) int {
			class := ObjectClass(lua.CheckInteger(l.State, 1))
			l.PushEntity(ScriptEntityClass, class)
			return 1
		},
	)
	l.DeclareEntityEqual(ScriptEntityClass, func(l *LuaInterpreter) int {
		if l.EntityTypeOf(1) != ScriptEntityClass || l.EntityTypeOf(2) != ScriptEntityClass {
			l.PushBoolean(false)
			return 1
		}
		a := l.CheckEntity(1, ScriptEntityClass).(ObjectClass)
		b := l.CheckEntity(2, ScriptEntityClass).(ObjectClass)
		l.PushBoolean(a == b)
		return 1
	})
	l.PushEntity(ScriptEntityClass, ObjectClassPerson)
	l.SetGlobal("PERSON")
	l.PushEntity(ScriptEntityClass, ObjectClassPickable)
	l.SetGlobal("PICKABLE")
	l.PushEntity(ScriptEntityClass, ObjectClassOpenable)
	l.SetGlobal("OPENABLE")
	l.PushEntity(ScriptEntityClass, ObjectClassCloseable)
	l.SetGlobal("CLOSEABLE")
	l.PushEntity(ScriptEntityClass, ObjectClassApplicable)
	l.SetGlobal("APPLICABLE")
}

// DeclareColorType declares the type of a Color in the Lua interpreter.
func (l *LuaInterpreter) DeclareColorType() {
	if l.DeclareEntityType(ScriptEntityColor) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityColor, "color",
		func(l *LuaInterpreter) int {
			r := l.CheckFieldInteger(1, "r")
			g := l.CheckFieldInteger(1, "g")
			b := l.CheckFieldInteger(1, "b")
			a, ok := l.ToIntegerField(1, "a")
			if !ok {
				a = 255
			}
			l.PushEntity(ScriptEntityColor, Color{byte(r), byte(g), byte(b), byte(a)})
			return 1
		},
	)
	l.DeclareEntityGetter(ScriptEntityColor, "r", func(l *LuaInterpreter) int {
		c := l.CheckEntity(1, ScriptEntityColor).(Color)
		l.PushInteger(int(c.R))
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityColor, "g", func(l *LuaInterpreter) int {
		c := l.CheckEntity(1, ScriptEntityColor).(Color)
		l.PushInteger(int(c.G))
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityColor, "b", func(l *LuaInterpreter) int {
		c := l.CheckEntity(1, ScriptEntityColor).(Color)
		l.PushInteger(int(c.B))
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityColor, "a", func(l *LuaInterpreter) int {
		c := l.CheckEntity(1, ScriptEntityColor).(Color)
		l.PushInteger(int(c.A))
		return 1
	})
	l.DeclareEntityToString(ScriptEntityColor, func(l *LuaInterpreter) int {
		c := l.CheckEntity(1, ScriptEntityColor).(Color)
		l.PushString(fmt.Sprintf("rgba(%d, %d, %d, %d)", c.R, c.G, c.B, c.A))
		return 1
	})
}

// DeclareControlType declares the type of a Control in the Lua interpreter.
func (l *LuaInterpreter) DeclareControlType(app *App) {
	if l.DeclareEntityType(ScriptEntityControl) {
		return
	}
	l.DeclareEntityMethod(ScriptEntityControl, "cursoron", func(l *LuaInterpreter) int {
		_, err := app.RunCommand(MouseCursorOn()).Wait()
		if err != nil {
			lua.Errorf(l.State, "error enabling cursor: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityControl, "cursoroff", func(l *LuaInterpreter) int {
		_, err := app.RunCommand(MouseCursorOff()).Wait()
		if err != nil {
			lua.Errorf(l.State, "error disabling cursor: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityControl, "paneon", func(l *LuaInterpreter) int {
		_, err := app.RunCommand(ControlPaneEnable()).Wait()
		if err != nil {
			lua.Errorf(l.State, "error enabling control panel: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityControl, "paneoff", func(l *LuaInterpreter) int {
		_, err := app.RunCommand(ControlPaneDisable()).Wait()
		if err != nil {
			lua.Errorf(l.State, "error disabling control panel: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityControl, "sentencechoice", func(l *LuaInterpreter) int {
		val, err := app.RunCommand(SentenceChoiceInit()).Wait()
		if err != nil {
			lua.Errorf(l.State, "error initializing sentence choice: %s", err.Error())
		}
		choice := val.(*ControlSentenceChoice)
		l.PushEntity(ScriptEntitySentenceChoice, choice)
		return 1
	})
	l.PushEntity(ScriptEntityControl, struct{}{})
	l.SetGlobal("CONTROL")
}

// DeclareDirectionType declares the type of a Direction in the Lua interpreter.
func (l *LuaInterpreter) DeclareDirectionType() {
	if l.DeclareEntityType(ScriptEntityDir) {
		return
	}
	l.DeclareEntityToString(ScriptEntityDir, func(l *LuaInterpreter) int {
		d := l.CheckEntity(1, ScriptEntityDir).(Direction)
		l.PushString(d.String())
		return 1
	})
	l.PushEntity(ScriptEntityDir, DirLeft)
	l.SetGlobal("LEFT")
	l.PushEntity(ScriptEntityDir, DirRight)
	l.SetGlobal("RIGHT")
	l.PushEntity(ScriptEntityDir, DirUp)
	l.SetGlobal("UP")
	l.PushEntity(ScriptEntityDir, DirDown)
	l.SetGlobal("DOWN")
}

// DeclareEntityType declares a new entity type in the Lua interpreter. If the entity is already
// declared, it returns true.
func (l *LuaInterpreter) DeclareEntityType(typ ScriptEntityType) bool {
	// Put the prototype metatable if it does not exist.
	if !lua.NewMetaTable(l.State, typ.RegistryName()) {
		l.Pop(1)
		return true
	}

	// Configure the __type metafield that indicates the entity type of the metatable.
	l.PushString(typ.String())
	l.SetField(-2, "__type")

	// Configure the __index metamethod. This will be called when a field is not found in the
	// entity. We have to look for methods and getters declared in the prototype.
	l.PushFunction(func(l *LuaInterpreter) int {
		lua.CheckType(l.State, 1, lua.TypeTable)
		key := lua.CheckString(l.State, 2)
		if !l.MetaTable(1) {
			lua.Errorf(l.State, "prototype __index called with a table without metatable")
		}

		// Look for getters in the prototype.
		lua.SubTable(l.State, -1, "__getters")
		l.Field(-1, key)
		if l.IsFunction(-1) {
			l.PushValue(1) // Pass the entity table as argument
			l.Call(1, 1)
			return 1
		}
		l.Pop(2)

		// Look for methods in the prototype.
		lua.SubTable(l.State, -1, "__methods")
		l.Field(-1, key)
		if l.IsFunction(-1) {
			return 1
		}

		// The field was not found.
		lua.ArgumentError(l.State, 2, fmt.Sprintf(
			"no getter or method '%s' declared for  entity type '%s'", key, typ))
		return 0
	})
	l.SetField(-2, "__index")

	// Configure the __newindex metamethod. This will be called when a new field is set in the
	// entity. We have to look for setters declared in the prototype.
	l.PushFunction(func(l *LuaInterpreter) int {
		lua.CheckType(l.State, 1, lua.TypeTable)
		key := lua.CheckString(l.State, 2)
		if !l.MetaTable(1) {
			lua.Errorf(l.State, "prototype __newindex called with a table without metatable")
		}

		// Look for setters in the prototype.
		lua.SubTable(l.State, -1, "__setters")
		l.Field(-1, key)
		if l.IsFunction(-1) {
			l.PushValue(1) // Pass the entity table as first argument
			l.PushValue(3) // Pass the value as second argument
			l.Call(2, 0)
			return 0
		}
		l.Pop(2)

		// Look for setters to fail if the field is read-only.
		lua.SubTable(l.State, -1, "__getters")
		l.Field(-1, key)
		if l.IsFunction(-1) {
			lua.Errorf(l.State, "field '%s' is read-only for entity type '%s'", key, typ)
		}
		l.Pop(2)

		// Look also for methods to fail if they are tried to be reassigned.
		lua.SubTable(l.State, -1, "__methods")
		l.Field(-1, key)
		if l.IsFunction(-1) {
			lua.Errorf(l.State, "cannot set method field '%s' of entity type '%s'", key, typ)
		}
		l.Pop(2)

		// The field was not found. Create a new one.
		l.PushValue(2) // key
		l.PushValue(3) // value
		l.RawSet(1)

		return 0
	})
	l.SetField(-2, "__newindex")

	// Prototype is ready. Remove it from the stack.
	l.Pop(1)

	return false
}

// DeclareEntityConstructor declares a constructor for an entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareEntityConstructor(typ ScriptEntityType, funcName string, f LuaFunction) {
	lua.MetaTableNamed(l.State, typ.RegistryName())
	if !l.IsTable(-1) {
		lua.Errorf(l.State, "entity type %s not declared", typ)
	}
	l.Pop(1)
	l.PushFunction(f)
	l.SetGlobal(funcName)
}

// DeclareEntityEqual declares an equality metamethod for an entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareEntityEqual(typ ScriptEntityType, f LuaFunction) {
	lua.MetaTableNamed(l.State, typ.RegistryName())
	if !l.IsTable(-1) {
		lua.Errorf(l.State, "entity type %s not declared", typ)
	}
	l.PushFunction(f)
	l.SetField(-2, "__eq")
}

// DeclareEntityGetter declares a getter for an entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareEntityGetter(typ ScriptEntityType, field string, f LuaFunction) {
	lua.MetaTableNamed(l.State, typ.RegistryName())
	if !l.IsTable(-1) {
		lua.Errorf(l.State, "entity type %s not declared", typ)
	}
	lua.SubTable(l.State, -1, "__getters")
	l.PushFunction(func(l *LuaInterpreter) int {
		l.CheckEntity(1, typ)
		return f(l)
	})
	l.SetField(-2, field)
	l.Pop(2)
}

// DeclareEntityMethod declares a method for an entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareEntityMethod(typ ScriptEntityType, method string, f LuaFunction) {
	lua.MetaTableNamed(l.State, typ.RegistryName())
	if !l.IsTable(-1) {
		lua.Errorf(l.State, "entity type %s not declared", typ)
	}
	lua.SubTable(l.State, -1, "__methods")
	l.PushFunction(func(l *LuaInterpreter) int {
		l.CheckEntity(1, typ)
		return f(l)
	})
	l.SetField(-2, method)
	l.Pop(2)
}

// DeclareEntitySetter declares a setter for an entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareEntitySetter(typ ScriptEntityType, field string, f LuaFunction) {
	lua.MetaTableNamed(l.State, typ.RegistryName())
	if !l.IsTable(-1) {
		lua.Errorf(l.State, "entity type %s not declared", typ)
	}
	lua.SubTable(l.State, -1, "__setters")
	l.PushFunction(func(li *LuaInterpreter) int {
		li.CheckEntity(1, typ)
		return f(li)
	})
	l.SetField(-2, field)
	l.Pop(2)
}

// DeclareEntityToString declares a __tostring metamethod for an entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareEntityToString(typ ScriptEntityType, f LuaFunction) {
	lua.MetaTableNamed(l.State, typ.RegistryName())
	if !l.IsTable(-1) {
		lua.Errorf(l.State, "entity type %s not declared", typ)
	}
	l.PushFunction(f)
	l.SetField(-2, "__tostring")
}

// DeclareExportFunction declares a function that can be called from Lua using the interpreter to
// export entity values to other modules.
func (l *LuaInterpreter) DeclareExportFunction(handler ScriptEntityHandler) {
	l.Global("export")
	exists := l.IsFunction(-1)
	l.Pop(1)
	if exists {
		return
	}

	l.PushFunction(func(l *LuaInterpreter) int {
		l.WithEachTableItem(1, func(key string) {
			typ := l.EntityTypeOf(-1)
			if typ == "" {
				lua.ArgumentError(l.State, -1, "expected entity")
			}
			val := l.CheckEntity(-1, typ)
			handler(ScriptNamedEntityValue{
				Name:              key,
				ScriptEntityValue: ScriptEntityValue{Type: typ, UserData: val},
			})

			// For convenience, each exported element is also declared as global in the script where
			// it is used.
			l.PushValue(-1)
			l.SetGlobal(key)
		})
		return 0
	})
	l.SetGlobal("export")
}

// DeclareFutureType declares the type of a Future in the Lua interpreter.
func (l *LuaInterpreter) DeclareFutureType() {
	if l.DeclareEntityType(ScriptEntityFuture) {
		return
	}
	l.DeclareEntityMethod(ScriptEntityFuture, "wait", func(l *LuaInterpreter) int {
		f := l.CheckEntity(1, ScriptEntityFuture).(Future)
		f.Wait()
		return 0
	})
}

// DeclareImportFunction declares a function that can be called from Lua using the interpreter to
// import entity values from other modules.
func (l *LuaInterpreter) DeclareImportFunction(handler ScriptImportHandler) {
	l.Global("import")
	exists := l.IsFunction(-1)
	l.Pop(1)
	if exists {
		return
	}

	l.PushFunction(func(l *LuaInterpreter) int {
		// For convenience, we admit also string for script reference parameter.
		var module ResourceRef
		if l.IsString(1) {
			var err error
			module, err = ParseResourceRef(lua.CheckString(l.State, 1))
			if err != nil {
				lua.ArgumentError(l.State, 1, "invalid script reference")
			}
		} else {
			module = l.CheckEntity(1, ScriptEntityRef).(ResourceRef)
		}

		l.NewTable()
		handler(module, func(exp ScriptNamedEntityValue) {
			l.PushEntity(exp.Type, exp.UserData)
			l.SetField(-2, exp.Name)

		})
		return 1
	})
	l.SetGlobal("import")
}

// DeclareMusicType declares the type of a Music in the Lua interpreter.
func (l *LuaInterpreter) DeclareMusicType(app *App) {
	if l.DeclareEntityType(ScriptEntityMusic) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityMusic, "music", func(l *LuaInterpreter) int {
		// Admit ref as a string for convenience.
		var ref ResourceRef
		if l.IsString(1) {
			var err error
			ref, err = ParseResourceRef(lua.CheckString(l.State, 1))
			if err != nil {
				lua.ArgumentError(l.State, 1, "invalid resource reference")
			}
		} else {
			ref = l.CheckEntity(1, ScriptEntityRef).(ResourceRef)
		}

		music := NewMusic(ref)
		l.PushEntity(ScriptEntityMusic, music)
		return 1
	})
	l.DeclareEntityMethod(ScriptEntityMusic, "play", func(l *LuaInterpreter) int {
		app.RunCommand(MusicPlay{
			Music: l.CheckEntity(1, ScriptEntityMusic).(*Music),
		})
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityMusic, "stop", func(l *LuaInterpreter) int {
		// TODO: music stopped from the music entity is odd. In fact, everything about music
		// is really odd. Music should be a member of room and sound as long as the room is active.
		// And stop music should be a command to the room.
		app.RunCommand(MusicStop{})
		return 0
	})
}

// DeclareObjectDefaultsType declares the type of an ObjectDefaults in the Lua interpreter.
func (l *LuaInterpreter) DeclareObjectDefaultsType(app *App, script *Script) {
	if l.DeclareEntityType(ScriptEntityObjectDefaults) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityObjectDefaults, "defaults", func(l *LuaInterpreter) int {
		defaults := new(ObjectDefaults)
		defaults.Script = script
		l.PushEntity(ScriptEntityObjectDefaults, defaults)
		defaults.CallRecv = l.RegisterCallReceiver(-1)
		if err := app.SetObjectDefaults(defaults); err != nil {
			lua.Errorf(l.State, "error setting object defaults: %s", err.Error())
		}
		return 1
	})
}

// DeclareObjectType declares the object entity type in the Lua interpreter.
func (l *LuaInterpreter) DeclareObjectType() {
	l.DeclareClassType()
	l.DeclareRectType()
	l.DeclareObjectStateType()
	l.DeclarePositionType()

	// Objects are constructed using the `object` method of `room` entity.
	if l.DeclareEntityType(ScriptEntityObject) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityObject, "object", func(l *LuaInterpreter) int {
		obj := NewObject()
		l.WithEachTableItem(1, func(key string) {
			switch key {
			case "class":
				obj.Class = l.CheckEntity(-1, ScriptEntityClass).(ObjectClass)
			case "hotspot":
				obj.Hotspot = l.CheckEntity(-1, ScriptEntityRect).(Rectangle)
			case "name":
				obj.Name = lua.CheckString(l.State, -1)
			case "pos":
				obj.Pos = l.CheckEntity(-1, ScriptEntityPos).(Position)
			case "sprites":
				obj.Sprites = l.CheckEntity(-1, ScriptEntityRef).(ResourceRef)
			case "state":
				obj.State = lua.CheckString(l.State, -1)
			case "usepos":
				obj.UsePos = l.CheckEntity(-1, ScriptEntityPos).(Position)
			case "usedir":
				obj.UseDir = l.CheckEntity(-1, ScriptEntityDir).(Direction)
			default:
				switch l.EntityTypeOf(-1) {
				case ScriptEntityState:
					st := l.CheckEntity(-1, ScriptEntityState).(*ObjectState)
					obj.States[key] = st
				default:
					lua.ArgumentError(l.State, -1, fmt.Sprintf(
						"unexpected field '%s' in object constructor", key))
				}
			}
		})
		l.PushEntity(ScriptEntityObject, obj)

		// Set the states as fields of the object with the same key as they was declared in the
		// constructor input.
		for k, st := range obj.States {
			l.PushEntity(ScriptEntityState, st)
			l.SetField(-2, k)
		}
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityObject, "class", func(l *LuaInterpreter) int {
		obj := l.CheckEntity(1, ScriptEntityObject).(*Object)
		l.PushEntity(ScriptEntityClass, obj.Class)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityObject, "hotspot", func(l *LuaInterpreter) int {
		obj := l.CheckEntity(1, ScriptEntityObject).(*Object)
		l.PushEntity(ScriptEntityRect, obj.Hotspot)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityObject, "name", func(l *LuaInterpreter) int {
		obj := l.CheckEntity(1, ScriptEntityObject).(*Object)
		l.PushString(obj.Name)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityObject, "owner", func(l *LuaInterpreter) int {
		obj := l.CheckEntity(1, ScriptEntityObject).(*Object)
		l.PushEntity(ScriptEntityActor, obj.Owner)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityObject, "pos", func(l *LuaInterpreter) int {
		obj := l.CheckEntity(1, ScriptEntityObject).(*Object)
		l.PushEntity(ScriptEntityPos, obj.Pos)
		return 1
	})

	l.DeclareEntityToString(ScriptEntityObject, func(l *LuaInterpreter) int {
		obj := l.CheckEntity(1, ScriptEntityObject).(*Object)
		l.PushString(fmt.Sprintf("object:(name=%s)", obj.Name))
		return 1
	})
}

// DeclareObjectStateType declares the type of an ObjectState in the Lua interpreter.
func (l *LuaInterpreter) DeclareObjectStateType() {
	l.DeclareAnimType()

	if l.DeclareEntityType(ScriptEntityState) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityState, "state", func(l *LuaInterpreter) int {
		st := &ObjectState{}
		l.WithOptionalField(1, "anim", func() {
			st.Anim = l.CheckEntity(-1, ScriptEntityAnimation).(*Animation)
		})
		l.PushEntity(ScriptEntityState, st)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityState, "anim", func(l *LuaInterpreter) int {
		st := l.CheckEntity(1, ScriptEntityState).(*ObjectState)
		l.PushEntity(ScriptEntityAnimation, st.Anim)
		return 1
	})
}

// DeclarePositionType declares the type of a Position in the Lua interpreter.
func (l *LuaInterpreter) DeclarePositionType() {
	if l.DeclareEntityType(ScriptEntityPos) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityPos, "pos",
		func(l *LuaInterpreter) int {
			x := l.CheckFieldInteger(1, "x")
			y := l.CheckFieldInteger(1, "y")
			l.PushEntity(ScriptEntityPos, Position{X: x, Y: y})
			return 1
		},
	)
	l.DeclareEntityGetter(ScriptEntityPos, "x", func(l *LuaInterpreter) int {
		pos := l.CheckEntity(1, ScriptEntityPos).(Position)
		l.PushInteger(pos.X)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityPos, "y", func(l *LuaInterpreter) int {
		pos := l.CheckEntity(1, ScriptEntityPos).(Position)
		l.PushInteger(pos.Y)
		return 1
	})
	l.DeclareEntityToString(ScriptEntityPos, func(l *LuaInterpreter) int {
		pos := l.CheckEntity(1, ScriptEntityPos).(Position)
		l.PushString(fmt.Sprintf("(%d, %d)", pos.X, pos.Y))
		return 1
	})
}

// DeclareRectType declares the type of a Rectangle in the Lua interpreter.
func (l *LuaInterpreter) DeclareRectType() {
	if l.DeclareEntityType(ScriptEntityRect) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityRect, "rect",
		func(l *LuaInterpreter) int {
			x := l.CheckFieldInteger(1, "x")
			y := l.CheckFieldInteger(1, "y")
			w := l.CheckFieldInteger(1, "w")
			h := l.CheckFieldInteger(1, "h")
			l.PushEntity(ScriptEntityRect, NewRect(x, y, w, h))
			return 1
		},
	)
	l.DeclareEntityGetter(ScriptEntityRect, "x", func(l *LuaInterpreter) int {
		r := l.CheckEntity(1, ScriptEntityRect).(Rectangle)
		l.PushInteger(r.Pos.X)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityRect, "y", func(l *LuaInterpreter) int {
		r := l.CheckEntity(1, ScriptEntityRect).(Rectangle)
		l.PushInteger(r.Pos.Y)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityRect, "w", func(l *LuaInterpreter) int {
		r := l.CheckEntity(1, ScriptEntityRect).(Rectangle)
		l.PushInteger(r.Size.W)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityRect, "h", func(l *LuaInterpreter) int {
		r := l.CheckEntity(1, ScriptEntityRect).(Rectangle)
		l.PushInteger(r.Size.H)
		return 1
	})
}

// DeclareReferenceType declares the type of a ResourceReference in the Lua interpreter.
func (l *LuaInterpreter) DeclareReferenceType() {
	if l.DeclareEntityType(ScriptEntityRef) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityRef, "ref",
		func(l *LuaInterpreter) int {
			ref, err := ParseResourceRef(lua.CheckString(l.State, 1))
			if err != nil {
				lua.ArgumentError(l.State, 1, fmt.Sprintf(
					"invalid format in resource reference: %s", err))
			}
			l.PushEntity(ScriptEntityRef, ref)
			return 1
		},
	)
	l.DeclareEntityEqual(ScriptEntityRef, func(l *LuaInterpreter) int {
		if l.EntityTypeOf(1) != ScriptEntityRef || l.EntityTypeOf(2) != ScriptEntityRef {
			l.PushBoolean(false)
			return 1
		}
		a := l.CheckEntity(1, ScriptEntityRef).(ResourceRef)
		b := l.CheckEntity(2, ScriptEntityRef).(ResourceRef)
		l.PushBoolean(a == b)
		return 1
	})
	l.DeclareEntityToString(ScriptEntityRef, func(l *LuaInterpreter) int {
		ref := l.CheckEntity(1, ScriptEntityRef).(ResourceRef)
		l.PushString(ref.String())
		return 1
	})
}

// DeclareRoomType declares the type of a Room in the Lua interpreter.
func (l *LuaInterpreter) DeclareRoomType(app *App, script *Script) {
	l.DeclareClassType()
	l.DeclareObjectType()
	l.DeclareReferenceType()

	if l.DeclareEntityType(ScriptEntityRoom) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityRoom, "room",
		func(l *LuaInterpreter) int {
			room := new(Room)
			room.script = script
			objects := make(map[string]*Object)
			l.WithEachTableItem(1, func(key string) {
				switch key {
				case "background":
					room.Background = l.CheckEntity(-1, ScriptEntityRef).(ResourceRef)
				case "walkboxes":
					var walkboxes []*WalkBox
					l.WithEachTableItem(-1, func(k string) {
						wb := l.CheckEntity(-1, ScriptEntityWalkBox).(*WalkBox)
						wb.walkBoxID = k
						wb.room = room
						walkboxes = append(walkboxes, wb)
					})
					matrix := NewWalkBoxMatrix(walkboxes)
					room.wbmatrix = matrix
				default:
					switch l.EntityTypeOf(-1) {
					case ScriptEntityObject:
						obj := l.CheckEntity(-1, ScriptEntityObject).(*Object)
						room.DeclareObject(obj)
						objects[key] = obj
					default:
						lua.ArgumentError(l.State, -1, fmt.Sprintf(
							"unexpected field '%s' in room constructor", key))
					}
				}
			})

			err := app.DeclareRoom(room)
			if err != nil {
				lua.Errorf(l.State, "error declaring room: %s", err)
			}

			// Set the objects as fields of the room with the same key as they was declared in the
			// constructor input.
			l.PushEntity(ScriptEntityRoom, room)
			for k, obj := range objects {
				l.PushEntity(ScriptEntityObject, obj)
				obj.CallRecv = l.RegisterCallReceiver(-1)
				l.SetField(-2, k)
			}

			// Do the same for the walkboxes.
			if room.wbmatrix != nil {
				for _, wb := range room.wbmatrix.walkBoxes {
					l.PushEntity(ScriptEntityWalkBox, wb)
					l.SetField(-2, wb.walkBoxID)
				}
			}

			// Declare the room as call receiver.
			room.callrecv = l.RegisterCallReceiver(-1)

			return 1
		},
	)
	l.DeclareEntityGetter(ScriptEntityRoom, "background", func(l *LuaInterpreter) int {
		room := l.CheckEntity(1, ScriptEntityRoom).(*Room)
		l.PushEntity(ScriptEntityRef, room.Background)
		return 1
	})
	l.DeclareEntityMethod(ScriptEntityRoom, "camfollow", func(l *LuaInterpreter) int {
		l.CheckEntity(1, ScriptEntityRoom)
		actor := l.CheckEntity(2, ScriptEntityActor).(*Actor)
		// TODO: hack obtaining viewport
		_, err := app.RunCommand(RoomCameraFollowActor(&app.viewport, actor)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error making camera follow actor: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityRoom, "camto", func(l *LuaInterpreter) int {
		l.CheckEntity(1, ScriptEntityRoom)
		pos := lua.CheckInteger(l.State, 2)
		// TODO: hack obtaining viewport
		_, err := app.RunCommand(RoomCameraTo(&app.viewport, pos)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error moving camera to position: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityRoom, "show", func(l *LuaInterpreter) int {
		app.RunCommand(RoomShow{
			Room: l.CheckEntity(1, ScriptEntityRoom).(*Room),
		})
		return 0
	})
}

// DeclareSentenceChoiceType declares the type of a SentenceChoice in the Lua interpreter.
func (l *LuaInterpreter) DeclareSentenceChoiceType(app *App) {
	if l.DeclareEntityType(ScriptEntitySentenceChoice) {
		return
	}
	l.DeclareEntityMethod(ScriptEntitySentenceChoice, "add", func(l *LuaInterpreter) int {
		choice := l.CheckEntity(1, ScriptEntitySentenceChoice).(*ControlSentenceChoice)
		sentence := lua.CheckString(l.State, 2)
		_, err := app.RunCommand(SentenceChoiceAdd(choice, sentence)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error adding sentence to choice: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntitySentenceChoice, "wait", func(l *LuaInterpreter) int {
		choice := l.CheckEntity(1, ScriptEntitySentenceChoice).(*ControlSentenceChoice)
		ret, err := app.RunCommand(SentenceChoiceWait(choice, false)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error waiting sentence choice: %s", err.Error())
		}
		sentence := ret.(IndexedSentence)
		l.PushInteger(sentence.Index + 1)
		l.PushString(sentence.Sentence)
		return 2
	})
	l.DeclareEntityMethod(ScriptEntitySentenceChoice, "waitsay", func(l *LuaInterpreter) int {
		choice := l.CheckEntity(1, ScriptEntitySentenceChoice).(*ControlSentenceChoice)
		ret, err := app.RunCommand(SentenceChoiceWait(choice, true)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error waiting sentence choice: %s", err.Error())
		}
		sentence := ret.(IndexedSentence)
		l.PushInteger(sentence.Index + 1)
		return 1
	})
}

// DeclareSizeType declares the type of a Size in the Lua interpreter.
func (l *LuaInterpreter) DeclareSizeType() {
	if l.DeclareEntityType(ScriptEntitySize) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntitySize, "size",
		func(l *LuaInterpreter) int {
			w := l.CheckFieldInteger(1, "w")
			h := l.CheckFieldInteger(1, "h")
			l.PushEntity(ScriptEntitySize, Size{w, h})
			return 1
		},
	)
	l.DeclareEntityGetter(ScriptEntitySize, "w", func(l *LuaInterpreter) int {
		s := l.CheckEntity(1, ScriptEntitySize).(Size)
		l.PushInteger(s.W)
		return 1
	})
	l.DeclareEntityGetter(ScriptEntitySize, "h", func(l *LuaInterpreter) int {
		s := l.CheckEntity(1, ScriptEntitySize).(Size)
		l.PushInteger(s.H)
		return 1
	})
	l.DeclareEntityToString(ScriptEntitySize, func(l *LuaInterpreter) int {
		s := l.CheckEntity(1, ScriptEntitySize).(Size)
		l.PushString(fmt.Sprintf("(%d, %d)", s.W, s.H))
		return 1
	})
}

// DeclareSoundType declares the type of a Sound in the Lua interpreter.
func (l *LuaInterpreter) DeclareSoundType(app *App) {
	l.DeclareReferenceType()

	if l.DeclareEntityType(ScriptEntitySound) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntitySound, "sound", func(l *LuaInterpreter) int {
		// Admit ref as a string for convenience.
		var ref ResourceRef
		if l.IsString(1) {
			var err error
			ref, err = ParseResourceRef(lua.CheckString(l.State, 1))
			if err != nil {
				lua.ArgumentError(l.State, 1, "invalid resource reference")
			}
		} else {
			ref = l.CheckEntity(1, ScriptEntityRef).(ResourceRef)
		}
		l.PushEntity(ScriptEntitySound, NewSound(ref))
		return 1
	})
	l.DeclareEntityMethod(ScriptEntitySound, "play", func(l *LuaInterpreter) int {
		sound := l.CheckEntity(1, ScriptEntitySound).(*Sound)
		app.RunCommand(SoundPlay{Sound: sound})
		return 0
	})
	l.DeclareEntityMethod(ScriptEntitySound, "stop", func(l *LuaInterpreter) int {
		sound := l.CheckEntity(1, ScriptEntitySound).(*Sound)
		app.RunCommand(SoundStop{Sound: sound})
		return 0
	})
}

// DeclareUtilityFunctions declares the utility functions in the Lua interpreter.
func (l *LuaInterpreter) DeclareUtilityFunctions(app *App) {
	l.PushFunction(func(l *LuaInterpreter) int {
		millis := lua.CheckInteger(l.State, 1)
		time.Sleep(time.Duration(millis) * time.Millisecond)
		return 0
	})
	l.SetGlobal("sleep")
}

// DeclareWalkBoxType declares the type of a Walkbox in the Lua interpreter.
func (l *LuaInterpreter) DeclareWalkBoxType(app *App) {
	l.DeclarePositionType()
	if l.DeclareEntityType(ScriptEntityWalkBox) {
		return
	}
	l.DeclareEntityConstructor(ScriptEntityWalkBox, "walkbox", func(l *LuaInterpreter) int {
		var vertices [4]Position
		scale := 1.0
		l.WithField(1, "vertices", func() {
			l.WithEachArrayItem(-1, func(idx int) {
				if idx > 4 {
					lua.ArgumentError(l.State, 1, "walkbox must have 4 vertices")
				}
				vertices[idx-1] = l.CheckEntity(-1, ScriptEntityPos).(Position)
			})
		})
		l.WithOptionalField(1, "scale", func() {
			scale = lua.CheckNumber(l.State, -1)
		})

		walkbox := NewWalkBox("", vertices, float32(scale))

		l.WithOptionalField(1, "enabled", func() {
			walkbox.enabled = l.State.ToBoolean(-1)
		})
		l.PushEntity(ScriptEntityWalkBox, walkbox)
		return 1
	})
	l.DeclareEntityMethod(ScriptEntityWalkBox, "enable", func(l *LuaInterpreter) int {
		w := l.CheckEntity(1, ScriptEntityWalkBox).(*WalkBox)
		_, err := app.RunCommand(EnableWalkBox(w)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error enabling walkbox: %s", err.Error())
		}
		return 0
	})
	l.DeclareEntityMethod(ScriptEntityWalkBox, "disable", func(l *LuaInterpreter) int {
		w := l.CheckEntity(1, ScriptEntityWalkBox).(*WalkBox)
		_, err := app.RunCommand(DisableWalkBox(w)).Wait()
		if err != nil {
			lua.Errorf(l.State, "error disabling walkbox: %s", err.Error())
		}
		return 0
	})
}

// EntityTypeOf returns the entity type of the entity at the given index.
func (l *LuaInterpreter) EntityTypeOf(index int) ScriptEntityType {
	index = l.AbsIndex(index)
	if !l.MetaTable(index) {
		return ""
	}
	l.Field(-1, "__type")
	if !l.IsString(-1) {
		lua.Errorf(l.State, "expected entity type")
	}
	typ := lua.CheckString(l.State, -1)
	l.Pop(2)
	return ScriptEntityType(typ)
}

// PushFunction pushes a LuaFunction into the stack.
func (l *LuaInterpreter) PushFunction(f LuaFunction) {
	l.PushGoFunction(func(state *lua.State) int {
		return f(WrapInterpreter(state))
	})
}

// PushEntity pushes an entity into the stack. The entity must be a user data. This function will
// configure the metatable of the entity to the one of the given entity type.
func (l *LuaInterpreter) PushEntity(typ ScriptEntityType, obj any) {
	l.NewTable()

	// Configure the user data of the entity table.
	l.PushUserData(obj)
	l.SetField(-2, "__userdata")

	// Configure the prototype as metatable of the entity table.
	lua.MetaTableNamed(l.State, typ.RegistryName())
	l.SetMetaTable(-2)
}

// RegisterCallReceiver registers the entity at index as a call receiver.
func (l *LuaInterpreter) RegisterCallReceiver(index int) (id ScriptCallReceiver) {
	index = l.AbsIndex(index)
	id = NewScriptCallReceiver()
	lua.NewMetaTable(l.State, "pctk.callreceivers")
	l.PushValue(index)
	l.SetField(-2, id.String())
	l.Pop(1)
	return
}

// ToIntegerField returns the field of the table at index as an integer.
func (l *LuaInterpreter) ToIntegerField(index int, name string) (i int, ok bool) {
	l.Field(index, name)
	i, ok = l.ToInteger(-1)
	l.Pop(1)
	return
}

// WithField calls the function f with the field of the table at index. It pops the field after the
// function is called.
func (l *LuaInterpreter) WithField(index int, name string, f func()) {
	l.Field(index, name)
	f()
	l.Pop(1)
}

// WithOptionalField calls the function f if the field of the table at index exists. It returns true
// if the field exists, and false otherwise. It pops the field after the function is called.
func (l *LuaInterpreter) WithOptionalField(index int, name string, f func()) bool {
	if index > l.Top() || l.IsNil(index) {
		return false
	}
	l.Field(index, name)
	if l.IsNil(-1) {
		l.Pop(1)
		return false
	}
	f()
	l.Pop(1)
	return true
}

// WithEachArrayItem calls the function f for each item in the array at index.
func (l *LuaInterpreter) WithEachArrayItem(index int, f func(idx int)) {
	index = l.AbsIndex(index)
	l.PushNil()
	for l.Next(index) {
		idx := lua.CheckInteger(l.State, -2)
		f(idx)
		l.Pop(1)
	}
}

// WithEachTableItem calls the function f for each item in the table at index.
func (l *LuaInterpreter) WithEachTableItem(index int, f func(key string)) {
	index = l.AbsIndex(index)
	l.PushNil()
	for l.Next(index) {
		key := lua.CheckString(l.State, -2)
		f(key)
		l.Pop(1)
	}
}
