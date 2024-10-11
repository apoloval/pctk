package pctk

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Camera struct {
	raw rl.Camera2D
}

func (c Camera) Target() Position {
	return positionFromRaylib(c.raw.Target)
}

func (c Camera) WithTarget(target Position) Camera {
	c.raw.Target = target.toRaylib()
	return c
}

func (c Camera) WithZoom(zoom float32) Camera {
	c.raw.Zoom = zoom
	return c
}

func (c Camera) ScreenToWorldPosition(pos Position) Position {
	return positionFromRaylib(rl.GetScreenToWorld2D(pos.toRaylib(), c.raw))
}

func (c Camera) Action(act func()) {
	rl.BeginMode2D(c.raw)
	act()
	rl.EndMode2D()
}
