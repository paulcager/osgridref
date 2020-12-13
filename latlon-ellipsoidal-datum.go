package osgrid

import (
	"fmt"
	"math"
	"strings"
)

/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
/* Geodesy tools for conversions between (historical) datums          (c) Chris Veness 2005-2019  */
/*                                                                                   MIT Licence  */
/* www.movable-type.co.uk/scripts/latlong-convert-coords.html                                     */
/* www.movable-type.co.uk/scripts/geodesy-library.html#latlon-ellipsoidal-datum                  */
/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */

/**
 * Historical geodetic datums: a latitude/longitude point defines a geographic location on or
 * above/below the  earth’s surface, measured in degrees from the equator & the International
 * Reference Meridian and metres above the ellipsoid, and based on a given datum. The datum is
 * based on a reference ellipsoid and tied to geodetic survey reference points.
 *
 * Modern geodesy is generally based on the WGS84 datum (as used for instance by GPS systems), but
 * previously various reference ellipsoids and datum references were used.
 *
 * This module extends the core latlon-ellipsoidal module to include ellipsoid parameters and datum
 * transformation parameters, and methods for converting between different (generally historical)
 * datums.
 *
 * It can be used for UK Ordnance Survey mapping (OS National Grid References are still based on the
 * otherwise historical OSGB36 datum), as well as for historical purposes.
 *
 * q.v. Ordnance Survey ‘A guide to coordinate systems in Great Britain’ Section 6,
 * www.ordnancesurvey.co.uk/docs/support/guide-coordinate-systems-great-britain.pdf, and also
 * www.ordnancesurvey.co.uk/blog/2014/12/2.
 *
 * @module latlon-ellipsoidal-datum
 */

/*
 * Ellipsoid parameters; exposed through static getter below.
 */
type Ellipseoid struct{ a, b, f float64 }

var (
	ellipsoids = map[string]Ellipseoid{
		"WGS84":         {a: 6378137, b: 6356752.314245, f: 1 / 298.257223563},
		"Airy1830":      {a: 6377563.396, b: 6356256.909, f: 1 / 299.3249646},
		"AiryModified":  {a: 6377340.189, b: 6356034.448, f: 1 / 299.3249646},
		"Bessel1841":    {a: 6377397.155, b: 6356078.962818, f: 1 / 299.1528128},
		"Clarke1866":    {a: 6378206.4, b: 6356583.8, f: 1 / 294.978698214},
		"Clarke1880IGN": {a: 6378249.2, b: 6356515.0, f: 1 / 293.466021294},
		"GRS80":         {a: 6378137, b: 6356752.314140, f: 1 / 298.257222101},
		"Intl1924":      {a: 6378388, b: 6356911.946, f: 1 / 297}, // aka Hayford
		"WGS72":         {a: 6378135, b: 6356750.5, f: 1 / 298.26},
	}
)

/**
 * Datums; with associated ellipsoid, and Helmert transform parameters to convert from WGS-84
 * into given datum.
 *
 * Note that precision of various datums will vary, and WGS-84 (original) is not defined to be
 * accurate to better than ±1 metre. No transformation should be assumed to be accurate to
 * better than a metre, for many datums somewhat less.
 *
 * This is a small sample of commoner datums from a large set of historical datums. I will add
 * new datums on request.
 *
 * @example
 *   const a = LatLon.datums.OSGB36.ellipsoid.a;                    // 6377563.396
 *   const tx = LatLon.datums.OSGB36.transform;                     // [ tx, ty, tz, s, rx, ry, rz ]
 *   const availableDatums = Object.keys(LatLon.datums).join(', '); // ED50, Irl1975, NAD27, ...
 */
type Datum struct {
	Name      string
	Ellipsoid Ellipseoid
	Transform [7]float64
}

