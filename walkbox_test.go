package pctk_test

import (
	"testing"

	"github.com/apoloval/pctk"
	"github.com/stretchr/testify/assert"
)

const (
	DefaultWalkBoxID = "walkbox"
)

func TestNewWalkBox(t *testing.T) {
	testCases := []struct {
		name        string
		vertices    [4]*pctk.Positionf
		shouldPanic bool
		message     string
	}{
		{
			name:        "Concave polygon should panic",
			vertices:    [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 2, Y: 1}, {X: 4, Y: 4}},
			shouldPanic: true,
			message:     "Expected panic because vertices form a concave polygon!",
		},
		{
			name:        "Collinear vertices should panic",
			vertices:    [4]*pctk.Positionf{{X: 1, Y: 2}, {X: 2, Y: 4}, {X: 3, Y: 6}, {X: 4, Y: 8}},
			shouldPanic: true,
			message:     "Expected panic because vertices are collinear!",
		},
		{
			name:        "Should successfully create a valid WalkBox with a convex polygon",
			vertices:    [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
			shouldPanic: false,
			message:     "Expected create a valid WalkBox, vertices form a convex polygon!",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.shouldPanic {
				assert.Panics(t, func() {
					pctk.NewWalkBox(DefaultWalkBoxID, testCase.vertices)
				}, testCase.message)
			} else {
				assert.NotPanics(t, func() {
					pctk.NewWalkBox(DefaultWalkBoxID, testCase.vertices)
				}, testCase.message)
			}
		})
	}

}

