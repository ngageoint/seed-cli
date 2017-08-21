package commands

import (
	//"os/exec"
	//"strings"
	"testing"
)

func TestDockerSearch(t *testing.T) {
	cases := []struct {
		search           string
		registry         string
		org              string
		expectedResult   string
		expectedErrorMsg string
	}{
		{"", "docker.io", "geoint",
			"docker.io/geoint/test-seed", ""},
	}

	for _, c := range cases { //TODO: Add test for search when testing registry is cleaned up
		_ = c
	}
}
