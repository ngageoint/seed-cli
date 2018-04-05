package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	// "github.com/howeyc/fsnotify"
	"github.com/ngageoint/seed-common/constants"
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
	"github.com/xeipuuv/gojsonschema"
)

type ServiceIO struct {
	LocalInputs []string
	LocalDir string
	MachineInputs []string
	MachineDir string
}

type ServiceOutput struct {
	Localoutdir string
	Volumeoutdir string
}

type Volume struct {
	Name string
	LocalDir string
	HostMachineDir string
	DataDir string
	HostMachine string
}

type Mount struct {
	Localpath string
	Machinepath string
	VolumeName string
	Mountargs []string
}

type Service struct {
	seed objects.Seed
	name string
	id string
	outdir string
}

// CHANGE THE target= TO POINT TO /data AND ONLY KEEP /VOLUME AND AFTER

//DockerBatchService runs an image as a service on a given cluster
func DockerBatchService(manager, batchDir, batchFile, imageName, outputDir, 
	metadataSchema, registry string, settings, mounts []string) error {
	util.CheckSudo()

	if imageName == "" {
		return errors.New("ERROR: No input image specified")
	}

	if exists, err := util.ImageExists(imageName); !exists {
		return err
	}

	if batchDir == "" {
		batchDir = "."
	}

	batchDir = util.GetFullPath(batchDir, "")

	// Extract the seed object
	seed := objects.SeedFromImageLabel(imageName)

	// Create and mount the output directory to the node
	// Inputs will be copied here
	// mountdir == path on manager
	// outdir === path on local machine
	// localOutDir := <outputDir>/volume
	// serviceOutput := getServiceOutputDir(outputDir, imageName)
	localVolumeDir := getServiceOutputDir(outputDir, imageName)
	util.PrintUtil("localVolumeDir after getServiceOutput: %s", localVolumeDir)
	home := util.MachineHome(manager)
	if home == "" {
		erString := fmt.Sprintf("ERROR: Error retrieving %s home directory\n", manager) 
		util.PrintUtil(erString)
		return errors.New(erString)
	}
	mountedVolumeSrcDir := filepath.Join(home, localVolumeDir) // /home/<USER>/$localVolumeDir
	err := util.Mount(manager, mountedVolumeSrcDir, localVolumeDir)

	// Creates inputs directory(s) under the localVolumeDir and copies the input files
	// inputs contains both the localVolumeDir path and the mountedVolumeDir path
	var inputs []ServiceIO
	if batchFile != "" {
		inputs, err = ProcessServiceBatchFile(seed, batchFile, localVolumeDir, mountedVolumeSrcDir)
		if err != nil {
			util.PrintUtil("ERROR: Error processing batch file: %s\n", err.Error())
			return err
		}
	} else {
		inputs, err = ProcessServiceDirectory(seed, batchDir, localVolumeDir, mountedVolumeSrcDir)
		if err != nil {
			util.PrintUtil("ERROR: Error processing batch directory: %s\n", err.Error())
			return err
		}
	}

	util.PrintUtil("Inputs:\n%v\n", inputs)
	////// Need to get the data from the local machine to the manager node //////

	// Create named volume that will be shared between service nodes
	volumeName := strings.Replace(strings.Replace(imageName, ".", "", -1), ":", "-", -1)+"-vol"
	mountedVolumeDataDir, err := util.CreateVolume(volumeName, manager)
	if err != nil {
		return err
	}
	volume := Volume{volumeName, localVolumeDir, mountedVolumeSrcDir, mountedVolumeDataDir, manager}

	// Create helper container to copy data to the volume
	_, err = exec.Command("docker-machine", "ssh", manager, 
		"docker", "create", "-v", volume.Name+":/volume", "--name", "helper", "busybox", "true").Output()
	if err != nil {
		util.PrintUtil("ERROR: Error creating helper container...%s\n", err.Error())
		return err
	}

	// Copy data from host machine's mounted directory to container
	_, err = exec.Command("docker-machine", "ssh", volume.HostMachine, "docker", "cp", volume.HostMachineDir, "helper:/volume").Output()
	if err != nil {
		util.PrintUtil("ERROR: Error copying input data to volume....%s\n", err.Error())
		return err
	}
	// Clean up helper container
	_, err = exec.Command("docker-machine", "ssh", volume.HostMachine, "docker", "rm", "helper").Output()
	if err != nil {
		util.PrintUtil("ERROR: Error removing helper container....%s\n", err.Error())
		return err
	}

	////// Move the docker image to the cluster //////
	// Exports image and SCPs it to the cluster
	registryImageName, err := RegistrySetup(manager, imageName, registry)
	if err != nil {
		util.PrintUtil("ERROR: Error pushing image to cluster registry.\n")
		util.PrintUtil(err.Error())
		return err
	}

	////// Service creation/running //////
	// Create and start a service for each input
	var wg sync.WaitGroup
	var services []Service
	var outDirs []string
	serviceName := strings.Replace(strings.Replace(imageName, ":", "-",-1), ".", "",-1)
	out := "Results: \n"
	for r, in := range inputs {
		serviceRunName := serviceName+"-"+strconv.Itoa(r)
		// exitCode, service, outputSize, err := DockerService(manager, serviceRunName, registryImageName,
		// 	imageName, localVolumeDir, mountedVolumeSrcDir, volumeName, metadataSchema, in.LocalInputs, 
		// 	in.MachineInputs, settings, mounts, false, r)
		exitCode, service, outputSize, err := DockerService(manager, serviceRunName, registryImageName,
			imageName, metadataSchema, volume, in.LocalInputs, in.MachineInputs, settings, mounts, false, r)
		services = append(services, service)
		outDirs = append(outDirs, service.outdir)

		// Check output of service. Need to wait until service has completed
		if service.seed.Job.Interface.Outputs.Files != nil ||
			service.seed.Job.Interface.Outputs.JSON != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				checkService(service, outputSize, metadataSchema, manager)
			}()
			wg.Wait()
		}

		//trim inputs to print only the key values and filenames
		truncatedInputs := []string{}
		for _, i := range in.LocalInputs {
			begin := strings.Index(i, "=") + 1
			end := strings.LastIndex(i, "/")
			truncatedInputs = append(truncatedInputs, i[0:begin]+"..."+i[end:])
		}

		//trim path to specified (or generated) batch output directory
		truncatedOut := "..." + strings.Replace(in.LocalDir, localVolumeDir, filepath.Base(localVolumeDir), 1)

		if err != nil {
			out += fmt.Sprintf("FAIL: Input = %v \t ExitCode = %d \t Error = %s \n", truncatedInputs, exitCode, err.Error())
		} else {
			out += fmt.Sprintf("PASS: Input = %v \t ExitCode = %d \t Output = %s \n", truncatedInputs, exitCode, truncatedOut)
		}
	}
	util.InitPrinter(util.PrintErr)
	util.PrintUtil("%v", out)

	// Unmount the volume directory
	// err = util.Unmount(manager, mountedVolumeSrcDir, localVolumeDir)
	// if err != nil {
	// 	return err
	// }
	// remove the local mounted directory
	// err = util.Remove(localVolumeDir)
	// if err != nil {
	// 	return err
	// }

	return err
}

