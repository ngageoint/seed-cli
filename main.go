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
	"flag"
	"fmt"
	"os"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/commands"
	"github.com/ngageoint/seed-cli/util"
)

var buildCmd flag.FlagSet
var listCmd flag.FlagSet
var publishCmd flag.FlagSet
var runCmd flag.FlagSet
var searchCmd flag.FlagSet
var validateCmd flag.FlagSet
var versionCmd *flag.FlagSet
var version string

func main() {
	// Parse input flags
	DefineFlags()

	// seed validate: Validate seed.manifest.json. Does not require docker
	if validateCmd.Parsed() {
		commands.Validate(validateCmd)
		os.Exit(0)
	}

	// Checks if Docker requires sudo access. Prints error message if so.
	util.CheckSudo()

	// seed list: Lists all seed compliant images on (default) local machine
	if listCmd.Parsed() {
		commands.DockerList()
		os.Exit(0)
	}

	// seed build: Build Docker image
	if buildCmd.Parsed() {
		commands.DockerBuild(buildCmd)
		os.Exit(0)
	}

	// seed run: Runs docker image provided or found in seed manifest
	if runCmd.Parsed() {
		commands.DockerRun(runCmd)
		os.Exit(0)
	}
	
	// seed search: Searches registry for seed images
	if searchCmd.Parsed() {
		commands.DockerSearch(searchCmd)
		os.Exit(0)
	}

	// seed publish: Publishes a seed compliant image
	if publishCmd.Parsed() {
		commands.DockerPublish(publishCmd)
		os.Exit(0)
	}
}

//DefineFlags defines the flags available for the seed runner.
func DefineFlags() {

	// seed build flags
	commands.DefineBuildFlags(&buildCmd)

	// seed run flags
	commands.DefineRunFlags(&runCmd)
	fmt.Println(runCmd)

	// seed list flags
	commands.DefineListFlags(&listCmd)

	// seed search flags
	commands.DefineSearchFlags(&searchCmd)

	// seed publish flags
	commands.DefinePublishFlags(&publishCmd)

	// seed validate flags
	commands.DefineValidateFlags(&validateCmd)

	// seed version flags
	versionCmd = flag.NewFlagSet(constants.VersionCommand, flag.ExitOnError)
	versionCmd.Usage = func() {
		PrintVersionUsage()
	}


	// Print usage if no command given
	if len(os.Args) == 1 {
		PrintUsage()
	}

	var cmd flag.FlagSet

	// Parse commands
	switch os.Args[1] {
	case constants.BuildCommand:
		buildCmd.Parse(os.Args[2:])
		cmd = buildCmd
		cmd.Usage = func() {
			commands.PrintBuildUsage()
		}

	case constants.RunCommand:
		cmd = runCmd

	case constants.SearchCommand:
		cmd = searchCmd

	case constants.ListCommand:
		cmd = listCmd

	case constants.PublishCommand:
		cmd = publishCmd

	case constants.ValidateCommand:
		cmd = validateCmd

	case constants.VersionCommand:
		versionCmd.Parse(os.Args[2:])
		PrintVersion()

	default:
		fmt.Fprintf(os.Stderr, "%q is not a valid command.\n", os.Args[1])
		PrintUsage()
		os.Exit(0)
	}

	cmd.Parse(os.Args[2:])
	fmt.Println(cmd.Usage)
	if len(cmd.Args()) == 0 {
		cmd.Usage()
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
