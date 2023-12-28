package documents

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// splitYAML splits a byte slice of YAML into individual documents.
func splitYAML(resources []byte) ([][]byte, error) {
	dec := yaml.NewDecoder(bytes.NewReader(resources))
	var documents [][]byte

	for {
		var doc interface{}
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding YAML: %w", err)
		}

		docBytes, err := yaml.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("error marshaling YAML: %w", err)
		}
		documents = append(documents, docBytes)
	}

	return documents, nil
}

// ReadYAMLWithDashDashDashSingle reads a YAML file and returns a slice of maps, each representing a YAML document.
// This allows to read yaml documentw with multiple documents container, separated by "---"
func ReadYAMLWithDashDashDashSingle(filename string) ([]map[string]interface{}, error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	yamlDocuments, err := splitYAML(fileContents)
	if err != nil {
		return nil, fmt.Errorf("error splitting YAML: %w", err)
	}

	var yamlMaps []map[string]interface{}
	for _, doc := range yamlDocuments {
		var yamlMap map[string]interface{}
		if err := yaml.Unmarshal(doc, &yamlMap); err != nil {
			return nil, fmt.Errorf("error unmarshaling YAML: %w", err)
		}
		yamlMaps = append(yamlMaps, yamlMap)
	}
	return yamlMaps, nil
}

// ProcessSectionCallback defines the type for the callback function.
type ProcessSectionCallback func(cmd string, section map[string]interface{})

// ProcessFiles processes multiple files using a callback function for each section.
func ProcessFiles(cmd string, processSection ProcessSectionCallback, filenames []string) error {
	for _, filename := range filenames {
		if err := processFile(filename, cmd, processSection); err != nil {
			return fmt.Errorf("error processing file %s: %w", filename, err)
		}
	}
	return nil
}

// processFile processes a single file, applying the callback to each section.
func processFile(filename, cmd string, processSection ProcessSectionCallback) error {
	sections, err := ReadYAMLWithDashDashDashSingle(filename)
	if err != nil {
		return fmt.Errorf("read YAML error: %w", err)
	}
	for _, section := range sections {
		processSection(cmd, section)
	}
	return nil
}
