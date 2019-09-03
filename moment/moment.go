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
	dedupZones := make([]tz.Zone, 0)
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

	usedIdx := make([]int, 0)
	usedIdxMap := make(map[int]struct{})
	usedMap := make(map[int]int)
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
		if _, ok := usedIdxMap[deduped]; !ok {
			usedIdx = append(usedIdx, deduped)
			usedIdxMap[deduped] = struct{}{}
			usedMap[deduped] = len(usedIdx) - 1
		}
		newIndex := uint8(usedMap[deduped])
		transitions = append(transitions, tz.ZoneTrans{
			When:  trans.When,
			Index: newIndex,
		})
	}

	// TODO: this is not always correct, see time.Location.lookupFirstZone in time/zoneinfo.go
	if _, ok := usedIdxMap[0]; !ok && (len(transitions) == 0 || transitions[0].When > alpha32) {
		// add transition from beginning of time
		transitions = append([]tz.ZoneTrans{{
			When:  alpha32,
			Index: 0,
		}}, transitions...)
		usedIdx = append([]int{0}, usedIdx...)
		usedIdxMap[0] = struct{}{}
		for i := 1; i < len(transitions); i++ {
			transitions[i].Index++
		}
	}

	// keep only used zones
	usedZones := make([]tz.Zone, 0)

	for _, idx := range usedIdx {
		zone := dedupZones[idx]
		usedZones = append(usedZones, zone)
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
		indices = append(indices, toBase60(int64(trans.Index)))
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
