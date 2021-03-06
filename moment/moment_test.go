package moment

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	tz "github.com/nkovacs/go-tz"
)

func TestPacked(t *testing.T) {
	cases := []struct {
		name     string
		expected string
	}{
		{
			name:     "America/Los_Angeles",
			expected: "America/Los_Angeles|PST PDT PWT PPT|80 70 70 70|010102301010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010|-261q0 1nX0 11B0 1nX0 SgN0 8x10 iy0 5Wp1 1VaX 3dA0 WM0 1qM0 11A0 1o00 11A0 1o00 11A0 1o00 11A0 1o00 11A0 1qM0 11A0 1o00 11A0 1o00 11A0 1o00 11A0 1o00 11A0 1qM0 WM0 1qM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1fA0 1a00 1fA0 1cN0 1cL0 1cN0 1cL0 1cN0 1cL0 1cN0 1cL0 1cN0 1fz0 1cN0 1cL0 1cN0 1cL0 s10 1Vz0 LB0 1BX0 1cN0 1fz0 1a10 1fz0 1cN0 1cL0 1cN0 1cL0 1cN0 1cL0 1cN0 1cL0 1cN0 1fz0 1a10 1fz0 1cN0 1cL0 1cN0 1cL0 1cN0 1cL0 14p0 1lb0 14p0 1nX0 11B0 1nX0 11B0 1nX0 14p0 1lb0 14p0 1lb0 14p0 1nX0 11B0 1nX0 11B0 1nX0 14p0 1lb0 14p0 1lb0 14p0 1lb0 14p0 1nX0 11B0 1nX0 11B0 1nX0 14p0 1lb0 14p0 1lb0 14p0 1nX0 11B0 1nX0 11B0 1nX0 Rd0 1zb0 Op0 1zb0 Op0 1zb0 Rd0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Rd0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Rd0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Rd0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Rd0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0 Op0 1zb0",
		},
		{
			name:     "Africa/Lagos",
			expected: "Africa/Lagos|LMT WAT|-d.A -10|01|-22y0d.A",
		},
		{
			name:     "Africa/Niamey",
			expected: "Africa/Niamey|LMT WAT|-d.A -10|01|-22y0d.A",
		},
		{
			name:     "Europe/Moscow",
			expected: "Europe/Moscow|MMT MMT MST MDST MSD MSK +05 EET EEST MSK|-2u.h -2v.j -3v.j -4v.j -40 -30 -50 -20 -30 -40|012132345464575454545454545454545458754545454545454545454545454545454545454595|-2ag2u.h 2pyW.W 1bA0 11X0 GN0 1Hb0 c4v.j ik0 3DA0 dz0 15A0 c10 2q10 iM10 23CL0 1db0 1cN0 1db0 1cN0 1db0 1dd0 1cO0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1fA0 1cM0 1cN0 IM0 rX0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1cM0 1fA0 1o00 11A0 1o00 11A0 1o00 11A0 1qM0 WM0 1qM0 WM0 1qM0 11A0 1o00 11A0 1o00 11A0 1qM0 WM0 1qM0 WM0 1qM0 WM0 1qM0 11A0 1o00 11A0 1o00 11A0 1qM0 WM0 8Hz0",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			locData, ok := tz.TZData(c.name)
			if !ok {
				t.Fatalf("error loading timezone data")
			}
			l, err := tz.ParseLocation(c.name, locData)
			if err != nil {
				t.Fatalf("error parsing location: %s", err)
			}

			mtz := Packed(l)
			if !tzMatches(mtz, c.expected) {
				t.Logf("expected: %v", c.expected)
				t.Logf("actual: %v", mtz)
				t.Fatalf("output differs from expected")
			}
		})
	}
}

type MomentData struct {
	Version string
	Zones   []string
}

func TestDavis(t *testing.T) {
	locData, ok := tz.TZData("Antarctica/Davis")
	if !ok {
		t.Fatalf("error loading timezone data")
	}
	l, err := tz.ParseLocation("Antarctica/Davis", locData)
	if err != nil {
		t.Fatalf("error parsing location: %s", err)
	}
	t.Logf("location: %#v", l)
	packed := Packed(l)
	expected := "Antarctica/Davis|-00 +07 +05|0 -70 -50|01012121|-vyo0 iXt0 alj0 1D7v0 VB0 3Wn0 KN0"
	t.Logf("actual: %v", packed)
	t.Logf("expected: %v", expected)
	if packed != expected {
		t.Fatalf("output differs from expected")
	}
}

type zone struct {
	name string
	offs string
}

