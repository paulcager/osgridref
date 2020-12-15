# Ordnance Survey Grid Refs in Golang

This is a Golang package to translate between Ordnance Survey (OS) Grid References and
Latitude / Longitude. OS Grid references are traditionally used in UK navigation, while
lat / lon is used by GPS systems and global mapping systems.

This package is a partial translation into Golang of the excellent
[Javascript library](https://github.com/chrisveness/geodesy) by Chris Veness.

## Quick start

```go
gridRef, err := ParseOsGridRef("SW 46760 28548")
if err != nil {
    panic(err)
}

lat, lon := gridRef.ToLatLon()
fmt.Printf("%.4f,%.4f\n", lat, lon)     // 50.1029,-5.5428
```
There are more detailed examples on [pkg.go.dev](
https://pkg.go.dev/github.com/paulcager/osgridref?readme=expanded#example-package), or
try it in the [Go Playground](https://play.golang.org/p/u_yvmPA1ZLf).

---

## Q & A


### What's an OS Grid Reference? ###

<img src="images/grid-refs.png" alt="grid refs" align="right" />

The Ordnance Survey have been producing maps of Great Britain since
[1791](https://www.ordnancesurvey.co.uk/about/history). They use a
[National Grid](https://en.wikipedia.org/wiki/Ordnance_Survey_National_Grid) system, distinct
from latitude and longitude, where grid references comprise two letters and a sequence of
digits, such as "SK127836".

OS grid references are ubiquitous in the great outdoors - guide books use them to tell you where to
park the car, hiking routes use them, and should you get into trouble the
local Mountain Rescue team would want to know the location as an OS grid ref. However, the OS grid
is only relevant in Great Britain; most electronic and global mapping systems instead
use Latitude and Longitude, as in, for example, this
[Google Maps URL](https://www.google.com/maps/place/51%C2%B030'11.9%22N+0%C2%B007'39.0%22W/).

So sometimes it is necessary to convert between OS grid refs and lat/lon references. This Golang
library can be used to perform the conversion.

### What Does a Grid Reference Look Like? ##

<img src="images/grid-letters.png" alt="grid letters" align="right" />

The normal, human-readable representation is two letters followed by two groups of digits,
for example `SZ 644 874`. The 2 letters define a 100 km by 100 km square, as in the diagram
on the right. The first group of digits isthe `eastings` and the second is the `northings`;
these digits define a coordinate _within_ the 100 km square.

An alternative notation is to omit the grid letters and provide just an `easting` and `northing`
separated by a comma. In this case these are coordinates relative to the _origin_ of the grid
as a whole, i.e. relative to the south-west corner of the grid.

The Ordnance Survey have created a friendly
[guide](https://getoutside.ordnancesurvey.co.uk/guides/beginners-guide-to-grid-references/)
with full details.

This library can parse and display both types of representation.

### How do you convert an OS Grid Reference? ##

It's difficult. Very, very difficult. Pages of this sort of stuff:

![img.png](images/scary-maths.png)
> (an excerpt from page 50 of the Ordnance Survey's
[reference guide](https://www.ordnancesurvey.co.uk/documents/resources/guide-coordinate-systems-great-britain.pdf))

Fortunately Chris Veness has already done the hard work of implementing this in his
[Javascript library](https://github.com/chrisveness/geodesy) (which does much more than
just converting grid ref to and from lat/lon). This package is a fairly mechanical translation
of the Javascript into Golang, without understanding how it works.

I am pleased to say that I don't understand the mathematics behind any of this.

### Why doesn't the code look like idiomatic Go? ###

This is deliberate, to make it easier to verify this implementation against the original
Javascript implementation. Where possible, each line of upstream code should match against
an equivalent line in the Golang code.

