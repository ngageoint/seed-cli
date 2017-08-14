package commands

import (
	"testing"
	"strings"
	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/util"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		seedFileName string
		expected error
		expectedErrorMsg string
	}{
		{"../examples/addition-algorithm/seed.manifest.json", nil, ""},
		{"../examples/extractor/seed.manifest.json", nil, ""},
	}

	for _, c := range cases {
		name := util.GetFullPath(c.seedFileName, "")
		err := ValidateSeedFile( "", name, constants.SchemaManifest)
		if (err != c.expected ) {
			t.Errorf("ValidateSeedFile(%q, %q, %q) == %v, expected %v", "", name, constants.SchemaManifest, err, c.expected)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErrorMsg) {
				t.Errorf("ValidateSeedFile(%q, %q, %q) == %v, expected %v", "", name, constants.SchemaManifest, err.Error(), c.expectedErrorMsg)
			}
		}
	}
}