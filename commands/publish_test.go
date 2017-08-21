package commands

import (
	//"os/exec"
	//"strings"
	"testing"
)

func TestDockerPublish(t *testing.T) {
	cases := []struct {
		directory        string
		imageName        string
		registry         string
		org              string
		deconflict       bool
		pkgmin           bool
		pkgmaj           bool
		algmin           bool
		algmaj           bool
		expectedImgName  string
		expected         bool
		expectedErrorMsg string
	}{
		{"../testdata/dummy-scratch/", "test-seed", "docker.io", "geoint",
			false, false, false, false, false,
			"docker.io/geoint/test-seed", true, ""},
	}

	for _, c := range cases { //TODO: Add test for publish when testing registry is cleaned up
		_ = c
		/*
			buildArgs := []string{"build", "-t", c.imageName, c.directory}
			cmd := exec.Command("docker", buildArgs...)
			cmd.Run()
			err := DockerPublish(c.imageName, c.registry, c.org, c.directory, c.deconflict, c.pkgmin, c.pkgmaj, c.algmin, c.algmaj)
			if err != nil {
				t.Errorf("DockerPublish returned an error: %v", err)
			}
			cmd = exec.Command("docker", "list")
			o, err := cmd.Output()
			paddedName := " " + c.imageName + " "
			if strings.Contains(string(o), paddedName) {
				t.Errorf("DockerPublish() did not remove local image %v", c.imageName)
			}
			if !strings.Contains(string(o), c.expectedImgName) {
				t.Errorf("DockerPublish() did not publish image %v", c.expectedImgName)
			}*/
	}
}
