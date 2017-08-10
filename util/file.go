package util

import (
	"fmt"
	"os"
)

//GetFullPath returns the full path of the given file. This expands relative file
// paths and verifes non-relative paths
// Validate path for file existance??
func GetFullPath(rFile string) string {

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
func SeedFileName(dir string) string {
/*
	// Get the proper job directory flag
	if runCmd.Parsed() {
		dir = runCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	} else if buildCmd.Parsed() {
		dir = buildCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	} else if validateCmd.Parsed() {
		dir = validateCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	} else if publishCmd.Parsed() {
		dir = publishCmd.Lookup(constants.JobDirectoryFlag).Value.String()
	}*/

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
	// If not, error and exit
	if _, err := os.Stat(seedFileName); os.IsNotExist(err) {

		// If no seed.manifest.json found, print the command usage and exit
		if len(os.Args) == 2 {
			PrintCommandUsage()
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s cannot be found. Exiting seed...\n",
				seedFileName)
			os.Exit(1)
		}
	}

	return seedFileName
}

func RemoveAll(v string) {
	err := os.RemoveAll(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing temporary directory: %s\n", err.Error())
	}
}
