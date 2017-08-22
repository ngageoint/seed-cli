package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ngageoint/seed-cli/util"
)

//DockerList - Simplified version of dockerlist - relies on name filter of
//  docker images command to search for images ending in '-seed'
func DockerList() (string, error) {
	var errs, out bytes.Buffer
	var cmd *exec.Cmd
	reference := util.DockerVersionHasReferenceFilter()
	if reference {
		cmd = exec.Command("docker", "images", "--filter=reference=*-seed*")
	} else {
		dCmd := exec.Command("docker", "images")
		cmd = exec.Command("grep", "-seed")
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

		cmd.Stdin = dOut
	}

	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = &out

	// run images
	err := cmd.Run()
	if reference && err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker images.\n%s\n",
			err.Error())
		return "", err
	}

	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n",
			errs.String())
		return "", errors.New(errs.String())
	}

	if !strings.Contains(out.String(), "seed") {
		fmt.Fprintf(os.Stderr, "No seed images found!\n")
		return "", nil
	}
	fmt.Fprintf(os.Stderr, "%s", out.String())
	return out.String(), nil
}

//PrintListUsage prints the seed list usage information, then exits the program
func PrintListUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed list\n")
	fmt.Fprintf(os.Stderr, "\nLists all Seed compliant docker images residing on the local system.\n")
	panic(util.Exit{0})
}