var Datums = map[string]Datum{
	// transforms: t in metres, s in ppm, r in arcseconds              tx       ty        tz       s        rx        ry        rz
	"ED50":       {Name: "ED50", Ellipsoid: ellipsoids["Intl1924"], Transform: [7]float64{89.5, 93.8, 123.1, -1.2, 0.0, 0.0, 0.156}},                        // epsg.io/1311
	"ETRS89":     {Name: "ETRS89", Ellipsoid: ellipsoids["GRS80"], Transform: [7]float64{0, 0, 0, 0, 0, 0, 0}},                                              // epsg.io/1149; @ 1-metre level
	"Irl1975":    {Name: "Irl1975", Ellipsoid: ellipsoids["AiryModified"], Transform: [7]float64{-482.530, 130.596, -564.557, -8.150, 1.042, 0.214, 0.631}}, // epsg.io/1954
	"NAD27":      {Name: "NAD27", Ellipsoid: ellipsoids["Clarke1866"], Transform: [7]float64{8, -160, -176, 0, 0, 0, 0}},
	"NAD83":      {Name: "NAD83", Ellipsoid: ellipsoids["GRS80"], Transform: [7]float64{0.9956, -1.9103, -0.5215, -0.00062, 0.025915, 0.009426, 0.011599}},
	"NTF":        {Name: "NTF", Ellipsoid: ellipsoids["Clarke1880IGN"], Transform: [7]float64{168, 60, -320, 0, 0, 0, 0}},
	"OSGB36":     {Name: "OSGB36", Ellipsoid: ellipsoids["Airy1830"], Transform: [7]float64{-446.448, 125.157, -542.060, 20.4894, -0.1502, -0.2470, -0.8421}}, // epsg.io/1314
	"Potsdam":    {Name: "Potsdam", Ellipsoid: ellipsoids["Bessel1841"], Transform: [7]float64{-582, -105, -414, -8.3, 1.04, 0.35, -3.08}},
	"TokyoJapan": {Name: "TokyoJapan", Ellipsoid: ellipsoids["Bessel1841"], Transform: [7]float64{148, -507, -685, 0, 0, 0, 0}},
	"WGS72":      {Name: "WGS72", Ellipsoid: ellipsoids["WGS72"], Transform: [7]float64{0, 0, -4.5, -0.22, 0, 0, 0.554}},
	"WGS84":      {Name: "WGS84", Ellipsoid: ellipsoids["WGS84"], Transform: [7]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0}},
}

var (
	OSGB36 = Datums["OSGB36"]
	WGS84  = Datums["WGS84"]
)

/* sources:
 * - ED50:       www.gov.uk/guidance/oil-and-gas-petroleum-operations-notices#pon-4
 * - Irl1975:    www.osi.ie/wp-content/uploads/2015/05/transformations_booklet.pdf
 * - NAD27:      en.wikipedia.org/wiki/Helmert_transformation
 * - NAD83:      www.uvm.edu/giv/resources/WGS84_NAD83.pdf [strictly, WGS84(G1150) -> NAD83(CORS96) @ epoch 1997.0]
 *               (note NAD83(1986) ≡ WGS84(Original); confluence.qps.nl/pages/viewpage.action?pageId=29855173)
 * - NTF:        Nouvelle Triangulation Francaise geodesie.ign.fr/contenu/fichiers/Changement_systeme_geodesique.pdf
 * - OSGB36:     www.ordnancesurvey.co.uk/docs/support/guide-coordinate-systems-great-britain.pdf
 * - Potsdam:    kartoweb.itc.nl/geometrics/Coordinate%20transformations/coordtrans.html
 * - TokyoJapan: www.geocachingtoolbox.com?page=datumEllipsoidDetails
 * - WGS72:      www.icao.int/safety/pbn/documentation/eurocontrol/eurocontrol wgs 84 implementation manual.pdf
 *
 * more transform parameters are available from earth-info.nga.mil/GandG/coordsys/datums/NATO_DT.pdf,
 * www.fieldenmaps.info/cconv/web/cconv_params.js
 */
/* note:
 * - ETRS89 reference frames are coincident with WGS-84 at epoch 1989.0 (ie null transform) at the one metre level.
 */

/* LatLon - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - */

/**
 * Latitude/longitude points on an ellipsoidal model earth, with ellipsoid parameters and methods
 * for converting between datums and to geocentric (ECEF) cartesian coordinates.
 *
 * @extends LatLonEllipsoidal
 */
type LatLonEllipsoidalDatum struct {
	Lat, Lon, Height float64
	Datum            Datum
}