func TestContainsPoint(t *testing.T) {
	testCases := []struct {
		name       string
		vertices   [4]*pctk.Positionf
		point      *pctk.Positionf
		assertFunc func(t *testing.T, isInside bool)
	}{
		{
			name:     "The point should be considered inside the polygon when it is on the edge",
			vertices: [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
			point:    &pctk.Positionf{X: 2, Y: 0}, // On the edge
			assertFunc: func(t *testing.T, isInside bool) {
				assert.True(t, isInside)
			},
		},
		{
			name:     "The point should be inside the polygon",
			vertices: [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
			point:    &pctk.Positionf{X: 2, Y: 2},
			assertFunc: func(t *testing.T, isInside bool) {
				assert.True(t, isInside)
			},
		},
		{
			name:     "The point should be outside the polygon",
			vertices: [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
			point:    &pctk.Positionf{X: 5, Y: 5},
			assertFunc: func(t *testing.T, isInside bool) {
				assert.False(t, isInside)
			},
		},
		{
			name:     "The point should be considered inside the polygon when it is on a vertex",
			vertices: [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
			point:    &pctk.Positionf{X: 0, Y: 0}, // On the vertex
			assertFunc: func(t *testing.T, isInside bool) {
				assert.True(t, isInside)
			},
		},
		{
			name:     "The point should be outside when it is far from the polygon",
			vertices: [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}},
			point:    &pctk.Positionf{X: 10, Y: 10}, // Clearly outside the polygon
			assertFunc: func(t *testing.T, isInside bool) {
				assert.False(t, isInside)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			walkBox := pctk.NewWalkBox(DefaultWalkBoxID, testCase.vertices)
			isInside := walkBox.ContainsPoint(testCase.point)
			testCase.assertFunc(t, isInside)
		})
	}
}

func TestWalkBoxMatrix(t *testing.T) {
	/*
		Polygons disposition:

		  +-------+-------+-------+
		  |       |       |       |
		  | box3  | box4  | box5  |
		  |       |       |       |
		  +-------+-------+-------+
		  |       |       |       |
		  | box0  | box1  | box2  |
		  |       |       |       |
		  +-------+-------+-------+

		Each box represents a square, with adjacent connections:
		- box0 is adjacent to box1, box3
		- box1 is adjacent to box0, box2, box3, box4
		- box2 is adjacent to box1, box4, box5
		- box3 is adjacent to box0, box1, box4
		- box4 is adjacent to box1, box2, box3, box5
		- box5 is adjacent to box2, box4
	*/

	box0 := pctk.NewWalkBox("walkbox0", [4]*pctk.Positionf{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	box1 := pctk.NewWalkBox("walkbox1", [4]*pctk.Positionf{{1, 0}, {2, 0}, {2, 1}, {1, 1}})
	box2 := pctk.NewWalkBox("walkbox2", [4]*pctk.Positionf{{2, 0}, {3, 0}, {3, 1}, {2, 1}})
	box3 := pctk.NewWalkBox("walkbox3", [4]*pctk.Positionf{{0, 1}, {1, 1}, {1, 2}, {0, 2}})
	box4 := pctk.NewWalkBox("walkbox4", [4]*pctk.Positionf{{1, 1}, {2, 1}, {2, 2}, {1, 2}})
	box5 := pctk.NewWalkBox("walkbox5", [4]*pctk.Positionf{{2, 1}, {3, 1}, {3, 2}, {2, 2}})

	wm := pctk.NewWalkBoxMatrix([]*pctk.WalkBox{box0, box1, box2, box3, box4, box5})

	assert.True(t, box0.IsAdjacent(box1), "box0 should be adjacent to box1")
	assert.True(t, box1.IsAdjacent(box2), "box1 should be adjacent to box2")
	assert.True(t, box1.IsAdjacent(box3), "box1 should be adjacent to box3")
	assert.True(t, box0.IsAdjacent(box3), "box0 should be adjacent to box3")
	assert.True(t, box0.IsAdjacent(box4), "box0 should be adjacent to box4")
	assert.True(t, box1.IsAdjacent(box4), "box1 should be adjacent to box4")
	assert.True(t, box3.IsAdjacent(box4), "box3 should be adjacent to box4")
	assert.True(t, box2.IsAdjacent(box4), "box2 should be adjacent to box4")
	assert.True(t, box2.IsAdjacent(box5), "box2 should be adjacent to box5")
	assert.True(t, box4.IsAdjacent(box5), "box4 should be adjacent to box5")
	// Test non-adjacency
	assert.False(t, box0.IsAdjacent(box2), "box0 should not be adjacent to box2")
	assert.False(t, box3.IsAdjacent(box2), "box3 should not be adjacent to box2")
	assert.False(t, box0.IsAdjacent(box5), "box0 should not be adjacent to box5")

	// Test pathfinding from box0 to box2 via box1
	from := 0
	to := 2
	next := wm.NextWalkBox(from, to)
	assert.Equal(t, 1, next, "Next walk box from 0 to 2 should be 1")

	next = wm.NextWalkBox(next, to)
	assert.Equal(t, to, next, "Next walk box from 1 to 2 should be 2")

	// Disable box1 and test new path from box0 to box2
	wm.EnableWalkBox(1, false)
	next = wm.NextWalkBox(from, to)
	assert.Equal(t, 4, next, "Next walk box from 0 to 2 with box1 disabled should be 4")

	next = wm.NextWalkBox(next, to)
	assert.Equal(t, to, next, "Next walk box from 4 to 2 should be 2")

	// Test path when destination (box1) is disabled
	to = 1
	next = wm.NextWalkBox(from, to)
	assert.Equal(t, pctk.InvalidWalkBox, next, "Next walk box from 0 to 1 should be invalid since box1 is disabled")

	// Re-enable box1 and test path again
	wm.EnableWalkBox(1, true)
	next = wm.NextWalkBox(from, to)
	assert.Equal(t, to, next, "Next walk box from 0 to 1 should be 1 after re-enabling box1")
}

func TestWalkBoxAt(t *testing.T) {
	box0 := pctk.NewWalkBox("walkbox0", [4]*pctk.Positionf{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}})
	box1 := pctk.NewWalkBox("walkbox1", [4]*pctk.Positionf{{X: 1, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 1}, {X: 1, Y: 1}})
	box2 := pctk.NewWalkBox("walkbox2", [4]*pctk.Positionf{{X: 2, Y: 0}, {X: 3, Y: 0}, {X: 3, Y: 1}, {X: 2, Y: 1}})
	box3 := pctk.NewWalkBox("walkbox3", [4]*pctk.Positionf{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 2}})
	box4 := pctk.NewWalkBox("walkbox4", [4]*pctk.Positionf{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 2, Y: 2}, {X: 1, Y: 2}})
	box5 := pctk.NewWalkBox("walkbox5", [4]*pctk.Positionf{{X: 2, Y: 1}, {X: 3, Y: 1}, {X: 3, Y: 2}, {X: 2, Y: 2}})

	wm := pctk.NewWalkBoxMatrix([]*pctk.WalkBox{box0, box1, box2, box3, box4, box5})

	testCases := []struct {
		name             string
		point            *pctk.Positionf
		expectedID       int
		expectedIncluded bool
	}{
		{
			name:             "Point inside a box",
			point:            &pctk.Positionf{X: 0.5, Y: 0.5},
			expectedID:       0,
			expectedIncluded: true,
		},
		{
			name:             "Point exactly on the edge of a box",
			point:            &pctk.Positionf{X: 0, Y: 1},
			expectedID:       0, // In this case, it returns the lowest box id
			expectedIncluded: true,
		},
		{
			name:             "Point at corner of a box",
			point:            &pctk.Positionf{X: 3, Y: 2},
			expectedID:       5,
			expectedIncluded: true,
		},
		{
			name:             "Point out and between two boxes, returns closest",
			point:            &pctk.Positionf{X: 0.8, Y: -0.2},
			expectedID:       0,
			expectedIncluded: false,
		},
		{
			name:             "Point far away, returns closest",
			point:            &pctk.Positionf{X: 10, Y: 10},
			expectedID:       5,
			expectedIncluded: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			id, included := wm.WalkBoxAt(testCase.point)
			assert.Equal(t, testCase.expectedID, id)
			assert.Equal(t, testCase.expectedIncluded, included)
		})
	}
}
