package pctk

// RoomDeclare is a command that will declare a new room with the given properties.
type RoomDeclare struct {
	Room *Room
}

func (cmd RoomDeclare) Execute(app *App, done *Promise) {
	if err := app.DeclareRoom(cmd.Room); err != nil {
		done.CompleteWithError(err)
		return
	}
	done.CompleteWithValue(cmd.Room)
}

// RoomShow is a command that will show the room with the given resource.
type RoomShow struct {
	Room *Room
}

func (cmd RoomShow) Execute(app *App, done *Promise) {
	done.Bind(app.StartRoom(cmd.Room))
}

// RoomCameraTo is a command that will move the camera to the given position.
func RoomCameraTo(vp *Viewport, pos int) CommandFunc {
	return func(a *App) (any, error) {
		vp.CameraMoveTo(pos)
		return nil, nil
	}
}

// RoomCameraFollowActor is a command that will make the camera follow the given actor.
func RoomCameraFollowActor(vp *Viewport, actor *Actor) CommandFunc {
	return func(a *App) (any, error) {
		vp.CameraFollowActor(actor)
		return nil, nil
	}
}

// RoomCameraOnLeftEdge is a command that will put the camera on the left edge of the room.
func RoomCameraOnLeftEdge(vp *Viewport) CommandFunc {
	return func(a *App) (any, error) {
		vp.CameraOnLeftEdge()
		return nil, nil
	}
}

// RoomCameraOnRightEdge is a command that will put the camera on the right edge of the room.
func RoomCameraOnRightEdge(vp *Viewport) CommandFunc {
	return func(a *App) (any, error) {
		vp.CameraOnRightEdge()
		return nil, nil
	}
}