func parseZones(parts []string) ([]zone, bool) {
	names := strings.Split(parts[1], " ")
	offs := strings.Split(parts[2], " ")
	if len(names) != len(offs) {
		return nil, false
	}
	zones := make([]zone, len(names))
	for i, name := range names {
		zones[i] = zone{
			name: name,
			offs: offs[i],
		}
	}
	return zones, true
}

// tzMatches checks whether the two packed timezones match,
// even if the order of the zones is not the same
func tzMatches(actual, expected string) bool {
	actualParts := strings.Split(actual, "|")
	expectedParts := strings.Split(expected, "|")
	if len(actualParts) < 5 || len(expectedParts) < 5 {
		return false
	}
	if actualParts[0] != expectedParts[0] {
		return false
	}
	// the transitions should be identical
	if actualParts[4] != expectedParts[4] {
		return false
	}

	expectedZones, ok := parseZones(expectedParts)
	if !ok {
		return false
	}
	actualZones, ok := parseZones(actualParts)
	if !ok {
		return false
	}

	if len(actualZones) != len(expectedZones) {
		return false
	}

	// order maps the actual zone index to expected zone index
	order := make([]int, len(expectedZones))
	for i := range expectedZones {
		found := false
		for j := range actualZones {
			if actualZones[j].name == expectedZones[i].name && actualZones[j].offs == expectedZones[i].offs {
				order[j] = i
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	actualReordered := ""
	for _, s := range actualParts[3] {
		idx := strings.IndexRune(digits, s)
		if idx < 0 || idx >= len(order) {
			return false
		}
		actualReordered += toBase60(int64(order[idx]))
	}

	return actualReordered == expectedParts[3]
}

func TestRarotonga(t *testing.T) {
	locData, ok := tz.TZData("Pacific/Rarotonga")
	if !ok {
		t.Fatalf("error loading timezone data")
	}
	l, err := tz.ParseLocation("Pacific/Rarotonga", locData)
	if err != nil {
		t.Fatalf("error parsing location: %s", err)
	}
	t.Logf("location: %#v", l)
	packed := Packed(l)
	expected := "Pacific/Rarotonga|-1030 -0930 -10|au 9u a0|012121212121212121212121212|lyWu IL0 1zcu Onu 1zcu Onu 1zcu Rbu 1zcu Onu 1zcu Onu 1zcu Onu 1zcu Onu 1zcu Onu 1zcu Rbu 1zcu Onu 1zcu Onu 1zcu Onu"
	t.Logf("actual: %v", packed)
	t.Logf("expected: %v", expected)
	if !tzMatches(packed, expected) {
		t.Fatalf("output differs from expected")
	}
}

func TestNoumea(t *testing.T) {
	locData, ok := tz.TZData("Pacific/Noumea")
	if !ok {
		t.Fatalf("error loading timezone data")
	}
	l, err := tz.ParseLocation("Pacific/Noumea", locData)
	if err != nil {
		t.Fatalf("error parsing location: %s", err)
	}
	t.Logf("location: %#v", l)
	packed := Packed(l)
	expected := "Pacific/Noumea|LMT +11 +12|-b5.M -b0 -c0|01212121|-2l9n5.M 2EqM5.M xX0 1PB0 yn0 HeP0 Ao0"
	t.Logf("actual: %v", packed)
	t.Logf("expected: %v", expected)
	if !tzMatches(packed, expected) {
		t.Fatalf("output differs from expected")
	}
}

func TestMomentPacked(t *testing.T) {

	f, err := os.Open("testdata/2019c.json")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	jsonS, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	var data MomentData
	err = json.Unmarshal(jsonS, &data)
	if err != nil {
		t.Error(err)
	}

	for _, zone := range data.Zones {
		nameIdx := strings.Index(zone, "|")
		popIdx := 0
		for i := 0; i < 5; i++ {
			if popIdx+1 > len(zone) {
				popIdx = -1
				break
			}
			addIdx := strings.Index(zone[popIdx:], "|")
			if addIdx == -1 {
				popIdx = -1
				break
			}
			popIdx += addIdx + 1
		}
		expected := zone
		if popIdx > 0 && popIdx <= len(expected) {
			expected = expected[:popIdx-1]
		}

		name := zone[:nameIdx]
		t.Run(name, func(t *testing.T) {
			locData, ok := tz.TZData(name)
			if !ok {
				t.Fatalf("error loading timezone data")
			}
			l, err := tz.ParseLocation(name, locData)
			if err != nil {
				t.Fatalf("error parsing location: %s", err)
			}

			mtz := Packed(l)
			if !tzMatches(mtz, expected) {
				t.Logf("expected: %v", expected)
				t.Logf("actual:   %v", mtz)
				t.Fatalf("output differs from expected")
			}
		})
	}
}