/**
* Parses a latitude/longitude point from a variety of formats.
*
* Latitude & longitude (in degrees) can be supplied as two separate parameters, as a single
* comma-separated lat/lon string, or as a single object with { lat, lon } or GeoJSON properties.
*
* The latitude/longitude values may be numeric or strings; they may be signed decimal or
* deg-min-sec (hexagesimal) suffixed by compass direction (NSEW); a variety of separators are
* accepted. Examples -3.62, '3 37 12W', '3°37′12″W'.
*
* Thousands/decimal separators must be comma/dot; use Dms.fromLocale to convert locale-specific
* thousands/decimal separators.
*
* @param   {number|string|Object} lat|latlon - Geodetic Latitude (in degrees) or comma-separated lat/lon or lat/lon object.
* @param   {number}               [lon] - Longitude in degrees.
* @param   {number}               [height=0] - Height above ellipsoid in metres.
* @param   {LatLon.datums}        [datum=WGS84] - Datum this point is defined within.
* @returns {LatLon} Latitude/longitude point on ellipsoidal model earth using given datum.
* @throws  {TypeError} Unrecognised datum.
*
* @example
*   const p1 = LatLon.parse('51.47736, 0.0000', 0, LatLon.datums.OSGB36);
*   const p2 = LatLon.parse('51°28′40″N, 000°00′05″W', 17);
 */
func ParseLatLon(latLon string, height float64, datum Datum) (LatLonEllipsoidalDatum, error) {
	errMessage := fmt.Errorf("invalid LatLon: '%s'", latLon)

	if datum.Name == "" {
		datum = WGS84
	}

	// single comma-separated lat/lon
	parts := strings.Split(latLon, ",")
	if len(parts) != 2 {
		return LatLonEllipsoidalDatum{}, errMessage
	}

	lat, err1 := ParseDegrees(parts[0])
	lat = Wrap90(lat)
	lon, err2 := ParseDegrees(parts[1])
	lon = Wrap180(lon)

	if err1 != nil || err2 != nil {
		return LatLonEllipsoidalDatum{}, errMessage
	}

	return LatLonEllipsoidalDatum{
		Lat:    lat,
		Lon:    lon,
		Height: height,
		Datum:  datum,
	}, nil
}

/**
 * Converts ‘this’ lat/lon coordinate to new coordinate system.
 *
 * @param   {LatLon.datums} toDatum - Datum this coordinate is to be converted to.
 * @returns {LatLon} This point converted to new datum.
 * @throws  {TypeError} Unrecognised datum.
 *
 * @example
 *   const pWGS84 = new LatLon(51.47788, -0.00147, 0, LatLon.datums.WGS84);
 *   const pOSGB = pWGS84.convertDatum(LatLon.datums.OSGB36); // 51.4773°N, 000.0001°E
 */
func (l LatLonEllipsoidalDatum) ConvertDatum(toDatum Datum) LatLonEllipsoidalDatum {
	oldCartesian := l.ToCartesian()                    // convert geodetic to cartesian
	newCartesian := oldCartesian.ConvertDatum(toDatum) // convert datum
	newLatLon := newCartesian.ToLatLon()               // convert cartesian back to geodetic

	return newLatLon
}

/**
 * Converts ‘this’ point from (geodetic) latitude/longitude coordinates to (geocentric) cartesian
 * (x/y/z) coordinates, based on the same datum.
 *
 * Shadow of LatLonEllipsoidal.toCartesian(), returning Cartesian augmented with
 * LatLonEllipsoidal_Datum methods/properties.
 *
 * @returns {Cartesian} Cartesian point equivalent to lat/lon point, with x, y, z in metres from
 *   earth centre, augmented with reference frame conversion methods and properties.
 */
func (l LatLonEllipsoidalDatum) ToCartesian() Cartesian {
	// x = (ν+h)⋅cosφ⋅cosλ, y = (ν+h)⋅cosφ⋅sinλ, z = (ν⋅(1-e²)+h)⋅sinφ
	// where ν = a/√(1−e²⋅sinφ⋅sinφ), e² = (a²-b²)/a² or (better conditioned) 2⋅f-f²
	ellipsoid := l.Datum.Ellipsoid

	φ := l.Lat * toRadians
	λ := l.Lon * toRadians
	h := l.Height
	a, f := ellipsoid.a, ellipsoid.f

	sinφ := math.Sin(φ)
	cosφ := math.Cos(φ)
	sinλ := math.Sin(λ)
	cosλ := math.Cos(λ)

	eSq := 2*f - f*f                    // 1st eccentricity squared ≡ (a²-b²)/a²
	ν := a / math.Sqrt(1-eSq*sinφ*sinφ) // radius of curvature in prime vertical

	x := (ν + h) * cosφ * cosλ
	y := (ν + h) * cosφ * sinλ
	z := (ν*(1-eSq) + h) * sinφ

	return Cartesian{
		X:     x,
		Y:     y,
		Z:     z,
		Datum: l.Datum,
	}
}

