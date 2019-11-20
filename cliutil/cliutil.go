package cliutil

import "runtime"

//DockerCommandArgsInit returns the initial command and args needed to run docker based on the OS seed CLI is running in.
func DockerCommandArgsInit() ([]string, string) {
	var dockerArgs []string
	var dockerCommand string

	if runtime.GOOS == "windows" { // windows does not recognise sudo as a command.
		dockerCommand = "docker"
	} else {
		dockerArgs = []string{"docker"}
		dockerCommand = "sudo"
	}

	return dockerArgs, dockerCommand
}
