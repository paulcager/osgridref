package osgrid

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
/* Ordnance Survey Grid Reference functions                           (c) Chris Veness 2005-2020  */
/*                                                                                   MIT Licence  */
/* www.movable-type.co.uk/scripts/latlong-gridref.html                                            */
/* www.movable-type.co.uk/scripts/geodesy-library.html#osgridref                                  */
/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */

/**
 * Ordnance Survey OSGB grid references provide geocoordinate references for UK mapping purposes.
 *
 * Formulation implemented here due to Thomas, Redfearn, etc is as published by OS, but is inferior
 * to Krüger as used by e.g. Karney 2011.
 *
 * www.ordnancesurvey.co.uk/documents/resources/guide-coordinate-systems-great-britain.pdf.
 *
 * Note OSGB grid references cover Great Britain only; Ireland and the Channel Islands have their
 * own references.
 *
 * Note that these formulae are based on ellipsoidal calculations, and according to the OS are
 * accurate to about 4–5 metres – for greater accuracy, a geoid-based transformation (OSTN15) must
 * be used.
 */

/*
 * Converted 2015 to work with WGS84 by default, OSGB36 as option;
 * www.ordnancesurvey.co.uk/blog/2014/12/confirmation-on-changes-to-latitude-and-longitude
 */

/* OsGridRef  - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - */

// Sample conversion: SJ9239552997	392395	352997	53.074149	-2.1149638
// OS Grid letters
// 		https://getoutside.ordnancesurvey.co.uk/site/uploads/images/assets/Web%20images/Diagram-A.jpg

const (
	toRadians = math.Pi / 180.0
	toDegrees = 180.0 / math.Pi

	// Airy 1830 major & minor semi-axes
	a = 6377563.396
	b = 6356256.909

	// NatGrid scale factor on central meridian
	F0 = 0.9996012717

	// NatGrid true origin is 49°N,2°W
	φ0 = 49 * toRadians
	λ0 = -2 * toRadians

	// northing & easting of true origin, metres
	N0 = -100e3
	E0 = 400e3

	// eccentricity squared
	e2 = 1.0 - (b*b)/(a*a)

	// n, n², n³
	n  = (a - b) / (a + b)
	n2 = n * n
	n3 = n * n * n
)

type OsGridRef struct {
	Easting, Northing int
}

var (
	commaSeparatedFormat = regexp.MustCompile(`^(\d+),\s*(\d+)$`)
	gridRefFormat        = regexp.MustCompile(`^[A-Z]{2}[0-9]+$`)
)

func ParseOsGridRef(s string) (OsGridRef, error) {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToUpper(s)

	matches := commaSeparatedFormat.FindStringSubmatch(s)
	if len(matches) > 0 {
		e, err1 := strconv.ParseFloat(matches[1], 32)
		n, err2 := strconv.ParseFloat(matches[2], 32)
		if err1 != nil || err2 != nil {
			return OsGridRef{}, fmt.Errorf("invalid comma-separated grid ref format: %q", s)
		}
		return OsGridRef{
			Easting:  int(e),
			Northing: int(n),
		}, nil
	}

	matches = gridRefFormat.FindStringSubmatch(s)
	if len(matches) == 0 {
		return OsGridRef{}, fmt.Errorf("invalid grid ref format: %q", s)
	}

	// get numeric values of letter references, mapping A->0, B->1, C->2, etc:
	l1 := int(s[0] - 'A')
	l2 := int(s[1] - 'A')
	// shuffle down letters after 'I' since 'I' is not used in grid:
	if s[0] == 'I' || s[1] == 'I' {
		return OsGridRef{}, fmt.Errorf("invalid grid ref format: %q", s)
	}

	if l1 > 7 {
		l1--
	}
	if l2 > 7 {
		l2--
	}

	// sanity check
	if l1 < 8 || l1 > 18 {
		return OsGridRef{}, fmt.Errorf(`invalid grid reference %q`, s)
	}

	// convert grid letters into 100km-square indexes from false origin (grid square SV):
	e100km := ((l1-2)%5)*5 + (l2 % 5)
	n100km := (19 - (l1/5)*5) - (l2 / 5)

	// skip grid letters to get numeric (easting/northing) part of ref
	digits := s[2:]
	// split half way
	e, n := digits[:len(digits)/2], digits[len(digits)/2:]
	if len(e) != len(n) {
		return OsGridRef{}, fmt.Errorf(`invalid grid reference %q`, s)
	}

	// standardise to 10-digit refs (metres)
	e = (e + "00000")[:5]
	n = (n + "00000")[:5]

	easting, _ := strconv.ParseInt(e, 10, 32)
	northing, _ := strconv.ParseInt(n, 10, 32)

	return OsGridRef{Easting: e100km*100000 + int(easting), Northing: n100km*100000 + int(northing)}, nil
}

func (o OsGridRef) Valid() bool {
	return o.Easting >= 0 && o.Easting <= 700e3 && o.Northing >= 0 && o.Northing <= 1300e3
}
func (o OsGridRef) assertValid() {
	if !o.Valid() {
		panic(fmt.Sprintf("Invalid OS grid ref: %+v", o))
	}
}

