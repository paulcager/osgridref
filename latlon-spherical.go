package osgridref

import "C"
import (
    "fmt"
    "math"
)

/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
/* Latitude/longitude spherical geodesy tools                         (c) Chris Veness 2002-2019  */
/*                                                                                   MIT Licence  */
/* www.movable-type.co.uk/scripts/latlong.html                                                    */
/* www.movable-type.co.uk/scripts/geodesy-library.html#latlon-spherical                           */
/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */

const (
	π                     = math.Pi
	metresToKm            = 1.0 / 1000
	metresToMiles         = 1.0 / 1609.344
	metresToNauticalMiles = 1.0 / 1852
	earthRadius           = 6_371_000.0 // Its equatorial radius is 6378 km, but its polar radius is 6357 km
)


/**
 * Library of geodesy functions for operations on a spherical earth model.
 *
 * Includes distances, bearings, destinations, etc, for both great circle paths and rhumb lines,
 * and other related functions.
 *
 * All calculations are done using simple spherical trigonometric formulae.
 *
 * @module latlon-spherical
 */

// note greek letters (e.g. φ, λ, θ) are used for angles in radians to distinguish from angles in
// degrees (e.g. lat, lon, brng)


/* LatLon - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */


/**
 * Latitude/longitude points on a spherical model earth, and methods for calculating distances,
 * bearings, destinations, etc on (orthodromic) great-circle paths and (loxodromic) rhumb lines.
 */
type LatLon struct {
	Lat, Lon float64
}


/**
 * Returns the distance along the surface of the earth from ‘this’ point to destination point.
 *
 * Uses haversine formula: a = sin²(Δφ/2) + cosφ1·cosφ2 · sin²(Δλ/2); d = 2 · atan2(√a, √(a-1)).
 *
 * @param   {LatLon} point - Latitude/longitude of destination point.
 * @param   {number} [radius=6371e3] - Radius of earth (defaults to mean radius in metres).
 * @returns {number} Distance between this point and destination point, in same units as radius.
 * @throws  {TypeError} Invalid radius.
 *
 * @example
 *   const p1 = new LatLon(52.205, 0.119);
 *   const p2 = new LatLon(48.857, 2.351);
 *   const d = p1.distanceTo(p2);       // 404.3×10³ m
 *   const m = p1.distanceTo(p2, 3959); // 251.2 miles
 */
func (ll LatLon) DistanceTo(point LatLon) float64 {

    // a = sin²(Δφ/2) + cos(φ1)⋅cos(φ2)⋅sin²(Δλ/2)
    // δ = 2·atan2(√(a), √(1−a))
    // see mathforum.org/library/drmath/view/51879.html for derivation

    R := earthRadius
    φ1 := ll.Lat * toRadians
    λ1 := ll.Lon * toRadians
    φ2 := point.Lat * toRadians
    λ2 := point.Lon * toRadians
    Δφ := φ2 - φ1
    Δλ := λ2 - λ1

    a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    d := R * c

    return d
}


/**
 * Returns the initial bearing from ‘this’ point to destination point.
 *
 * @param   {LatLon} point - Latitude/longitude of destination point.
 * @returns {number} Initial bearing in degrees from north (0°..360°).
 *
 * @example
 *   const p1 = new LatLon(52.205, 0.119);
 *   const p2 = new LatLon(48.857, 2.351);
 *   const b1 = p1.initialBearingTo(p2); // 156.2°
 */
func (ll LatLon) InitialBearingTo(point LatLon) float64 {
    // tanθ = sinΔλ⋅cosφ2 / cosφ1⋅sinφ2 − sinφ1⋅cosφ2⋅cosΔλ
    // see mathforum.org/library/drmath/view/55417.html for derivation

    φ1 := ll.Lat * toRadians
    φ2 := point.Lat * toRadians
    Δλ := (point.Lon - ll.Lon) * toRadians

    x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
    y := math.Sin(Δλ) * math.Cos(φ2)
    θ := math.Atan2(y, x)

    bearing := θ* toDegrees

    return Wrap360(bearing)
}

