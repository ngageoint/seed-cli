package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
	"github.com/xeipuuv/gojsonschema"
)

// seed validate: Validate seed.manifest.json. Does not require docker
func Validate(schemaFile, dir string) error {
	var err error = nil
	var seedFileName = ""

	seedFileName, err = util.SeedFileName(dir)
	if err != nil {
		return err
	}

	if schemaFile != "" {
		schemaFile = "file://" + util.GetFullPath(schemaFile, dir)
	}

	err = ValidateSeedFile(schemaFile, seedFileName, constants.SchemaManifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}

	return err
}

//PrintValidateUsage prints the seed validate usage, then exits the program
func PrintValidateUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage:\tseed validate [OPTIONS] \n")
	fmt.Fprintf(os.Stderr, "\nValidates the given %s by verifying it is compliant with the Seed spec.\n",
		constants.SeedFileName)
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	fmt.Fprintf(os.Stderr, "  -%s -%s\tSpecifies directory in which Seed is located (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	fmt.Fprintf(os.Stderr, "  -%s -%s   \tExternal Seed schema file; Overrides built in schema to validate Seed spec against\n",
		constants.ShortSchemaFlag, constants.SchemaFlag)
	os.Exit(0)
}

//ValidateSeedFile Validates the seed.manifest.json file based on the given schema
func ValidateSeedFile(schemaFile string, seedFileName string, schemaType constants.SchemaType) error {
	var result *gojsonschema.Result
	var err error

	typeStr := "manifest"
	if schemaType == constants.SchemaMetadata {
		typeStr = "metadata"
	}

	// Load supplied schema file
	if schemaFile != "" {
		fmt.Fprintf(os.Stderr, "INFO: Validating seed %s file %s against schema file %s...\n",
			typeStr, seedFileName, schemaFile)
		schemaLoader := gojsonschema.NewReferenceLoader(schemaFile)
		docLoader := gojsonschema.NewReferenceLoader("file://" + seedFileName)
		result, err = gojsonschema.Validate(schemaLoader, docLoader)

		// Load baked-in schema file
	} else {
		fmt.Fprintf(os.Stderr, "INFO: Validating seed %s file %s against schema...\n",
			typeStr, seedFileName)
		schemaBytes, _ := constants.Asset("schema/seed.manifest.schema.json")
		if schemaType == constants.SchemaMetadata {
			schemaBytes, _ = constants.Asset("schema/seed.metadata.schema.json")
		}
		schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
		docLoader := gojsonschema.NewReferenceLoader("file://" + seedFileName)
		result, err = gojsonschema.Validate(schemaLoader, docLoader)
	}

	// Error occurred loading the schema or seed.manifest.json
	if err != nil {
		return errors.New("ERROR: Error validating seed file against schema. Error is:" + err.Error() + "\n")
	}

	// Validation failed. Print results
	var buffer bytes.Buffer
	if !result.Valid() {
		buffer.WriteString("ERROR:" + seedFileName + " is not valid. See errors:\n")
		for _, e := range result.Errors() {
			buffer.WriteString("-ERROR " + e.Description() + "\n")
			buffer.WriteString("\tField: " + e.Field() + "\n")
			buffer.WriteString("\tContext: " + e.Context().String() + "\n")
		}
	}

	//Identify any name collisions for the follwing reserved variables:
	//		OUTPUT_DIR, ALLOCATED_CPUS, ALLOCATED_MEM, ALLOCATED_SHARED_MEM, ALLOCATED_STORAGE
	fmt.Fprintf(os.Stderr, "INFO: Checking for variable name collisions...\n")
	seed := objects.SeedFromManifestFile(seedFileName)

	// Grab all scalar resource names (verify none are set to OUTPUT_DIR)
	var allocated []string
	// var vars map[string]string
	vars := make(map[string][]string)
	if seed.Job.Resources.Scalar != nil {
		for _, s := range seed.Job.Resources.Scalar {
			name := util.GetNormalizedVariable(s.Name)
			allocated = append(allocated, "ALLOCATED_"+strings.ToUpper(name))
			if util.IsReserved(s.Name, nil) {
				buffer.WriteString("ERROR: job.resources.scalar Name " +
					s.Name + " is a reserved variable. Please choose a different name value.\n")
			}

			util.IsInUse(s.Name, "job.resources.scalar", vars)
		}
	}

	if seed.Job.Interface.InputData.Files != nil {
		for _, f := range seed.Job.Interface.InputData.Files {
			// check against the ALLOCATED_* and OUTPUT_DIR
			if util.IsReserved(f.Name, allocated) {
				buffer.WriteString("ERROR: job.interface.inputData.files Name " +
					f.Name + " is a reserved variable. Please choose a different name value.\n")
			}

			util.IsInUse(f.Name, "job.interface.inputData.files", vars)
		}
	}

	if seed.Job.Interface.InputData.Json != nil {
		for _, f := range seed.Job.Interface.InputData.Json {
			if util.IsReserved(f.Name, allocated) {
				buffer.WriteString("ERROR: job.interface.inputData.json Name " +
					f.Name + " is a reserved variable. Please choose a different name value.\n")
			}

			util.IsInUse(f.Name, "job.interface.inputData.json", vars)
		}
	}

	if seed.Job.Interface.OutputData.Files != nil {
		for _, f := range seed.Job.Interface.OutputData.Files {
			// check against the ALLOCATED_* and OUTPUT_DIR
			if util.IsReserved(f.Name, allocated) {
				buffer.WriteString("ERROR: job.interface.outputData.files Name " +
					f.Name + " is a reserved variable. Please choose a different name value.\n")
			}
			util.IsInUse(f.Name, "job.interface.outputData.files", vars)
		}
	}

	if seed.Job.Interface.OutputData.JSON != nil {
		for _, f := range seed.Job.Interface.OutputData.JSON {
			// check against the ALLOCATED_* and OUTPUT_DIR
			if util.IsReserved(f.Name, allocated) {
				buffer.WriteString("ERROR: job.interface.outputData.json Name " +
					f.Name + " is a reserved variable. Please choose a different name value.\n")
			}
			util.IsInUse(f.Name, "job.interface.outputData.json", vars)
		}
	}

	if seed.Job.Interface.Mounts != nil {
		for _, m := range seed.Job.Interface.Mounts {
			// check against the ALLOCATED_* and OUTPUT_DIR
			if util.IsReserved(m.Name, allocated) {
				buffer.WriteString("ERROR: job.interface.mounts Name " + m.Name +
					" is a reserved variable. Please choose a different name value.\n")
			}
			util.IsInUse(m.Name, "job.interface.mounts", vars)
		}
	}

	if seed.Job.Interface.Settings != nil {
		for _, s := range seed.Job.Interface.Settings {
			// check against the ALLOCATED_* and OUTPUT_DIR
			if util.IsReserved(s.Name, allocated) {
				buffer.WriteString("ERROR: job.interface.settings Name " + s.Name +
					" is a reserved variable. Please choose a different name value.\n")
			}
			util.IsInUse(s.Name, "job.interface.settings", vars)
		}
	}

	// Find any name collisions
	for key, val := range vars {
		if len(val) > 1 {
			buffer.WriteString("ERROR: Multiple Name values are assigned the same " +
				key + " Name value. Each Name value must be unique.\n")
			for _, v := range val {
				buffer.WriteString("\t" + v + "\n")
			}
		}
	}

	// Return error if issues found
	if buffer.String() != "" {
		return errors.New(buffer.String())
	}

	// Validation succeeded
	fmt.Fprintf(os.Stderr, "SUCCESS: No errors found. %s is valid.\n\n", seedFileName)
	return nil
}