//checkService Waits for specified service to complete then validates output
func checkService(service Service, diskLimit float64, metadataSchema, manager string ) {
	util.PrintUtil("Waiting for %s to complete...\n", service.name)
	complete := false

	for !complete {
		args := []string{"ssh", manager, "docker", "service", "ps", service.name, "--format", "\"{{.CurrentState}}\""}
		cmd, err := exec.Command("docker-machine", args...).Output()
		if err != nil {
			util.PrintUtil("ERROR: Error checking service status for %s\n", service.name)
			break
		} else if string(cmd) != "" {
			if strings.Contains(string(cmd), "Complete") {
				util.PrintUtil("INFO: Service %s completed!\n", service.name)
				complete = true
				break
			} else if strings.Contains(string(cmd), "Rejected") {
				util.PrintUtil("ERROR: Service %s has been rejected! %s\n", service.name, string(cmd))
				break
			} else if strings.Contains(string(cmd), "Shutdown") {
				util.PrintUtil("ERROR: Service %s has been shutdown! %s\n", service.name, string(cmd))
				break
			} else if strings.Contains(string(cmd), "Failed") {
				util.PrintUtil("ERROR: Service %s has failed! %s\n", service.name, string(cmd))
				break
			}
		}
	}

	// Copy service output from mounted volume to output directory on local machine
	if complete {
		outDir := service.outdir
		machineDir := outDir
		outdir := filepath.Join(path.Dir(path.Dir(outDir)), filepath.Base(machineDir))
		util.PrintUtil("Time to copy output!\n%s -> %s", machineDir, outdir)
		err := CopyServiceOutput(machineDir, outdir)
		if err != nil {
			util.PrintUtil("ERROR: Error copying service %s output: %s\n", service.name, err.Error())
		}
		CheckServiceOutput(&service.seed, outdir, metadataSchema, diskLimit)

	// Inform user service did not complete and try to show the logs
	} else {
		util.PrintUtil("ERROR: Service did not complete. Output will not be verified.\n")
		args := []string{"ssh", manager, "docker", "service", "logs", service.name}
		logCmd := exec.Command("docker-machine", args...)
		var cmd bytes.Buffer
		logCmd.Stderr = io.MultiWriter(&cmd)
		logCmd.Stdout = io.MultiWriter(&cmd)

		err := logCmd.Run()
		if err != nil {
			util.PrintUtil("ERROR: Error retrieving service %s logs: %s\n", service.name, err.Error())
		} else if cmd.String() != "" {
			util.PrintUtil("INFO: %s logs:\n%s", service.name, cmd.String())
		}
	}

	// Remove service
	// args := []string{"ssh", manager, "docker", "service", "rm", service.name}
	// outs, err := exec.Command("docker-machine", args...).Output()
	// if err != nil {
	// 	util.PrintUtil("ERROR: Error removing completed service: %s\n%s\n", string(outs), err.Error())
	// }

}

