package shared

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
)

// RenderOnConsole displays data in a table format on the console.
func RenderOnConsoleWithFilter(outputs []map[string]interface{}, headers []string, filterColumn, filterString string) {
	if len(outputs) == 0 {
		return
	}
	filteredOutputs, _ := filterOutputs(outputs, filterColumn, filterString)

	RenderOnConsoleNew(filteredOutputs, headers, nil)
}

func RenderOnConsoleNew(outputs []map[string]interface{}, headers []string, rightAlignments []string) {
	if len(outputs) == 0 {
		return
	}

	if len(headers) == 0 {
		headers = extractHeadersFromOutputs(outputs)
	}

	if rightAlignments == nil {
		rightAlignments = determineRightAlignments(outputs, headers)
	}

	columnWidths := calculateColumnWidths(outputs, headers)

	PrintTable(headers, outputs, columnWidths, rightAlignments)
}

func extractHeadersFromOutputs(outputs []map[string]interface{}) []string {
	var headers []string
	for key := range outputs[0] {
		headers = append(headers, key)
	}
	return headers
}

// heuristic, deterimine right alignment by looking at first line
func determineRightAlignments(outputs []map[string]interface{}, headers []string) []string {
	var rightAlignments []string

	// get first line from outputs
	firstLine := outputs[0]

	for _, key := range headers {
		//fmt.Fprintf(os.Stderr, "header: %s\n", key)
		value, found := firstLine[key]
		if !found {
			continue
		}
		_, ok := value.(string)
		if ok {
			continue
		}
		rightAlignments = append(rightAlignments, key)
	}
	return rightAlignments
}

func filterOutputs(outputs []map[string]interface{}, filterColumn, filterString string) ([]map[string]interface{}, error) {
	if filterColumn == "" {
		return nil, fmt.Errorf("filterColumn cannot be empty")
	}

	var filtered []map[string]interface{}
	for _, output := range outputs {
		value, ok := output[filterColumn]
		if !ok {
			return nil, fmt.Errorf("filterColumn %s does not exist in output", filterColumn)
		}

		valueStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("value for %s is not a string", filterColumn)
		}

		if !strings.HasPrefix(valueStr, filterString) {
			continue
		}

		filtered = append(filtered, output)
	}
	return filtered, nil
}

func calculateColumnWidths(outputs []map[string]interface{}, headers []string) []int {
	columnWidths := make([]int, len(headers))
	for i, header := range headers {
		columnWidths[i] = len(header)
		for _, output := range outputs {
			value := fmt.Sprintf("%v", output[header])
			if len(value) < columnWidths[i] {
				continue
			}
			columnWidths[i] = len(value)
		}
	}
	return columnWidths
}

// GenerateTableString builds a table string from the given data.
func GenerateTableString(headers []string, outputs []map[string]interface{}, columnWidths []int, rightAlignments []string) string {
	var sb strings.Builder
	lineEnding := getLineEnding()

	// Header
	for i, header := range headers {
		sb.WriteString(formatColumn(strings.ToUpper(header), columnWidths[i], isRightAligned(header, rightAlignments)))
	}
	sb.WriteString(lineEnding)

	// Rows
	for _, output := range outputs {
		for i, header := range headers {
			var value string
			if output[header] == nil {
				value = "~"
			} else {
				value = fmt.Sprintf("%v", output[header])
			}

			sb.WriteString(formatColumn(value, columnWidths[i], isRightAligned(header, rightAlignments)))
		}
		sb.WriteString(lineEnding)
	}

	return sb.String()
}

// PrintTable prints the table to the console.
// It uses GenerateTableString to build the table string.
func PrintTable(headers []string, outputs []map[string]interface{}, columnWidths []int, rightAlignments []string) {
	tableString := GenerateTableString(headers, outputs, columnWidths, rightAlignments)
	fmt.Print(tableString)
}

// getLineEnding returns the appropriate line ending character sequence based on the operating system.
func getLineEnding() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func formatColumn(value string, width int, rightAlign bool) string {
	if rightAlign {
		return fmt.Sprintf("%*s ", width, value)
	}
	return fmt.Sprintf("%-*s ", width, value)
}

func isRightAligned(header string, rightAlignments []string) bool {
	return slices.Contains(rightAlignments, header)
}
