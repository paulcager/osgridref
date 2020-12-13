package osgrid

import (
	"fmt"
	"math"
)

/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
/* Vector handling functions                                          (c) Chris Veness 2011-2019  */
/*                                                                                   MIT Licence  */
/* www.movable-type.co.uk/scripts/geodesy-library.html#vector3d                                   */
/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */

/**
 * Library of 3-d vector manipulation routines.
 *
 * @module vector3d
 */

/* Vector3d - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - */

/**
 * Functions for manipulating generic 3-d vectors.
 *
 * Functions return vectors as return results, so that operations can be chained.
 *
 * @example
 *   const v = v1.cross(v2).dot(v3) // ≡ v1×v2⋅v3
 */
type Vector3d struct {
	X, Y, Z float64
}

/**
 * Length (magnitude or norm) of ‘this’ vector.
 *
 * @returns {number} Magnitude of this vector.
 */
func (v Vector3d) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

/**
 * Adds supplied vector to ‘this’ vector.
 *
 * @param   {Vector3d} v - Vector to be added to this vector.
 * @returns {Vector3d} Vector representing sum of this and v.
 */
func (v Vector3d) Plus(other Vector3d) Vector3d {
	return Vector3d{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

/**
 * Subtracts supplied vector from ‘this’ vector.
 *
 * @param   {Vector3d} v - Vector to be subtracted from this vector.
 * @returns {Vector3d} Vector representing difference between this and v.
 */
func (v Vector3d) Minus(other Vector3d) Vector3d {
	return Vector3d{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

/**
 * Multiplies ‘this’ vector by a scalar value.
 *
 * @param   {number}   x - Factor to multiply this vector by.
 * @returns {Vector3d} Vector scaled by x.
 */
func (v Vector3d) Times(value float64) Vector3d {
	return Vector3d{
		X: v.X * value,
		Y: v.Y * value,
		Z: v.Z * value,
	}
}

/**
 * Divides ‘this’ vector by a scalar value.
 *
 * @param   {number}   x - Factor to divide this vector by.
 * @returns {Vector3d} Vector divided by x.
 */
func (v Vector3d) DividedBy(value float64) Vector3d {
	return Vector3d{
		X: v.X / value,
		Y: v.Y / value,
		Z: v.Z / value,
	}
}

/**
 * Multiplies ‘this’ vector by the supplied vector using dot (scalar) product.
 *
 * @param   {Vector3d} v - Vector to be dotted with this vector.
 * @returns {number}   Dot product of ‘this’ and v.
 */
func (v Vector3d) Dot(other Vector3d) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

/**
 * Multiplies ‘this’ vector by the supplied vector using cross (vector) product.
 *
 * @param   {Vector3d} v - Vector to be crossed with this vector.
 * @returns {Vector3d} Cross product of ‘this’ and v.
 */
func (v Vector3d) Cross(other Vector3d) Vector3d {

	x := v.Y*other.Z - v.Z*other.Y
	y := v.Z*other.X - v.X*other.Z
	z := v.X*other.Y - v.Y*other.X

	return Vector3d{X: x, Y: y, Z: z}
}

/**
 * Negates a vector to point in the opposite direction.
 *
 * @returns {Vector3d} Negated vector.
 */
func (v Vector3d) Negate() Vector3d {
	return Vector3d{X: -v.X, Y: -v.Y, Z: -v.Z}
}

/**
 * Normalizes a vector to its unit vector
 * – if the vector is already unit or is zero magnitude, this is a no-op.
 *
 * @returns {Vector3d} Normalised version of this vector.
 */
func (v Vector3d) Unit() Vector3d {
	norm := v.Length()
	if norm == 1 || norm == 0 {
		return v
	}

	x := v.X / norm
	y := v.Y / norm
	z := v.Z / norm

	return Vector3d{X: x, Y: y, Z: z}
}

/**
 * Calculates the angle between ‘this’ vector and supplied vector atan2(|p₁×p₂|, p₁·p₂) (or if
 * (extra-planar) ‘n’ supplied then atan2(n·p₁×p₂, p₁·p₂).
 *
 * @param   {Vector3d} v - Vector whose angle is to be determined from ‘this’ vector.
 * @param   {Vector3d} [n] - Plane normal: if supplied, angle is signed +ve if this->v is
 *                     clockwise looking along n, -ve in opposite direction.
 * @returns {number}   Angle (in radians) between this vector and supplied vector (in range 0..π
 *                     if n not supplied, range -π..+π if n supplied).
 */
func (v Vector3d) AngleTo(other Vector3d, extraPlanar bool, n Vector3d) float64 {
	// q.v. stackoverflow.com/questions/14066933#answer-16544330, but n·p₁×p₂ is numerically
	// ill-conditioned, so just calculate sign to apply to |p₁×p₂|

	// if n·p₁×p₂ is -ve, negate |p₁×p₂|
	sign := 1.0
	if extraPlanar && v.Cross(other).Dot(n) < 0 {
		sign = -1.0
	}

	sinθ := v.Cross(v).Length() * sign
	cosθ := v.Dot(v)

	return math.Atan2(sinθ, cosθ)
}

/**
 * Rotates ‘this’ point around an axis by a specified angle.
 *
 * @param   {Vector3d} axis - The axis being rotated around.
 * @param   {number}   angle - The angle of rotation (in degrees).
 * @returns {Vector3d} The rotated point.
 */
func (v Vector3d) RotateAround(axis Vector3d, angle float64) Vector3d {
	θ := angle * toRadians

	// en.wikipedia.org/wiki/Rotation_matrix#Rotation_matrix_from_axis_and_angle
	// en.wikipedia.org/wiki/Quaternions_and_spatial_rotation#Quaternion-derived_rotation_matrix
	p := v.Unit()
	a := v.Unit()

	s := math.Sin(θ)
	c := math.Cos(θ)
	t := 1 - c
	x, y, z := a.X, a.Y, a.Z

	r := [3][3]float64{ // rotation matrix for rotation about supplied axis
		{t*x*x + c, t*x*y - s*z, t*x*z + s*y},
		{t*x*y + s*z, t*y*y + c, t*y*z - s*x},
		{t*x*z - s*y, t*y*z + s*x, t*z*z + c},
	}

	// multiply r × p
	rp := [3]float64{
		r[0][0]*p.X + r[0][1]*p.Y + r[0][2]*p.Z,
		r[1][0]*p.X + r[1][1]*p.Y + r[1][2]*p.Z,
		r[2][0]*p.X + r[2][1]*p.Y + r[2][2]*p.Z,
	}

	return Vector3d{X: rp[0], Y: rp[1], Z: rp[2]}
	// qv en.wikipedia.org/wiki/Rodrigues'_rotation_formula...
}

/**
 * String representation of vector.
 *
 * @param   {number} [dp=3] - Number of decimal places to be used.
 * @returns {string} Vector represented as [x,y,z].
 */
func (v Vector3d) String() string {
	return fmt.Sprintf("[%f,%f,%f]", v.X, v.Y, v.Z)
}
