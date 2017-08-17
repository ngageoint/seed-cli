package commands

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
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

func TestDefineInputs(t *testing.T) {
	cases := []struct {
		seedFileName     string
		inputs           []string
		expectedVol      string
		expectedSize     float64
		expectedTempDir  string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json",
			[]string{"INPUT_FILE=../examples/addition-algorithm/inputs.txt"},
			"[-v INPUT_FILE:INPUT_FILE]", 2.288818359375e-05,
			"map[]", true, ""},
		{"../examples/extractor/seed.manifest.json",
			[]string{"ZIP=../testdata/seed-scale.zip", "MULTIPLE=../testdata/"},
			"[-v ZIP:ZIP -v MULTIPLE:$MULTIPLETEMP$]", 0.0762338638305664,
			"map[MULTIPLE:$MULTIPLETEMP$]", true, ""},
	}

	for _, c := range cases {
		seedFileName := util.GetFullPath(c.seedFileName, "")
		seed := objects.SeedFromManifestFile(seedFileName)
		volumes, size, tempDir, err := DefineInputs(&seed, c.inputs)

		if c.expected != (err == nil) {
			t.Errorf("DefineInputs(%q, %q) == %v, expected %v", seedFileName, c.inputs, err, nil)
		}

		expectedVol := c.expectedVol
		expectedTempDir := c.expectedTempDir
		for _, f := range c.inputs {
			x := strings.Split(f, "=")
			path := util.GetFullPath(x[1], "")
			expectedVol = strings.Replace(expectedVol, x[0], path, -1)

			tempDir, ok := tempDir[x[0]]
			if ok {
				defer util.RemoveAllFiles(tempDir)
				tempVarStr := fmt.Sprintf("$%sTEMP$", x[0])
				expectedVol = strings.Replace(expectedVol, tempVarStr, tempDir, -1)
				expectedTempDir = strings.Replace(expectedTempDir, tempVarStr, tempDir, -1)
			}
		}
		tempStr := fmt.Sprintf("%v", volumes)
		if expectedVol != tempStr {
			t.Errorf("DefineInputs(%q, %q) == %v, expected %v", seedFileName, c.inputs, tempStr, expectedVol)
		}

		if c.expectedSize != size {
			t.Errorf("DefineInputs(%q, %q) == %v, expected %v", seedFileName, c.inputs, size, c.expectedSize)
		}

		tempStr = fmt.Sprintf("%v", tempDir)
		if expectedTempDir != tempStr {
			t.Errorf("DefineInputs(%q, %q) == %v, expected %v", seedFileName, c.inputs, tempStr, expectedTempDir)
		}

	}
}

/*
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
