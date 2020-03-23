// Package moment converts timezone data
// to the format used by moment-timezone.js.
package moment

import (
	"fmt"
	"strings"

	tz "github.com/nkovacs/go-tz"
)

const digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alpha = -1 << 63 // math.MinInt64

func toBase60(n int64) string {
	s := ""
	if n < 0 {
		s = "-"
		n = -n
	}
	u := uint64(n)
	b := uint64(60)
	var a [13 + 1]byte // 13 bytes is enough for base 32, +1 for sign
	i := len(a)
	for u >= b {
		i--
		q := u / b
		a[i] = digits[uint(u-q*b)]
		u = q
	}
	// u < base
	i--
	a[i] = digits[uint(u)]

	s += string(a[i:])
	return s
}

func packMinutes(sec int64) string {
	mins := sec / 60
	fraction := sec - mins*60
	if fraction < 0 {
		fraction = -fraction
	}
	s := ""
	if mins != 0 || fraction == 0 {
		s = toBase60(mins)
	}
	if fraction != 0 {
		// since we're dividing whole numbers by 60, fraction can only be 0..59, which is a single digit in base 60
		s += "." + toBase60(fraction)
	}
	return s
}

// Packed converts the location to moment-timezone.js packed format.
func Packed(l *tz.Location) string {
	// dedupZones are the deduplicated zones
	dedupZones := make([]tz.Zone, 0)
	// dedupMap maps the original zone index in transitions to the deduplicated zone's index
	dedupMap := make(map[int]int)

	for i, zone := range l.Zone {
		found := false
		for j := range dedupZones {
			if dedupZones[j].Name == zone.Name && dedupZones[j].Offset == zone.Offset {
				found = true
				dedupMap[i] = j
				break
			}
		}
		if found {
			continue
		}
		dedupZones = append(dedupZones, zone)
		dedupMap[i] = len(dedupZones) - 1
	}

	used := make(map[int]struct{})
	transitions := make([]tz.ZoneTrans, 0)

	// cut off 64 bit dates to generate the same dates as momentjs timezone
	// this is only necessary to match the data from momentjs
	const alpha32 = -1 << 31
	const omega32 = 1<<31 - 1

	for i, trans := range l.Tx {
		if i+1 < len(l.Tx) && l.Tx[i+1].When < alpha32 {
			continue
		}
		if trans.When > omega32 {
			break
		}
		deduped, ok := dedupMap[int(trans.Index)]
		if !ok {
			panic(fmt.Sprintf("zone not found: %v", trans.Index))
		}
		used[deduped] = struct{}{}
		newIndex := uint8(deduped)
		when := trans.When
		if when < alpha32 {
			when = alpha32
		}
		transitions = append(transitions, tz.ZoneTrans{
			When:  when,
			Index: newIndex,
		})
	}

	if len(transitions) == 0 || transitions[0].When > alpha32 {
		findFirstZone := func() int {
			// case 1: if the first zone is unused, use it
			if _, ok := used[0]; !ok {
				return 0
			}

			// case 2: if the first transition is to a dst zone,
			// find the first zone before it that is not dst
			if len(transitions) > 0 && dedupZones[int(transitions[0].Index)].IsDST {
				for zi := int(transitions[0].Index) - 1; zi >= 0; zi-- {
					if !dedupZones[zi].IsDST {
						return zi
					}
				}
			}

			// case 3: use the first one that is not dst
			for zi := range dedupZones {
				if !dedupZones[zi].IsDST {
					return zi
				}
			}

			// case 4: use the first zone
			return 0
		}

		firstZone := findFirstZone()
		used[firstZone] = struct{}{}
		transitions = append([]tz.ZoneTrans{{
			When:  alpha32,
			Index: uint8(firstZone),
		}}, transitions...)
	}

	// keep only used zones
	usedZones := make([]tz.Zone, 0)
	// usedMap maps deduped idx to used idx
	usedMap := make(map[int]int)
	for i, zone := range dedupZones {
		if _, ok := used[i]; ok {
			usedZones = append(usedZones, zone)
			usedMap[i] = len(usedZones) - 1
		}
	}

	var abbrevMap []string
	var offsetMap []string

	for _, zone := range usedZones {
		abbrevMap = append(abbrevMap, zone.Name)
		offsetMap = append(offsetMap, packMinutes(-int64(zone.Offset)))
	}

	var untils []string
	var indices []string

	lastTimeStamp := int64(0)
	first := true
	prevZoneIdx := -1

	for _, trans := range transitions {
		if prevZoneIdx == int(trans.Index) {
			continue
		}
		prevZoneIdx = int(trans.Index)
		indices = append(indices, toBase60(int64(usedMap[int(trans.Index)])))
		if first {
			first = false
			continue
		}
		ts := trans.When - lastTimeStamp
		lastTimeStamp = trans.When
		untils = append(untils, packMinutes(ts))
	}

	return fmt.Sprintf("%v|%v|%v|%v|%v",
		l.Name,
		strings.Join(abbrevMap, " "),
		strings.Join(offsetMap, " "),
		strings.Join(indices, ""),
		strings.Join(untils, " "),
	)
}