//DockerService creates and runs the image as a service on the given swarm
// func DockerService(manager, serviceName, registryImageName, imageName, localdir, mountdir, volumeName, metadataSchema string,
// 	localinputs, machineInputs, settings, mounts []string, quiet bool, run int) (int, Service, float64, error) {
//DockerService creates and runs the image as a service on the given swarm
func DockerService(manager, serviceName, registryImageName, imageName, metadataSchema string, volume Volume,
	localinputs, machineInputs, settings, mounts []string, quiet bool, run int) (int, Service, float64, error) {
		
	var service Service
	util.InitPrinter(util.PrintErr)
	if quiet {
		util.InitPrinter(util.Quiet)
	}

	if imageName == "" {
		return 0, service, 0, errors.New("ERROR: No input image specified")
	}

	if exists, err := util.MachineImageExists(manager, imageName); !exists {
		return 0, service, 0, err
	}

	seed := objects.SeedFromImageLabel(imageName)

	// start building the 'docker service create...' command
	dockerArgs := []string{"ssh", manager, "docker", "service", "create"}


	dockerArgs = append(dockerArgs, "--name")
	dockerArgs = append(dockerArgs, serviceName)
	dockerArgs = append(dockerArgs, "--with-registry-auth")
	dockerArgs = append(dockerArgs, "--detach")
	dockerArgs = append(dockerArgs, "--restart-condition")
	dockerArgs = append(dockerArgs, "none")

	// We only want one of each service so it doesn't continuously
	dockerArgs = append(dockerArgs, "--replicas")
	dockerArgs = append(dockerArgs, "1")
	// dockerArgs = append(dockerArgs, "--mode")
	// dockerArgs = append(dockerArgs, "global")

	// var mountPaths []string
	var Mounts []Mount
	var envArgs []string
	var resourceArgs []string
	var inputSize float64
	var outputSize float64

	// mount the batch volume directory
	// Mounts = append(Mounts, Mount{localdir, mountdir, []string{"--mount", "type=bind,src="+mountdir+",destination="+mountdir}})
	Mounts = append(Mounts, Mount{volume.LocalDir, volume.HostMachineDir, volume.Name, 
		[]string{"--mount", "source="+volume.Name+",target=/data"}}) //+volume.HostMachineDir}})

	// expand INPUT_FILEs to specified Inputs files
	if seed.Job.Interface.Inputs.Files != nil {
		inmounts, size, temp, err := DefineServiceInputs(&seed, localinputs, machineInputs)
		for _, v := range temp {
			defer util.RemoveAllFiles(v)
		}
		if err != nil {
			util.PrintUtil("ERROR: Error occurred processing inputs arguments.\n%s", err.Error())
			util.PrintUtil("Exiting seed...\n")
			panic(util.Exit{1})

		} else if inmounts != nil {
			// mountsArgs = append(mountsArgs, inMounts...)
			inputSize = size
		}
	}

	if len(seed.Job.Resources.Scalar) > 0 {
		inResources, diskSize, err := DefineServiceResources(&seed, inputSize)
		if err != nil {
			util.PrintUtil("ERROR: Error occurred processing resources\n%s", err.Error())
			util.PrintUtil("Exiting seed...\n")
			panic(util.Exit{1})
		} else if inResources != nil {
			resourceArgs = append(resourceArgs, inResources...)
			outputSize = diskSize
		}
	}

	// mount the JOB_OUTPUT_DIR (outDir flag)
	var outDir string
	if strings.Contains(seed.Job.Interface.Command, "OUTPUT_DIR") {
		// Create the output directory
		localout, machineOutDir := SetServiceOutputDir(imageName, serviceName, &seed, volume.LocalDir, volume.HostMachineDir)
		if outDir != "" {
			Mounts = append(Mounts, Mount{localout, machineOutDir, volume.Name, []string{"--mount", 
				"source="+volume.Name+",destination="+localout}})
		}
		outDir = localout
	}

	// Settings
	if seed.Job.Interface.Settings != nil {
		inSettings, err := DefineServiceSettings(&seed, settings)
		if err != nil {
			util.PrintUtil("ERROR: Error occurred processing settings arguments.\n%s", err.Error())
			util.PrintUtil("Exiting seed...\n")
			panic(util.Exit{1})
		} else if inSettings != nil {
			envArgs = append(envArgs, inSettings...)
		}
	}

	// Additional Mounts defined in seed.json
	if seed.Job.Interface.Mounts != nil {
		inMounts, err := DefineServiceMounts(&seed, volume, mounts)
		if err != nil {
			util.PrintUtil("ERROR: Error occurred processing mount arguments.\n%s", err.Error())
			util.PrintUtil("Exiting seed...\n")
			panic(util.Exit{1})
		} else if inMounts != nil {
			Mounts = append(Mounts, inMounts...)
		}
	}

	// Build Docker command arguments:
	// 		service
	//		env injection
	// 		all mounts
	//		image name
	//		Job.Interface.Command

	// Add the mounts args
	for _, m := range Mounts {
		dockerArgs = append(dockerArgs, m.Mountargs...)
	}
	dockerArgs = append(dockerArgs, envArgs...)
	dockerArgs = append(dockerArgs, resourceArgs...)
	dockerArgs = append(dockerArgs, registryImageName)

	// Parse out command arguments from seed.Job.Interface.Command
	args := strings.Split(seed.Job.Interface.Command, " ")
	dockerArgs = append(dockerArgs, args...)

	// Run
	var cmd bytes.Buffer
	cmd.WriteString("docker-machine ")
	for _, s := range dockerArgs {
		cmd.WriteString(s + " ")
	}
	util.PrintUtil("INFO: Starting Docker service:\n%s\n", cmd.String())

	// Run Docker command and capture output
	dockerService := exec.Command("docker-machine", dockerArgs...)
	var errs bytes.Buffer
	var serviceid bytes.Buffer
	if !quiet {
		dockerService.Stderr = io.MultiWriter(&errs)
		dockerService.Stdout = io.MultiWriter(&serviceid)
	}

	service = Service{seed, serviceName, serviceid.String(), outDir}
	
	// Run docker run
	runTime := time.Now()
	err := dockerService.Run()
	util.TimeTrack(runTime, "INFO: "+imageName+" run")
	exitCode := 0
	if err != nil {
		util.PrintUtil("Error running service...\n%s\n", err.Error())
		exitError, ok := err.(*exec.ExitError)
		util.PrintUtil("ERROR bytes %s\n", string(exitError.Stderr))
		if ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
			util.PrintUtil("Exited with error code %v\n", exitCode)
			match := false
			for _, e := range seed.Job.Errors {
				if e.Code == exitCode {
					util.PrintUtil("Title: \t %s\n", e.Title)
					util.PrintUtil("Description: \t %s\n", e.Description)
					util.PrintUtil("Category: \t %s \n \n", e.Category)
					match = true
					util.PrintUtil("Exiting seed...\n")
					return exitCode, service, outputSize, err
				}
			}
			if !match {
				util.PrintUtil("No matching error code found in Seed manifest\n")
			}
		} else {
			util.PrintUtil("ERROR: error executing docker run. %s\n",
				err.Error())
		}
	}
	
	if errs.String() != "" {
		util.PrintUtil("ERROR: Error running service '%s':\n%s\n",
			imageName, errs.String())
		util.PrintUtil("Exiting seed...\n")
		return exitCode, service, outputSize, errors.New(errs.String())
	}

	// // Validate output against pattern
	// if seed.Job.Interface.Outputs.Files != nil ||
	// 	seed.Job.Interface.Outputs.JSON != nil {

		// go CheckService(service, outputSize, metadataSchema, manager)
		
		// // Copy output from volume directory to output directory so we don't lose it
		// machineDir := outDir
		// outdir := filepath.Join(path.Dir(path.Dir(outDir)), filepath.Base(machineDir))
		// util.PrintUtil("Listing files in %s\n", machineDir)
		// // err = CopyServiceOutput(machineDir, outDir)		
		// files, _ := ioutil.ReadDir(machineDir)
		// for _, file := range files {
		// 	util.PrintUtil("\t%s\n",file.Name())
		// }
		// lsDir(machineDir)
		

		// util.PrintUtil("Copying %s to %s\n", machineDir, outdir)
		// cpCmd, err := exec.Command("cp", "-R", machineDir, outdir).Output()
		// if err != nil {
		// 	util.PrintUtil("ERROR copying output: %s\n", err.Error())
		// }
		// if string(cpCmd) != "" {
		// 	util.PrintUtil("cp output: %s\n", string(cpCmd))
		// }

		// CheckServiceOutput(&seed, outDir, metadataSchema, outputSize)
	// }


	return exitCode, service, outputSize, err
}

