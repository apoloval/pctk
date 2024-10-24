package pctk

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

// ObjectSetState set the state  of the object.
func ObjectSetState(st *ObjectState) CommandFunc {
	return func(a *App) (any, error) {
		st.Object.State = st
		return nil, nil
	}
}

// ObjectEnableClass is a command that will enable a class of an object.
func ObjectEnableClass(obj *Object, class ObjectClass) CommandFunc {
	return func(a *App) (any, error) {
		obj.EnableClass(class)
		return nil, nil
	}
}

// ObjectDisableClass is a command that will disable a class of an object.
func ObjectDisableClass(obj *Object, class ObjectClass) CommandFunc {
	return func(a *App) (any, error) {
		obj.DisableClass(class)
		return nil, nil
	}
}
