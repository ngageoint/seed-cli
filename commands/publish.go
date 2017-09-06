package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

//DockerPublish executes the seed publish command
func DockerPublish(origImg, registry, org, username, password, jobDirectory string,
	force, P, pm, pp, J, jm, jp bool) error {

	if origImg == "" {
		err := errors.New("ERROR: No input image specified.")
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	if exists, err := util.ImageExists(origImg); !exists {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}


	if username != "" {
		//set config dir so we don't stomp on other users' logins with sudo
		configDir := constants.DockerConfigDir + time.Now().Format(time.RFC3339)
		os.Setenv(constants.DockerConfigKey, configDir)
		defer util.RemoveAllFiles(configDir)
		defer os.Unsetenv(constants.DockerConfigKey)

		err := util.Login(registry, username, password)
		if err != nil {
			fmt.Println(err)
		}
	}

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
	images, err := DockerSearch(registry, org, "", username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error searching for matching tag names.\n%s\n",
			err.Error())
	}
	conflict := util.ContainsString(images, origImg)
	if conflict {
		fmt.Fprintf(os.Stderr, "INFO: Image %s exists on registry %s\n", img, registry)
	}

	// If it conflicts, bump specified version number
	if conflict && !force {
		fmt.Fprintf(os.Stderr, "INFO: Force flag not specified, attempting to rebuild with new version number.\n")

		//1. Verify we have a valid manifest (-d option or within the current directory)
		seedFileName, err := util.SeedFileName(jobDirectory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
			return err
		}
		ValidateSeedFile("", seedFileName, constants.SchemaManifest)
		seed := objects.SeedFromManifestFile(seedFileName)

		fmt.Fprintf(os.Stderr, "INFO: An image with the name %s already exists. ", img)
		// Bump the package patch version
		if pp {
			pkgVersion := strings.Split(seed.Job.PackageVersion, ".")
			patchVersion, _ := strconv.Atoi(pkgVersion[2])
			pkgVersion[2] = strconv.Itoa(patchVersion + 1)
			seed.Job.PackageVersion = strings.Join(pkgVersion, ".")
			fmt.Fprintf(os.Stderr, "The package patch version will be increased to %s.\n",
				seed.Job.PackageVersion)

			// Bump the package minor verion
		} else if  pm {
			pkgVersion := strings.Split(seed.Job.PackageVersion, ".")
			minorVersion, _ := strconv.Atoi(pkgVersion[1])
			pkgVersion[1] = strconv.Itoa(minorVersion + 1)
			pkgVersion[2] = "0"
			seed.Job.PackageVersion = strings.Join(pkgVersion, ".")

			fmt.Fprintf(os.Stderr, "The package version will be increased to %s.\n",
				seed.Job.PackageVersion)

			// Bump the package major version
		} else if P {
			pkgVersion := strings.Split(seed.Job.PackageVersion, ".")
			majorVersion, _ := strconv.Atoi(pkgVersion[0])
			pkgVersion[0] = strconv.Itoa(majorVersion + 1)
			pkgVersion[1] = "0"
			pkgVersion[2] = "0"
			seed.Job.PackageVersion = strings.Join(pkgVersion, ".")

			fmt.Fprintf(os.Stderr, "The major package version will be increased to %s.\n",
				seed.Job.PackageVersion)
		}
		// Bump the job patch version
		if jp {
			jobVersion := strings.Split(seed.Job.JobVersion, ".")
			patchVersion, _ := strconv.Atoi(jobVersion[2])
			jobVersion[2] = strconv.Itoa(patchVersion + 1)
			seed.Job.JobVersion = strings.Join(jobVersion, ".")
			fmt.Fprintf(os.Stderr, "The job patch version will be increased to %s.\n",
				seed.Job.JobVersion)

			// Bump the job minor verion
		} else if jm {
			jobVersion := strings.Split(seed.Job.JobVersion, ".")
			minorVersion, _ := strconv.Atoi(jobVersion[1])
			jobVersion[1] = strconv.Itoa(minorVersion + 1)
			jobVersion[2] = "0"
			seed.Job.JobVersion = strings.Join(jobVersion, ".")
			fmt.Fprintf(os.Stderr, "The minor job version will be increased to %s.\n",
				seed.Job.JobVersion)

			// Bump the job major verion
		} else if J {
			jobVersion := strings.Split(seed.Job.JobVersion, ".")
			majorVersion, _ := strconv.Atoi(jobVersion[0])
			jobVersion[0] = strconv.Itoa(majorVersion + 1)
			jobVersion[1] = "0"
			jobVersion[2] = "0"
			seed.Job.JobVersion = strings.Join(jobVersion, ".")

			fmt.Fprintf(os.Stderr, "The major job version will be increased to %s.\n",
				seed.Job.JobVersion)
		}
		if !J && !jm && !jp && !P && !pm && !pp{
			fmt.Fprintf(os.Stderr, "ERROR: No tag deconfliction method specified. Aborting seed publish.\n")
			fmt.Fprintf(os.Stderr, "Exiting seed...\n")
			return errors.New("Image exists and no tag deconfliction method specified.")
		}

		img = objects.BuildImageName(&seed)
		fmt.Fprintf(os.Stderr, "\nNew image name: %s\n", img)

		// write version back to the seed manifest
		seedJSON, _ := json.Marshal(&seed)
		err = ioutil.WriteFile(seedFileName, seedJSON, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Error occurred writing updated seed version to %s.\n%s\n",
				seedFileName, err.Error())
			return errors.New("Error updating seed version in manifest.")
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
			return errors.New(errs.String())
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
			return errors.New(errs.String())
		}
	}

	// docker tag
	fmt.Fprintf(os.Stderr, "INFO: Performing docker tag %s\n", img)
	errs.Reset()
	tagCmd := exec.Command("docker", "tag", origImg, img)
	tagCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	tagCmd.Stdout = os.Stdout

	// Run docker tag
	if err := tagCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker tag. %s\n",
			err.Error())
		return err
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
		return err
	}

	// Check for errors. Exit if error occurs
	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error pushing image '%s':\n%s\n", img,
			errs.String())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		return errors.New(errs.String())
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
		return err
	}

	// check for errors on stderr
	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error removing image '%s':\n%s\n", img,
			errs.String())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

