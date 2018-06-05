package tz

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		name string
	}{
		{"Etc/GMT"},
		{"Etc/GMT-0"},
		{"Etc/GMT-1"},
		{"Etc/GMT-2"},
		{"Etc/GMT-4"},
		{"Etc/GMT-5"},
		{"Etc/GMT-6"},
		{"Etc/GMT-7"},
		{"Etc/GMT-8"},
		{"Etc/GMT-9"},
		{"Etc/GMT-10"},
		{"Etc/GMT-11"},
		{"Etc/GMT+0"},
		{"Etc/GMT+1"},
		{"Etc/GMT+2"},
		{"Etc/GMT+4"},
		{"Etc/GMT+5"},
		{"Etc/GMT+6"},
		{"Etc/GMT+7"},
		{"Etc/GMT+8"},
		{"Etc/GMT+9"},
		{"Etc/GMT+10"},
		{"Etc/GMT+11"},
		{"Australia/Sydney"},
		{"America/New_York"},
		{"US/Central"},
		{"US/Eastern"},
		{"US/Pacific"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			locData, ok := TZData(c.name)
			if !ok {
				t.Fatalf("error loading timezone data")
			}
			_, err := ParseLocation(c.name, locData)
			if err != nil {
				t.Fatalf("error parsing location: %s", err)
			}
		})
	}
}

func BenchmarkParseLocation(b *testing.B) {
	name := "US/Central"
	for i := 0; i < b.N; i++ {
		data, ok := TZData(name)
		if !ok {
			b.Fatal("error loading timezone data")
		}
		ParseLocation(name, data)
	}
}
