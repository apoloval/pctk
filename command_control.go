package pctk

import "errors"

// ControlPanelChangeMode is a command that will enable or disable the control panel.
type ControlPanelChangeMode struct {
	Mode ControlPaneMode
}

func (cmd ControlPanelChangeMode) Execute(app *App, done *Promise) {
	app.control.Mode = cmd.Mode
	done.Complete()
}

// EnableMouseCursor is a command that will enable or disable the mouse control.
type EnableMouseCursor struct {
	Enable bool
}

func (cmd EnableMouseCursor) Execute(app *App, done *Promise) {
	app.control.cursor.Enabled = cmd.Enable
	done.Complete()
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
