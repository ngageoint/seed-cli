package commands

import (
	"strings"
	"testing"
)

func TestDockerBuild(t *testing.T) {
	cases := []struct {
		directory        string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/", true, ""},
		{"../examples/extractor/", true, ""},
		{"", false, "no such file or directory"},
	}

	for _, c := range cases {
		err := DockerBuild(c.directory)
		success := err == nil
		if success != c.expected {
			t.Errorf("DockerBuild(%q) == %v, expected %v", c.directory, success, c.expected)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErrorMsg) {
				t.Errorf("DockerBuild(%q) == %v, expected %v", c.directory, err.Error(), c.expectedErrorMsg)
			}
		}
	}
}