/**
 * Returns final bearing arriving at destination point from ‘this’ point; the final bearing will
 * differ from the initial bearing by varying degrees according to distance and latitude.
 *
 * @param   {LatLon} point - Latitude/longitude of destination point.
 * @returns {number} Final bearing in degrees from north (0°..360°).
 *
 * @example
 *   const p1 = new LatLon(52.205, 0.119);
 *   const p2 = new LatLon(48.857, 2.351);
 *   const b2 = p1.finalBearingTo(p2); // 157.9°
 */
func (ll LatLon) FinalBearingTo(point LatLon) float64 {
    // get initial bearing from destination point to this point & reverse it by adding 180°

    bearing := point.InitialBearingTo(ll) + 180

    return Wrap360(bearing)
}


/**
 * Returns the midpoint between ‘this’ point and destination point.
 *
 * @param   {LatLon} point - Latitude/longitude of destination point.
 * @returns {LatLon} Midpoint between this point and destination point.
 *
 * @example
 *   const p1 = new LatLon(52.205, 0.119);
 *   const p2 = new LatLon(48.857, 2.351);
 *   const pMid = p1.midpointTo(p2); // 50.5363°N, 001.2746°E
 */
//func (ll LatLon) MidpointTo(point LatLon) LatLon{
//    // φm = atan2( sinφ1 + sinφ2, √( (cosφ1 + cosφ2⋅cosΔλ)² + cos²φ2⋅sin²Δλ ) )
//    // λm = λ1 + atan2(cosφ2⋅sinΔλ, cosφ1 + cosφ2⋅cosΔλ)
//    // midpoint is sum of vectors to two points: mathforum.org/library/drmath/view/51822.html
//
//    φ1 := ll.Lat* toRadians;
//    λ1 := ll.Lon* toRadians;
//    φ2 := point.Lat* toRadians;
//    Δλ := (point.Lon - ll.Lon)* toRadians;
//
//    // get cartesian coordinates for the two points
//    A := { x: math.Cos(φ1), y: 0, z: math.Sin(φ1) }; // place point A on prime meridian y=0
//    B := { x: math.Cos(φ2)*math.Cos(Δλ), y: math.Cos(φ2)*math.Sin(Δλ), z: math.Sin(φ2) };
//
//// vector to midpoint is sum of vectors to two points (no need to normalise)
//    C := { x: A.x + B.x, y: A.y + B.y, z: A.z + B.z };
//
//    φm := math.Atan2(C.z, math.Sqrt(C.x*C.x + C.y*C.y));
//    λm := λ1 + math.Atan2(C.y, C.x);
//
//    lat := φm* toDegrees;
//    lon := λm* toDegrees;
//
//    return LatLon{Lat: lat, Lon: lon}
//}


/**
 * Returns the point at given fraction between ‘this’ point and given point.
 *
 * @param   {LatLon} point - Latitude/longitude of destination point.
 * @param   {number} fraction - Fraction between the two points (0 = this point, 1 = specified point).
 * @returns {LatLon} Intermediate point between this point and destination point.
 *
 * @example
 *   const p1 = new LatLon(52.205, 0.119);
 *   const p2 = new LatLon(48.857, 2.351);
 *   const pInt = p1.intermediatePointTo(p2, 0.25); // 51.3721°N, 000.7073°E
 */
