/*
Seed implements a command line interface library to build and run
docker images defined by a seed.manifest.json file.
usage is as folllows:
	seed build [OPTIONS]
		Options:
		-d, -directory	The directory containing the seed spec and Dockerfile
										(default is current directory)

	seed list [OPTIONS]
		Not yet implemented

	seed publish [OPTIONS]
		Not yet implemented

	seed run [OPTIONS]
		Options:
		-i, -inputData  The input data. May be multiple -id flags defined
										(seedfile: Job.Interface.InputData.Files)
		-in, -imageName The name of the Docker image to run (overrides image name
										pulled from seed spec)
		-o, -outDir			The job output directory. Output defined in
										seedfile: Job.Interface.OutputData.Files and
										Job.Interface.OutputData.Json will be stored relative to
										this directory.
		-s, -schema     The Seed Metadata Schema file; Overrides built in schema to validate
									side-car metadata files against

		-rm				Automatically remove the container when it exits (same as
										docker run --rm)
	seed search [OPTIONS]
		Options:
			-r, -registry	The registry to search
			
			-o, -org		Limit results to this organization/user (for docker hub, this arg is required as images cannot be listed except by org
			
			-u, -user		Optional username to use for authentication
			
			-p, -password	Optional password to use for authentication

	seed validate [OPTIONS]
		Options:
			-d, -directory	The directory containing the seed spec
											(default is current directory)
			-s, -schema			Seed Schema file; Overrides built in schema to validate
											spec against.

	seed version
*/
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ngageoint/seed/cli/constants"
	"github.com/ngageoint/seed/cli/objects"
	"github.com/ngageoint/seed/cli/dockerHubRegistry"
	"github.com/xeipuuv/gojsonschema"
	
	"github.com/heroku/docker-registry-client/registry"
)

var buildCmd *flag.FlagSet
var listCmd *flag.FlagSet
var publishCmd *flag.FlagSet
var runCmd *flag.FlagSet
var searchCmd *flag.FlagSet
var validateCmd *flag.FlagSet
var versionCmd *flag.FlagSet
var directory string
var version string

func main() {

	// Parse input flags
	DefineFlags()

	// seed validate: Validate seed.manifest.json. Does not require docker
	if validateCmd.Parsed() {
		Validate()
		os.Exit(0)
	}

	// Checks if Docker requires sudo access. Prints error message if so.
	CheckSudo()

	// seed list: Lists all seed compliant images on (default) local machine
	if listCmd.Parsed() {
		DockerList()
		os.Exit(0)
	}

	// seed build: Build Docker image
	if buildCmd.Parsed() {
		DockerBuild("")
		os.Exit(0)
	}

	// seed run: Runs docker image provided or found in seed manifest
	if runCmd.Parsed() {
		DockerRun()
		os.Exit(0)
	}
	
	// seed search: Searches registry for seed images
	if searchCmd.Parsed() {
		DockerSearch()
		os.Exit(0)
	}

	// seed publish: Publishes a seed compliant image
	if publishCmd.Parsed() {
		DockerPublish()
		os.Exit(0)
	}
}

//DefineFlags defines the flags available for the seed runner.
func DefineFlags() {

	// seed build flags
	DefineBuildFlags()

	// seed run flags
	DefineRunFlags()

	// seed list flags
	listCmd = flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.Usage = func() {
		PrintListUsage()
	}

	// seed search flags
	DefineSearchFlags()

	// seed publish flags
	DefinePublishFlags()

	// seed validate flags
	DefineValidateFlags()

	// seed version flags
	versionCmd = flag.NewFlagSet(constants.VersionCommand, flag.ExitOnError)
	versionCmd.Usage = func() {
		PrintVersionUsage()
	}


	// Print usage if no command given
	if len(os.Args) == 1 {
		PrintUsage()
	}

	// Parse commands
	switch os.Args[1] {
	case constants.BuildCommand:
		buildCmd.Parse(os.Args[2:])
		if len(buildCmd.Args()) == 1 {
			directory = buildCmd.Args()[0]
		}

	case constants.RunCommand:
		runCmd.Parse(os.Args[2:])
		if len(runCmd.Args()) == 0 {
			PrintRunUsage()
		} else if len(runCmd.Args()) == 1 {
			directory = runCmd.Args()[0]
		}

	case constants.SearchCommand:
		searchCmd.Parse(os.Args[2:])

	case constants.ListCommand:
		listCmd.Parse(os.Args[2:])

	case constants.PublishCommand:
		publishCmd.Parse(os.Args[2:])

		if len(publishCmd.Args()) == 0 {
			PrintPublishUsage()
		} else if len(publishCmd.Args()) == 1 {
			directory = publishCmd.Args()[0]
		}

	case constants.ValidateCommand:
		validateCmd.Parse(os.Args[2:])
		if len(validateCmd.Args()) == 1 {
			directory = validateCmd.Args()[0]
		}

	case constants.VersionCommand:
		versionCmd.Parse(os.Args[2:])
		PrintVersion()

	default:
		fmt.Fprintf(os.Stderr, "%q is not a valid command.\n", os.Args[1])
		PrintUsage()
		os.Exit(0)
	}
}

//PrintCommandUsage prints usage of parsed command, or seed usage if no command
// parsed
func PrintCommandUsage() {
	if buildCmd.Parsed() {
		PrintBuildUsage()
	} else if listCmd.Parsed() {
		PrintListUsage()
	} else if publishCmd.Parsed() {
		PrintPublishUsage()
	} else if runCmd.Parsed() {
		PrintRunUsage()
	} else if searchCmd.Parsed() {
		PrintSearchUsage()
	} else if validateCmd.Parsed() {
		PrintValidateUsage()
	} else {
		PrintUsage()
	}
}

//PrintUsage prints the seed usage arguments
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed COMMAND\n\n")
	fmt.Fprintf(os.Stderr, "A test runner for seed spec compliant algorithms\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  build \tBuilds Seed compliant Docker image\n")
	fmt.Fprintf(os.Stderr, "  list  \tAllows for listing of all Seed compliant images residing on the local system\n")
	fmt.Fprintf(os.Stderr, "  publish\tAllows for publish of Seed compliant images to remote Docker registry\n")
	fmt.Fprintf(os.Stderr, "  run   \tExecutes Seed compliant Docker docker image\n")
	fmt.Fprintf(os.Stderr, "  search\tAllows for discovery of Seed compliant images hosted within a Docker registry (default is docker.io)\n")
	fmt.Fprintf(os.Stderr, "  validate\tValidates a Seed spec\n")
	fmt.Fprintf(os.Stderr, "  version\tPrints the version of Seed spec\n")
	fmt.Fprintf(os.Stderr, "\nRun 'seed COMMAND --help' for more information on a command.\n")
	os.Exit(0)
}


//PrintVersionUsage prints the seed version usage, then exits the program
func PrintVersionUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed version \n")
	fmt.Fprintf(os.Stderr, "\nOutputs the version of the Seed CLI and specification.\n")
	os.Exit(0)
}

//PrintVersion prints the seed CLI version
func PrintVersion() {
	fmt.Fprintf(os.Stderr, "Seed v%s\n", version)
	os.Exit(0)
}
