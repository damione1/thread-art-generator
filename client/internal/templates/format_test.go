package templates

import "testing"

func TestHumanInt(t *testing.T) {
	for _, c := range []struct {
		in   int32
		want string
	}{{0, "0"}, {7, "7"}, {450, "450"}, {3148, "3 148"}, {12000, "12 000"}, {2412000, "2 412 000"}, {-1500, "-1 500"}} {
		if got := humanInt(c.in); got != c.want {
			t.Errorf("humanInt(%d) = %q, want %q", c.in, got, c.want)
		}
	}
	if got := threadMetres(2412000); got != "2 412 m" {
		t.Errorf("threadMetres = %q", got)
	}
}
