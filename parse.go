package tz

import "errors"

// Zone represents a single time zone such as CEST or CET.
type Zone struct {
	Name   string // abbreviated name, "CET"
	Offset int    // seconds east of UTC
	IsDST  bool   // is this zone Daylight Savings Time?
}

// ZoneTrans represents a single time zone transition.
type ZoneTrans struct {
	When  int64 // transition time, in seconds since 1970 GMT
	Index uint8 // the index of the zone that goes into effect at that time
}

// alpha and omega are the beginning and end of time for zone
// transitions.
const (
	alpha = -1 << 63  // math.MinInt64
	omega = 1<<63 - 1 // math.MaxInt64
)

// Location represents the collection of time offsets
// in use in a geographical area, such as CEST and CET for central Europe.
type Location struct {
	Name string
	Zone []Zone
	Tx   []ZoneTrans
}

// Simple I/O interface to binary blob of data.
type dataIO struct {
	p     []byte
	error bool
}

func (d *dataIO) read(n int) []byte {
	if len(d.p) < n {
		d.p = nil
		d.error = true
		return nil
	}
	p := d.p[0:n]
	d.p = d.p[n:]
	return p
}

func (d *dataIO) big4() (n uint32, ok bool) {
	p := d.read(4)
	if len(p) < 4 {
		d.error = true
		return 0, false
	}
	return uint32(p[0])<<24 | uint32(p[1])<<16 | uint32(p[2])<<8 | uint32(p[3]), true
}

func (d *dataIO) byte() (n byte, ok bool) {
	p := d.read(1)
	if len(p) < 1 {
		d.error = true
		return 0, false
	}
	return p[0], true
}

// Make a string by stopping at the first NUL
func byteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

var errBadData = errors.New("malformed time zone information")

// ParseLocation returns a Location with the given name
// initialized from the IANA Time Zone database-formatted data.
// The data should be in the format of a standard IANA time zone file
// (for example, the content of /etc/localtime on Unix systems).
func ParseLocation(name string, data []byte) (*Location, error) {
	d := dataIO{data, false}

	// 4-byte magic "TZif"
	if magic := d.read(4); string(magic) != "TZif" {
		return nil, errBadData
	}

	// 1-byte version, then 15 bytes of padding
	var p []byte
	if p = d.read(16); len(p) != 16 || p[0] != 0 && p[0] != '2' && p[0] != '3' {
		return nil, errBadData
	}

	// six big-endian 32-bit integers:
	//	number of UTC/local indicators
	//	number of standard/wall indicators
	//	number of leap seconds
	//	number of transition times
	//	number of local time zones
	//	number of characters of time zone abbrev strings
	const (
		NUTCLocal = iota
		NStdWall
		NLeap
		NTime
		NZone
		NChar
	)
	var n [6]int
	for i := 0; i < 6; i++ {
		nn, ok := d.big4()
		if !ok {
			return nil, errBadData
		}
		n[i] = int(nn)
	}

	// Transition times.
	txtimes := dataIO{d.read(n[NTime] * 4), false}

	// Time zone indices for transition times.
	txzones := d.read(n[NTime])

	// Zone info structures
	zonedata := dataIO{d.read(n[NZone] * 6), false}

	// Time zone abbreviations.
	abbrev := d.read(n[NChar])

	// Leap-second time pairs
	d.read(n[NLeap] * 8)

	// Whether tx times associated with local time types
	// are specified as standard time or wall time.
	/* isstd := */
	_ = d.read(n[NStdWall])

	// Whether tx times associated with local time types
	// are specified as UTC or local time.
	/* isutc := */
	_ = d.read(n[NUTCLocal])

	if d.error { // ran out of data
		return nil, errBadData
	}

	// If version == 2 or 3, the entire file repeats, this time using
	// 8-byte ints for txtimes and leap seconds.
	// We won't need those until 2106.

	// Now we can build up a useful data structure.
	// First the zone information.
	//	utcoff[4] isdst[1] nameindex[1]
	zone := make([]Zone, n[NZone])
	for i := range zone {
		var ok bool
		var n uint32
		if n, ok = zonedata.big4(); !ok {
			return nil, errBadData
		}
		zone[i].Offset = int(int32(n))
		var b byte
		if b, ok = zonedata.byte(); !ok {
			return nil, errBadData
		}
		zone[i].IsDST = b != 0
		if b, ok = zonedata.byte(); !ok || int(b) >= len(abbrev) {
			return nil, errBadData
		}
		zone[i].Name = byteString(abbrev[b:])
	}

	// Now the transition time info.
	tx := make([]ZoneTrans, n[NTime])
	for i := range tx {
		var ok bool
		var n uint32
		if n, ok = txtimes.big4(); !ok {
			return nil, errBadData
		}
		tx[i].When = int64(int32(n))
		if int(txzones[i]) >= len(zone) {
			return nil, errBadData
		}
		tx[i].Index = txzones[i]
		// These are ignored.
		/*
			if i < len(isstd) {
				tx[i].Isstd = isstd[i] != 0
			}
			if i < len(isutc) {
				tx[i].Isutc = isutc[i] != 0
			}
		*/
	}

	if len(tx) == 0 {
		// Build fake transition to cover all time.
		// This happens in fixed locations like "Etc/GMT0".
		tx = append(tx, ZoneTrans{When: alpha, Index: 0})
	}

	// Committed to succeed.
	l := &Location{Zone: zone, Tx: tx, Name: name}

	return l, nil
}
