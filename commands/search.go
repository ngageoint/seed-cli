package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/dockerHubRegistry"
	"github.com/ngageoint/seed-cli/util"

	"github.com/heroku/docker-registry-client/registry"
)

//DockerSearch executes the seed search command
func DockerSearch(url, org, filter, username, password string) ([]string, error) {
	_ = filter //TODO: add filter

	if url == "" {
		url = constants.DefaultRegistry
	}

	httpFallback := ""
	if !strings.HasPrefix(url, "http") {
		httpFallback = "http://" + url
		url = "https://" + url
	}

	if org == "" {
		org = constants.DefaultOrg
	}

	dockerHub := false
	if strings.Contains(url, "hub.docker.com") || strings.Contains(url, "index.docker.io") || strings.Contains(url, "registry-1.docker.io") {
		url = "https://hub.docker.com"
		dockerHub = true
	}

	var images []string
	var err error
	if dockerHub { //_catalog is disabled on docker hub, cannot get list of images so get all of the images for the org (if specified)
		hub, err := dockerHubRegistry.New(url)
		if err != nil {
			hub, err = dockerHubRegistry.New(httpFallback)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return nil, err
			}
		}
		images, err = hub.UserRepositories(org)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return nil, err
		}
	} else {
		hub, err := registry.New(url, username, password)
		if err != nil {
			hub, err = registry.New(httpFallback, username, password)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				return nil, err
			}
		}
		repositories, err := hub.Repositories()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return nil, err
		}
		for _, repo := range repositories {
			tags, err := hub.Tags(repo)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				continue
			}
			for _, tag := range tags {
				images = append(images, repo + ":" + tag)
			}
		}

	}
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return nil, err
	}

	var stringImages []string
	for _, image := range images {
		if strings.Contains(image, "-seed") {
			stringImages = append(stringImages, image)
		}
	}

	return stringImages, nil
}

//PrintSearchUsage prints the seed search usage information, then exits the program
func PrintSearchUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed search [-r REGISTRY_NAME] [-o ORGANIZATION_NAME] [-f FILTER] [-u Username] [-p password]\n")
	fmt.Fprintf(os.Stderr, "\nAllows for discovery of seed compliant images hosted within a Docker registry.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific registry to search (default is index.docker.io).\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific organization to filter (default is no filter).\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a filter to apply (default is no filter).\n",
		constants.ShortFilterFlag, constants.FilterFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tUsername to login to remote registry (default is anonymous).\n",
		constants.ShortUserFlag, constants.UserFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tPassword to login to remote registry (default is anonymous).\n",
		constants.ShortPassFlag, constants.PassFlag)
	panic(util.Exit{0})
}
