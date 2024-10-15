package pctk

// EnableWalkBox is a command that will enable a walkbox in the room.
func EnableWalkBox(w *WalkBox) CommandFunc {
	return func(a *App) (any, error) {
		w.room.wbmatrix.EnableWalkBox(w.walkBoxID, true)
		return nil, nil
	}
}

// DisableWalkBox is a command that will disable a walkbox in the room.
func DisableWalkBox(w *WalkBox) CommandFunc {
	return func(a *App) (any, error) {
		w.room.wbmatrix.EnableWalkBox(w.walkBoxID, false)
		return nil, nil
	}
}
