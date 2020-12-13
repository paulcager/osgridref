package osgrid

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */
/* Geodesy representation conversion functions                        (c) Chris Veness 2002-2020  */
/*                                                                                   MIT Licence  */
/* www.movable-type.co.uk/scripts/latlong.html                                                    */
/* www.movable-type.co.uk/scripts/js/geodesy/geodesy-library.html#dms                             */
/* - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -  */

/**
 * Latitude/longitude points may be represented as decimal degrees, or subdivided into sexagesimal
 * minutes and seconds. This module provides methods for parsing and representing degrees / minutes
 * / seconds.
 *
 * @module dms
 */

var (
	// /[^0-9., ]+/
	separatorChars = regexp.MustCompile(`[^0-9.]+`)
)

/**
 * Constrain degrees to range -90..+90 (for latitude); e.g. -91 => -89, 91 => 89.
 *
 * @private
 * @param {number} degrees
 * @returns degrees within range -90..+90.
 */
func Wrap90(degrees float64) float64 {
	// avoid rounding due to arithmetic ops if within range
	if -90 <= degrees && degrees <= 90 {
		return degrees
	}

	// latitude wrapping requires a triangle wave function; a general triangle wave is
	//     f(x) = 4a/p ⋅ | (x-p/4)%p - p/2 | - a
	// where a = amplitude, p = period, % = modulo; however, JavaScript '%' is a remainder operator
	// not a modulo operator - for modulo, replace 'x%n' with '((x%n)+n)%n'
	var (
		x = degrees
		a = 90.0
		p = 360.0
	)

	return 4*a/p*math.Abs(
		math.Mod(math.Mod(x-p/4, p)+p, p)-p/2) - a
}

/**
 * Constrain degrees to range -180..+180 (for longitude); e.g. -181 => 179, 181 => -179.
 *
 * @private
 * @param {number} degrees
 * @returns degrees within range -180..+180.
 */
func Wrap180(degrees float64) float64 {
	// avoid rounding due to arithmetic ops if within range
	if -180 <= degrees && degrees <= 180 {
		return degrees
	}

	// longitude wrapping requires a sawtooth wave function; a general sawtooth wave is
	//     f(x) = (2ax/p - p/2) % p - a
	// where a = amplitude, p = period, % = modulo; however, JavaScript '%' is a remainder operator
	// not a modulo operator - for modulo, replace 'x%n' with '((x%n)+n)%n'
	var (
		x = degrees
		a = 180.0
		p = 360.0
	)
	return math.Mod((math.Mod(2*a*x/p-p/2, p))+p, p) - a
}

/**
 * Constrain degrees to range 0..360 (for bearings); e.g. -1 => 359, 361 => 1.
 *
 * @private
 * @param {number} degrees
 * @returns degrees within range 0..360.
 */
func Wrap360(degrees float64) float64 {
	// avoid rounding due to arithmetic ops if within range
	if 0 <= degrees && degrees <= 360 {
		return degrees
	}

	// bearing wrapping requires a sawtooth wave function with a vertical offset equal to the
	// amplitude and a corresponding phase shift; this changes the general sawtooth wave function from
	//     f(x) = (2ax/p - p/2) % p - a
	// to
	//     f(x) = (2ax/p) % p
	// where a = amplitude, p = period, % = modulo; however, JavaScript '%' is a remainder operator
	// not a modulo operator - for modulo, replace 'x%n' with '((x%n)+n)%n'
	var (
		x = degrees
		a = 180.0
		p = 360.0
	)
	return math.Mod((math.Mod(2*a*x/p, p))+p, p)
}

func invalid(s string) error {
	return fmt.Errorf("invalid degree: '%s'", s)
}

/**
 * Parses string representing degrees/minutes/seconds into numeric degrees.
 *
 * This is very flexible on formats, allowing signed decimal degrees, or deg-min-sec optionally
 * suffixed by compass direction (NSEW); a variety of separators are accepted. Examples -3.62,
 * '3 37 12W', '3°37′12″W'.
 *
 * Thousands/decimal separators must be comma/dot; use Dms.fromLocale to convert locale-specific
 * thousands/decimal separators.
 *
 * @param   {string|number} dms - Degrees or deg/min/sec in variety of formats.
 * @returns {number}        Degrees as decimal number.
 *
 * @example
 *   const lat = Dms.parse('51° 28′ 40.37″ N');
 *   const lon = Dms.parse('000° 00′ 05.29″ W');
 *   const p1 = new LatLon(lat, lon); // 51.4779°N, 000.0015°W
 */
func ParseDegrees(s string) (float64, error) {
	orig := s
	s = strings.TrimSpace(s)
	// check for signed decimal degrees without NSEW, if so return it directly
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f, nil
	}

	if len(s) == 0 {
		return 0, invalid(orig)
	}
	// strip off any sign or compass dir'n & split out separate d/m/s
	negative := s[0] == '-'
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return 0, invalid(orig)
	}

	switch s[len(s)-1] {
	case 'S', 'W':
		negative = true
		s = s[:len(s)-1]
	case 'N', 'E':
		s = s[:len(s)-1]
	}
	s = strings.TrimSpace(s)

	dmsParts := separatorChars.Split(s, -1)
	if dmsParts[0] == "" {
		return 0, invalid(orig)
	}
	if dmsParts[len(dmsParts)-1] == "" {
		dmsParts=dmsParts[:len(dmsParts)-1]
	}
	multiplier := 1.0
	sum := 0.0
	for i := range dmsParts {
		f, err := strconv.ParseFloat(dmsParts[i], 64)
		if err != nil {
			return 0, invalid(orig)
		}
		sum += f *multiplier
		multiplier /= 60.0
	}

	if negative {
		sum = -sum
	}
	return sum, nil
}
