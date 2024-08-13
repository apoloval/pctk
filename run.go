package pctk

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (a *App) Run() {
	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		a.run()
	}
}

func (a *App) run() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	rl.BeginMode2D(a.cam)
	a.drawBackgroud()
	a.drawControlPanel()
	a.drawDialogs()
	rl.EndMode2D()
	rl.EndDrawing()
}
