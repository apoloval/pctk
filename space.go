package pctk

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// ScreenWidth is the width of the screen (ignoring zoom).
	ScreenWidth = 320

	// ScreenHeight is the height of the screen (ignoring zoom).
	ScreenHeight = 200

	// ViewportHeight is the width of the room section of the screen.
	ViewportHeight = 144

	// ControlPaneHeight is the height of the control pane of the screen.
	ControlPaneHeight = 56
)

var (
	// ViewportRect is the rectangle that represents the viewport of the scene.
	ViewportRect = NewRect(0, 0, ScreenWidth, ViewportHeight)

	// ControlPaneRect is the rectangle that represents the control pane.
	ControlPaneRect = NewRect(0, 0, ScreenWidth, ControlPaneHeight)
)

// Position represents a 2D position.
type Position struct {
	X, Y int
}

// NewPos creates a new position.
func NewPos(x, y int) Position {
	return Position{x, y}
}

func positionFromRaylib(v rl.Vector2) Position {
	return Position{int(v.X), int(v.Y)}
}

// String returns a string representation of the position.
func (p Position) String() string {
	return fmt.Sprintf("(X:%d, Y:%d)", p.X, p.Y)
}

// ToPosf converts an integer position to a floating point position.
func (p Position) ToPosf() Positionf {
	return Positionf{float32(p.X), float32(p.Y)}
}

// Add adds two positions.
func (p Position) Add(other Position) Position {
	return Position{p.X + other.X, p.Y + other.Y}
}

// Sub subtracts two positions.
func (p Position) Sub(other Position) Position {
	return Position{p.X - other.X, p.Y - other.Y}
}

// Above returns a position above the current one.
func (p Position) Above(h int) Position {
	return Position{p.X, p.Y - h}
}

// Distance returns the distance between two positions.
func (p Position) Distance(other Position) Size {
	s := Size{other.X - p.X, other.Y - p.Y}
	if s.W < 0 {
		s.W = -s.W
	}
	if s.H < 0 {
		s.H = -s.H
	}
	return s
}

// DirectionTo returns the direction from the current position to another.
func (p Position) DirectionTo(other Position) Direction {
	dist := p.Distance(other)
	if dist.W > dist.H {
		// Horizontal direction
		if other.X < p.X {
			return DirLeft
		}
		return DirRight
	} else {
		// Vertical direction
		if other.Y < p.Y {
			return DirUp
		}
		return DirDown
	}
}

// Equals returns true if both Position instances have the same X and Y coordinates.
func (p Position) Equals(p1 Position) bool {
	return p.X == p1.X && p.Y == p1.Y
}

func (p Position) toRaylib() rl.Vector2 {
	return rl.NewVector2(float32(p.X), float32(p.Y))
}

// Positionf represents a 2D position with floating point coordinates for fractional positions.
type Positionf struct {
	X, Y float32
}

// NewPosf creates a new position with floating point coordinates.
func NewPosf(x, y float32) Positionf {
	return Positionf{x, y}
}

// String returns a string representation of the position.
func (p Positionf) String() string {
	return fmt.Sprintf("(X:%.2f, Y:%.2f)", p.X, p.Y)
}

// ToPos converts a floating point position to an integer position.
func (p Positionf) ToPos() Position {
	return Position{int(p.X), int(p.Y)}
}

// Scale scales the position by a factor.
func (p Positionf) Scale(s float32) Positionf {
	return Positionf{p.X * s, p.Y * s}
}

// ScaleBy scales the position by another position.
func (p Positionf) ScaleBy(other Positionf) Positionf {
	return Positionf{p.X * other.X, p.Y * other.Y}
}

// Add adds two positions.
func (p Positionf) Add(other Positionf) Positionf {
	return Positionf{p.X + other.X, p.Y + other.Y}
}

// Sub subtracts two positions.
func (p Positionf) Sub(other Positionf) Positionf {
	return Positionf{p.X - other.X, p.Y - other.Y}
}

// Move moves the position towards another position by a given speed.
func (p Positionf) Move(to Positionf, speed Positionf) Positionf {
	dist := to.Sub(p)
	if dist.X > 0 {
		if speed.X < dist.X {
			p.X += speed.X
		} else {
			p.X = to.X
		}
	} else if dist.X < 0 {
		if speed.X < -dist.X {
			p.X -= speed.X
		} else {
			p.X = to.X
		}
	}
	if dist.Y > 0 {
		if speed.Y < dist.Y {
			p.Y += speed.Y
		} else {
			p.Y = to.Y
		}
	} else if dist.Y < 0 {
		if speed.Y < -dist.Y {
			p.Y -= speed.Y
		} else {
			p.Y = to.Y
		}
	}
	return p
}

// CrossProduct calculates the 2D cross product (determinant) of vectors p->p1 and p1->p2 to determine their orientation.
func (p Positionf) CrossProduct(p1, p2 Positionf) float32 {
	return (p1.X-p.X)*(p2.Y-p1.Y) - (p1.Y-p.Y)*(p2.X-p1.X)
}