func (l LatLonEllipsoidalDatum) ToOsGridRef() OsGridRef {
	// if necessary convert to OSGB36 first
	point := l
	if point.Datum.Name != OSGB36.Name {
		point = point.ConvertDatum(OSGB36)
	}

	φ := point.Lat * toRadians
	λ := point.Lon * toRadians

	cosφ := math.Cos(φ)
	sinφ := math.Sin(φ)
	ν := a * F0 / math.Sqrt(1-e2*sinφ*sinφ)                // nu = transverse radius of curvature
	ρ := a * F0 * (1 - e2) / math.Pow(1-e2*sinφ*sinφ, 1.5) // rho = meridional radius of curvature
	η2 := ν/ρ - 1                                          // eta = ?

	Ma := (1 + n + (5/4)*n2 + (5/4)*n3) * (φ - φ0)
	Mb := (3*n + 3*n*n + (21/8)*n3) * math.Sin(φ-φ0) * math.Cos(φ+φ0)
	Mc := ((15/8)*n2 + (15/8)*n3) * math.Sin(2*(φ-φ0)) * math.Cos(2*(φ+φ0))
	Md := (35 / 24) * n3 * math.Sin(3*(φ-φ0)) * math.Cos(3*(φ+φ0))
	M := b * F0 * (Ma - Mb + Mc - Md) // meridional arc

	cos3φ := cosφ * cosφ * cosφ
	cos5φ := cos3φ * cosφ * cosφ
	tan2φ := math.Tan(φ) * math.Tan(φ)
	tan4φ := tan2φ * tan2φ

	I := M + N0
	II := (ν / 2) * sinφ * cosφ
	III := (ν / 24) * sinφ * cos3φ * (5 - tan2φ + 9*η2)
	IIIA := (ν / 720) * sinφ * cos5φ * (61 - 58*tan2φ + tan4φ)
	IV := ν * cosφ
	V := (ν / 6) * cos3φ * (ν/ρ - tan2φ)
	VI := (ν / 120) * cos5φ * (5 - 18*tan2φ + tan4φ + 14*η2 - 58*tan2φ*η2)

	Δλ := λ - λ0
	Δλ2 := Δλ * Δλ
	Δλ3 := Δλ2 * Δλ
	Δλ4 := Δλ3 * Δλ
	Δλ5 := Δλ4 * Δλ
	Δλ6 := Δλ5 * Δλ

	N := I + II*Δλ2 + III*Δλ4 + IIIA*Δλ6
	E := E0 + IV*Δλ + V*Δλ3 + VI*Δλ5

	return OsGridRef{
		Easting:  int(math.Round(E)),
		Northing: int(math.Round(N)),
	}
}

/* Cartesian  - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - */

/**
 * Creates cartesian coordinate representing ECEF (earth-centric earth-fixed) point, on a given
 * datum. The datum will identify the primary meridian (for the x-coordinate), and is also
 * useful in transforming to/from geodetic (lat/lon) coordinates.
 *
 */
type Cartesian struct {
	X, Y, Z float64
	Datum   Datum
}

/**
 * Converts ‘this’ (geocentric) cartesian (x/y/z) coordinate to (geodetic) latitude/longitude
 * point (based on the same datum, or WGS84 if unset).
 *
 * Uses Bowring’s (1985) formulation for μm precision in concise form; ‘The accuracy of geodetic
 * latitude and height equations’, B R Bowring, Survey Review vol 28, 218, Oct 1985.
 *
 * @returns {LatLon} Latitude/longitude point defined by cartesian coordinates.
 * @throws  {TypeError} Unrecognised datum
 *
 * @example
 *   const c = new Cartesian(4027893.924, 307041.993, 4919474.294);
 *   const p = c.toLatLon(); // 50.7978°N, 004.3592°E
 */
