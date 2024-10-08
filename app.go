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
	debugMode     bool
	debugEnabled  bool

	actors   []*Actor
	defaults *ObjectDefaults
	dialogs  []Dialog
	objects  []*Object
	rooms    []*Room
	room     *Room
	scripts  map[ResourceRef]*Script

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
		scripts: make(map[ResourceRef]*Script),
	}

	opts = append(defaultAppOptions, opts...)
	for _, opt := range opts {
		opt(app)
	}

	app.init()

	return app
}

// Close closes the application.
func (a *App) Close() {
	a.StopMusic()
	rl.CloseAudioDevice()
	rl.CloseWindow()
}

// Run starts the application.
func (a *App) Run() {
	defer a.Close()

	for !rl.WindowShouldClose() {
		a.run()
	}
}

func (a *App) init() {
	rl.InitWindow(ScreenWidth*a.screenZoom, ScreenHeight*a.screenZoom, a.screenCaption)
	rl.InitAudioDevice()
	rl.SetTargetFPS(60)
	rl.HideCursor()

	a.cam.Zoom = float32(a.screenZoom)
	a.control.Init(&a.cam)
}

func (a *App) run() {
	a.updateMusic()
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	rl.BeginMode2D(a.cam)
	a.drawSceneViewport()
	a.control.Draw(a)
	a.drawDialogs()
	rl.EndMode2D()
	rl.EndDrawing()
	a.control.processControlInputs(a)
	a.commands.Execute(a)
}