// IsIntersecting checks if a horizontal ray from point p intersects the line segment p1->p2 (Ray-Casting method).
func (p Positionf) IsIntersecting(p1, p2 Positionf) bool {
	if (p1.Y > p.Y) != (p2.Y > p.Y) {
		return p.X < (p2.X-p1.X)*(p.Y-p1.Y)/(p2.Y-p1.Y)+p1.X
	}

	return false
}

// Distance calculates the Euclidean distance between two points.
func (p Positionf) Distance(p1 Positionf) float32 {
	dx := p.X - p1.X
	dy := p.Y - p1.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

// ClosestPointOnSegment returns the closest point on the segment defined by p1 and p2 to the point p.
func (p Positionf) ClosestPointOnSegment(p1, p2 Positionf) Positionf {
	if p1.Equals(p2) {
		return p1
	}

	px := p.X - p1.X
	py := p.Y - p1.Y

	vx := p2.X - p1.X
	vy := p2.Y - p1.Y

	lengthSq := vx*vx + vy*vy // how long the segment is, squared.

	// calculate point's position on the segment.
	// If t is less than 0, it’s closer to the start of the segment.
	// If it’s greater than 1, it’s closer to the end.
	t := float32(math.Max(0, math.Min(1, float64((px*vx+py*vy)/lengthSq))))

	// Find the closest point on the segment.
	closestX := p1.X + t*vx
	closestY := p1.Y + t*vy

	return Positionf{closestX, closestY}
}

// Equals returns true if both Positionf instances have the same X and Y coordinates.
func (p Positionf) Equals(p1 Positionf) bool {
	return p.X == p1.X && p.Y == p1.Y
}

// Converts a Positionf to a Raylib Vector2.
func (p Positionf) toRaylib() rl.Vector2 {
	return rl.NewVector2(p.X, p.Y)
}

// Size represents a 2D size.
type Size struct {
	W, H int
}

// NewSize creates a new size.
func NewSize(w, h int) Size {
	return Size{w, h}
}

func sizeFromRaylib(v rl.Vector2) Size {
	return Size{int(v.X), int(v.Y)}
}

func (s Size) String() string {
	return fmt.Sprintf("(W:%d, H:%d)", s.W, s.H)
}

func (s Size) FlipH() Size {
	return Size{W: -s.W, H: s.H}
}

func (s Size) toRaylib() rl.Vector2 {
	return rl.NewVector2(float32(s.W), float32(s.H))
}

// Rectangle represents a 2D rectangle.
type Rectangle struct {
	Pos  Position
	Size Size
}

func NewRect(x, y, w, h int) Rectangle {
	return Rectangle{
		Pos:  Position{x, y},
		Size: Size{w, h},
	}
}

// IsZero returns true if the rectangle is empty.
func (r Rectangle) IsZero() bool {
	return r.Pos.X == 0 && r.Pos.Y == 0 && r.Size.W == 0 && r.Size.H == 0
}

// LeftEdge returns the X coordinate of the left edge of the rectangle.
func (r Rectangle) LeftEdge() int {
	return r.Pos.X
}

// RightEdge returns the X coordinate of the right edge of the rectangle.
func (r Rectangle) RightEdge() int {
	return r.Pos.X + r.Size.W
}

// TopEdge returns the Y coordinate of the top edge of the rectangle.
func (r Rectangle) TopEdge() int {
	return r.Pos.Y
}

// BottomEdge returns the Y coordinate of the bottom edge of the rectangle.
func (r Rectangle) BottomEdge() int {
	return r.Pos.Y + r.Size.H
}

func (r Rectangle) String() string {
	return fmt.Sprintf("(Pos:%v, Size:%v)", r.Pos, r.Size)
}

func (r Rectangle) toRaylib() rl.Rectangle {
	return rl.NewRectangle(
		float32(r.Pos.X), float32(r.Pos.Y), float32(r.Size.W), float32(r.Size.H),
	)
}

// Contains returns true if the mouse is into the given rectangle.
func (r Rectangle) Contains(pos Position) bool {
	return rl.CheckCollisionPointRec(pos.toRaylib(), r.toRaylib())
}

// Direction represents a direction in 2D space.
type Direction byte

// Inverse returns the opposite direction.
func (d Direction) Inverse() Direction {
	switch d {
	case DirUp:
		return DirDown
	case DirDown:
		return DirUp
	case DirLeft:
		return DirRight
	case DirRight:
		return DirLeft
	default:
		return d
	}
}

func (d Direction) String() string {
	switch d {
	case DirUp:
		return "Up"
	case DirDown:
		return "Down"
	case DirLeft:
		return "Left"
	case DirRight:
		return "Right"
	default:
		return "Unknown"
	}
}

const (
	DirRight Direction = iota
	DirLeft
	DirUp
	DirDown
)
