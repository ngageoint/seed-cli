package commands

//DockerSearch executes the seed search command
func DockerSearch() {

	url := searchCmd.Lookup(constants.RegistryFlag).Value.String()
	org := searchCmd.Lookup(constants.OrgFlag).Value.String()
	//filter := searchCmd.Lookup(constants.FilterFlag).Value.String()
	username := searchCmd.Lookup(constants.UserFlag).Value.String()
	password := searchCmd.Lookup(constants.PassFlag).Value.String()
	
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

//DefineSearchFlags defines the flags for the seed search command
func DefineSearchFlags() {
	// Search command
	searchCmd = flag.NewFlagSet(constants.SearchCommand, flag.ExitOnError)
	var registry string
	searchCmd.StringVar(&registry, constants.RegistryFlag, "", "Specifies registry to search (default is index.docker.io).")
	searchCmd.StringVar(&registry, constants.ShortRegistryFlag, "", "Specifies registry to search (default is index.docker.io).")

	var org string
	searchCmd.StringVar(&org, constants.OrgFlag, "", "Specifies organization to filter (default is no filter, search all orgs).")
	searchCmd.StringVar(&org, constants.ShortOrgFlag, "", "Specifies organization to filter (default is no filter, search all orgs).")
	
	var filter string
	searchCmd.StringVar(&filter, constants.FilterFlag, "", "Specifies filter to apply (default is no filter).")
	searchCmd.StringVar(&filter, constants.ShortFilterFlag, "", "Specifies filter to apply (default is no filter).")
	
	var user string
	searchCmd.StringVar(&user, constants.UserFlag, "", "Specifies filter to apply (default is no filter).")
	searchCmd.StringVar(&user, constants.ShortUserFlag, "", "Specifies filter to apply (default is no filter).")
	
	var password string
	searchCmd.StringVar(&password, constants.PassFlag, "", "Specifies filter to apply (default is no filter).")
	searchCmd.StringVar(&password, constants.ShortPassFlag, "", "Specifies filter to apply (default is no filter).")
	
	searchCmd.Usage = func() {
		PrintSearchUsage()
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