func (c Cartesian) ToLatLon() LatLonEllipsoidalDatum {
	x, y, z := c.X, c.Y, c.Z
	a, b, f := c.Datum.Ellipsoid.a, c.Datum.Ellipsoid.b, c.Datum.Ellipsoid.f

	e2 := 2*f - f*f           // 1st eccentricity squared ≡ (a²−b²)/a²
	ε2 := e2 / (1 - e2)       // 2nd eccentricity squared ≡ (a²−b²)/b²
	p := math.Sqrt(x*x + y*y) // distance from minor axis
	R := math.Sqrt(p*p + z*z) // polar radius

	// parametric latitude (Bowring eqn.17, replacing tanβ = z·a / p·b)
	tanβ := (b * z) / (a * p) * (1 + ε2*b/R)
	sinβ := tanβ / math.Sqrt(1+tanβ*tanβ)
	cosβ := sinβ / tanβ

	// geodetic latitude (Bowring eqn.18: tanφ = z+ε²⋅b⋅sin³β / p−e²⋅cos³β)
	φ := 0.0
	if !math.IsNaN(cosβ) {
		φ = math.Atan2(z+ε2*b*sinβ*sinβ*sinβ, p-e2*a*cosβ*cosβ*cosβ)
	}

	// longitude
	λ := math.Atan2(y, x)

	// height above ellipsoid (Bowring eqn.7)
	sinφ := math.Sin(φ)
	cosφ := math.Cos(φ)
	ν := a / math.Sqrt(1-e2*sinφ*sinφ) // length of the normal terminated by the minor axis
	h := p*cosφ + z*sinφ - (a * a / ν)

	return LatLonEllipsoidalDatum{
		Lat:    φ * toDegrees,
		Lon:    λ * toDegrees,
		Height: h,
		Datum:  c.Datum,
	}
}

/**
 * Converts ‘this’ cartesian coordinate to new datum using Helmert 7-parameter transformation.
 *
 * @param   {LatLon.datums} toDatum - Datum this coordinate is to be converted to.
 * @returns {Cartesian} This point converted to new datum.
 * @throws  {Error} Undefined datum.
 *
 * @example
 *   const c = new Cartesian(3980574.247, -102.127, 4966830.065, LatLon.datums.OSGB36);
 *   c.convertDatum(LatLon.datums.Irl1975); // [??,??,??]
 */
func (c Cartesian) ConvertDatum(toDatum Datum) Cartesian {
	if c.Datum.Name == toDatum.Name {
		return c
	}

	// TODO: what if datum is not geocentric?

	var (
		oldCartesian Cartesian
		transform    [7]float64
	)

	if c.Datum.Name == "WGS84" {
		// converting from WGS 84
		oldCartesian = c
		transform = toDatum.Transform
	} else if toDatum.Name == "WGS84" {
		// converting to WGS 84; use inverse transform
		oldCartesian = c
		for i := range c.Datum.Transform {
			transform[i] = -c.Datum.Transform[i]
		}
	} else {
		// neither this.datum nor toDatum are WGS84: convert this to WGS84 first
		oldCartesian = c.ConvertDatum(WGS84)
		transform = toDatum.Transform
	}

	newCartesian := oldCartesian.applyTransform(transform)
	newCartesian.Datum = toDatum

	return newCartesian
}

/**
 * Applies Helmert 7-parameter transformation to ‘this’ coordinate using transform parameters t.
 *
 * This is used in converting datums (geodetic->cartesian, apply transform, cartesian->geodetic).
 *
 * @private
 * @param   {number[]} t - Transformation to apply to this coordinate.
 * @returns {Cartesian} Transformed point.
 */
func (c Cartesian) applyTransform(t [7]float64) Cartesian {
	// this point
	x1, y1, z1 := c.X, c.Y, c.Z

	// transform parameters
	tx := t[0]                      // x-shift in metres
	ty := t[1]                      // y-shift in metres
	tz := t[2]                      // z-shift in metres
	s := t[3]/1e6 + 1               // scale: normalise parts-per-million to (s+1)
	rx := (t[4] / 3600) * toRadians // x-rotation: normalise arcseconds to radians
	ry := (t[5] / 3600) * toRadians // y-rotation: normalise arcseconds to radians
	rz := (t[6] / 3600) * toRadians // z-rotation: normalise arcseconds to radians

	// apply transform
	x2 := tx + x1*s - y1*rz + z1*ry
	y2 := ty + x1*rz + y1*s - z1*rx
	z2 := tz - x1*ry + y1*rx + z1*s

	return Cartesian{
		X: x2,
		Y: y2,
		Z: z2,
	}
}
