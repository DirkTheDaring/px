package documents

import (
	"fmt"
)

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
