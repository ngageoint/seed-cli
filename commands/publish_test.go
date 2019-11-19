package commands

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/ngageoint/seed-common/objects"
	RegistryFactory "github.com/ngageoint/seed-common/registry"
	"github.com/ngageoint/seed-common/util"
)

func init() {
	util.InitPrinter(util.Quiet, nil, nil)
}

func TestDockerPublish(t *testing.T) {
	util.RestartRegistry()

	//build images to be used for testing in advance
	imgDirs := []string{"../testdata/complete/"}
	imgNames := []string{"my-job-0.1.0-seed:0.1.0"}
	version := "1.0.0"
	for _, dir := range imgDirs {
		_, err := DockerBuild(dir, version, "", "", ".", ".", "", false)
		if err != nil {
			t.Errorf("Error building image %v for DockerPublish test", dir)
		}
	}

	cases := []struct {
		directory        string
		imageName        string
		manifest         string
		registry         string
		org              string
		username         string
		password         string
		force            bool
		pkgpatch         bool
		pkgmin           bool
		pkgmaj           bool
		jobpatch         bool
		jobmin           bool
		jobmaj           bool
		expectedImgName  string
		jobVersion       string
		packageVersion   string
		expected         bool
		expectedErrorMsg string
	}{
		{imgDirs[0], imgNames[0], "", "localhost:5000", "test", "testuser", "testpassword",
			false, false, false, false, false, false, false,
			"localhost:5000/test/my-job-0.1.0-seed:0.1.0", "0.1.0", "0.1.0", true, ""},
		{imgDirs[0], imgNames[0], "", "localhost:5000", "", "testuser", "testpassword",
			false, false, false, false, false, false, false,
			"localhost:5000/my-job-0.1.0-seed:0.1.0", "0.1.0", "0.1.0", true, ""},
		{imgDirs[0], imgNames[0], "", "localhost:5000", "test", "testuser", "testpassword",
			true, false, false, false, false, false, false,
			"localhost:5000/test/my-job-0.1.0-seed:0.1.0", "0.1.0", "0.1.0", true, ""},
		{imgDirs[0], imgNames[0], "", "localhost:5000", "test", "testuser", "testpassword",
			false, false, false, false, false, false, false,
			"localhost:5000/test/my-job-0.1.0-seed:0.1.0", "0.1.0", "0.1.0", false, "Image exists and no tag deconfliction method specified."},
		{imgDirs[0], imgNames[0], "", "localhost:5000", "test", "", "",
			false, false, false, false, false, false, false,
			"localhost:5000/test/my-job-0.1.0-seed:0.1.0", "0.1.0", "0.1.0", false, "The specified registry requires a login.  Please try again with a username (-u) and password (-p)."},
		{imgDirs[0], imgNames[0], "", "localhost:5000", "test", "testuser", "testpassword",
			false, false, false, true, true, false, false,
			"localhost:5000/test/my-job-0.1.1-seed:1.0.0", "0.1.1", "1.0.0", true, ""},
		{imgDirs[0], "", "", "localhost:5000", "test", "testuser", "testpassword",
			false, false, false, true, true, false, false,
			"localhost:5000/test/my-job-0.1.2-seed:2.0.0", "0.1.2", "2.0.0", true, ""},
	}

	for i, c := range cases {
		img, err := DockerPublish(c.imageName, c.manifest, c.registry, c.org, c.username, c.password, c.directory,
			c.force, c.pkgmaj, c.pkgmin, c.pkgpatch, c.jobmaj, c.jobmin, c.jobpatch)

		reg, err2 := RegistryFactory.CreateRegistry(c.registry, c.org, c.username, c.password)
		var seed objects.Seed
		var prefix string
		if c.registry != "" {
			prefix = c.registry + "/"
		}
		tempStr := strings.Replace(c.expectedImgName, prefix, "", 1)
		temp := strings.Split(tempStr, ":")
		repoName := temp[0]
		repoTag := temp[1]
		if c.expected && err2 == nil && reg != nil {
			manifest, _ := reg.GetImageManifest(repoName, repoTag)
			seed, _ = objects.SeedFromManifestString(manifest)
		}

		if err != nil && c.expected == true {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerPublish returned an error: %v\n", err)
		}
		if err != nil && !strings.Contains(err.Error(), c.expectedErrorMsg) {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerPublish returned an error: %v\n expected %v", err, c.expectedErrorMsg)
		}
		if err == nil && c.expected == false {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerPublish did not receive an expected error: %v\n", c.expectedErrorMsg)
		}
		if c.expected && c.expectedImgName != img {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerPublish returned image %v instead of %v\n", img, c.expectedImgName)
		}
		if c.expected && c.jobVersion != seed.Job.JobVersion {
			t.Errorf("test %v failed\n", i)
			t.Errorf("Job version %v is wrong; expected %v\n", seed.Job.JobVersion, c.jobVersion)
		}
		if c.expected && c.packageVersion != seed.Job.PackageVersion {
			t.Errorf("test %v failed\n", i)
			t.Errorf("Package version %v is wrong; expected %v\n", seed.Job.PackageVersion, c.packageVersion)
		}
		cmd := exec.Command("docker", "list")
		o, err := cmd.Output()
		paddedName := " " + c.expectedImgName + " "
		if strings.Contains(string(o), paddedName) {
			t.Errorf("test %v failed\n", i)
			t.Errorf("DockerPublish() did not remove local image %v after publishing it", c.imageName)
		}
	}
}
