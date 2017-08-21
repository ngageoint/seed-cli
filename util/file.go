package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ngageoint/seed-cli/constants"
)

//GetFullPath returns the full path of the given file. This expands relative file
// paths and verifes non-relative paths
// Validate path for file existance??
func GetFullPath(rFile, directory string) string {

	// Normalize
	rFile = filepath.Clean(filepath.ToSlash(rFile))

	if !filepath.IsAbs(rFile) {

		// Expand relative file path
		// Define the current working directory
		curDir, _ := os.Getwd()

		// Test relative to current directory
		dir := filepath.Join(curDir, rFile)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			rFile = filepath.Clean(dir)

			// see if parent directory exists. If so, assume this directory will be created
		} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
			rFile = filepath.Clean(dir)
		}

		// Test relative to working directory
		if directory != "." {
			dir = filepath.Join(directory, rFile)
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)

				// see if parent directory exists. If so, assume this directory will be created
			} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)
			}
		}
	}

	return rFile
}

//SeedFileName Finds and returns the full filepath to the seed.manifest.json
func SeedFileName(dir string) (string, error) {
	// Define the current working directory
	curDirectory, _ := os.Getwd()

	// set path to seed file -
	// 	Either relative to current directory or located in given directory
	//  	-d directory might be a relative path to current directory
	seedFileName := constants.SeedFileName
	if dir == "." {
		seedFileName = filepath.Join(curDirectory, seedFileName)
	} else {
		if filepath.IsAbs(dir) {
			seedFileName = filepath.Join(dir, seedFileName)
		} else {
			seedFileName = filepath.Join(curDirectory, dir, seedFileName)
		}
	}

	// Verify seed.json exists within specified directory.
	_, err := os.Stat(seedFileName)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: %s cannot be found.\n",
			seedFileName)
		fmt.Fprintf(os.Stderr, "Make sure you have specified the correct directory.\n")
	}

	return seedFileName, err
}

func RemoveAllFiles(v string) {
	err := os.RemoveAll(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing directory: %s\n", err.Error())
	}
}
