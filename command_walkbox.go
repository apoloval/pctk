package pctk

// EnableWalkBox is a command that will enable or disable a walkbox in the room.
type EnableWalkBox struct {
	WalkBox *WalkBox
	Enabled bool
}

func (cmd EnableWalkBox) Execute(app *App, done *Promise) {
	cmd.WalkBox.room.wbmatrix.EnableWalkBox(cmd.WalkBox.walkBoxID, cmd.Enabled)
	done.Complete()
}