//PrintPublishUsage prints the seed publish usage information, then exits the program
func PrintPublishUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed publish -in IMAGE_NAME [-r REGISTRY_NAME] [-o ORG_NAME] [-u username] [-p password] [Conflict Options]\n")
	fmt.Fprintf(os.Stderr, "\nAllows for the publish of seed compliant images.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s Docker image name to publish\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific registry to publish the image\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific organization to publish the image\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tUsername to login if needed to publish images (default anonymous).\n",
		constants.ShortUserFlag, constants.UserFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tPassword to login if needed to publish images (default anonymous).\n",
		constants.ShortPassFlag, constants.PassFlag)
	fmt.Fprintf(os.Stderr, "  -%s\t\tOverwrite remote image if publish conflict found\n",
		constants.ForcePublishFlag)


	fmt.Fprintf(os.Stderr, "\nIf the force flag is not set, the following options specify how a publish conflict is handled:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s Specifies the directory containing the seed.manifest.json and dockerfile to rebuild the image.\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Patch version bump of 'packageVersion' in manifest on disk if publish conflict found\n",
		constants.PkgVersionPatch)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Minor version bump of 'packageVersion' in manifest on disk if publish conflict found\n",
		constants.PkgVersionMinor)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Major version bump of 'packageVersion' in manifest on disk if publish conflict found\n",
		constants.PkgVersionMajor)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Patch version bump of 'jobVersion' in manifest on disk if publish conflict found\n",
		constants.JobVersionPatch)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Minor version bump of 'jobVersion' in manifest on disk if publish conflict found\n",
		constants.JobVersionMinor)
	fmt.Fprintf(os.Stderr, "  -%s\t\tForce Major version bump of 'jobVersion' in manifest on disk if publish conflict found\n",
		constants.JobVersionMajor)

	fmt.Fprintf(os.Stderr, "\nExample: seed publish -in example-0.1.3-seed:0.1.3 -r hub.docker.com -o geoint -j path/to/example -jm -P\n")
	fmt.Fprintf(os.Stderr, "Will build a new image example-0.2.0-seed:1.0.0 and publish it to hub.docker.com/geoint\n")
	panic(util.Exit{0})
}
