package processor

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/schema"
	"os"
	"path/filepath"
	"strings"
)

const (
	outputFileName = "output.json"
	outputDir      = "module-schema/processor"
)

func ProcessBatchOutput(dirPath string) ([]schema.TerraformObject, error) {
	var terraformObjects []schema.TerraformObject
	variablesDir := filepath.Join(dirPath, "variables")
	examplesDir := filepath.Join(dirPath, "examples")
	filepath.WalkDir(variablesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through directory: %w", err)
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), "_variables.tf") {
			return nil
		}

		objName := strings.TrimSuffix(filepath.Base(path), "_variables.tf")
		hclBlock, err := schema.FromHCLFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse HCL file %s: %w", path, err)
		}

		parsedBlock, err := FindFieldsFromBlock(hclBlock, objName)

		// Read the example HCL file content
		exampleFilePath := filepath.Join(examplesDir, fmt.Sprintf("%s_example.tf", objName))
		exampleHclContent, err := os.ReadFile(exampleFilePath)
		var exampleHcl string
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error reading example file %s: %w", exampleFilePath, err)
			}
			// If the file doesn't exist, we just leave the string empty.
			exampleHcl = ""
		} else {
			exampleHcl = string(exampleHclContent)
		}

		singleObj := schema.TerraformObject{
			Name:       objName,
			Fields:     parsedBlock.Attributes,
			ExampleHCL: exampleHcl,
		}

		terraformObjects = append(terraformObjects, singleObj)
		return nil
	})
	jsonOutput, err := json.MarshalIndent(terraformObjects, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal terraform objects to JSON: %w", err)
	}

	err = os.WriteFile(filepath.Join(outputDir, outputFileName), jsonOutput, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write output to file: %w", err)
	}

	fmt.Printf("Output written to %s\n", outputFileName)
	return terraformObjects, nil
}

func ProcessOutput(fileName string) (*schema.TerraformObject, error) {
	// Created a block from the output file, but the format is not ideal, we still need to process the `AttributeType`, `Default`
	hclBlock, err := schema.FromHCLFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL file: %w", err)
	}
	objName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	parsedBlock, err := FindFieldsFromBlock(hclBlock, objName)
	if err != nil {
		return nil, fmt.Errorf("failed to find fields in block: %w", err)
	}
	// Marshal the parsed block to JSON
	jsonOutput, err := json.MarshalIndent(parsedBlock, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parsed block to JSON: %w", err)
	}

	err = os.WriteFile(outputFileName, jsonOutput, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write output to file: %w", err)
	}

	fmt.Printf("Output written to %s\n", outputFileName)
	return &schema.TerraformObject{}, nil
}