//intermediatePointTo(point, fraction) {
//if (!(point instanceof LatLon)) point = LatLon.parse(point); // allow literal forms
//if (this.equals(point)) return new LatLon(ll.Lat, ll.Lon); // coincident points
//
//    φ1 = ll.Lat* toRadians, λ1 = ll.Lon* toRadians;
//    φ2 = point.lat* toRadians, λ2 = point.lon* toRadians;
//
//// distance between points
//    Δφ = φ2 - φ1;
//    Δλ = λ2 - λ1;
//    a = math.Sin(Δφ/2) * math.Sin(Δφ/2)
//+ math.Cos(φ1) * math.Cos(φ2) * math.Sin(Δλ/2) * math.Sin(Δλ/2);
//    δ = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a));
//
//    A = math.Sin((1-fraction)*δ) / math.Sin(δ);
//    B = math.Sin(fraction*δ) / math.Sin(δ);
//
//    x = A * math.Cos(φ1) * math.Cos(λ1) + B * math.Cos(φ2) * math.Cos(λ2);
//    y = A * math.Cos(φ1) * math.Sin(λ1) + B * math.Cos(φ2) * math.Sin(λ2);
//    z = A * math.Sin(φ1) + B * math.Sin(φ2);
//
//    φ3 = math.Atan2(z, math.Sqrt(x*x + y*y));
//    λ3 = math.Atan2(y, x);
//
//    lat = φ3* toDegrees;
//    lon = λ3* toDegrees;
//
//return new LatLon(lat, lon);
//}


/**
 * Returns the destination point from ‘this’ point having travelled the given distance on the
 * given initial bearing (bearing normally varies around path followed).
 *
 * @param   {number} distance - Distance travelled, in same units as earth radius (default: metres).
 * @param   {number} bearing - Initial bearing in degrees from north.
 * @param   {number} [radius=6371e3] - (Mean) radius of earth (defaults to radius in metres).
 * @returns {LatLon} Destination point.
 *
 * @example
 *   const p1 = new LatLon(51.47788, -0.00147);
 *   const p2 = p1.destinationPoint(7794, 300.7); // 51.5136°N, 000.0983°W
 */
func (ll LatLon) DestinationPoint(distance float64, bearing float64) LatLon {
    // sinφ2 = sinφ1⋅cosδ + cosφ1⋅sinδ⋅cosθ
    // tanΔλ = sinθ⋅sinδ⋅cosφ1 / cosδ−sinφ1⋅sinφ2
    // see mathforum.org/library/drmath/view/52049.html for derivation

    δ := distance / earthRadius // angular distance in radians
    θ := bearing * toRadians

    φ1 := ll.Lat * toRadians
    λ1 := ll.Lon * toRadians

    sinφ2 := math.Sin(φ1)*math.Cos(δ) + math.Cos(φ1)*math.Sin(δ)*math.Cos(θ)
    φ2 := math.Asin(sinφ2)
    y := math.Sin(θ) * math.Sin(δ) * math.Cos(φ1)
    x := math.Cos(δ) - math.Sin(φ1)*sinφ2
    λ2 := λ1 + math.Atan2(y, x)

    lat := φ2 * toDegrees
    lon := λ2 * toDegrees

    return LatLon{Lat: lat, Lon: lon}
}


/**
 * Returns the point of intersection of two paths defined by point and bearing.
 *
 * @param   {LatLon}      p1 - First point.
 * @param   {number}      brng1 - Initial bearing from first point.
 * @param   {LatLon}      p2 - Second point.
 * @param   {number}      brng2 - Initial bearing from second point.
 * @returns {LatLon|null} Destination point (null if no unique intersection defined).
 *
 * @example
 *   const p1 = new LatLon(51.8853, 0.2545), brng1 = 108.547;
 *   const p2 = new LatLon(49.0034, 2.5735), brng2 =  32.435;
 *   const pInt = LatLon.intersection(p1, brng1, p2, brng2); // 50.9078°N, 004.5084°E
 */
