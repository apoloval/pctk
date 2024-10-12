package pctk

import (
	"log"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Walkbox refers to a convex polygonal area that defines the walkable space for actors.
type WalkBox struct {
	walkBoxID string
	enabled   bool
	vertices  [4]Positionf
	scale     float32
}

// NewWalkBox creates a new WalkBox with the given ID and vertices.
// It ensures the polygon formed by the vertices is convex. If not, it will cause a panic.
// Why convex? Because you can draw a straight line/path between any two vertices inside the polygon
// without needing to implement complex pathfinding algorithms.
func NewWalkBox(id string, vertices [4]Position, scale float32) *WalkBox {
	var verticesf [4]Positionf
	for i, v := range vertices {
		verticesf[i] = v.ToPosf()
	}
	w := &WalkBox{
		walkBoxID: id,
		vertices:  verticesf,
		enabled:   true,
		scale:     scale,
	}

	if !w.isConvex() {
		log.Panicf("walkbox must be a convex polygon: %v", vertices)
	}
	return w
}

// Scale returns the scale factor of the WalkBox for camera zoom effects.
func (w *WalkBox) Scale() float32 {
	return w.scale
}

// Draws the edges of the WalkBox.
func (w *WalkBox) draw() {
	numVertices := len(w.vertices)
	for i := 0; i < numVertices; i++ {
		p1 := w.vertices[i]
		p2 := w.vertices[(i+1)%numVertices]
		rl.DrawLineEx(p1.toRaylib(), p2.toRaylib(), 1.2, rl.NewColor(0x55, 0xFF, 0x55, 0x7D))
	}
}

// isConvex check if the current WalkBox is a convex poligon.
func (w *WalkBox) isConvex() bool {
	numVertices := len(w.vertices)

	var totalCrossProduct float32
	var polygonDirection bool // true if clockwise, false if counter-clockwise
	for i := 0; i < numVertices; i++ {
		// Get three consecutive vertices (cyclically)
		p1 := w.vertices[i]
		p2 := w.vertices[(i+1)%numVertices]
		p3 := w.vertices[(i+2)%numVertices]

		cp := p1.CrossProduct(p2, p3)

		if cp == 0 {
			continue // Skip collinear vertices
		}

		totalCrossProduct += cp

		if i == 0 {
			polygonDirection = cp > 0
		} else {
			if (cp > 0) != polygonDirection {
				return false // If direction changes, the polygon is not convex
			}
		}
	}
	return totalCrossProduct != 0
}

// containsPoint check if the provided position is in the boundaries defined by the WalkBox.
func (w *WalkBox) containsPoint(p Positionf) bool {
	// Check if the position is one of the vertices
	for _, vertex := range w.vertices {
		if p.Equals(vertex) {
			return true
		}
	}
	numberOfIntersections := 0
	numVertices := len(w.vertices)

	for i := 0; i < numVertices; i++ {
		p1 := w.vertices[i]
		p2 := w.vertices[(i+1)%numVertices]

		if p.IsIntersecting(p1, p2) {
			numberOfIntersections++
		}
	}

	return numberOfIntersections%2 == 1 // Odd count means inside
}

// isAdjacent checks if two WalkBoxes are adjacent. It returns false if either WalkBox is disabled.
func (w *WalkBox) isAdjacent(otherWalkBox *WalkBox) bool {
	if w.enabled && otherWalkBox.enabled {
		for _, vertex := range otherWalkBox.vertices {
			if w.containsPoint(vertex) {
				return true
			}
		}

		// two-way verification
		for _, vertex := range w.vertices {
			if otherWalkBox.containsPoint(vertex) {
				return true
			}
		}
	}
	return false
}

// distance calculates the shortest distance from the WalkBox to the given position.
func (wb *WalkBox) distance(p Positionf) float32 {
	numVertices := len(wb.vertices)

	var minDistance float32 = math.MaxFloat32
	for i := 0; i < numVertices; i++ {
		p1 := wb.vertices[i]
		p2 := wb.vertices[(i+1)%numVertices]
		closestPoint := p.ClosestPointOnSegment(p1, p2)
		currentDistance := p.Distance(closestPoint)

		if currentDistance < minDistance {
			minDistance = currentDistance
		}
	}

	return minDistance
}

// WalkBoxMatrix represents a collection of WalkBoxes and their adjacency relationships.
type WalkBoxMatrix struct {
	walkBoxes       []*WalkBox
	walkBoxesMap    map[string]*WalkBox
	itineraryMatrix [][]int
}

const (
	// infinityDistance represents a maximum distance value used for unconnected paths.
	infinityDistance = 255
	// InvalidWalkBox indicates an invalid WalkBox ID, typically used to signify non-existence.
	InvalidWalkBox = -1
	// MaxScaleDifference represents the maximum scale difference allowed between two points.
	MaxScaleDifference = 0.04
)

// NewWalkBoxMatrix creates and returns a new WalkBoxMatrix instance
func NewWalkBoxMatrix(walkboxes []*WalkBox) *WalkBoxMatrix {
	wm := &WalkBoxMatrix{
		walkBoxes:    walkboxes,
		walkBoxesMap: make(map[string]*WalkBox),
	}

	for _, w := range walkboxes {
		wm.walkBoxesMap[w.walkBoxID] = w
	}

	wm.resetItinerary()
	return wm
}

// WalkBoxes draw walkable boxes of the WalkBoxMatrix.
func (wm *WalkBoxMatrix) Draw() {
	for _, wb := range wm.walkBoxes {
		wb.draw()
	}
}

// resetItinerary computes the shortest paths between WalkBoxes and returns the resulting itinerary matrix.
func (wm *WalkBoxMatrix) resetItinerary() {
	numBoxes := len(wm.walkBoxes)
	distanceMatrix := make([][]int, numBoxes)
	itineraryMatrix := make([][]int, numBoxes)

	for i, walkbox := range wm.walkBoxes {
		itineraryMatrix[i] = make([]int, numBoxes)
		distanceMatrix[i] = make([]int, numBoxes)

		// Initialize the distance matrix: each box has distance 0 to itself,
		// and distance 1 to its direct neighbors. Initially, it has distance
		// 255 (= infinityDistance) to all other boxes.
		for j, otherWalkBox := range wm.walkBoxes {
			if i == j {
				distanceMatrix[i][j] = 0
				itineraryMatrix[i][j] = i
			} else if walkbox.isAdjacent(otherWalkBox) {
				distanceMatrix[i][j] = 1
				itineraryMatrix[i][j] = j
			} else {
				distanceMatrix[i][j] = infinityDistance
				itineraryMatrix[i][j] = InvalidWalkBox
			}
		}
	}

	// Compute the shortest routes between boxes via Kleene's algorithm.
	for i := range wm.walkBoxes {
		for j := range wm.walkBoxes {
			for k := range wm.walkBoxes {
				distIK := distanceMatrix[i][k]
				distKJ := distanceMatrix[k][j]
				if distanceMatrix[i][j] > distIK+distKJ {
					distanceMatrix[i][j] = distIK + distKJ
					itineraryMatrix[i][j] = k
				}
			}
		}
	}

	wm.itineraryMatrix = itineraryMatrix
}

// EnableWalkBox enables or disables the specified walk box and recalculates the itinerary matrix.
func (wm *WalkBoxMatrix) EnableWalkBox(id string, enabled bool) {
	if w, ok := wm.walkBoxesMap[id]; ok {
		w.enabled = enabled
		wm.resetItinerary()
	}
}

// WayPoint represents a point and its associated WalkBox.
type WayPoint struct {
	Walkbox  *WalkBox
	Position Position
}

// FindPath calculates and returns a path as a sequence of waypoints from the
// starting point 'from' to the destination 'to' within the walk box matrix.
// The path is returned as a slice of waypoints.
func (wm *WalkBoxMatrix) FindPath(from, to Position) []*WayPoint {
	var path []*WayPoint
	fromf := from.ToPosf()
	tof := to.ToPosf()
	current, _ := wm.walkBoxIndex(fromf)
	target, _ := wm.walkBoxIndex(tof)
	currentWp := &WayPoint{Walkbox: wm.walkBoxes[current], Position: from}

	path = append(path, currentWp)
	for current != target {
		next := wm.nextWalkBox(current, target)
		if next == InvalidWalkBox {
			break
		}
		nextPosition := wm.closestPositionToWalkBox(fromf, next)
		nextWp := &WayPoint{Walkbox: wm.walkBoxes[next], Position: nextPosition.ToPos()}
		path = append(path, append(currentWp.interpolate(nextWp), nextWp)...)
		current = next
		currentWp = nextWp
		fromf = nextPosition
	}

	path = append(path, &WayPoint{Walkbox: wm.walkBoxes[current], Position: wm.closestPositionOnWalkBox(current, tof).ToPos()})
	return path
}

// interpolate generates intermediate waypoints between a starting and ending waypoint.
// We need interpolate points if scale difference is higher.
func (from *WayPoint) interpolate(to *WayPoint) []*WayPoint {
	scaleDifference := math.Abs(float64(from.Walkbox.Scale()) - float64(to.Walkbox.Scale()))
	num := int(math.Ceil(scaleDifference / MaxScaleDifference))

	if num < 2 {
		return []*WayPoint{}
	}

	wayPoints := []*WayPoint{}

	deltaX := to.Position.X - from.Position.X
	deltaY := to.Position.Y - from.Position.Y
	deltaScale := scaleDifference / float64((num))

	for i := 1; i < num; i++ {
		t := float32(i) / float32(num)
		interpPos := Positionf{
			X: float32(from.Position.X) + t*float32(deltaX),
			Y: float32(from.Position.Y) + t*float32(deltaY),
		}
		interpScale := from.Walkbox.Scale()
		if from.Walkbox.Scale() > to.Walkbox.Scale() {
			interpScale -= float32(i) * float32(deltaScale)
		} else {
			interpScale += float32(i) * float32(deltaScale)
		}

		clonedWalkBox := *from.Walkbox // copy
		clonedWalkBox.scale = interpScale
		wayPoints = append(wayPoints, &WayPoint{
			Walkbox:  &clonedWalkBox,
			Position: interpPos.ToPos(),
		})
	}

	return wayPoints
}

// nextWalkBox returns the next walk box in the path from the source to the destination.
func (wm *WalkBoxMatrix) nextWalkBox(from, to int) int {
	if from < 0 || from >= len(wm.walkBoxes) || to < 0 || to >= len(wm.walkBoxes) {
		return InvalidWalkBox
	}

	if wm.itineraryMatrix[from][to] == to {
		return to
	}

	next := wm.itineraryMatrix[from][to]
	return wm.nextWalkBox(from, next)
}

// walkBoxAt returns the walk box identifier at the given position or the closest one,
// along with a boolean indicating inclusion. If the point is located between two or more
// boxes, it returns the lowest walk box ID among them.
func (wm *WalkBoxMatrix) walkBoxIndex(p Positionf) (id int, included bool) {
	var minDistance float32 = math.MaxFloat32
	id = InvalidWalkBox
	for i, wb := range wm.walkBoxes {
		if included = wb.containsPoint(p); included {
			return i, included
		}

		current := wb.distance(p)
		if current < minDistance {
			id = i
			minDistance = current
		}
	}

	return id, false
}

// walkBoxAt returns the walk box at the given position or the closest one,
func (wm *WalkBoxMatrix) walkBoxAt(p Positionf) (w *WalkBox, included bool) {
	id, include := wm.walkBoxIndex(p)
	if id != InvalidWalkBox {
		return wm.walkBoxes[id], include
	}

	return nil, false
}

// closestPositionOnWalkBox returns the closest point on the walkbox at a given position.
func (wm *WalkBoxMatrix) closestPositionOnWalkBox(from int, p Positionf) Positionf {
	wb := wm.walkBoxes[from]
	if wb.containsPoint(p) {
		return p
	}

	return wm.closestPositionToWalkBox(p, from) // looking for the closest edge point
}

// closestPositionToWalkBox returns the nearest point on the edge of the specified walkbox
// from a given position p.
func (wm *WalkBoxMatrix) closestPositionToWalkBox(p Positionf, to int) Positionf {
	var minDistance float32 = math.MaxFloat32
	var closestPoint Positionf
	wb := wm.walkBoxes[to]
	numVertices := len(wb.vertices)

	for i := 0; i < numVertices; i++ {
		p1 := wb.vertices[i]
		p2 := wb.vertices[(i+1)%numVertices]
		currentPoint := p.ClosestPointOnSegment(p1, p2)
		currentDistance := p.Distance(currentPoint)

		if currentDistance < minDistance {
			minDistance = currentDistance
			closestPoint = currentPoint
		}
	}

	return closestPoint
}
