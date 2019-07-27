package commands

import (
	"os"
	"strings"
	"testing"
	"time"

	common_const "github.com/ngageoint/seed-common/constants"
	RegistryFactory "github.com/ngageoint/seed-common/registry"
	"github.com/ngageoint/seed-common/util"
)

func init() {
	util.InitPrinter(util.Quiet, nil, nil)
}

func TestDockerUnpublish(t *testing.T) {
	util.RestartRegistry()

	registry := "localhost:5000"
	username := "testuser"
	password := "testpassword"

	//set config dir so we don't stomp on other users' logins with sudo
	configDir := common_const.DockerConfigDir + time.Now().Format(time.RFC3339)
	os.Setenv(common_const.DockerConfigKey, configDir)
	defer util.RemoveAllFiles(configDir)
	defer os.Unsetenv(common_const.DockerConfigKey)

	err := util.Login(registry, username, password)
	if err != nil {
		util.PrintUtil(err.Error())
	}

	//build images to be used for testing in advance
	imgDirs := []string{"../testdata/complete/"}
	imgNames := []string{"my-job-0.1.0-seed:0.1.0", "my-job-1.0.0-seed:1.0.0", "not-a-valid-image:latest"}
	manifests := []string{"../testdata/complete/seed.manifest.json"}
	origImg := "my-job-0.1.0-seed:0.1.0"
	remoteImg := []string{"localhost:5000/my-job-0.1.0-seed:0.1.0", "localhost:5000/my-job-1.0.0-seed:1.0.0", "localhost:5000/not-a-valid-image"}
	version := "1.0.0"

	for _, dir := range imgDirs {
		_, err := DockerBuild(dir, version, "", "", ".", ".", "")
		if err != nil {
			t.Errorf("Error building image %v for DockerUnpublish test", dir)
		}
	}

	for _, img := range remoteImg {
		err := util.Tag(origImg, img)
		if err != nil {
			t.Errorf("Error tagging image %v for DockerUnpublish test: %v", img, err)
		}

		err = util.Push(img)
		if err != nil {
			t.Errorf("Error pushing image %v for DockerUnpublish test: %v", img, err)
		}
	}

	cases := []struct {
		imageName        string
		manifest         string
		registry         string
		org              string
		username         string
		password         string
		expected         bool
		expectedErrorMsg string
		repoName         string
		repoTag          string
	}{
		// test missing image name and missing manifest
		{"", "", "localhost:5000", "", "testuser", "testpassword",
			false, "seed.manifest.json cannot be found", "", ""},
		// test manifest without login
		{"", manifests[0], "localhost:5000", "", "", "",
			false, "The specified registry requires a login", "", ""},
		// test manifest with bad login
		{"", manifests[0], "localhost:5000", "", "wrong", "bad",
			false, "Incorrect username/password", "", ""},
		// test successful manifest
		{"", manifests[0], "localhost:5000", "", "testuser", "testpassword",
			true, "", "my-job-0.1.0-seed", "0.1.0"},
		// test trying to remove the same image twice, this time with image name
		{imgNames[0], "", "localhost:5000", "", "testuser", "testpassword",
			false, "status=404", "", ""},
		// test image name with no login and bad login
		{imgNames[2], manifests[0], "localhost:5000", "", "", "",
			false, "The specified registry requires a login.", "", ""},
		{imgNames[2], manifests[0], "localhost:5000", "", "wrong", "bad",
			false, "Incorrect username/password", "", ""},
		// test successful image name
		{imgNames[2], manifests[0], "localhost:5000", "", "testuser", "testpassword",
			true, "", "not-a-valid-image", "latest"},
	}

	for i, c := range cases {
		err := DockerUnpublish(c.imageName, c.manifest, c.registry, c.org, c.username, c.password)

		if err != nil && c.expected == true {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerUnpublish returned an error: %v\n", err)
		}
		if err != nil && !strings.Contains(err.Error(), c.expectedErrorMsg) {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerUnpublish returned an error: %v\n expected %v", err, c.expectedErrorMsg)
		}
		if err == nil && c.expected == false {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerUnpublish did not receive an expected error: %v\n", c.expectedErrorMsg)
		}

		if err == nil {
			reg, err2 := RegistryFactory.CreateRegistry(c.registry, c.org, c.username, c.password)
			if reg != nil && err2 == nil {
				_, err2 = reg.GetImageManifest(c.repoName, c.repoTag)
				if err2 == nil {
					t.Errorf("test %v failed\n", i)
					t.Errorf("DockerUnpublish did not remove image %v from registry %v\n", c.imageName, c.registry)
				}
			} else {
				t.Errorf("test %v failed\n", i)
				t.Errorf("Error creating registry to check if image was removed: %v\n", err2.Error())
			}
		}
	}
}