func Intersection(p1 LatLon, brng1 float64, p2 LatLon, brng2 float64) (LatLon, bool) {
    // see www.edwilliams.org/avform.htm#Intersection

    φ1, λ1 := p1.Lat*toRadians, p1.Lon*toRadians
    φ2, λ2 := p2.Lat*toRadians, p2.Lon*toRadians
    θ13, θ23 := brng1*toRadians, brng2*toRadians
    Δφ := φ2 - φ1
    Δλ := λ2 - λ1

    // angular distance p1-p2
    δ12 := 2 * math.Asin(math.Sqrt(math.Sin(Δφ/2)*math.Sin(Δφ/2)+math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)))
    if math.Abs(δ12) <= math.SmallestNonzeroFloat64 {
        return p1, true
    }

    // initial/final bearings between points
    cosθa := (math.Sin(φ2) - math.Sin(φ1)*math.Cos(δ12)) / (math.Sin(δ12) * math.Cos(φ1))
    cosθb := (math.Sin(φ1) - math.Sin(φ2)*math.Cos(δ12)) / (math.Sin(δ12) * math.Cos(φ2))
    θa := math.Acos(math.Min(math.Max(cosθa, -1), 1)) // protect against rounding errors
    θb := math.Acos(math.Min(math.Max(cosθb, -1), 1)) // protect against rounding errors

    θ12 := θa
    if math.Sin(λ2-λ1) <= 0 {
        θ12 = 2*π - θa
    }

    θ21 := 2*π - θb
    if math.Sin(λ2-λ1) <= 0 {
        θ21 = θb
    }

    α1 := θ13 - θ12 // angle 2-1-3
    α2 := θ21 - θ23 // angle 1-2-3

    if math.Sin(α1) == 0 && math.Sin(α2) == 0 {
        // infinite intersections
        return LatLon{}, false
    }

    if math.Sin(α1)*math.Sin(α2) < 0 {
        // ambiguous intersection (antipodal?)
        return LatLon{}, false
    }

    cosα3 := -math.Cos(α1)*math.Cos(α2) + math.Sin(α1)*math.Sin(α2)*math.Cos(δ12)

    δ13 := math.Atan2(math.Sin(δ12)*math.Sin(α1)*math.Sin(α2), math.Cos(α2)+math.Cos(α1)*cosα3)

    φ3 := math.Asin(math.Min(math.Max(math.Sin(φ1)*math.Cos(δ13)+math.Cos(φ1)*math.Sin(δ13)*math.Cos(θ13), -1), 1))

    Δλ13 := math.Atan2(math.Sin(θ13)*math.Sin(δ13)*math.Cos(φ1), math.Cos(δ13)-math.Sin(φ1)*math.Sin(φ3))
    λ3 := λ1 + Δλ13

    lat := φ3 * toDegrees
    lon := λ3 * toDegrees

    return LatLon{Lat: lat, Lon: lon}, true
}


