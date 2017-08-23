package commands

import (
	"strings"
	"testing"
)

func TestSeedInit(t *testing.T) {
	cases := []struct {
		directory string
		expected  string
	}{
		{"../testdata/dummy-scratch/", ""},
		{"../testdata/complete/", "sExisting file left unmodified."},
	}

	for _, c := range cases {
		err := SeedInit(c.directory)

		if c.expected == "" && err != nil {
			t.Errorf("SeedInit(%q) == %v, expected %v", c.directory, err.Error(), c.expected)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expected) {
				t.Errorf("SeedInit(%q) == %v, expected %v", c.directory, err.Error(), c.expected)
			}
		}
	}
}
