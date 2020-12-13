package osgrid

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		want    float64
		wantErr bool
	}{
		{name: "0.0", want: 0, wantErr: false},
		{name: "0°", want: 0, wantErr: false},
		{name: "000°00′00.0″", want: 0, wantErr: false},
		{name: "45.76260", want: 45.76260, wantErr: false},
		{name: " 45.76260 ", want: 45.76260, wantErr: false},
		{name: "45°45.756′", want: 45.76260, wantErr: false},
		{name: `45° 45.756′ 0"`, want: 45.76260, wantErr: false},
		{name: "45° 45’ 45.36", want: 45.76260, wantErr: false},
		{name: `45° 45’ 45.36"`, want: 45.76260, wantErr: false},
		{name: `45 45 45.36`, want: 45.76260, wantErr: false},
		{name: "45.76260N", want: 45.76260, wantErr: false},
		{name: "45.76260S", want: -45.76260, wantErr: false},
		{name: "45.76260E", want: 45.76260, wantErr: false},
		{name: "45.76260W", want: -45.76260, wantErr: false},
		{name: "-45.76260", want: -45.76260, wantErr: false},
		{name: "+45.76260", want: +45.76260, wantErr: false},
		{name: "", wantErr: true},
		{name: "    ", wantErr: true},
		{name: "7.2.1", wantErr: true},
		{name: "7..18", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDegrees(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDegrees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDegrees() got = %v, want %v", got, tt.want)
			}
		})
	}
}
