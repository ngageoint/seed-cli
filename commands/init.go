package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/util"
)

//SeedInit places a sample seed.manifest.json within given directory (defaults to CWD)
// Should check for file existence in given directory
// If file exists, warn and exit
// If file does not exist, write sample to given directory
func SeedInit(directory string) error {
	seedFileName, exists := util.GetSeedFileName(directory)
	if exists {
		return errors.New("Pre-existing %s found.", seedFileName)
	}

	// TODO: We need to support init of all supported schema versions in the future
	exampleSeedJson, _ := constants.Asset("schema/0.1.0/seed.manifest.example.json")

	err = ioutil.WriteFile(seedFileName, exampleSeedJson, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error occurred writing example Seed manifest to %s.\n%s\n",
			seedFileName, err.Error())
		return errors.New("Error writing example Seed manifest.")
	}

	return nil
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintInitUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed init [-d JOB_DIRECTORY]\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory to place seed.manifest.json example. (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	panic(util.Exit{0})
}
