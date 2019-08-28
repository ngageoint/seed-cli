package commands

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/ngageoint/seed-cli/constants"
	common_const "github.com/ngageoint/seed-common/constants"
	"github.com/ngageoint/seed-common/objects"
	RegistryFactory "github.com/ngageoint/seed-common/registry"
	"github.com/ngageoint/seed-common/util"
)

//DockerUnpublish executes the seed unpublish command
func DockerUnpublish(image, manifest, registry, org, username, password string) error {

	if image == "" {
		util.PrintUtil("INFO: Image name not specified. Attempting to use manifest: %v\n", manifest)
		temp, err := objects.GetImageNameFromManifest(manifest, "")
		if err != nil {
			return err
		}
		image = temp
	}

	if image == "" {
		err := errors.New("ERROR: No image specified. Unable to determine image name.\n")
		util.PrintUtil("%s\n", err.Error())
		return err
	}

	util.PrintUtil("INFO: Attempting to remove image %s from registry %s\n", image, registry)

	temp := strings.Split(image, ":")
	if len(temp) != 2 {
		err := errors.New("ERROR: Invalid seed name: %s. Unable to split into name/tag pair\n")
		return err
	}
	repoName := temp[0]
	repoTag := temp[1]

	if org != "" {
		repoName = org + "/" + repoName
	}

	if username != "" {
		//set config dir so we don't stomp on other users' logins with sudo
		configDir := common_const.DockerConfigDir + time.Now().Format(time.RFC3339)
		os.Setenv(common_const.DockerConfigKey, configDir)
		defer util.RemoveAllFiles(configDir)
		defer os.Unsetenv(common_const.DockerConfigKey)

		err := util.Login(registry, username, password)
		if err != nil {
			util.PrintUtil(err.Error())
		}
	}

	reg, err := RegistryFactory.CreateRegistry(registry, org, username, password)
	if err != nil {
		err = errors.New(checkError(err, registry, username, password))
		return err
	}
	if reg == nil {
		err = errors.New("Unknown error connecting to registry")
		return err
	}

	if reg != nil && err == nil {
		err = reg.RemoveImage(repoName, repoTag)
	}

	if err == nil {
		if org != "" {
			util.PrintUtil("%v removed from %v/%v\n", image, registry, org)
		} else {
			util.PrintUtil("%v removed from %v\n", image, registry)
		}
	}

	return err
}

//PrintPublishUsage prints the seed publish usage information, then exits the program
func PrintUnpublishUsage() {
	util.PrintUtil("\nUsage:\tseed unpublish [-in IMAGE_NAME] [-M MANIFEST] [-v VERSION] [-r REGISTRY_NAME] [-O ORG_NAME] [-u username] [-p password] [Conflict Options]\n")
	util.PrintUtil("\nAllows for the removal of seed compliant images from a registry.\n")
	util.PrintUtil("\nOptions:\n")
	util.PrintUtil("  -%s -%s Docker image name to publish\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	util.PrintUtil("  -%s  -%s\t Manifest file to use if an image name is not specified (default is seed.manifest.json within the current directory).\n",
		constants.ShortManifestFlag, constants.ManifestFlag)
	util.PrintUtil("  -%s  -%s\t Specifies a specific registry to publish the image\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	util.PrintUtil("  -%s  -%s\t Specifies a specific organization to publish the image\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	util.PrintUtil("  -%s  -%s\t Username to login if needed to publish images (default anonymous).\n",
		constants.ShortUserFlag, constants.UserFlag)
	util.PrintUtil("  -%s  -%s\t Password to login if needed to publish images (default anonymous).\n",
		constants.ShortPassFlag, constants.PassFlag)

	util.PrintUtil("\nExample: \tseed unpublish -in example-0.1.3-seed:1.0.0 -r my.registry.address\n")
	util.PrintUtil("\nThis will remove the tag 1.0.0 for the repository example-0.1.3-seed from the registry my.registry.address\n")
	return
}