func lsDir(path string) {
	lsCmd, err := exec.Command("ls", path).Output()
	if err != nil {
		util.PrintUtil("ERROR listing output: %s\n", err.Error())
	}
	if string(lsCmd) != "" {
		util.PrintUtil("ls output: %s\n", string(lsCmd))
	}
}

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
  } 

//SetServiceOutputDir sets the output directory for the service
// returns the output directory relative to the local machine and the manage machine
func SetServiceOutputDir(imageName, serviceName string, seed *objects.Seed, localOutputDir, machineOutputDir string) (string, string) {
	if !strings.Contains(seed.Job.Interface.Command, "OUTPUT_DIR") {
		return "", ""
	}

	// #37: if -o is not specified, and OUTPUT_DIR is in the command args,
	//	auto create a time-stamped subdirectory with the name of the form:
	//		imagename-iso8601timestamp
	outputDir := localOutputDir
	if outputDir == "" {
		outputDir = "output-" + imageName + "-" + time.Now().Format(time.RFC3339)
		outputDir = strings.Replace(outputDir, ":", "_", -1)
	}

	outdir := util.GetFullPath(outputDir, "")

	// Check if outputDir exists. Create if not
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		// Create the directory
		// Didn't find the specified directory
		util.PrintUtil("INFO: %s not found; creating directory...\n",
			outdir)
		os.Mkdir(outdir, os.ModePerm)
	}

	// Check if outdir is empty. Create time-stamped subdir if not
	f, err := os.Open(outdir)
	if err != nil {
		// complain
		util.PrintUtil("ERROR: Error with %s. %s\n", outdir, err.Error())
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err != io.EOF {
		// Directory is not empty
		// t := time.Now().Format("20060102_150405")
		util.PrintUtil(
			"INFO: Output directory %s is not empty. Creating sub-directory %s for Job Output Directory.\n",
			outdir, serviceName)
		outdir = filepath.Join(outdir, serviceName)
		outputDir = filepath.Join(machineOutputDir, serviceName)
		os.Mkdir(outdir, os.ModePerm)
	}
	util.PrintUtil("Setting OUTPUT_DIR in seed to %s\n", 
		outputDir)
	seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command,
		"$OUTPUT_DIR", "/data/volume/"+serviceName, -1)
	seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command,
		"${OUTPUT_DIR}", "/data/volume/"+serviceName, -1)

	return outdir, outputDir
}

//mountVolumeDir creates directory inside of mounted output directory
func getInputDir(outputDir string) (string, string) {

	// Create directory inside of mounted output directory
	dirname := "inputs-"+time.Now().Format(time.RFC3339)
	dirname = strings.Replace(dirname, ":", "_", -1)
	mountDir := path.Join(outputDir, dirname)
	if _, err := os.Stat(mountDir); os.IsNotExist(err) {
		util.PrintUtil("INFO: %s not found; creating directory...\n",
			mountDir)
		os.Mkdir(mountDir, os.ModePerm)
	}
	
	return mountDir, dirname
}

//ProcessServiceBatchFile Creates timestampped input directory under the output directory
// copies input files to input directory and returns inputs list
func ProcessServiceBatchFile(seed objects.Seed, batchFile, outdir, mountdir string) ([]ServiceIO, error) {
	lines, err := util.ReadLinesFromFile(batchFile)
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, errors.New("ERROR: Empty batch file")
	}

	keys := strings.Split(lines[0], ",")
	extraKeys := keys

	if len(keys) == 0 || len(keys[0]) == 0 {
		return nil, errors.New("ERROR: Empty keys list on first line of batch file")
	}

	for _, f := range seed.Job.Interface.Inputs.Files {
		hasKey := util.ContainsString(keys, f.Name)
		if f.Required && !hasKey {
			msg := fmt.Sprintf("ERROR: Batch file is missing required key %v", f.Name)
			return nil, errors.New(msg)
		} else if !hasKey {
			fmt.Println("WARN: Missing input for key " + f.Name)
		}
		extraKeys = util.RemoveString(extraKeys, f.Name)
	}

	if len(extraKeys) > 0 {
		msg := fmt.Sprintf("WARN: These input keys don't match any specified keys in the Seed manifest: %v\n", extraKeys)
		fmt.Println(msg)
	}

	// Create inputs directory
	dir, inputdir := getInputDir(outdir)

	batchIO := []ServiceIO{}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		values := strings.Split(line, ",")
		machineInputs := []string{}
		fileInputs := []string{}
		inputNames := fmt.Sprintf("%d", i)
		for j, file := range values {
			if j > len(keys) {
				fmt.Println("WARN: More files provided than keys")
			}

			// copy file to new location
			cpFile := filepath.Join(dir, filepath.Base(file))
			machineFile := filepath.Join(mountdir, inputdir, file)
			if cp, err := util.CopyFiles(file, cpFile); !cp || err != nil {
				util.PrintUtil("ERROR: Error copying file %s\n", file)
				if err != nil {
					util.PrintUtil("%s\n", err.Error())
				}
				continue
			}
			// if err := os.Symlink(file, cpFile); err != nil {
			// 	util.PrintUtil("ERROR: ERROR Linking %s to volume directory\n%s", file, err.Error())
			// 	continue
			// }
			machineInputs = append(machineInputs, keys[j]+"="+machineFile)
			inputNames += "-" + filepath.Base(cpFile)
		}
		fileDir := filepath.Join(dir, inputNames)
		machineDir := filepath.Join(mountdir, inputNames)
		row := ServiceIO{fileInputs, fileDir, machineInputs, machineDir}
		batchIO = append(batchIO, row)
	}

	util.PrintUtil("Batch Input = %s \t", batchFile)
	util.PrintUtil("Batch Output Dir = %s \n", outdir)

	return batchIO, err
}