func (o OsGridRef) ToLatLon() (float64, float64) {
	E := float64(o.Easting)
	N := float64(o.Northing)

	φ := φ0
	M := float64(0)

	for {
		φ = (N-N0-M)/(a*F0) + φ

		Ma := (1 + n + (5/4)*n2 + (5/4)*n3) * (φ - φ0)
		Mb := (3*n + 3*n*n + (21/8)*n3) * math.Sin(φ-φ0) * math.Cos(φ+φ0)
		Mc := ((15/8)*n2 + (15/8)*n3) * math.Sin(2*(φ-φ0)) * math.Cos(2*(φ+φ0))
		Md := (35 / 24) * n3 * math.Sin(3*(φ-φ0)) * math.Cos(3*(φ+φ0))
		M = b * F0 * (Ma - Mb + Mc - Md) // meridional arc

		// until < 0.01mm
		if math.Abs(N-N0-M) < 0.00001 {
			break
		}
	}

	cosφ := math.Cos(φ)
	sinφ := math.Sin(φ)
	ν := a * F0 / math.Sqrt(1-e2*sinφ*sinφ)                // nu = transverse radius of curvature
	ρ := a * F0 * (1 - e2) / math.Pow(1-e2*sinφ*sinφ, 1.5) // rho = meridional radius of curvature
	η2 := ν/ρ - 1                                          // eta = ?

	tanφ := math.Tan(φ)
	tan2φ := tanφ * tanφ
	tan4φ := tan2φ * tan2φ
	tan6φ := tan4φ * tan2φ
	secφ := 1 / cosφ
	ν3 := ν * ν * ν
	ν5 := ν3 * ν * ν
	ν7 := ν5 * ν * ν
	VII := tanφ / (2 * ρ * ν)
	VIII := tanφ / (24 * ρ * ν3) * (5 + 3*tan2φ + η2 - 9*tan2φ*η2)
	IX := tanφ / (720 * ρ * ν5) * (61 + 90*tan2φ + 45*tan4φ)
	X := secφ / ν
	XI := secφ / (6 * ν3) * (ν/ρ + 2*tan2φ)
	XII := secφ / (120 * ν5) * (5 + 28*tan2φ + 24*tan4φ)
	XIIA := secφ / (5040 * ν7) * (61 + 662*tan2φ + 1320*tan4φ + 720*tan6φ)

	dE := E - E0
	dE2 := dE * dE
	dE3 := dE2 * dE
	dE4 := dE2 * dE2
	dE5 := dE3 * dE2
	dE6 := dE4 * dE2
	dE7 := dE5 * dE2
	φ = φ - VII*dE2 + VIII*dE4 - IX*dE6
	λ := λ0 + X*dE - XI*dE3 + XII*dE5 - XIIA*dE7

	// That has calculated the lat/lon in OSGB36; we want WGS84
	φ, λ = osgb36ToWGS84(φ*toDegrees, λ*toDegrees)

	return φ, λ
}

func (o OsGridRef) String() string {
	return o.StringN(8)
}

func (o OsGridRef) StringNCompact(digits int) string {
	return o.stringN(digits, false)
}

func (o OsGridRef) StringN(digits int) string {
	return o.stringN(digits, true)
}

func (o OsGridRef) stringN(digits int, spaces bool) string {
	e, n := o.Easting, o.Northing
	// get the 100km-grid indices
	e100km := e / 100_000
	n100km := n / 100_000

	// translate those into numeric equivalents of the grid letters
	l1 := (19 - n100km) - (19-n100km)%5 + (e100km+10)/5
	l2 := (19-n100km)*5%25 + e100km%5

	// compensate for skipped 'I' and build grid letter-pairs
	if l1 > 7 {
		l1++
	}
	if l2 > 7 {
		l2++
	}
	letterPair := string([]byte{byte(l1 + 'A'), byte(l2 + 'A')})

	pow := func(n int) int {
		ret := 1
		for i := 0; i < n; i++ {
			ret *= 10
		}
		return ret
	}

	// strip 100km-grid indices from easting & northing, and reduce precision
	e = (e % 100000) / pow(5-digits/2)
	n = (n % 100000) / pow(5-digits/2)

	// pad eastings & northings with leading zeros
	if spaces {
		return fmt.Sprintf("%s %0*d %0*d", letterPair, digits/2, e, digits/2, n)
	}
	return fmt.Sprintf("%s%0*d%0*d", letterPair, digits/2, e, digits/2, n)
}

func (o OsGridRef) NumericString() string {
	return fmt.Sprintf("%d,%d", o.Easting, o.Northing)
}

func osgb36ToWGS84(lat, lon float64) (float64, float64) {
	latLon := LatLonEllipsoidalDatum{
		Lat:    lat,
		Lon:    lon,
		Height: 0,
		Datum:  OSGB36,
	}

	converted := latLon.ConvertDatum(WGS84)
	return converted.Lat, converted.Lon
}
