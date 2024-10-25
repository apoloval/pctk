package pctk

// ShowDialog is a command that will show a dialog with the given text.
type ShowDialog struct {
	Actor    *Actor
	Text     string
	Position Position
	Color    Color
	Speed    float32
}

func (cmd ShowDialog) Execute(app *App, done *Promise) {
	dialog := NewDialog(cmd.Actor, cmd.Text, cmd.Position, cmd.Color, cmd.Speed)
	if room := cmd.Actor.Room; room != nil {
		dialog.SetBounds(room.Rect())
	}

	app.viewport.BeginDialog(dialog)
	done.Bind(dialog.Done())
}

// CancelDialogs is a command that will cancel all dialogs in the viewport.
func CancelDialogs() CommandFunc {
	return func(a *App) (any, error) {
		a.viewport.CancelDialogs()
		return nil, nil
	}
}
