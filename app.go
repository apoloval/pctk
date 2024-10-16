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
	ego      *Actor
	music    *Music
	objects  []*Object
	rooms    []*Room
	scripts  map[ResourceRef]*Script
	sound    *Sound

	cam      Camera
	control  ControlPane
	commands CommandQueue
	mouse    *Mouse
	frame    *Frame
	viewport Viewport
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
		a.processFrame()
	}
}

func (a *App) init() {
	rl.InitWindow(ScreenWidth*a.screenZoom, ScreenHeight*a.screenZoom, a.screenCaption)
	rl.InitAudioDevice()
	rl.SetTargetFPS(60)
	rl.HideCursor()

	a.mouse = NewMouseCursor()
	a.frame = NewFrame(a.mouse, a.debugEnabled)
	a.cam = a.cam.WithZoom(float32(a.screenZoom))
	a.mouse.SetRelativePosition(&a.cam, Position{X: ScreenWidth / 2, Y: ScreenHeight / 2})
	a.viewport.Init(a.cam)
	a.control.Init(a, a.cam, &a.viewport)
}

func (a *App) processFrame() {
	a.frame.Num++
	a.frame.DebugEnabled = a.debugEnabled

	a.updateMusic()
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	a.viewport.ProcessFrame(a.frame)
	a.control.ProcessFrame(a, a.frame)
	a.frame.WithCamera(&a.cam, func(f *Frame) {
		a.mouse.Draw(f)
	})
	rl.EndDrawing()
	a.commands.Execute(a)
}
