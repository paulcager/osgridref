package osgridref

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector3d_AngleTo(t *testing.T) {
	tests := []struct {
		name        string
		v1          Vector3d
		v2          Vector3d
		extraPlanar bool
		n           Vector3d
		wantRadians float64
	}{
		{
			name:        "90 degrees between x and y axis",
			v1:          Vector3d{X: 1, Y: 0, Z: 0},
			v2:          Vector3d{X: 0, Y: 1, Z: 0},
			extraPlanar: false,
			wantRadians: math.Pi / 2, // 90 degrees
		},
		{
			name:        "180 degrees - opposite vectors",
			v1:          Vector3d{X: 1, Y: 0, Z: 0},
			v2:          Vector3d{X: -1, Y: 0, Z: 0},
			extraPlanar: false,
			wantRadians: math.Pi, // 180 degrees
		},
		{
			name:        "0 degrees - same direction",
			v1:          Vector3d{X: 1, Y: 0, Z: 0},
			v2:          Vector3d{X: 2, Y: 0, Z: 0}, // same direction, different magnitude
			extraPlanar: false,
			wantRadians: 0,
		},
		{
			name:        "45 degrees",
			v1:          Vector3d{X: 1, Y: 0, Z: 0},
			v2:          Vector3d{X: 1, Y: 1, Z: 0},
			extraPlanar: false,
			wantRadians: math.Pi / 4, // 45 degrees
		},
		{
			name:        "signed angle - positive",
			v1:          Vector3d{X: 1, Y: 0, Z: 0},
			v2:          Vector3d{X: 0, Y: 1, Z: 0},
			extraPlanar: true,
			n:           Vector3d{X: 0, Y: 0, Z: 1}, // normal pointing up
			wantRadians: math.Pi / 2,                // +90 degrees (counterclockwise)
		},
		{
			name:        "signed angle - negative",
			v1:          Vector3d{X: 1, Y: 0, Z: 0},
			v2:          Vector3d{X: 0, Y: -1, Z: 0},
			extraPlanar: true,
			n:           Vector3d{X: 0, Y: 0, Z: 1}, // normal pointing up
			wantRadians: -math.Pi / 2,               // -90 degrees (clockwise)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.AngleTo(tt.v2, tt.extraPlanar, tt.n)
			assert.InDelta(t, tt.wantRadians, got, 1e-10, "angle should be %f radians (%f degrees), got %f radians (%f degrees)",
				tt.wantRadians, tt.wantRadians*180/math.Pi, got, got*180/math.Pi)
		})
	}
}

func TestVector3d_RotateAround(t *testing.T) {
	tests := []struct {
		name   string
		vector Vector3d
		axis   Vector3d
		angle  float64 // in degrees
		want   Vector3d
	}{
		{
			name:   "rotate x-axis 90° around z-axis -> y-axis",
			vector: Vector3d{X: 1, Y: 0, Z: 0},
			axis:   Vector3d{X: 0, Y: 0, Z: 1},
			angle:  90,
			want:   Vector3d{X: 0, Y: 1, Z: 0},
		},
		{
			name:   "rotate y-axis -90° around z-axis -> x-axis",
			vector: Vector3d{X: 0, Y: 1, Z: 0},
			axis:   Vector3d{X: 0, Y: 0, Z: 1},
			angle:  -90,
			want:   Vector3d{X: 1, Y: 0, Z: 0},
		},
		{
			name:   "rotate x-axis 180° around z-axis -> -x-axis",
			vector: Vector3d{X: 1, Y: 0, Z: 0},
			axis:   Vector3d{X: 0, Y: 0, Z: 1},
			angle:  180,
			want:   Vector3d{X: -1, Y: 0, Z: 0},
		},
		{
			name:   "rotate x-axis 90° around y-axis -> -z-axis",
			vector: Vector3d{X: 1, Y: 0, Z: 0},
			axis:   Vector3d{X: 0, Y: 1, Z: 0},
			angle:  90,
			want:   Vector3d{X: 0, Y: 0, Z: -1},
		},
		{
			name:   "rotate around arbitrary axis",
			vector: Vector3d{X: 1, Y: 0, Z: 0},
			axis:   Vector3d{X: 1, Y: 1, Z: 0}, // 45° diagonal in XY plane
			angle:  180,
			want:   Vector3d{X: 0, Y: 1, Z: 0},
		},
		{
			name:   "no rotation (0 degrees)",
			vector: Vector3d{X: 1, Y: 2, Z: 3},
			axis:   Vector3d{X: 0, Y: 0, Z: 1},
			angle:  0,
			want:   Vector3d{X: 1, Y: 2, Z: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vector.RotateAround(tt.axis, tt.angle)
			// Normalize to unit vectors for comparison
			gotNorm := got.Unit()
			wantNorm := tt.want.Unit()
			assert.InDelta(t, wantNorm.X, gotNorm.X, 1e-10, "X component mismatch")
			assert.InDelta(t, wantNorm.Y, gotNorm.Y, 1e-10, "Y component mismatch")
			assert.InDelta(t, wantNorm.Z, gotNorm.Z, 1e-10, "Z component mismatch")
		})
	}
}

func TestVector3d_BasicOperations(t *testing.T) {
	v1 := Vector3d{X: 1, Y: 2, Z: 3}
	v2 := Vector3d{X: 4, Y: 5, Z: 6}

	// Test Length
	assert.InDelta(t, math.Sqrt(14), v1.Length(), 1e-10)

	// Test Dot product
	assert.Equal(t, 32.0, v1.Dot(v2)) // 1*4 + 2*5 + 3*6 = 32

	// Test Cross product
	cross := v1.Cross(v2)
	// i(2*6 - 3*5) - j(1*6 - 3*4) + k(1*5 - 2*4)
	// i(12-15) - j(6-12) + k(5-8)
	// = -3i + 6j - 3k
	assert.Equal(t, -3.0, cross.X)
	assert.Equal(t, 6.0, cross.Y)
	assert.Equal(t, -3.0, cross.Z)

	// Test Unit
	unit := v1.Unit()
	assert.InDelta(t, 1.0, unit.Length(), 1e-10)
}
