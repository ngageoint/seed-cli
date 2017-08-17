package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

//DockerPublish executes the seed publish command
func DockerPublish(origImg, registry, org, jobDirectory string, deconflict,
	increasePkgMinor, increasePkgMajor, increaseAlgMinor, increaseAlgMajor bool) {

	//1. Check names and verify it doesn't conflict
	tag := ""
	img := origImg

	// docker tag if registry and/or org specified
	if registry != "" || org != "" {
		if org != "" {
			tag = org + "/"
		}
		if registry != "" {
			tag = registry + "/" + tag
		}

		img = tag + img
	}

	// Check for image confliction.
	conflict := false //TODO - Need to call seed search when implemented

	// If it conflicts, bump specified version number
	if conflict && deconflict {
		//1. Verify we have a valid manifest (-d option or within the current directory)
		seedFileName, err := util.SeedFileName(jobDirectory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: no %s found in %s.\nExiting seed...\n", constants.SeedFileName, jobDirectory)
			panic(util.Exit{1})
		}
		ValidateSeedFile("", seedFileName, constants.SchemaManifest)
		seed := objects.SeedFromManifestFile(seedFileName)

		fmt.Fprintf(os.Stderr, "INFO: An image with the name %s already exists. ", img)
		// Bump the package minor version
		if increasePkgMinor {
			pkgVersion := strings.Split(seed.Job.PackageVersion, ".")
			minorVersion, _ := strconv.Atoi(pkgVersion[1])
			pkgVersion[1] = strconv.Itoa(minorVersion + 1)
			seed.Job.PackageVersion = strings.Join(pkgVersion, ".")

			fmt.Fprintf(os.Stderr, "The package version will be increased to %s.\n",
				seed.Job.PackageVersion)

			// Bump the package major version
		} else if increasePkgMajor {
			pkgVersion := strings.Split(seed.Job.PackageVersion, ".")
			majorVersion, _ := strconv.Atoi(pkgVersion[0])
			pkgVersion[0] = strconv.Itoa(majorVersion + 1)
			seed.Job.PackageVersion = strings.Join(pkgVersion, ".")

			fmt.Fprintf(os.Stderr, "The major package version will be increased to %s.\n",
				seed.Job.PackageVersion)

			// Bump the algorithm minor version
		} else if increaseAlgMinor {

			algVersion := strings.Split(seed.Job.AlgorithmVersion, ".")
			minorVersion, _ := strconv.Atoi(algVersion[1])
			algVersion[1] = strconv.Itoa(minorVersion + 1)
			seed.Job.AlgorithmVersion = strings.Join(algVersion, ".")

			fmt.Fprintf(os.Stderr, "The minor algorithm version will be increased to %s.\n",
				seed.Job.AlgorithmVersion)

			// Bump the algorithm major verion
		} else if increaseAlgMajor {
			algVersion := strings.Split(seed.Job.AlgorithmVersion, ".")
			majorVersion, _ := strconv.Atoi(algVersion[0])
			algVersion[0] = strconv.Itoa(majorVersion + 1)
			seed.Job.AlgorithmVersion = strings.Join(algVersion, ".")

			fmt.Fprintf(os.Stderr, "The major algorithm version will be increased to %s.\n",
				seed.Job.AlgorithmVersion)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: No tag deconfliction method specified. Aborting seed publish.\n")
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			panic(util.Exit{1})
		}

		img = objects.BuildImageName(&seed)
		fmt.Fprintf(os.Stderr, "\nNew image name: %s\n", img)

		// write version back to the seed manifest
		seedJSON, _ := json.Marshal(&seed)
		err = ioutil.WriteFile(seedFileName, seedJSON, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error occurred writing updated seed version to %s.\n%s\n",
				seedFileName, err.Error())
		}

		// Build Docker image
		fmt.Fprintf(os.Stderr, "INFO: Building %s\n", img)
		buildArgs := []string{"build", "-t", img, jobDirectory}
		if util.DockerVersionHasLabel() {
			// Set the seed.manifest.json contents as an image label
			label := "com.ngageoint.seed.manifest=" + objects.GetManifestLabel(seedFileName)
			buildArgs = append(buildArgs, "--label", label)
		}
		rebuildCmd := exec.Command("docker", buildArgs...)
		var errs bytes.Buffer
		rebuildCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
		rebuildCmd.Stdout = os.Stderr

		// Run docker build
		rebuildCmd.Run()

		// check for errors on stderr
		if errs.String() != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Error re-building image '%s':\n%s\n",
				img, errs.String())
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			panic(util.Exit{1})
		}

		// Set final image name to tag + image
		img = tag + img
	}

	var errs bytes.Buffer

	// Run docker tag
	if img != origImg {
		fmt.Fprintf(os.Stderr, "INFO: Tagging image %s as %s\n", origImg, img)
		tagCmd := exec.Command("docker", "tag", origImg, img)
		tagCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
		tagCmd.Stdout = os.Stderr

		if err := tagCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error executing docker tag. %s\n",
				err.Error())
		}
		if errs.String() != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Error tagging image '%s':\n%s\n", origImg, errs.String())
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			panic(util.Exit{1})
		}
	}

	// docker push
	fmt.Fprintf(os.Stderr, "INFO: Performing docker push %s\n", img)
	errs.Reset()
	pushCmd := exec.Command("docker", "push", img)
	pushCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	pushCmd.Stdout = os.Stdout

	// Run docker push
	if err := pushCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker push. %s\n",
			err.Error())
	}

	// Check for errors. Exit if error occurs
	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error pushing image '%s':\n%s\n", img,
			errs.String())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		panic(util.Exit{1})
	}

	// docker rmi
	errs.Reset()
	fmt.Fprintf(os.Stderr, "INFO: Removing local image %s\n", img)
	rmiCmd := exec.Command("docker", "rmi", img)
	rmiCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	rmiCmd.Stdout = os.Stdout

	if err := rmiCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker rmi. %s\n",
			err.Error())
	}

	// check for errors on stderr
	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error removing image '%s':\n%s\n", img,
			errs.String())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		panic(util.Exit{1})
	}
}

//PrintPublishUsage prints the seed publish usage information, then exits the program
func PrintPublishUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed publish [-r REGISTRY_NAME] [-o ORG_NAME] [-f] [-p | -P | -a | -A] -d DIRECTORY IMAGE_NAME\n")
	fmt.Fprintf(os.Stderr, "\nAllows for the publish of seed compliant images.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s Specifies the directory containing the seed.manifest.json and dockerfile\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific registry to which to publish the image\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific registry to which to publish the image\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Minor version bump of 'packageVersion' in manifest on disk if publish conflict found\n",
		constants.PkgVersionMinor)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Major version bump of 'packageVersion' in manifest on disk if publish conflict found\n",
		constants.PkgVersionMajor)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Minor version bump of 'algorithmVersion' in manifest on disk if publish conflict found\n",
		constants.AlgVersionMinor)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Major version bump of 'algorithmVersion' in manifest on disk if publish conflict found\n",
		constants.AlgVersionMajor)
	panic(util.Exit{0})
}
