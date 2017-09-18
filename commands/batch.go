package commands

import (
	"errors"

	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"

	"os"
	"fmt"
	"github.com/ngageoint/seed-cli-bak/commands"
	"time"
	"strings"
	"github.com/ngageoint/seed-cli/constants"
	"io/ioutil"
	"path/filepath"
)

func BatchRun(batchDir, imageName, outputDir, metadataSchema string, settings, mounts []string, rmFlag bool) error {
	if imageName == "" {
		return errors.New("ERROR: No input image specified.")
	}

	if exists, err := util.ImageExists(imageName); !exists {
		return err
	}

	if batchDir == "" {
		batchDir = "."
	}

	batchDir = util.GetFullPath(batchDir, "")

	seed := objects.SeedFromImageLabel(imageName)

	key := ""
	unrequired := ""
	for _, f := range seed.Job.Interface.Inputs.Files {
		if f.Multiple {
			continue
		}
		if f.Required {
			if key != "" {
				return errors.New("ERROR: Batch processing does not support multiple required inputs.")
			}
			key = f.Name
		} else if unrequired == ""{
			unrequired = f.Name
		}
	}

	if key == "" {
		key = unrequired
	}

	if key == "" {
		return errors.New("ERROR: Could not determine which input to use from Seed manifest.")
	}

	files, err := ioutil.ReadDir(batchDir)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if outputDir == "" {
		outputDir = "batch-" + imageName + "-" + time.Now().Format(time.RFC3339)
		outputDir = strings.Replace(outputDir, ":", "_", -1)
	}

	outdir := util.GetFullPath(outputDir, "")

	// Check if outputDir exists. Create if not
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		// Create the directory
		// Didn't find the specified directory
		fmt.Fprintf(os.Stderr, "INFO: %s not found; creating directory...\n",
			outdir)
		os.Mkdir(outdir, os.ModePerm)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileDir := filepath.Join(outdir, file.Name())
		filePath := filepath.Join(batchDir, file.Name())
		fmt.Println(file)
		fmt.Println(fileDir)
		inputs := []string{}
		inputs = append(inputs, key + "=" + filePath)
		fmt.Println(inputs)
		commands.DockerRun(imageName, fileDir, metadataSchema, inputs, settings, mounts, rmFlag)
	}

	return err
}

//PrintBatchUsage prints the seed batch usage arguments, then exits the program
func PrintBatchUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed batch -in IMAGE_NAME [OPTIONS] \n")

	fmt.Fprintf(os.Stderr, "\nRuns Docker image defined by seed spec.\n")

	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s Docker image name to run\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s Specifies the directory of files to batch process\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s \t Specifies the key/value setting values of the seed spec in the format SETTING_KEY=VALUE\n",
		constants.ShortSettingFlag, constants.SettingFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s \t Specifies the key/value mount values of the seed spec in the format MOUNT_KEY=HOST_PATH\n",
		constants.ShortMountFlag, constants.MountFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s \t Job Output Directory Location\n",
		constants.ShortJobOutputDirFlag, constants.JobOutputDirFlag)
	fmt.Fprintf(os.Stderr, "  -%s \t\t Automatically remove the container when it exits (docker run --rm)\n",
		constants.RmFlag)
	fmt.Fprintf(os.Stderr, "  -%s  -%s \t External Seed metadata schema file; Overrides built in schema to validate side-car metadata files\n",
		constants.ShortSchemaFlag, constants.SchemaFlag)
	panic(util.Exit{0})
}
