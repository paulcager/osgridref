package osgrid

import (
	"net/http"
	"testing"

	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"
)

var (
	vm = otto.New()
)

// Run the Javascript in an Otto VM, so we have a reference copy we can test against. As you'd expect,
// the Otto version runs much more slowly:
// 	BenchmarkOttoImpl-16               13500            440363 ns/op
// 	BenchmarkGoImpl-16              12458200               468 ns/op

func init() {
	modules := []string{
		"https://cdn.jsdelivr.net/npm/geodesy@1/vector3d.js",
		"https://cdn.jsdelivr.net/npm/geodesy@1/dms.js",
		"https://cdn.jsdelivr.net/npm/geodesy@1/latlon-ellipsoidal.js",
		"https://cdn.jsdelivr.net/npm/geodesy@1/osgridref.js",
	}

	for _, mod := range modules {
		resp, err := http.Get(mod)
		must(err)
		_, err = vm.Run(resp.Body)
		resp.Body.Close()
		must(err)
	}
}

func OttoGridToLatLon(osgrid string) (float64, float64, error) {
	vm.Set("osgrid", osgrid)
	ret, err := vm.Run(`OsGridRef.osGridToLatLon(OsGridRef.parse(osgrid), LatLon.datum.WGS84);`)
	if err != nil {
		return 0, 0, err
	}
	obj, _ := ret.Export()
	latLon := obj.(map[string]interface{})
	return latLon["lat"].(float64), latLon["lon"].(float64), nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}


func BenchmarkOttoImpl(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := OttoGridToLatLon("TL 44982 57869")
		assert.NoError(b, err)
	}
}

func BenchmarkGoImpl(b *testing.B) {
	for i := 0; i < b.N; i++ {
		o, err := ParseOsGridRef("TL 44982 57869")
		assert.NoError(b, err)
		_, _ = o.ToLatLon()
	}
}

//func BenchmarkAPICall(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		resp, err := http.Get("http://192.168.0.123:9090/?gridRef=SJ121689")
//		require.NoError(b, err)
//		reply, err := ioutil.ReadAll(resp.Body)
//		resp.Body.Close()
//		require.NoError(b, err)
//		assert.Contains(b, string(reply), "53")
//	}
//}
