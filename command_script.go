package pctk

import "log"

// ScriptRun is a command to run a script.
type ScriptRun struct {
	ScriptRef ResourceRef
}

func (c ScriptRun) Execute(app *App, prom *Promise) {
	script, ok := app.scripts[c.ScriptRef]
	if ok {
		prom.CompleteWithValue(script)
		return
	}

	// The script was not loaded yet, so we load and execute it now.
	script = app.res.LoadScript(c.ScriptRef)
	if script == nil {
		log.Panicf("Script not found: %s", c.ScriptRef)
	}
	app.scripts[c.ScriptRef] = script
	script.Run(app)
	prom.CompleteWithValue(c)
}

// ScriptCall is a command to call a script function.
type ScriptCall struct {
	ScriptRef ResourceRef
	Receiver  ScriptCallReceiver
	Method    string
	Args      []ScriptEntityValue
}

func (c ScriptCall) Execute(app *App, prom *Promise) {
	script, ok := app.scripts[c.ScriptRef]
	if !ok {
		log.Panicf("Script not found: %s", c.ScriptRef)
	}

	prom.Bind(script.CallMethod(c.Receiver, c.Method, c.Args))
}

// ScriptImport is a command to import a script.
type ScriptImport struct {
	ScriptRef ResourceRef
	Handler   ScriptEntityHandler
}

func (c ScriptImport) Execute(app *App, prom *Promise) {
	script, ok := app.scripts[c.ScriptRef]
	if !ok {
		script = app.res.LoadScript(c.ScriptRef)
		if script == nil {
			log.Panicf("Script not found: %s", c.ScriptRef)
		}
		script.Run(app)
	}

	for _, exp := range script.Exports() {
		c.Handler(exp)
	}

	prom.CompleteWithValue(script)
}
