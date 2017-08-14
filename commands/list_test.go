package commands

import (
	"testing"
)

func TestDockerList(t *testing.T) {
	cases := []struct {
		expected error
		expectedErrorMsg string
	}{
		{ nil, ""},
	}

	for _, c := range cases {
		err := DockerList()
		if (err != c.expected ) {
			t.Errorf("DockerBuild() == %v, expected %v", err, c.expected)
		}
	}
}