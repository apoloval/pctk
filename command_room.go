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
func RoomCameraTo(room *Room, pos int) CommandFunc {
	return func(a *App) (any, error) {
		room.CameraMoveTo(pos)
		return nil, nil
	}
}

func RoomCameraFollowActor(room *Room, actor *Actor) CommandFunc {
	return func(a *App) (any, error) {
		room.CameraFollowActor(actor)
		return nil, nil
	}
}