//ProcessServiceDirectory - batch input files are provided in a given directory;
// this method creates new input directory under the mounted volume and copies all input files to the new directory
func ProcessServiceDirectory(seed objects.Seed, batchDir, outdir, mountdir string) ([]ServiceIO, error) {

	key := ""
	unrequired := ""
	for _, f := range seed.Job.Interface.Inputs.Files {
		if f.Multiple {
			continue
		}
		if f.Required {
			if key != "" {
				return nil, errors.New("ERROR: Multiple required inputs are not supported when batch processing directories")
			}
			key = f.Name
		} else if unrequired == "" {
			unrequired = f.Name
		}
	}

	if key == "" {
		key = unrequired
	}

	if key == "" {
		return nil, errors.New("ERROR: Could not determine which input to use from Seed manifest")
	}

	files, err := ioutil.ReadDir(batchDir)
	if err != nil {
		return nil, err
	}

	// Create inputs directory
	dir, inputdir := getInputDir(outdir)

	batchIO := []ServiceIO{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileDir := filepath.Join(dir, file.Name())
		filePath := filepath.Join(dir, file.Name())
		srcFile := filepath.Join(batchDir, file.Name())

		machinePath := filepath.Join("volume", inputdir, file.Name())
		if cp, err := util.CopyFile(srcFile, filePath); !cp || err != nil {
			util.PrintUtil("ERROR: Error copying files to volume input directory\n")
			if err != nil {
				util.PrintUtil("%s\n", err.Error())
			}
			continue
		}
		// if err := os.Symlink(srcFile, filePath); err != nil {
		// 	util.PrintUtil("ERROR: ERROR Linking %s to volume directory\n%s\n", file.Name(), err.Error())
		// 	continue
		// }
		fileInputs := []string{}
		fileInputs = append(fileInputs, key+"="+fileDir)
		machineInputs := []string{}
		machineInputs = append(machineInputs, key+"="+machinePath)
		row := ServiceIO{fileInputs, fileDir, machineInputs, filepath.Join("volume", inputdir)}
		batchIO = append(batchIO, row)
	}
	util.PrintUtil("Batch Input Dir = %v \t Batch Output Dir = %v \n", batchDir, outdir)

	return batchIO, err
}

//DefineServiceInputs extracts the paths to any input data given by the 'cluster' command
// flags 'inputs' and sets the path in the json object. Returns:
// 	[]string: docker command args for input files in the format:
//	"--mount src=/path/to/file1,destination=/path/to/file1"
func DefineServiceInputs(seed *objects.Seed, localinputs, machineinputs []string) ([]string, float64, map[string]string, error) {
	// Validate inputs given vs. inputs defined in manifest
	util.PrintUtil("Define service inputs:\n%v\n%v\n", localinputs, machineinputs)

	var mountArgs []string
	var sizeMiB float64

	localinMap := inputMap(localinputs)

	// Valid by default
	valid := true
	var keys []string
	var unrequired []string
	var tempDirectories map[string]string
	tempDirectories = make(map[string]string)
	for _, f := range seed.Job.Interface.Inputs.Files {
		if f.Multiple {
			tempDir := "temp-" + time.Now().Format(time.RFC3339)
			tempDir = strings.Replace(tempDir, ":", "_", -1)
			os.Mkdir(tempDir, os.ModePerm)
			tempDirectories[f.Name] = tempDir
			mountArgs = append(mountArgs, "--mount")
			mountArgs = append(mountArgs, "type=bind,src="+util.GetFullPath(tempDir,"")+",destination="+tempDir)
		}
		if f.Required == false {
			unrequired = append(unrequired, f.Name)
			continue
		}
		keys = append(keys, f.Name)
		if _, prs := localinMap[f.Name]; !prs {
			valid = false
		}
	}

	if !valid {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR: Incorrect input data files key/values provided. -i arguments should be in the form:\n")
		buffer.WriteString("  seed run -i KEY1=path/to/file1 -i KEY2=path/to/file2 ...\n")
		buffer.WriteString("The following input file keys are expected:\n")
		for _, n := range keys {
			buffer.WriteString("  " + n + "\n")
		}
		buffer.WriteString("\n")
		return nil, 0.0, tempDirectories, errors.New(buffer.String())
	}

	for i, f := range localinputs {
		x := strings.SplitN(f, "=", 2)
		if len(x) != 2 {
			util.PrintUtil("ERROR: Input files should be specified in KEY=VALUE format.\n")
			util.PrintUtil("ERROR: Unknown key for input %v encountered.\n",
				localinputs)
			continue
		}

		key := x[0]
		val := x[1]

		// Expand input VALUE
		val = util.GetFullPath(val, "")

		//get total size of input files in MiB
		info, err := os.Stat(val)
		if os.IsNotExist(err) {
			util.PrintUtil("ERROR: Input file %s not found\n", val)
		}
		sizeMiB += (1.0 * float64(info.Size())) / (1024.0 * 1024.0) //fileinfo's Size() returns bytes, convert to MiB


		// get the value from the machine list
		y := strings.SplitN(machineinputs[i], "=", 2)[1]
		// Replace key if found in args strings
		// Handle replacing KEY or ${KEY} or $KEY
		util.PrintUtil("Value is %s\n", val)
		// value := val
		value := filepath.Join("/data", y)
		if directory, ok := tempDirectories[key]; ok {
			value = directory //replace with the temp directory if multiple files
		}
		seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command,
			"${"+key+"}", value, -1)
		seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command, "$"+key,
			value, -1)
		seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command, key, value,
			-1)

		// for _, k := range seed.Job.Interface.Inputs.Files {
		// 	if k.Name == key {
		// 		if k.Multiple {
					//directory has already been added to mount args, just link file into that directory
					// os.Symlink(y, filepath.Join(tempDirectories[key], info.Name()))
			// 	} else {
			// 		mountArgs = append(mountArgs, "--mount")
			// 		mountArgs = append(mountArgs, "type=bind,src="++",destination="+y)
			// 	}
			// }
		// }
	}

	//remove unspecified unrequired inputs from cmd string
	for _, k := range unrequired {
		key := k
		value := ""
		seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command,
			"${"+key+"}", value, -1)
		seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command, "$"+key,
			value, -1)
		seed.Job.Interface.Command = strings.Replace(seed.Job.Interface.Command, key, value,
			-1)
	}

	return mountArgs, sizeMiB, tempDirectories, nil
}

