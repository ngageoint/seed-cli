package commands

import (
	// 	"fmt"
	// 	"strings"
	"io/ioutil"
	"testing"

	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
)

func init() {
	util.InitPrinter(util.PrintErr)
}

func TestGetManifestLabel(t *testing.T) {
	cases := []struct {
		manifest    string
		version     string
		expectedErr error
	}{
		{"../testdata/escape-chars/pass.seed.manifest.json", "1.0.0", nil},
		// 		"../testdata/escape-chars/fail.seed.manifest.json", "1.0.0", errors.New("Error "),
	}

	for _, c := range cases {

		buf, err := ioutil.ReadFile(c.manifest)
		if err != nil {
			util.PrintUtil("Error reading manifest file %v\n", c.manifest)
		}

		manifest := string(buf)
		util.PrintUtil("manifest from file:\n%s\n\n", manifest)
		// _, err = objects.SeedFromManifestString(manifest)
		// if c.expectedErr == nil && err != nil {
		// 	t.Errorf("SeedFromManifestString(manifest) == %v, expected %v", err.Error(), c.expectedErr)
		// } else {
		//     util.PrintUtil("objects.SeedFromManifestString OK\n\n")
		// }
		seedManifest := objects.GetManifestLabel(c.manifest)
		util.PrintUtil("seed manifest:\n%s\n", seedManifest)
		_, err = objects.SeedFromManifestString(seedManifest)
		if c.expectedErr == nil && err != nil {
			t.Errorf("SeedFromManifestString(seedManifest) == %v, expected %v", err.Error(), c.expectedErr)
		}
	}

}