///**
// * Returns (signed) distance from ‘this’ point to great circle defined by start-point and
// * end-point.
// *
// * @param   {LatLon} pathStart - Start point of great circle path.
// * @param   {LatLon} pathEnd - End point of great circle path.
// * @param   {number} [radius=6371e3] - (Mean) radius of earth (defaults to radius in metres).
// * @returns {number} Distance to great circle (-ve if to left, +ve if to right of path).
// *
// * @example
// *   const pCurrent = new LatLon(53.2611, -0.7972);
// *   const p1 = new LatLon(53.3206, -1.7297);
// *   const p2 = new LatLon(53.1887, 0.1334);
// *   const d = pCurrent.crossTrackDistanceTo(p1, p2);  // -307.5 m
// */
//crossTrackDistanceTo(pathStart, pathEnd, radius=6371e3) {
//if (!(pathStart instanceof LatLon)) pathStart = LatLon.parse(pathStart); // allow literal forms
//if (!(pathEnd instanceof LatLon)) pathEnd = LatLon.parse(pathEnd);       // allow literal forms
//    R = radius;
//
//if (this.equals(pathStart)) return 0;
//
//    δ13 = pathStart.distanceTo(this, R) / R;
//    θ13 = pathStart.initialBearingTo(this)* toRadians;
//    θ12 = pathStart.initialBearingTo(pathEnd)* toRadians;
//
//    δxt = math.Asin(math.Sin(δ13) * math.Sin(θ13 - θ12));
//
//return δxt * R;
//}
//
//
///**
// * Returns how far ‘this’ point is along a path from from start-point, heading towards end-point.
// * That is, if a perpendicular is drawn from ‘this’ point to the (great circle) path, the
// * along-track distance is the distance from the start point to where the perpendicular crosses
// * the path.
// *
// * @param   {LatLon} pathStart - Start point of great circle path.
// * @param   {LatLon} pathEnd - End point of great circle path.
// * @param   {number} [radius=6371e3] - (Mean) radius of earth (defaults to radius in metres).
// * @returns {number} Distance along great circle to point nearest ‘this’ point.
// *
// * @example
// *   const pCurrent = new LatLon(53.2611, -0.7972);
// *   const p1 = new LatLon(53.3206, -1.7297);
// *   const p2 = new LatLon(53.1887,  0.1334);
// *   const d = pCurrent.alongTrackDistanceTo(p1, p2);  // 62.331 km
// */
//alongTrackDistanceTo(pathStart, pathEnd, radius=6371e3) {
//if (!(pathStart instanceof LatLon)) pathStart = LatLon.parse(pathStart); // allow literal forms
//if (!(pathEnd instanceof LatLon)) pathEnd = LatLon.parse(pathEnd);       // allow literal forms
//    R = radius;
//
//if (this.equals(pathStart)) return 0;
//
//    δ13 = pathStart.distanceTo(this, R) / R;
//    θ13 = pathStart.initialBearingTo(this)* toRadians;
//    θ12 = pathStart.initialBearingTo(pathEnd)* toRadians;
//
//    δxt = math.Asin(math.Sin(δ13) * math.Sin(θ13-θ12));
//
//    δat = math.Acos(math.Cos(δ13) / Math.abs(math.Cos(δxt)));
//
//return δat*Math.sign(math.Cos(θ12-θ13)) * R;
//}
//
//
///**
// * Returns maximum latitude reached when travelling on a great circle on given bearing from
// * ‘this’ point (‘Clairaut’s formula’). Negate the result for the minimum latitude (in the
// * southern hemisphere).
// *
// * The maximum latitude is independent of longitude; it will be the same for all points on a
// * given latitude.
// *
// * @param   {number} bearing - Initial bearing.
// * @returns {number} Maximum latitude reached.
// */
//maxLatitude(bearing) {
//    θ = Number(bearing)* toRadians;
//
//    φ = ll.Lat* toRadians;
//
//    φMax = math.Acos(Math.abs(math.Sin(θ) * math.Cos(φ)));
//
//return φMax* toDegrees;
//}
//
//
///**
// * Returns the pair of meridians at which a great circle defined by two points crosses the given
// * latitude. If the great circle doesn't reach the given latitude, null is returned.
// *
// * @param   {LatLon}      point1 - First point defining great circle.
// * @param   {LatLon}      point2 - Second point defining great circle.
// * @param   {number}      latitude - Latitude crossings are to be determined for.
// * @returns {Object|null} Object containing { lon1, lon2 } or null if given latitude not reached.
// */
//static crossingParallels(point1, point2, latitude) {
//if (point1.equals(point2)) return null; // coincident points
//
//    φ = Number(latitude)* toRadians;
//
//    φ1 = point1.lat* toRadians;
//    λ1 = point1.lon* toRadians;
//    φ2 = point2.lat* toRadians;
//    λ2 = point2.lon* toRadians;
//
//    Δλ = λ2 - λ1;
//
//    x = math.Sin(φ1) * math.Cos(φ2) * math.Cos(φ) * math.Sin(Δλ);
//    y = math.Sin(φ1) * math.Cos(φ2) * math.Cos(φ) * math.Cos(Δλ) - math.Cos(φ1) * math.Sin(φ2) * math.Cos(φ);
//    z = math.Cos(φ1) * math.Cos(φ2) * math.Sin(φ) * math.Sin(Δλ);
//
//if (z * z > x * x + y * y) return null; // great circle doesn't reach latitude
//
//    λm = math.Atan2(-y, x);               // longitude at max latitude
//    Δλi = math.Acos(z / math.Sqrt(x*x + y*y)); // Δλ from λm to intersection points
//
//    λi1 = λ1 + λm - Δλi;
//    λi2 = λ1 + λm + Δλi;
//
//    lon1 = λi1* toDegrees;
//    lon2 = λi2* toDegrees;
//
//return {
//lon1: Dms.wrap180(lon1),
//lon2: Dms.wrap180(lon2),
//};
//}


