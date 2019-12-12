package cliutil

import (
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ngageoint/seed-common/util"
)

//DockerCommandArgsInit returns the initial command and args needed to run docker based on the OS seed CLI is running in.
func DockerCommandArgsInit() ([]string, string) {
	var dockerArgs []string
	var dockerCommand string

	if runtime.GOOS == "windows" { // windows does not recognise sudo as a command.
		dockerCommand = "docker"
	} else if CheckDockerAccess() { // we can call docker no problem, no need to use sudo
		dockerCommand = "docker"
	} else {
		util.PrintUtil("\033[30;43mWARNING: Unable to access docker command. Attempting with sudo. You may be prompted for your password.")
		dockerArgs = []string{"docker"}
		dockerCommand = "sudo"
	}

	return dockerArgs, dockerCommand
}

func CheckDockerAccess() bool {

	var returnValue bool

	cmd := exec.Command("docker", "info")
	errPipe, _ := cmd.StderrPipe()
	cmd.Start()

	slurperr, _ := ioutil.ReadAll(errPipe)
	er := string(slurperr)
	if er != "" {
		if strings.Contains(er, "Cannot connect to the Docker daemon. Is the docker daemon running on this host?") ||
			strings.Contains(er, "dial unix /var/run/docker.sock: connect: permission denied") {
			returnValue = false
		}
	} else {
		returnValue = true
	}
	return returnValue
}
