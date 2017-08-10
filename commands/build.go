package commands

import (

)

//DockerBuild Builds the docker image with the given image tag.
func DockerBuild(imageName string) {

	seedFileName := SeedFileName()

	// Validate seed file
	err := ValidateSeedFile("", seedFileName, constants.SchemaManifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: seed file could not be validated. See errors for details.\n")
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}

	// retrieve seed from seed manifest
	seed := SeedFromManifestFile(seedFileName)

	// Retrieve docker image name
	if imageName == "" {
		imageName = BuildImageName(&seed)
	}

	jobDirectory := buildCmd.Lookup(constants.JobDirectoryFlag).Value.String()

	// Build Docker image
	fmt.Fprintf(os.Stderr, "INFO: Building %s\n", imageName)
	buildArgs := []string{"build", "-t", imageName, jobDirectory}
	if DockerVersionHasLabel() {
		// Set the seed.manifest.json contents as an image label
		label := "com.ngageoint.seed.manifest=" + GetManifestLabel(seedFileName)
		buildArgs = append(buildArgs, "--label", label)
	}
	buildCmd := exec.Command("docker", buildArgs...)
	var errs bytes.Buffer
	buildCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	buildCmd.Stdout = os.Stderr

	// Run docker build
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker build. %s\n",
			err.Error())
	}

	// check for errors on stderr
	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error building image '%s':\n%s\n",
			imageName, errs.String())
		fmt.Fprintf(os.Stderr, "Exiting seed...\n")
		os.Exit(1)
	}
}

//GetManifestLabel returns the seed.manifest.json as LABEL
//  com.ngageoint.seed.manifest contents
func GetManifestLabel(seedFileName string) string {
	// read the seed.manifest.json into a string
	seedbytes, err := ioutil.ReadFile(seedFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Eror reading %s. %s\n", seedFileName,
			err.Error())
		os.Exit(1)
	}
	var seedbuff bytes.Buffer
	json.Compact(&seedbuff, seedbytes)
	seedbytes, err = json.Marshal(seedbuff.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error marshalling seed manifest. %s\n",
			err.Error())
	}

	// Escape forward slashes and dollar signs
	seed := string(seedbytes)
	seed = strings.Replace(seed, "$", "\\$", -1)
	seed = strings.Replace(seed, "/", "\\/", -1)

	return seed
}

//DefineBuildFlags defines the flags for the seed build command
func DefineBuildFlags() {
	// build command flags
	buildCmd = flag.NewFlagSet(constants.BuildCommand, flag.ContinueOnError)
	buildCmd.StringVar(&directory, constants.JobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")
	buildCmd.StringVar(&directory, constants.ShortJobDirectoryFlag, ".",
		"Directory of seed spec and Dockerfile (default is current directory).")

	// Print usage function
	buildCmd.Usage = func() {
		PrintBuildUsage()
	}
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintBuildUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed build [-d JOB_DIRECTORY]\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr,
		"  -%s  -%s\tDirectory containing Seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	os.Exit(0)
}