///* Rhumb - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
//
//
///**
// * Returns the distance travelling from ‘this’ point to destination point along a rhumb line.
// *
// * @param   {LatLon} point - Latitude/longitude of destination point.
// * @param   {number} [radius=6371e3] - (Mean) radius of earth (defaults to radius in metres).
// * @returns {number} Distance in km between this point and destination point (same units as radius).
// *
// * @example
// *   const p1 = new LatLon(51.127, 1.338);
// *   const p2 = new LatLon(50.964, 1.853);
// *   const d = p1.distanceTo(p2); //  40.31 km
// */
//rhumbDistanceTo(point, radius=6371e3) {
//if (!(point instanceof LatLon)) point = LatLon.parse(point); // allow literal forms
//
//// see www.edwilliams.org/avform.htm#Rhumb
//
//    R = radius;
//    φ1 = ll.Lat* toRadians;
//    φ2 = point.lat* toRadians;
//    Δφ = φ2 - φ1;
//let Δλ = Math.abs(point.lon - ll.Lon)* toRadians;
//// if dLon over 180° take shorter rhumb line across the anti-meridian:
//if (Math.abs(Δλ) > π) Δλ = Δλ > 0 ? -(2 * π - Δλ) : (2 * π + Δλ);
//
//// on Mercator projection, longitude distances shrink by latitude; q is the 'stretch factor'
//// q becomes ill-conditioned along E-W line (0/0); use empirical tolerance to avoid it
//    Δψ = Math.log(math.Tan(φ2 / 2 + π / 4) / math.Tan(φ1 / 2 + π / 4));
//    q = Math.abs(Δψ) > 10e-12 ? Δφ / Δψ : math.Cos(φ1);
//
//// distance is pythagoras on 'stretched' Mercator projection, √(Δφ² + q²·Δλ²)
//    δ = math.Sqrt(Δφ*Δφ + q*q * Δλ*Δλ); // angular distance in radians
//    d = δ * R;
//
//return d;
//}
//
//
///**
// * Returns the bearing from ‘this’ point to destination point along a rhumb line.
// *
// * @param   {LatLon}    point - Latitude/longitude of destination point.
// * @returns {number}    Bearing in degrees from north.
// *
// * @example
// *   const p1 = new LatLon(51.127, 1.338);
// *   const p2 = new LatLon(50.964, 1.853);
// *   const d = p1.rhumbBearingTo(p2); // 116.7°
// */
//rhumbBearingTo(point) {
//if (!(point instanceof LatLon)) point = LatLon.parse(point); // allow literal forms
//if (this.equals(point)) return NaN; // coincident points
//
//    φ1 = ll.Lat* toRadians;
//    φ2 = point.lat* toRadians;
//let Δλ = (point.lon - ll.Lon)* toRadians;
//// if dLon over 180° take shorter rhumb line across the anti-meridian:
//if (Math.abs(Δλ) > π) Δλ = Δλ > 0 ? -(2 * π - Δλ) : (2 * π + Δλ);
//
//    Δψ = Math.log(math.Tan(φ2 / 2 + π / 4) / math.Tan(φ1 / 2 + π / 4));
//
//    θ = math.Atan2(Δλ, Δψ);
//
//    bearing = θ* toDegrees;
//
//return Dms.wrap360(bearing);
//}
//
//
///**
// * Returns the destination point having travelled along a rhumb line from ‘this’ point the given
// * distance on the given bearing.
// *
// * @param   {number} distance - Distance travelled, in same units as earth radius (default: metres).
// * @param   {number} bearing - Bearing in degrees from north.
// * @param   {number} [radius=6371e3] - (Mean) radius of earth (defaults to radius in metres).
// * @returns {LatLon} Destination point.
// *
// * @example
// *   const p1 = new LatLon(51.127, 1.338);
// *   const p2 = p1.rhumbDestinationPoint(40300, 116.7); // 50.9642°N, 001.8530°E
// */
//rhumbDestinationPoint(distance, bearing, radius=6371e3) {
//    φ1 = ll.Lat* toRadians, λ1 = ll.Lon* toRadians;
//    θ = Number(bearing)* toRadians;
//
//    δ = distance / radius; // angular distance in radians
//
//    Δφ = δ * math.Cos(θ);
//let φ2 = φ1 + Δφ;
//
//// check for some daft bugger going past the pole, normalise latitude if so
//if (Math.abs(φ2) > π / 2) φ2 = φ2 > 0 ? π - φ2 : -π - φ2;
//
//    Δψ = Math.log(math.Tan(φ2 / 2 + π / 4) / math.Tan(φ1 / 2 + π / 4));
//    q = Math.abs(Δψ) > 10e-12 ? Δφ / Δψ : math.Cos(φ1); // E-W course becomes ill-conditioned with 0/0
//
//    Δλ = δ * math.Sin(θ) / q;
//    λ2 = λ1 + Δλ;
//
//    lat = φ2* toDegrees;
//    lon = λ2* toDegrees;
//
//return new LatLon(lat, lon);
//}
//
//
///**
// * Returns the loxodromic midpoint (along a rhumb line) between ‘this’ point and second point.
// *
// * @param   {LatLon} point - Latitude/longitude of second point.
// * @returns {LatLon} Midpoint between this point and second point.
// *
// * @example
// *   const p1 = new LatLon(51.127, 1.338);
// *   const p2 = new LatLon(50.964, 1.853);
// *   const pMid = p1.rhumbMidpointTo(p2); // 51.0455°N, 001.5957°E
// */
//rhumbMidpointTo(point) {
//if (!(point instanceof LatLon)) point = LatLon.parse(point); // allow literal forms
//
//// see mathforum.org/kb/message.jspa?messageID=148837
//
//    φ1 = ll.Lat* toRadians; let λ1 = ll.Lon* toRadians;
//    φ2 = point.lat* toRadians, λ2 = point.lon* toRadians;
//
//if (Math.abs(λ2 - λ1) > π) λ1 += 2 * π; // crossing anti-meridian
//
//    φ3 = (φ1 + φ2) / 2;
//    f1 = math.Tan(π / 4 + φ1 / 2);
//    f2 = math.Tan(π / 4 + φ2 / 2);
//    f3 = math.Tan(π / 4 + φ3 / 2);
//let λ3 = ((λ2 - λ1) * Math.log(f3) + λ1 * Math.log(f2) - λ2 * Math.log(f1)) / Math.log(f2 / f1);
//
//if (!isFinite(λ3)) λ3 = (λ1 + λ2) / 2; // parallel of latitude
//
//    lat = φ3* toDegrees;
//    lon = λ3* toDegrees;
//
//return new LatLon(lat, lon);
//}


