package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

//DockerList lists all seed compliant images (ending with -seed) on the local
//	system
func DockerList() error {
	dCmd := exec.Command("docker", "images")
	gCmd := exec.Command("grep", "seed")
	var dErr bytes.Buffer
	dCmd.Stderr = &dErr
	dOut, err := dCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error attaching to std output pipe. %s\n",
			err.Error())
	}

	dCmd.Start()
	if string(dErr.Bytes()) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n",
			string(dErr.Bytes()))
	}

	gCmd.Stdin = dOut
	var gErr bytes.Buffer
	gCmd.Stderr = &gErr

	o, err := gCmd.Output()
	fmt.Fprintf(os.Stderr, string(gErr.Bytes()))
	if string(gErr.Bytes()) != "" {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", string(gErr.Bytes()))
	}
	if err != nil && err.Error() != "exit status 1" {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing seed list: %s\n", err.Error())
	}
	if string(o) == "" {
		fmt.Fprintf(os.Stderr, "No Seed Images found!\n")
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", string(o))
	}

	return err
}

//PrintListUsage prints the seed list usage information, then exits the program
func PrintListUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed list\n")
	fmt.Fprintf(os.Stderr, "\nLists all Seed compliant docker images residing on the local system.\n")
	os.Exit(0)
}
