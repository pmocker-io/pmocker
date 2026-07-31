package scope

import "testing"

func TestPathLevel(t *testing.T) {
	cases := []struct {
		path string
		want int
	}{
		{"1", 1},
		{"1.2", 2},
		{"1.2.3", 3},
	}
	for _, c := range cases {
		if got := PathLevel(c.path); got != c.want {
			t.Errorf("PathLevel(%s) = %d, want %d", c.path, got, c.want)
		}
	}
}
