package commands

import (
	"testing"
	"strings"
)

func TestDockerBuild(t *testing.T) {
	cases := []struct {
		directory string
		expected bool
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/", true, ""},
		{"../examples/extractor/", true, ""},
		{"", false, "no such file or directory"},
	}

	for _, c := range cases {
		result, err := DockerBuild(c.directory)
		if (result != c.expected ) {
			t.Errorf("DockerBuild(%q) == %v, expected %v", c.directory, result, c.expected)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErrorMsg) {
				t.Errorf("DockerBuild(%q) == %v, expected %v", c.directory, err.Error(), c.expectedErrorMsg)
			}
		}
	}
}