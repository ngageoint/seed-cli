package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/dockerHubRegistry"

	"github.com/heroku/docker-registry-client/registry"
)
//DockerSearch executes the seed search command
func DockerSearch(url, org, filter, username, password string) {
	_ = filter //TODO: add filter

	if url == "" {
		url = constants.DefaultRegistry
	}
	
	if org == "" {
		org = constants.DefaultOrg
	}
	
	dockerHub := false
	if strings.Contains(url, "hub.docker.com") || strings.Contains(url, "index.docker.io") || strings.Contains(url, "registry-1.docker.io") {
		url = "https://hub.docker.com"
		dockerHub = true
	}
	

	var repositories []string
	var err error
	if dockerHub { //_catalog is disabled on docker hub, cannot get list of images so get all of the images for the org (if specified)
		hub, err := dockerHubRegistry.New(url)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		repositories, err = hub.UserRepositories(org)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		hub, err := registry.New(url, username, password)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		repositories, err = hub.Repositories()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	
	for _, repo := range repositories {
		if strings.HasSuffix(repo, "-seed") {
			fmt.Println(repo)
		}
	}
}

//PrintSearchUsage prints the seed search usage information, then exits the program
func PrintSearchUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed search [-r REGISTRY_NAME] [-o ORGANIZATION_NAME] [-f FILTER] \n")
	fmt.Fprintf(os.Stderr, "\nAllows for discovery of seed compliant images hosted within a Docker registry.\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific registry to search (default is index.docker.io).\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a specific organization to filter (default is no filter).\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies a filter to apply (default is no filter).\n",
		constants.ShortFilterFlag, constants.FilterFlag)
	os.Exit(0)
}
