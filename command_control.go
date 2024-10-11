package pctk

import (
	"errors"
)

// ControlPaneEnable is a command that will enable the control panel.
func ControlPaneEnable() CommandFunc {
	return func(a *App) (any, error) {
		a.control.Enable()
		return nil, nil
	}
}

// ControlPaneDisable is a command that will disable the control panel.
func ControlPaneDisable() CommandFunc {
	return func(a *App) (any, error) {
		a.control.Disable()
		return nil, nil
	}
}

// MouseCursorOn is a command that will enable the mouse cursor.
func MouseCursorOn() CommandFunc {
	return func(a *App) (any, error) {
		a.mouse.Enabled = true
		return nil, nil
	}
}

// MouseCursorOff is a command that will disable the mouse cursor.
func MouseCursorOff() CommandFunc {
	return func(a *App) (any, error) {
		a.mouse.Enabled = false
		return nil, nil
	}
}

// SentenceChoiceInit is a command that will initialize a new sentence choice dialog.
func SentenceChoiceInit() CommandFunc {
	return func(a *App) (any, error) {
		return a.control.NewSentenceChoice(), nil
	}
}

// SentenceChoiceAdd is a command that will add a sentence to the sentence choice dialog.
func SentenceChoiceAdd(choice *ControlSentenceChoice, sentence string) CommandFunc {
	return func(app *App) (any, error) {
		if app.control.choice != choice {
			return nil, errors.New("sentence choice dialog mismatch")
		}
		app.control.choice.Add(sentence)
		return nil, nil
	}
}

// SentenceChoiceWait is a command that will wait for the user to choose a sentence.
func SentenceChoiceWait(choice *ControlSentenceChoice, sayit bool) CommandAsyncFunc {
	return func(app *App) Future {
		if app.control.choice != choice {
			return AlreadyFailed(errors.New("sentence choice dialog mismatch"))
		}

		return Continue(
			app.control.choice.Done(),
			func(val any) Future {
				choice := val.(IndexedSentence)
				if !sayit {
					return AlreadySucceeded(choice)
				}
				return FutureMap(
					app.RunCommand(ActorSpeak{
						Actor: app.ego,
						Text:  choice.Sentence,
					}),
					func(any) any { return choice },
				)
			},
		)
	}
}
