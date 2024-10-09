package pctk

import (
	"testing"

	"github.com/Shopify/go-lua"
	"github.com/stretchr/testify/assert"
)

func TestDeclareColorType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareColorType()

	assert.NoError(t, lua.DoString(l.State, `
		c = color { r=255, g=128, b=64, a=32 }
		assert(c.r == 255)
		assert(c.g == 128)
		assert(c.b == 64)
		assert(c.a == 32)
		assert(tostring(c) == "rgba(255, 128, 64, 32)")
	`))
}

func TestDirectionType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareDirectionType()

	assert.NoError(t, lua.DoString(l.State, `
		print(tostring(UP))
		assert(tostring(UP) == "Up")
		assert(tostring(DOWN) == "Down")
		assert(tostring(LEFT) == "Left")
		assert(tostring(RIGHT) == "Right")
	`))
}

func TestDeclareExportFunction(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareColorType()
	l.DeclarePositionType()

	exported := make(map[string]ScriptEntityValue)
	l.DeclareExportFunction(func(exp ScriptNamedEntityValue) {
		exported[exp.Name] = exp.ScriptEntityValue
	})

	assert.NoError(t, lua.DoString(l.State, `
		mycol = color { r=255, g=128, b=64, a=32 }
		mypos = pos { x=42, y=24 }
		export {
			thecolor = mycol,
			thepos = mypos,
		}
	`))
	assert.Equal(t, ScriptEntityColor, exported["thecolor"].Type)
	assert.Equal(t, Color{R: 255, G: 128, B: 64, A: 32}, exported["thecolor"].UserData)
	assert.Equal(t, ScriptEntityPos, exported["thepos"].Type)
	assert.Equal(t, Position{X: 42, Y: 24}, exported["thepos"].UserData)
}

func TestDeclareImportFunction(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareColorType()
	l.DeclarePositionType()

	l.DeclareImportFunction(func(script ResourceRef, handler ScriptEntityHandler) {
		assert.Equal(t, "resources:/scripts/mymodule", script.String())
		handler(ScriptNamedEntityValue{
			Name: "thecolor",
			ScriptEntityValue: ScriptEntityValue{
				Type:     ScriptEntityColor,
				UserData: Color{R: 255, G: 128, B: 64, A: 32},
			},
		})
		handler(ScriptNamedEntityValue{
			Name: "thepos",
			ScriptEntityValue: ScriptEntityValue{
				Type:     ScriptEntityPos,
				UserData: Position{X: 42, Y: 24},
			},
		})
	})

	assert.NoError(t, lua.DoString(l.State, `
		mymodule = import("resources:/scripts/mymodule")

		assert(mymodule.thecolor.r == 255)
		assert(mymodule.thecolor.g == 128)
		assert(mymodule.thecolor.b == 64)
		assert(mymodule.thecolor.a == 32)
		assert(mymodule.thepos.x == 42)
		assert(mymodule.thepos.y == 24)
	`))
}

func TestObjectType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareObjectType()

	assert.NoError(t, lua.DoString(l.State, `
		o = object { 
			name = "bar",
			class = APPLICABLE,
			pos = pos { x=42, y=24 },
			hotspot = rect { x=0, y=0, w=32, h=48 },
			default = state {
				anim = fixedanim { row = 3, col = 5 },
			},
		}
		assert(o.name == "bar")
		assert(o.class == APPLICABLE)
		assert(o.pos.x == 42)
		assert(o.pos.y == 24)
		assert(o.hotspot.x == 0)
		assert(o.hotspot.y == 0)
		assert(o.hotspot.w == 32)
		assert(o.hotspot.h == 48)
		assert(o.default ~= nil)
		assert(tostring(o) == "object:(name=bar)")
	`))
}

func TestDeclarePositionType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclarePositionType()

	assert.NoError(t, lua.DoString(l.State, `
		p = pos { x=42, y=24 }
		assert(p.x == 42)
		assert(p.y == 24)
		assert(tostring(p) == "(42, 24)")
	`))
}

func TestDeclareReferenceType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareReferenceType()

	assert.NoError(t, lua.DoString(l.State, `
		r = ref("foo:/bar")
		assert(tostring(r) == "foo:/bar")
	`))
}

func TestDeclareRoomType(t *testing.T) {
	// TODO
	t.Skip("Fix the deadlock running app.init() --> rl.InitWindow(...)")
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)
	app := New(NewResourceBundle())
	l.DeclareRoomType(app, nil)

	assert.NoError(t, lua.DoString(l.State, `
		r = room { 
			background = ref("resources:/backgrounds/foobar"),
			window = object {
				name = "broken window",
			},
		}
		assert(r.background == ref("resources:/backgrounds/foobar"))
		assert(r.window.name == "broken window")
	`))
}

func TestDeclareSizeType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	l.DeclareSizeType()

	assert.NoError(t, lua.DoString(l.State, `
		s = size { w=42, h=24 }
		assert(s.w == 42)
		assert(s.h == 24)
		assert(tostring(s) == "(42, 24)")
	`))
}

func TestDeclareObjectType(t *testing.T) {
	l := NewLuaInterpreter()
	lua.BaseOpen(l.State)

	type userData struct {
		str string
	}
	val := &userData{str: "Hello World!"}
	l.DeclareEntityType(ScriptEntityType("foobar"))
	l.DeclareEntityConstructor(ScriptEntityType("foobar"), "foobar",
		func(l *LuaInterpreter) int {
			l.PushEntity(ScriptEntityType("foobar"), val)
			return 1
		},
	)
	l.DeclareEntityMethod(ScriptEntityType("foobar"), "len", func(l *LuaInterpreter) int {
		v := l.CheckEntity(1, ScriptEntityType("foobar")).(*userData)
		l.PushInteger(len(v.str))
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityType("foobar"), "first", func(l *LuaInterpreter) int {
		v := l.CheckEntity(1, ScriptEntityType("foobar")).(*userData)
		l.PushString(string(v.str[0]))
		return 1
	})
	l.DeclareEntityGetter(ScriptEntityType("foobar"), "str", func(l *LuaInterpreter) int {
		v := l.CheckEntity(1, ScriptEntityType("foobar")).(*userData)
		l.PushString(v.str)
		return 1
	})
	l.DeclareEntitySetter(ScriptEntityType("foobar"), "str", func(l *LuaInterpreter) int {
		v := l.CheckEntity(1, ScriptEntityType("foobar")).(*userData)
		v.str = lua.CheckString(l.State, 2)
		return 0
	})
	err := lua.DoString(l.State, `
		f1 = foobar()
		assert(f1:len() == 12)				-- Call method
		assert(f1.first == "H")				-- Call getter
		assert(f1.str == "Hello World!")	-- Call getter with setter

		f1.bar = "baz"						-- Set custom field
		assert(f1.bar == "baz")			    -- Get custom field	

		f1.str = "Hello"					-- Call setter
		assert(f1:len() == 5)
		assert(f1.str == "Hello")
	`)
	assert.NoError(t, err)

	err = lua.DoString(l.State, `
		f2 = foobar()
		f2.len = 42 -- Set method
	`)
	assert.Error(t, err)

	err = lua.DoString(l.State, `
		f3 = foobar()
		f3.first = "X" -- Set getter without setter (read-only)
	`)
	assert.Error(t, err)
}
