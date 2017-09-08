package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	RegistryFactory "github.com/ngageoint/seed-cli/registry"
	"github.com/ngageoint/seed-cli/util"
)

//DockerSearch executes the seed search command
func DockerSearch(url, org, filter, username, password string) ([]string, error) {
	_ = filter //TODO: add filter

	if url == "" {
		url = constants.DefaultRegistry
	}

	if org == "" {
		org = constants.DefaultOrg
	}

	registry, err := RegistryFactory.CreateRegistry(url, username, password)
	if registry != nil && err == nil {
		images, err := registry.Images(org)
		return images, err
	}

	err = errors.New(checkError(err, url, username, password))

	return nil, err
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

func checkError(err error, url, username, password string) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	humanError := ""

	if strings.Contains(errStr, "status=401") {
		if username == "" || password == "" {
			humanError = "The specified registry requires a login.  Please try again with a username (-u) and password (-p)."
		} else {
			humanError = "Incorrect username/password."
		}
	} else if strings.Contains(errStr, "status=404") {
		humanError = "Connected to registry but received a 404 error. Please check the url and try again."
	} else {
		humanError = "Could not connect to the specified registry. Please check the url and try again."
	}
	return humanError
}