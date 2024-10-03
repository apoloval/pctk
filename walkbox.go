package pctk

import (
	"log"
	"math"
)

// Walkbox refers to a convex polygonal area that defines the walkable space for actors.
type WalkBox struct {
	walkBoxID string
	enabled   bool
	vertices  [4]*Positionf
}

// NewWalkBox creates a new WalkBox with the given ID and vertices.
// It ensures the polygon formed by the vertices is convex. If not, it will cause a panic.
// Why convex? Because you can draw a straight line/path between any two vertices inside the polygon
// without needing to implement complex pathfinding algorithms.
func NewWalkBox(id string, vertices [4]*Positionf) *WalkBox {
	w := &WalkBox{
		walkBoxID: id,
		vertices:  vertices,
		enabled:   true,
	}

	if !w.isConvex() {
		log.Panicf("walkbox must be a convex polygon: %v", vertices)
	}
	return w
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

// ContainsPoint check if the provided position is in the boundaries defined by the WalkBox.
func (w *WalkBox) ContainsPoint(p *Positionf) bool {
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

// IsAdjacent checks if two WalkBoxes are adjacent. It returns false if either WalkBox is disabled.
func (w *WalkBox) IsAdjacent(otherWalkBox *WalkBox) bool {
	return w.GateWith(otherWalkBox) != nil
}

// GateWith returns the gate (shared point) between two adjacent walk boxes, if any.
func (w *WalkBox) GateWith(otherWalkBox *WalkBox) *Positionf {
	if w.enabled && otherWalkBox.enabled {
		for _, vertex := range otherWalkBox.vertices {
			if w.ContainsPoint(vertex) {
				return vertex
			}
		}

		// two-way verification
		for _, vertex := range w.vertices {
			if otherWalkBox.ContainsPoint(vertex) {
				return vertex
			}
		}
	}

	return nil // WalkBoxes aren't adjacent
}

// Distance calculates the shortest distance from the WalkBox to the given position.
func (wb *WalkBox) Distance(p *Positionf) float32 {
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
	itineraryMatrix [][]int
}

const (
	// infinityDistance represents a maximum distance value used for unconnected paths.
	infinityDistance = 255
	// InvalidWalkBox indicates an invalid WalkBox ID, typically used to signify non-existence.
	InvalidWalkBox = -1
)

// NewWalkBoxMatrix creates and returns a new WalkBoxMatrix instance
func NewWalkBoxMatrix(walkboxes []*WalkBox) *WalkBoxMatrix {
	wm := &WalkBoxMatrix{
		walkBoxes: walkboxes,
	}

	wm.resetItinerary()
	return wm
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
			} else if walkbox.IsAdjacent(otherWalkBox) {
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
func (wm *WalkBoxMatrix) EnableWalkBox(id int, enabled bool) {
	if id >= 0 && id < len(wm.walkBoxes) {
		wm.walkBoxes[id].enabled = enabled
		wm.resetItinerary()
	}
}

// FindPath calculates and returns a path as a sequence of positions from the
// starting point 'from' to the destination 'to' within the walk box matrix.
// The path is returned as a slice of positions representing waypoints.
func (wm *WalkBoxMatrix) FindPath(from, to *Positionf) []*Positionf {
	var path []*Positionf
	current, _ := wm.walkBoxAt(from)
	target, _ := wm.walkBoxAt(to)

	for current != target {
		next := wm.nextWalkBox(current, target)
		if next == InvalidWalkBox {
			break
		}
		path = append(path, wm.closestPositionToWalkBox(current, next))
		current = next
	}
	path = append(path, wm.closestPositionOnWalkBox(current, to))
	return path
}

// nextWalkBox returns the next walk box in the path from the source to the destination.
func (wm *WalkBoxMatrix) nextWalkBox(from, to int) int {
	if from < 0 || from >= len(wm.walkBoxes) || to < 0 || to >= len(wm.walkBoxes) {
		return InvalidWalkBox
	}
	return wm.itineraryMatrix[from][to]
}

// walkBoxAt returns the walk box identifier at the given position or the closest one,
// along with a boolean indicating inclusion. If the point is located between two or more
// boxes, it returns the lowest walk box ID among them.
func (wm *WalkBoxMatrix) walkBoxAt(p *Positionf) (id int, included bool) {
	var minDistance float32 = math.MaxFloat32
	id = InvalidWalkBox
	for i, wb := range wm.walkBoxes {
		if included = wb.ContainsPoint(p); included {
			return i, included
		}

		current := wb.Distance(p)
		if current < minDistance {
			id = i
			minDistance = current
		}
	}

	return id, false
}

// closestPositionToWalkBox returns the closest point to the specified walkbox identifiers from
// the origin.
func (wm *WalkBoxMatrix) closestPositionToWalkBox(from, to int) *Positionf {
	fromWb := wm.walkBoxes[from]
	toWb := wm.walkBoxes[to]
	// GateWith can return nil, but in this case is not expected, should panic?
	return fromWb.GateWith(toWb)
}

// closestPositionOnWalkBox returns the closest point on the walkbox at a given position.
func (wm *WalkBoxMatrix) closestPositionOnWalkBox(from int, p *Positionf) *Positionf {
	wb := wm.walkBoxes[from]
	if wb.ContainsPoint(p) {
		return p
	}

	var minDistance float32 = math.MaxFloat32
	var closestPoint *Positionf
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
