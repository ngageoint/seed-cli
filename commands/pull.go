package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/util"
)

//Dockerpull pulls specified image from remote repository (default docker.io)
func DockerPull(image, registry, org, username, password string) error {
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

	if registry == "" {
		registry = constants.DefaultRegistry
	}

	if org == "" {
		org = constants.DefaultOrg
	}

	remoteImage := fmt.Sprintf("%s/%s/%s", registry, org, image)

	var errs, out bytes.Buffer
	// pull image
	pullArgs := []string{"pull", remoteImage}
	pullCmd := exec.Command("docker", pullArgs...)
	pullCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	pullCmd.Stdout = &out

	err := pullCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker pull.\n%s\n",
			err.Error())
		return err
	}

	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n",
			errs.String())
		return errors.New(errs.String())
	}

	// tag image
	tagArgs := []string{"tag", remoteImage, image}
	tagCmd := exec.Command("docker", tagArgs...)
	tagCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	tagCmd.Stdout = &out

	err = tagCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker tag.\n%s\n",
			err.Error())
		return err
	}

	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n",
			errs.String())
		return errors.New(errs.String())
	}

	return nil
}

//PrintListUsage prints the seed list usage information, then exits the program
func PrintPullUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed pull -in IMAGE_NAME [-r REGISTRY_NAME] [-o ORGANIZATION_NAME] [-u Username] [-p password]\n")
	fmt.Fprintf(os.Stderr, "\nPulls seed image from remote repository.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s Docker image name to pull\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific registry (default is index.docker.io).\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific organization (default is no organization).\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tUsername to login to remote registry (default anonymous).\n",
		constants.ShortUserFlag, constants.UserFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tPassword to login to remote registry (default anonymous).\n",
		constants.ShortPassFlag, constants.PassFlag)
	panic(util.Exit{0})
}
