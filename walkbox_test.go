package pctk_test

import (
	"fmt"
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

func TestFindPath(t *testing.T) {
	/*
		Polygons disposition:

		          +-------+
		          |       |
		  +-------|       |+-------
		  |       |       |       |
		  | box6  | box7  | box8  |
		  |       |       |       |
		  +-------|       |+-------
		          |       |
		  +-------+-------+-------+
		  |       |       |       |
		  | box3  | box4  | box5  |
		  |       |       |       |
		  +-------+-------+-------+
		  |       |       |       |
		  | box0  | box1  | box2  |
		  |       |       |       |
		  +-------+-------+-------+

		Each box represents a square, except box7, which is three times taller.
		- box0 is adjacent to box1, box3
		- box1 is adjacent to box0, box2, box3, box4
		- box2 is adjacent to box1, box5
		- box3 is adjacent to box0, box1, box4
		- box4 is adjacent to box1, box3, box5, box7
		- box5 is adjacent to box2, box4
		- box6 is adjacent to box7 (positioned above box3 but not connected)
		- box7 is adjacent to box4, box6 (taller and positioned above box4)
		- box8 is adjacent to box7 (positioned above box5 but not connected)
	*/

	testCases := []struct {
		name       string
		from       *pctk.Positionf
		to         *pctk.Positionf
		expectedTo *pctk.Positionf
		assertFunc func(t *testing.T, path []*pctk.WayPoint, expectedTo *pctk.Positionf)
	}{
		{
			name:       "Should return a valid path when 'from' and 'to' are inside walk boxes",
			from:       &pctk.Positionf{X: 0.5, Y: 0.5}, // inside box0
			to:         &pctk.Positionf{X: 1.5, Y: 2.5}, // inside box7
			expectedTo: &pctk.Positionf{X: 1.5, Y: 2.5}, // expected return to
			assertFunc: func(t *testing.T, path []*pctk.WayPoint, expectedTo *pctk.Positionf) {
				assert.True(t, path[len(path)-1].Position.Equals(expectedTo))
			},
		},
		{
			name:       "Should return the closest point when 'to' is outside the closest walk box",
			from:       &pctk.Positionf{X: 0.5, Y: 0.5}, // inside box0
			to:         &pctk.Positionf{X: 3.5, Y: 1.5}, // outside all boxes, close to box5
			expectedTo: &pctk.Positionf{X: 3, Y: 1.5},   // expected to return the closest point inside box5
			assertFunc: func(t *testing.T, path []*pctk.WayPoint, expectedTo *pctk.Positionf) {
				assert.True(t, path[len(path)-1].Position.Equals(expectedTo))
			},
		},
		{
			name:       "Should return the closest point avoiding zigzags",
			from:       &pctk.Positionf{X: 0.5, Y: 3.4}, // inside box6
			to:         &pctk.Positionf{X: 3.5, Y: 1.5}, // outside all boxes, close to box5
			expectedTo: &pctk.Positionf{X: 3, Y: 1.5},   // expected to return the closest point inside box5
			assertFunc: func(t *testing.T, path []*pctk.WayPoint, expectedTo *pctk.Positionf) {
				assert.True(t, path[0].Position.Equals(&pctk.Positionf{X: 1, Y: 3.4}), fmt.Sprintf("should be the closest point but was %v", path[0].Position))
				assert.True(t, path[len(path)-1].Position.Equals(expectedTo))
			},
		},
	}

	box0 := pctk.NewWalkBox("walkbox0", [4]*pctk.Positionf{{0, 0}, {1, 0}, {1, 1}, {0, 1}})
	box1 := pctk.NewWalkBox("walkbox1", [4]*pctk.Positionf{{1, 0}, {2, 0}, {2, 1}, {1, 1}})
	box2 := pctk.NewWalkBox("walkbox2", [4]*pctk.Positionf{{2, 0}, {3, 0}, {3, 1}, {2, 1}})
	box3 := pctk.NewWalkBox("walkbox3", [4]*pctk.Positionf{{0, 1}, {1, 1}, {1, 2}, {0, 2}})
	box4 := pctk.NewWalkBox("walkbox4", [4]*pctk.Positionf{{1, 1}, {2, 1}, {2, 2}, {1, 2}})
	box5 := pctk.NewWalkBox("walkbox5", [4]*pctk.Positionf{{2, 1}, {3, 1}, {3, 2}, {2, 2}})
	box6 := pctk.NewWalkBox("walkbox6", [4]*pctk.Positionf{{1, 4}, {0, 4}, {0, 3}, {1, 3}}) // starts in top right vertex on purpose
	box7 := pctk.NewWalkBox("walkbox7", [4]*pctk.Positionf{{1, 2}, {2, 2}, {2, 5}, {1, 5}})
	box8 := pctk.NewWalkBox("walkbox8", [4]*pctk.Positionf{{2, 3}, {3, 3}, {3, 4}, {2, 4}})

	walkBoxMatrix := pctk.NewWalkBoxMatrix([]*pctk.WalkBox{box0, box1, box2, box3, box4, box5, box6, box7, box8})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path := walkBoxMatrix.FindPath(testCase.from, testCase.to)
			testCase.assertFunc(t, path, testCase.expectedTo)
		})
	}
}
