package pctk

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// App is the pctk application. It is the main struct that holds all the context necessary to run
// the application.
type App struct {
	res ResourceLoader

	screenCaption string
	screenZoom    int32

	actors  map[string]*Actor
	dialogs []Dialog
	objects []*Object
	rooms   map[string]*Room
	room    *Room
	scripts map[ResourceRef]*Script

	control  ControlPane
	commands CommandQueue

	cam         rl.Camera2D
	cursorTx    rl.Texture2D
	cursorColor Color
	music       *Music
	sound       *Sound
	ego         *Actor
}

// New creates a new pctk application.
func New(resources ResourceLoader, opts ...AppOption) *App {
	app := &App{
		res:     resources,
		actors:  make(map[string]*Actor),
		rooms:   make(map[string]*Room),
		scripts: make(map[ResourceRef]*Script),
	}

	opts = append(defaultAppOptions, opts...)
	for _, opt := range opts {
		opt(app)
	}

	app.init()

	return app
}

func (a *App) init() {
	rl.InitWindow(ScreenWidth*a.screenZoom, ScreenHeight*a.screenZoom, a.screenCaption)
	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	a.cam.Zoom = float32(a.screenZoom)
	a.initMouse()
	a.control.Init()
}

func (a *App) Close() {
	a.unloadMusic()
	rl.CloseAudioDevice()
	rl.CloseWindow()
}
