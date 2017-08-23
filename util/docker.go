package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//CheckSudo Checks error for telltale sign seed command should be run as sudo
func CheckSudo() {
	cmd := exec.Command("docker", "info")

	// attach stderr pipe
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: Error attaching to version command stderr. %s\n", err.Error())
	}

	// Run docker build
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker version. %s\n",
			err.Error())
	}

	slurperr, _ := ioutil.ReadAll(errPipe)
	er := string(slurperr)
	if er != "" {
		if strings.Contains(er, "Cannot connect to the Docker daemon. Is the docker daemon running on this host?") ||
			strings.Contains(er, "dial unix /var/run/docker.sock: connect: permission denied") {
			fmt.Fprintf(os.Stderr, "Elevated permissions are required by seed to run Docker. Try running the seed command again as sudo.\n")
			panic(Exit{1})
		}
	}
}

//DockerVersionHasLabel returns if the docker version is greater than 1.11.1
func DockerVersionHasLabel() bool {
	return DockerVersionGreaterThan(1, 11, 1)
}

//DockerVersionHasLabel returns if the docker version is greater than 1.13.0
func DockerVersionHasReferenceFilter() bool {
	return DockerVersionGreaterThan(1, 13, 0)
}

//DockerVersionGreaterThan returns if the docker version is greater than the specified version
func DockerVersionGreaterThan(major, minor, patch int) bool {
	cmd := exec.Command("docker", "version", "-f", "{{.Client.Version}}")

	// Attach stdout pipe
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"ERROR: Error attaching to version command stdout. %s\n", err.Error())
	}

	// Run docker version
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker version. %s\n",
			err.Error())
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		version := strings.Split(string(slurp), ".")

		// check each part of version. Return false if 1st < 1, 2nd < 11, 3rd < 1
		if len(version) > 1 {
			v1, _ := strconv.Atoi(version[0])
			v2, _ := strconv.Atoi(version[1])

			// check for minimum of 1.11.1
			if v1 == major {
				if v2 > minor {
					return true
				} else if v2 == minor && len(version) == 3 {
					v3, _ := strconv.Atoi(version[2])
					if v3 >= patch {
						return true
					}
				}
			} else if v1 > major {
				return true
			}

			return false
		}
	}

	return false
}

//ImageExists returns true if a local image already exists, false otherwise
func ImageExists(imageName string) (bool, error) {
	// Test if image has been built; Rebuild if not
	imgsArgs := []string{"images", "-q", imageName}
	imgOut, err := exec.Command("docker", imgsArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker %v\n", imgsArgs)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return false, err
	} else if string(imgOut) == "" {
		fmt.Fprintf(os.Stderr, "INFO: No docker image found for image name %s. Building image now...\n",
			imageName)
		return false, nil
	}
	return true, nil
}

//ImageCpuUsage displays CPU usage of image
func ImageCpuUsage(imageName string) {

}

//ImageMemoryUsage displays memory usage of image
func ImageMemoryUsage(imageName string) {

}

func Login(username, password, registry string) error {
	var errs, out bytes.Buffer
	args := []string{"login", "-u", username, "-p", password, registry}
	cmd := exec.Command("docker", args...)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = &out

	// run images
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Error executing docker images.\n%s\n",
			err.Error())
		return err
	}

	if errs.String() != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Error reading stderr %s\n",
			errs.String())
		return errors.New(errs.String())
	}

	fmt.Fprintf(os.Stderr, "%s", out.String())
	return nil
}
