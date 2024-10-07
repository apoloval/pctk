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
	call := obj.Room.script.CallMethod(
		cmd.Object.CallRecv,
		cmd.Function,
		cmd.Args,
	)
	call = Recover(call, func(err error) Future {
		if app.defaults == nil {
			return AlreadyFailed(err)
		}
		return app.defaults.CallFunction(cmd.Function, cmd.Args)
	})
	done.Bind(call)
}
