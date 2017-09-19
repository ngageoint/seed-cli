package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

type BatchIO struct {
	Inputs []string
	Outdir string
}

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

	outdir := getOutputDir(outputDir, imageName)

	inputs, err := ParseDirectory(seed, batchDir, outdir)

	_ = inputs

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

func getOutputDir(outputDir, imageName string) string {
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
	return outdir
}

func ProcessDirectory(seed objects.Seed, batchDir, outdir string) ([]BatchIO, error) {
	key := ""
	unrequired := ""
	for _, f := range seed.Job.Interface.Inputs.Files {
		if f.Multiple {
			continue
		}
		if f.Required {
			if key != "" {
				return nil, errors.New("ERROR: Batch processing does not support multiple required inputs.")
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
		return nil, errors.New("ERROR: Could not determine which input to use from Seed manifest.")
	}

	files, err := ioutil.ReadDir(batchDir)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	batchIO := []BatchIO{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileDir := filepath.Join(outdir, file.Name())
		filePath := filepath.Join(batchDir, file.Name())
		fileInputs := []string{}
		fileInputs = append(fileInputs, key+"="+filePath)
		row := BatchIO{fileInputs, fileDir}
		batchIO = append(batchIO, row)
	}

	return batchIO, err
}

func ProcessBatchFile(seed objects.Seed, batchFile, outdir string) ([]BatchIO, error) {
	lines, err := util.ReadLinesFromFile(batchFile)
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, errors.New("ERROR: Empty batch file")
	}

	keys := strings.Split(lines[0], ",")
	extraKeys := keys

	if len(keys) == 0 {
		return nil, errors.New("ERROR: Empty keys list on first line of batch file.")
	}

	for _, f := range seed.Job.Interface.Inputs.Files {
		hasKey := util.ContainsString(keys, f.Name)
		if f.Required && !hasKey {
			msg := fmt.Sprintf("ERROR: Batch file is missing required key %v", f.Name)
			return nil, errors.New(msg)
		} else if !hasKey {
			fmt.Println("WARN: Missing input for key " + f.Name)
		}
		util.RemoveString(extraKeys, f.Name)
	}

	if len(extraKeys) > 0 {
		msg := fmt.Sprintf("WARN: These input keys don't match any specified keys in the Seed manifest: %v\n", extraKeys)
		fmt.Println(msg)
	}

	batchIO := []BatchIO{}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		values := strings.Split(line, ",")
		fileInputs := []string{}
		for j, file := range values {
			if j > len(keys) {
				fmt.Println("WARN: More files provided than keys")
			}
			fileInputs = append(fileInputs, keys[j]+"="+file)
		}
		fileDir := filepath.Join(outdir, fmt.Sprintf("%s", i))
		row := BatchIO{fileInputs, fileDir}
		batchIO = append(batchIO, row)
	}

	return batchIO, err
}
