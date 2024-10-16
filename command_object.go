package pctk

// ObjectDeclare is a command that will declare a new object with the given properties.
type ObjectDeclare struct {
	Room   *Room
	Object *Object
}

func (cmd ObjectDeclare) Execute(app *App, done *Promise) {
	cmd.Room.DeclareObject(cmd.Object)
	done.Complete()
}

// ObjectCall is a command that will execute a script function of an object.
type ObjectCall struct {
	Object   *Object
	Function string
	Args     []ScriptEntityValue
}

func (cmd ObjectCall) Execute(app *App, done *Promise) {
	obj := cmd.Object
	cb := obj.FindCallback(cmd.Function)
	if cb != nil {
		done.Bind(cb.Invoke(cmd.Args))
		return
	}
	args := append([]ScriptEntityValue{{
		Type:     ScriptEntityObject,
		UserData: cmd.Object,
	}}, cmd.Args...)
	done.Bind(app.defaults.CallFunction(cmd.Function, args))
}
