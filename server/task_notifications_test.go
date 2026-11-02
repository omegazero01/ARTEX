package server

import "testing"

// The cursor goes straight into ::pg_snapshot; malformed input must be rejected
// before the query so a bad browser value is a 400, not a 500 from Postgres.
func TestValidNotificationSnapshot(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},           // no baseline yet
		{"2289:2289:", true}, // empty in-progress list
		{"2289:2295:2290,2293", true},
		{"1:3:2", true},
		{"0:2:", false},                // xmin must be non-zero
		{"5:2:", false},                // xmin > xmax
		{"2289:2295:2290,2290", false}, // duplicate xip
		{"2289:2295:2289", false},      // xip out of range
		{"2289:2295:2296", false},      // xip must be < xmax
		{"abc", false},
		{"2289-2295", false},
		{"2289:2295:1,", false},
	}
	for _, c := range cases {
		if got := validNotificationSnapshot(c.in); got != c.want {
			t.Errorf("validNotificationSnapshot(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
