package bid754

import "testing"

func TestIs754VersionPredicates(t *testing.T) {
	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "1985", got: Is754Version1985(), want: false},
		{name: "2008", got: Is754Version2008(), want: true},
		{name: "2019", got: Is754Version2019(), want: true},
	}

	for _, tc := range tests {
		if tc.got != tc.want {
			t.Fatalf("Is754Version%s() = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}