//DefineServiceMounts defines any seed specified mounts.
func DefineServiceMounts(seed *objects.Seed, volume Volume, inputs []string) ([]Mount, error) {
	util.PrintUtil("Defining service mounts: %v\n", inputs)

	inMap := inputMap(inputs)

	// Valid by default
	valid := true
	var keys []string
	for _, f := range seed.Job.Interface.Mounts {
		keys = append(keys, f.Name)
		if _, prs := inMap[f.Name]; !prs {
			valid = false
		}
	}

	if !valid {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR: Incorrect mount key/values provided. -m arguments should be in the form:\n")
		buffer.WriteString("  seed run -m MOUNT=path/to ...\n")
		buffer.WriteString("The following mount keys are expected:\n")
		for _, n := range keys {
			buffer.WriteString("  " + n + "\n")
		}
		buffer.WriteString("\n")
		return nil, errors.New(buffer.String())
	}

	var mounts []Mount
	if seed.Job.Interface.Mounts != nil {
		for _, mount := range seed.Job.Interface.Mounts {

			localpath := util.GetFullPath(inMap[mount.Name], "")
			machinepath := filepath.Join(volume.HostMachineDir, filepath.Base(localpath))
			localout := filepath.Join(volume.LocalDir, filepath.Base(localpath))
			util.CopyFiles(localpath, localout)
			// mountPath := "type=bind,src="+machinepath +",destination="+mount.Path
			mountPath := "source="+volume.Name+",target="+localpath

			if mount.Mode == "" {
			// 	mountPath += "," + mount.Mode
			// } else {
				mountPath += ",ro"
			}
			m := Mount{localpath, machinepath, volume.Name, []string{"--mount", mountPath}}
			mounts = append(mounts, m)
		}
		return mounts, nil
	}

	return mounts, nil
}

//DefineServiceSettings defines any seed specified docker settings.
// Return []string of docker command arguments in form of:
//	"--env setting1=val1 --env setting2=val2 etc"
func DefineServiceSettings(seed *objects.Seed, inputs []string) ([]string, error) {
	inMap := inputMap(inputs)

	// Valid by default
	valid := true
	var keys []string
	for _, s := range seed.Job.Interface.Settings {
		keys = append(keys, s.Name)
		if _, prs := inMap[s.Name]; !prs {
			valid = false
		}
	}

	if !valid {
		var buffer bytes.Buffer
		buffer.WriteString("ERROR: Incorrect setting key/values provided. -e arguments should be in the form:\n")
		buffer.WriteString("  seed run -e SETTING=somevalue ...\n")
		buffer.WriteString("The following settings are expected:\n")
		for _, n := range keys {
			buffer.WriteString("  " + n + "\n")
		}
		buffer.WriteString("\n")
		return nil, errors.New(buffer.String())
	}

	var settings []string
	for _, key := range keys {
		settings = append(settings, "--env")
		settings = append(settings, util.GetNormalizedVariable(key)+"="+inMap[key])
	}

	return settings, nil
}

//DefineServiceResources defines any seed specified docker resource requirements
//based on the seed spec and the size of the input in MiB
// returns array of arguments to pass to docker to restrict/specify the resources required
// returns the total disk space requirement to be checked when validating output
func DefineServiceResources(seed *objects.Seed, inputSizeMiB float64) ([]string, float64, error) {
	var resources []string
	var disk float64

	for _, s := range seed.Job.Resources.Scalar {
		if s.Name == "mem" {
			//resourceRequirement = inputVolume * inputMultiplier + constantValue
			mem := (s.InputMultiplier * inputSizeMiB) + s.Value
			mem = math.Max(mem, 4.0)        //docker memory requirement must be > 4MiB
			intMem := int64(math.Ceil(mem)) //docker expects integer, get the ceiling of the specified value and convert
			resources = append(resources, "--limit-memory")
			resources = append(resources, fmt.Sprintf("%dm", intMem))
		}
		if s.Name == "disk" {
			//resourceRequirement = inputVolume * inputMultiplier + constantValue
			disk = (s.InputMultiplier * inputSizeMiB) + s.Value
		}
		if s.Name == "sharedMem" {
			//resourceRequirement = inputVolume * inputMultiplier + constantValue
			mem := ((s.InputMultiplier * inputSizeMiB) + s.Value) * 9.5367431640625E-7
			intMem := int64(math.Ceil(mem)) //docker expects integer, get the ceiling of the specified value and convert
			resources = append(resources, fmt.Sprintf("--mount type=tmpfs,destination=/dev/shm,tmpfs-size=%d", intMem))
		}		
	}

	resources = []string{}

	return resources, disk, nil
}

