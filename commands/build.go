package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

//DockerBuild Builds the docker image with the given image tag.
func DockerBuild(jobDirectory, username, password string) error {
	if username != "" {
		registry, err := util.DockerfileBaseRegistry(jobDirectory)
		if err != nil {
			fmt.Fprintf(os.Stderr,"Error getting registry from dockerfile: %s\n", err.Error())
		}
		err = util.Login(registry, username, password)
		if err != nil {
			fmt.Fprintf(os.Stderr,"Error calling docker login: %s\n", err.Error())
		}
	}

	seedFileName, err := util.SeedFileName(jobDirectory)
	if err != nil && os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: %s cannot be found.\n",
			seedFileName)
		fmt.Fprintf(os.Stderr, "Make sure you have specified the correct directory.\n")
		return err
	}

	// Validate seed file
	err = ValidateSeedFile("", seedFileName, constants.SchemaManifest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: seed file could not be validated. See errors for details.")
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		return err
	}

	// retrieve seed from seed manifest
	seed := objects.SeedFromManifestFile(seedFileName)

	// Retrieve docker image name
	imageName := objects.BuildImageName(&seed)

	// Build Docker image
	fmt.Fprintf(os.Stderr, "INFO: Building %s\n", imageName)
	buildArgs := []string{"build", "-t", imageName, jobDirectory}
	if util.DockerVersionHasLabel() {
		// Set the seed.manifest.json contents as an image label
		label := "com.ngageoint.seed.manifest=" + objects.GetManifestLabel(seedFileName)
		buildArgs = append(buildArgs, "--label", label)
	}
	cmd := exec.Command("docker", buildArgs...)
	var errs bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = os.Stderr

	// Run docker build
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker build. %s\n",
			err.Error())
		return err
	}

	// check for errors on stderr
	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error building image '%s':\n%s\n",
			imageName, errs.String())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintBuildUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed build [-d JOB_DIRECTORY]\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory containing Seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tUsername to login if needed to pull images (will use cached login if available and not specified).\n",
		constants.ShortUserFlag, constants.UserFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tPassword to login if needed to pull images (will use cached login if available and not specified).\n",
		constants.ShortPassFlag, constants.PassFlag)
	panic(util.Exit{0})
}
