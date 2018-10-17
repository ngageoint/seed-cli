package commands

import (
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
		{"../testdata/escape-chars/seed.manifest.json", "1.0.0", nil},
	}

	for _, c := range cases {
		seedManifest := objects.GetManifestLabel(c.manifest)
		_, err := objects.SeedFromManifestString(seedManifest)
		if c.expectedErr == nil && err != nil {
			t.Errorf("SeedFromManifestString(seedManifest) == %v, expected %v", err.Error(), c.expectedErr)
		}
	}

}
