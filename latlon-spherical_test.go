package osgridref

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

var (
	cambridge = LatLon{Lat: 52.205, Lon: 0.119}
	paris     = LatLon{Lat: 48.857, Lon: 2.351}
	valley    = LatLon{Lat: 53.248303, Lon: -4.535474}
	caernafon = LatLon{Lat: 53.102846, Lon: -4.332533}
	greenwich = LatLon{Lat: 51.47788, Lon: -0.00147}
	stansted  = LatLon{Lat: 51.8853, Lon: 0.2545}
	cdg       = LatLon{Lat: 49.0034, Lon: 2.5735}
	bxl       = LatLon{Lat: 50.9078, Lon: 004.5084}
)

func TestLatLon_DistanceTo(t *testing.T) {
	tests := []struct {
		name     string
		from, to LatLon
		want     float64
	}{
		{name: "self", from: cambridge, to: cambridge, want: 0},
		{name: "Paris", from: cambridge, to: paris, want: 404279},
		{name: "valley-to-caernafon", from: valley, to: caernafon, want: 21084.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.DistanceTo(tt.to)
			assert.InDelta(t, tt.want, got, 1)
		})
	}
}

func TestLatLon_BearingTo(t *testing.T) {
	justNorthOfCambridge := LatLon{Lat: 52.206, Lon: 0.119}
	justWestOfCambridge := LatLon{Lat: 52.205, Lon: 0.118}

	tests := []struct {
		name     string
		from, to LatLon
		init     float64
		final    float64
	}{
		{name: "justNorthOfCambridge", from: cambridge, to: justNorthOfCambridge, init: 0, final: 0},
		{name: "justWestOfCambridge", from: cambridge, to: justWestOfCambridge, init: 270, final: 270},
		{name: "paris", from: cambridge, to: paris, init: 156.2, final: 157.9},
		{name: "north pole", from: cambridge, to: LatLon{Lat: 90, Lon: 0}, init: 0, final: 360},
		{name: "south pole", from: cambridge, to: LatLon{Lat: -90, Lon: 0}, init: 180, final: 180},
		{name: "pole-to-pole", from: LatLon{Lat: 90, Lon: 0}, to: LatLon{Lat: -90, Lon: 0}, init: 180, final: 180},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			init := tt.from.InitialBearingTo(tt.to)
			final := tt.from.FinalBearingTo(tt.to)
			assert.InDelta(t, tt.init, init, .5)
			assert.InDelta(t, tt.final, final, .5)
		})
	}
}

func TestLatLon_DestinationPoint(t *testing.T) {
	tests := []struct {
		name     string
		from     LatLon
		distance float64
		bearing  float64
		want     LatLon
	}{
		{name: "no-op", from: cambridge, distance: 0, bearing: 77, want: cambridge},
		{name: "greenwich", from: greenwich, distance: 7794, bearing: 300.7, want: LatLon{Lat: 51.5136, Lon: -0.0983}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.DestinationPoint(tt.distance, tt.bearing)
			assert.InDelta(t, tt.want.Lat, got.Lat, 5e-5)
			assert.InDelta(t, tt.want.Lon, got.Lon, 5e-5)
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name  string
		p1    LatLon
		brng1 float64
		p2    LatLon
		brng2 float64
		want  LatLon
		ok    bool
	}{
		{name: "Same", p1: cambridge, brng1: 17, p2: cambridge, brng2: 360 - 17, want: cambridge, ok: true},
		{name: "stn-cdg-bxl", p1: stansted, brng1: 108.547, p2: cdg, brng2: 32.435, want: bxl, ok: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Intersection(tt.p1, tt.brng1, tt.p2, tt.brng2)
			assert.Equal(t, tt.ok, ok)
			assert.InDelta(t, tt.want.Lat, got.Lat, 0.0001)
			assert.InDelta(t, tt.want.Lon, got.Lon, 0.0001)
		})
	}
}

func poly(t *testing.T, name string, s string) []LatLon {
	points := strings.Split(s, " ")
	poly := make([]LatLon, len(points))

	for i := range points {
		ll := strings.Split(points[i], ",")
		require.Len(t, ll, 2, "Bad lat.lon for test %s: %s", name, s)
		lat, err := strconv.ParseFloat(ll[0], 64)
		require.NoError(t, err)
		lon, err := strconv.ParseFloat(ll[1], 64)
		require.NoError(t, err)
		poly[i] = LatLon{Lat: lat, Lon: lon}
	}

	return poly
}

func TestLatLon_AreaOf(t *testing.T) {
	tests := []struct {
		name    string
		polygon string
		want    float64
	}{
		{name: "line", polygon: "45,45 50,50", want: 0},
		{name: "triangle-open", polygon: "1,1 2,1 1,2", want: 6181527888},
		{name: "triangle-closed", polygon: "1,1 2,1 1,2 1,1", want: 6181527888},
		{name: "square cw", polygon: "1,1 2,1 2,2 1,2", want: 12360230987},
		{name: "square ccw", polygon: "1,1 1,2 2,2 2,1", want: 12360230987},
		{name: "pole", polygon: "89,0 89,120 89,-120", want: 16063139192},
		{name: "concave", polygon: "1,1 5,1 5,3 1,3 3,2", want: 74042699236},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AreaOf(poly(t, tt.name, tt.polygon))
			assert.InDelta(t, tt.want, got, 1.0)
		})
	}
}
