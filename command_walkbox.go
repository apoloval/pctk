package pctk

// EnableWalkBox is a command that will enable or disable a walkbox in the room.
func EnableWalkBox(w *WalkBox, enabled bool) CommandFunc {
	return func(a *App) (any, error) {
		w.room.wbmatrix.EnableWalkBox(w.walkBoxID, enabled)
		return nil, nil
	}
}
