package commands

import (
	"strings"
	"testing"
)

func TestDockerRun(t *testing.T) {
	cases := []struct {
		directory        string
		imageName        string
		inputs           []string
		settings         []string
		mounts           []string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/", "addition-algorithm-0.0.1-seed:1.0.0",
			[]string{"INPUT_FILE=../examples/addition-algorithm/inputs.txt"},
			[]string{"SETTING_ONE=one", "SETTING_TWO=two"},
			[]string{"MOUNT_BIN=../testdata", "MOUNT_TMP=../testdata"},
			true, ""},
		{"../examples/extractor/", "extractor-0.1.0-seed:0.1.0",
			[]string{"ZIP=../testdata/seed-scale.zip", "MULTIPLE=../testdata/"},
			[]string{"HELLO=Hello"}, []string{"MOUNTAIN=../examples/"},
			true, ""},
	}

	for _, c := range cases {
		//make sure the image exists
		outputDir := "output"
		metadataSchema := ""
		DockerBuild(c.directory)
		err := DockerRun(c.imageName, outputDir, metadataSchema,
			c.inputs, c.settings, c.mounts, true)
		success := err == nil
		if success != c.expected {
			t.Errorf("DockerRun(%q, %q, %q, %q, %q, %q, %q) == %v, expected %v", c.imageName, outputDir, metadataSchema, c.inputs, c.settings, c.mounts, err, nil)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErrorMsg) {
				t.Errorf("DockerRun(%q, %q, %q, %q, %q, %q, %q) == %v, expected %v", c.imageName, outputDir, metadataSchema, c.inputs, c.settings, c.mounts, err.Error(), c.expectedErrorMsg)
			}
		}
	}
}
/*
func TestDefineInputs(t *testing.T) {
	cases := []struct {
		directory        string
		imageName        string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/", "addition-algorithm-0.0.1-seed:1.0.0", true, ""},
		{"../examples/extractor/", "extractor-0.1.0-seed:0.1.0", true, ""},
	}

	for _, c := range cases {
		DockerBuild(c.directory)
		seedFileName, _ := util.SeedFileName(c.directory)

		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		seed2 := objects.SeedFromImageLabel(c.imageName)
		seedStr1 := fmt.Sprintf("%v", seed)
		seedStr2 := fmt.Sprintf("%v", seed2)

		success := seedStr1 != "" && seedStr1 == seedStr2
		if success != c.expected {
			t.Errorf("SeedFromImageLabel(%q) == %v, expected %v", seedFileName, seedStr1, seedStr2)
		}
	}
}

func TestSetOutputDir(t *testing.T) {
	cases := []struct {
		filename         string
		expected         string
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json", "addition-algorithm-0.0.1-seed:1.0.0", ""},
		{"../examples/extractor/seed.manifest.json", "extractor-0.1.0-seed:0.1.0", ""},
	}

	for _, c := range cases {

		seedFileName := util.GetFullPath(c.filename, "")
		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		// Retrieve docker image name
		imageName := objects.BuildImageName(&seed)

		if imageName != c.expected {
			t.Errorf("BuildImageName(%q) == %v, expected %v", seedFileName, imageName, c.expected)
		}
	}
}

func TestDefineMounts(t *testing.T) {
	cases := []struct {
		filename         string
		expected         string
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json", "addition-algorithm-0.0.1-seed:1.0.0", ""},
		{"../examples/extractor/seed.manifest.json", "extractor-0.1.0-seed:0.1.0", ""},
	}

	for _, c := range cases {

		seedFileName := util.GetFullPath(c.filename, "")
		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		// Retrieve docker image name
		imageName := objects.BuildImageName(&seed)

		if imageName != c.expected {
			t.Errorf("BuildImageName(%q) == %v, expected %v", seedFileName, imageName, c.expected)
		}
	}
}

func TestDefineResources(t *testing.T) {
	cases := []struct {
		filename         string
		expected         string
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json", "addition-algorithm-0.0.1-seed:1.0.0", ""},
		{"../examples/extractor/seed.manifest.json", "extractor-0.1.0-seed:0.1.0", ""},
	}

	for _, c := range cases {

		seedFileName := util.GetFullPath(c.filename, "")
		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		// Retrieve docker image name
		imageName := objects.BuildImageName(&seed)

		if imageName != c.expected {
			t.Errorf("BuildImageName(%q) == %v, expected %v", seedFileName, imageName, c.expected)
		}
	}
}

func TestDefineSettings(t *testing.T) {
	cases := []struct {
		filename         string
		expected         string
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json", "addition-algorithm-0.0.1-seed:1.0.0", ""},
		{"../examples/extractor/seed.manifest.json", "extractor-0.1.0-seed:0.1.0", ""},
	}

	for _, c := range cases {

		seedFileName := util.GetFullPath(c.filename, "")
		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		// Retrieve docker image name
		imageName := objects.BuildImageName(&seed)

		if imageName != c.expected {
			t.Errorf("BuildImageName(%q) == %v, expected %v", seedFileName, imageName, c.expected)
		}
	}
}

func TestCheckRunOutput(t *testing.T) {
	cases := []struct {
		filename         string
		expected         string
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json", "addition-algorithm-0.0.1-seed:1.0.0", ""},
		{"../examples/extractor/seed.manifest.json", "extractor-0.1.0-seed:0.1.0", ""},
	}

	for _, c := range cases {

		seedFileName := util.GetFullPath(c.filename, "")
		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		// Retrieve docker image name
		imageName := objects.BuildImageName(&seed)

		if imageName != c.expected {
			t.Errorf("BuildImageName(%q) == %v, expected %v", seedFileName, imageName, c.expected)
		}
	}
}*/