/* Area - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - */


/**
 * Calculates the area of a spherical polygon where the sides of the polygon are great circle
 * arcs joining the vertices.
 *
 * @param   {LatLon[]} polygon - Array of points defining vertices of the polygon.
 * @param   {number}   [radius=6371e3] - (Mean) radius of earth (defaults to radius in metres).
 * @returns {number}   The area of the polygon in the same units as radius.
 *
 * @example
 *   const polygon = [new LatLon(0,0), new LatLon(1,0), new LatLon(0,1)];
 *   const area = LatLon.areaOf(polygon); // 6.18e9 m²
 */
func AreaOf(polygon []LatLon) float64 {
    // uses method due to Karney: osgeo-org.1560.x6.nabble.com/Area-of-a-spherical-polygon-td3841625.html;
    // for each edge of the polygon, tan(E/2) = tan(Δλ/2)·(tan(φ₁/2)+tan(φ₂/2)) / (1+tan(φ₁/2)·tan(φ₂/2))
    // where E is the spherical excess of the trapezium obtained by extending the edge to the equator
    // (Karney's method is probably more efficient than the more widely known L’Huilier’s Theorem)

    const R = earthRadius

    // close polygon so that last point equals first point
    closed := polygon[0] == polygon[len(polygon)-1]
    if !closed {
        polygon = append(polygon, polygon[0])
    }
    nVertices := len(polygon) - 1

    var S float64 // spherical excess in steradians
    for v := 0; v < nVertices; v++ {
        φ1 := polygon[v].Lat * toRadians
        φ2 := polygon[v+1].Lat * toRadians
        Δλ := (polygon[v+1].Lon - polygon[v].Lon) * toRadians
        E := 2 * math.Atan2(math.Tan(Δλ/2)*(math.Tan(φ1/2)+math.Tan(φ2/2)), 1+math.Tan(φ1/2)*math.Tan(φ2/2))
        S += E
    }

    if isPoleEnclosedBy(polygon) {
        S = math.Abs(S) - 2*π
    }

    A := math.Abs(S * R * R) // area in units of R

    if !closed {
        polygon = polygon[:len(polygon)-1]
    }

    return A
}