//CheckServiceOutput validates the output of the docker run command. Output data is
// validated as defined in the seed.Job.Interface.Outputs.
func CheckServiceOutput(seed *objects.Seed, outDir, metadataSchema string, diskLimit float64) {
	// Validate any Outputs.Files
	if seed.Job.Interface.Outputs.Files != nil {
		util.PrintUtil("INFO: Validating output files found under %s...\n",
			outDir)

		var dirSize int64
		readSize := func(path string, file os.FileInfo, err error) error {
			if file != nil && !file.IsDir() {
				dirSize += file.Size()
			}

			return nil
		}
		filepath.Walk(outDir, readSize)
		sizeMB := float64(dirSize) / (1024.0 * 1024.0)
		if diskLimit > 0 && sizeMB > diskLimit {
			util.PrintUtil("ERROR: Output directory exceeds disk space limit (%f MiB vs. %f MiB)\n", sizeMB, diskLimit)
		}

		// For each defined Outputs file:
		//	#1 Check file media type
		// 	#2 Check file names match output pattern
		//  #3 Check number of files (if defined)
		for _, f := range seed.Job.Interface.Outputs.Files {
			// find all pattern matches in OUTPUT_DIR
			matches, _ := filepath.Glob(path.Join(outDir, f.Pattern))

			// Check media type of matches
			count := 0
			var matchList []string
			for _, match := range matches {
				ext := filepath.Ext(match)
				mType := mime.TypeByExtension(ext)
				if strings.Contains(mType, f.MediaType) ||
					strings.Contains(f.MediaType, mType) {
					count++
					matchList = append(matchList, "\t"+match+"\n")
					metadata := match + ".metadata.json"
					if _, err := os.Stat(metadata); err == nil {
						schema := metadataSchema
						if schema != "" {
							schema = util.GetFullPath(schema, "")
						}
						err := ValidateSeedFile(schema, metadata, constants.SchemaMetadata)
						if err != nil {
							util.PrintUtil("ERROR: Side-car metadata file %s validation error: %s", metadata, err.Error())
						}
					}
				}
			}

			// Validate that any required fields are present
			if f.Required {
				util.PrintUtil("ERROR: Required file expected for output %v, %v found.\n",
					f.Name, strconv.Itoa(len(matchList)))
			} else if !f.Multiple && len(matchList) < 1 {
				util.PrintUtil("ERROR: Multiple files found for single output %v, %v found.\n",
					f.Name, strconv.Itoa(len(matchList)))
			} else {

				util.PrintUtil("SUCCESS: Files found for output %v: %s\n",
					f.Name, strconv.Itoa(len(matchList)))
				for _, s := range matchList {
					util.PrintUtil(s)
				}
			}
		}
	}

	// Validate any defined Outputs.Json
	// Look for ResultsFileManifestName.json in the root of the OUTPUT_DIR
	// and then validate any keys identified in Outputs exist
	if seed.Job.Interface.Outputs.JSON != nil {
		util.PrintUtil("INFO: Validating %s...\n",
			filepath.Join(outDir, constants.ResultsFileManifestName))
		// look for results manifest
		manfile := filepath.Join(outDir, constants.ResultsFileManifestName)
		if _, err := os.Stat(manfile); os.IsNotExist(err) {
			util.PrintUtil("ERROR: %s specified but cannot be found. %s\n Exiting testrunner.\n\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}

		bites, err := ioutil.ReadFile(filepath.Join(outDir,
			constants.ResultsFileManifestName))
		if err != nil {
			util.PrintUtil("ERROR: Error reading %s.%s\n\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}

		documentLoader := gojsonschema.NewStringLoader(string(bites))
		_, err = documentLoader.LoadJSON()
		if err != nil {
			util.PrintUtil("ERROR: Error loading results manifest file: %s. %s\n Exiting testrunner.\n\n",
				constants.ResultsFileManifestName, err.Error())
			return
		}

		schemaFmt := "{ \"type\": \"object\", \"properties\": { %s }, \"required\": [ %s ] }"
		schema := ""
		required := ""

		// Loop through defined name/key values to extract from seed.outputs.json
		for _, jsonStr := range seed.Job.Interface.Outputs.JSON {
			key := jsonStr.Name
			if jsonStr.Key != "" {
				key = jsonStr.Key
			}

			schema += fmt.Sprintf("\"%s\": { \"type\": \"%s\" },", key, jsonStr.Type)

			if jsonStr.Required {
				required += fmt.Sprintf("\"%s\",", key)
			}
		}
		//remove trailing commas
		if len(schema) > 0 {
			schema = schema[:len(schema)-1]
		}
		if len(required) > 0 {
			required = required[:len(required)-1]
		}

		schema = fmt.Sprintf(schemaFmt, schema, required)

		schemaLoader := gojsonschema.NewStringLoader(schema)
		schemaResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			util.PrintUtil("ERROR: Error running validator: %s\n Exiting testrunner.\n\n",
				err.Error())
			return
		}

		if len(schemaResult.Errors()) == 0 {
			util.PrintUtil("SUCCESS: Results manifest file is valid.\n")
		}

		for _, desc := range schemaResult.Errors() {
			util.PrintUtil("ERROR: %s is invalid: - %s\n", constants.ResultsFileManifestName, desc)
		}
	}
}

//CopyServiceOutput Copies the output directory to outside the mounted directory
func CopyServiceOutput(mountoutdir, outdir string) error {
	util.PrintUtil("INFO: Copying service output from %s to %s\n", mountoutdir, outdir)
	copied, err := util.CopyFiles(mountoutdir, outdir)
	if err != nil {
		util.PrintUtil("Error copying service output: %s\n", err.Error())
		return err
	} else if !copied {
		util.PrintUtil("Copied returned false... what?\n")
	}

	util.PrintUtil("INFO: Service output files copied...\n\n")
	return nil
}

//getServiceOutputDir returns the service output directory. Creates an output folder
func getServiceOutputDir(outputDir, imageName string) string {

	// check for slashes in imageName
	imageName = strings.Replace(imageName, "/", "-", -1)

	if outputDir == "" {
		outputDir = "batch-" + imageName + "-" + time.Now().Format(time.RFC3339)
		outputDir = strings.Replace(outputDir, ":", "_", -1)
	}

	outdir := util.GetFullPath(outputDir, "")

	// Check if outputDir exists. Create if not
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		// Create the directory
		// Didn't find the specified directory
		util.PrintUtil("INFO: %s not found; creating directory...\n",
			outdir)
		os.Mkdir(outdir, os.ModePerm)
	}

	// Create a volume directory withing to mount to the machine. 
	// This directory's contents will be removed when the folder is unmounted
	localoutdir := filepath.Join(outdir, "volume")
	if _, err := os.Stat(localoutdir); os.IsNotExist(err) {
		// Create the directory
		// Didn't find the specified directory
		util.PrintUtil("INFO: %s not found; creating directory...\n",
			localoutdir)
		os.Mkdir(localoutdir, os.ModePerm)
	}

	return localoutdir
}


// RegistrySetup Pushes the local image to the given registry
func RegistrySetup(manager, imageName, registry string) (string, error) {

	if registry == "" {
		registry = "localhost:5000"
	}
	tag := registry+"/"+imageName

	if exist, _ := util.MachineImageExists(manager, imageName); !exist {
		// Save local image (not on cluster) to tar file
		util.PrintUtil("INFO: Exporting local image %s to .tar file...\n", imageName)
		imgFile, err := util.SaveImage(imageName)
		if err != nil {
			return "", err
		}
		
		// SCP saved .tar to cluster manager node
		util.PrintUtil("INFO: Local image %s saved to %s.\nMoving file to cluster node %s...\n", 
			imageName, imgFile, manager)
		if _, err = util.MachineSCP(imgFile, "/tmp/"+imgFile, manager); err != nil {
			return "", err
		}

		// Load image to manager node registry
		util.PrintUtil("INFO: Image file moved to cluster node %s.\nLoading image to %s local registry...\n",
			manager, manager)
		if _, err = util.MachineLoad(imgFile, manager); err != nil {
			return "", err
		}
		
		// Remove the exported image file from local directory
		util.PrintUtil("INFO: Image loaded into registry %s on %s. Removing tar file from host.\n", registry, manager)
		err = os.Remove(imgFile)
		if err != nil {
			util.PrintUtil("ERROR: Error removing image file %s.", imgFile)
			return "", err
		}
	}

	// Tag the image on the cluster
	util.PrintUtil("INFO: Tagging %s to %s\n", imageName, tag)
	err := util.MachineTag(manager, imageName, tag)
	if err != nil {
		return tag, err
	}

	// Push the image to the cluster registry
	util.PrintUtil("INFO: Pushing tagged image to %s registry\n", manager)
	err = util.MachinePush(manager, tag)
	if err != nil {
		return tag,err
	}
	return tag, nil
}

//PrintClusterUsage prints the seed batch usage arguments, then exits the program
func PrintClusterUsage() {
	// seed cluster -in IMAGE_NAME \
	// <-b BATCH FILE || -d BATCH DIRECTORY> \
	// -e ENV_VAR SETTING \
	// -in IMAGE_NAME
	// -m MOUNT \
	// -ma MANAGER_NODE_NAME
	// -o OUTPUT_DIR \
	// -r REGISTRY \
	// -s SCHEMA \
	util.PrintUtil("\nUsage:\tseed cluster -in IMAGE_NAME [OPTIONS] \n")

	util.PrintUtil("\nRuns Docker image defined by seed spec as a service on a cluster.\n")

	util.PrintUtil("\nOptions:\n")
	util.PrintUtil("  -%s  -%s \t Optional file specifying input keys and file mapping for batch processing. Supersedes directory flag.\n",
		constants.ShortBatchFlag, constants.BatchFlag)
	util.PrintUtil("  -%s  -%s Alternative to batch file. Specifies a directory of files to batch process (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	util.PrintUtil("  -%s  -%s \t Specifies the key/value setting values of the seed spec in the format SETTING_KEY=VALUE\n",
		constants.ShortSettingFlag, constants.SettingFlag)
	util.PrintUtil("  -%s -%s Docker image name to run\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	// util.PrintUtil("  -%s     \t Defines if running in batch mode; Not specifying this option defaults to single mode; valid options are 'batch' or 'single'\n", 
	// 	constants.ModeFlag)
	util.PrintUtil("  -%s  -%s \t Specifies the key/value mount values of the seed spec in the format MOUNT_KEY=HOST_PATH\n",
		constants.ShortMountFlag, constants.MountFlag)
	util.PrintUtil("  -%s -%s \t Defines the cluster's manager node\n", 
		constants.ShortManagerFlag, constants.ManagerFlag)
	util.PrintUtil("  -%s  -%s \t Job Output Directory Location\n",
		constants.ShortJobOutputDirFlag, constants.JobOutputDirFlag)
	util.PrintUtil("  -%s  -%s Optionally specifies registry to push local image to on the cluster; Default is localhost:5000 if not specified\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	util.PrintUtil("  -%s  -%s \t External Seed metadata schema file; Overrides built in schema to validate side-car metadata files\n",
		constants.ShortSchemaFlag, constants.SchemaFlag)
		
	panic(util.Exit{0})
}