// returns whether polygon encloses pole: sum of course deltas around pole is 0° rather than
// normal ±360°: blog.element84.com/determining-if-a-spherical-polygon-contains-a-pole.html
func isPoleEnclosedBy(p []LatLon) bool {
    // TODO: any better test than this?
    ΣΔ := 0.0
    prevBrng := p[0].InitialBearingTo(p[1])
    for v := 0; v < len(p)-1; v++ {
        initBrng := p[v].InitialBearingTo(p[v+1])
        finalBrng := p[v].FinalBearingTo(p[v+1])
        ΣΔ += math.Mod(initBrng-prevBrng+540, 360) - 180
        ΣΔ += math.Mod(finalBrng-initBrng+540,360) - 180
        prevBrng = finalBrng
    }
    initBrng := p[0].InitialBearingTo(p[1])
    ΣΔ += float64(int(initBrng-prevBrng+540)%360 - 180)
    // TODO: fix (intermittant) edge crossing pole - eg (85,90), (85,0), (85,-90)
    enclosed := math.Abs(ΣΔ) < 90 // 0°-ish
    return enclosed
}



/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */



/**
 * Returns a string representation of ‘this’ point, formatted as degrees, degrees+minutes, or
 * degrees+minutes+seconds.
 *
 * @param   {string} [format=d] - Format point as 'd', 'dm', 'dms', or 'n' for signed numeric.
 * @param   {number} [dp=4|2|0] - Number of decimal places to use: default 4 for d, 2 for dm, 0 for dms.
 * @returns {string} Comma-separated formatted latitude/longitude.
 * @throws  {RangeError} Invalid format.
 *
 * @example
 *   const greenwich = new LatLon(51.47788, -0.00147);
 *   const d = greenwich.toString();                        // 51.4779°N, 000.0015°W
 *   const dms = greenwich.toString('dms', 2);              // 51°28′40.37″N, 000°00′05.29″W
 *   const [lat, lon] = greenwich.toString('n').split(','); // 51.4779, -0.0015
 */
func (ll LatLon)String() string {
    return fmt.Sprintf("%f,%f", ll.Lat, ll.Lon)
}
/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
