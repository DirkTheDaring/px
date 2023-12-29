package shared

import (
	"fmt"
	"slices"
	"strings"
)

// RenderOnConsole displays data in a table format on the console.
func RenderOnConsole(outputs []map[string]interface{}, headers []string, filterColumn, filterString string) {
	RenderOnConsoleNew(outputs, headers, filterColumn, filterString, nil)
}

// RenderOnConsoleNew is an enhanced version of RenderOnConsole with right alignment options.
func RenderOnConsoleNew(outputs []map[string]interface{}, headers []string, filterColumn, filterString string, rightAlignments []string) {
	if len(headers) == 0 {
		headers = extractHeadersFromOutputs(outputs)
	}

	filteredOutputs := filterOutputs(outputs, filterColumn, filterString)
	if len(filteredOutputs) == 0 {
		return
	}

	columnWidths := calculateColumnWidths(filteredOutputs, headers)
	printTable(headers, filteredOutputs, columnWidths, rightAlignments)
}

func extractHeadersFromOutputs(outputs []map[string]interface{}) []string {
	if len(outputs) == 0 {
		return nil
	}

	var headers []string
	for key := range outputs[0] {
		headers = append(headers, key)
	}
	return headers
}

func filterOutputs(outputs []map[string]interface{}, filterColumn, filterString string) []map[string]interface{} {
	if filterString == "" {
		return outputs
	}

	var filtered []map[string]interface{}
	for _, output := range outputs {
		if value, ok := output[filterColumn].(string); ok && strings.HasPrefix(value, filterString) {
			filtered = append(filtered, output)
		}
	}
	return filtered
}

func calculateColumnWidths(outputs []map[string]interface{}, headers []string) []int {
	columnWidths := make([]int, len(headers))
	for i, header := range headers {
		columnWidths[i] = len(header)
		for _, output := range outputs {
			value := fmt.Sprintf("%v", output[header])
			if len(value) > columnWidths[i] {
				columnWidths[i] = len(value)
			}
		}
	}
	return columnWidths
}

func printTable(headers []string, outputs []map[string]interface{}, columnWidths []int, rightAlignments []string) {
	var sb strings.Builder

	// Header
	for i, header := range headers {
		sb.WriteString(formatColumn(strings.ToUpper(header), columnWidths[i], isRightAligned(header, rightAlignments)))
	}
	sb.WriteRune('\n')

	// Rows
	for _, output := range outputs {
		for i, header := range headers {
			value := fmt.Sprintf("%v", output[header])
			sb.WriteString(formatColumn(value, columnWidths[i], isRightAligned(header, rightAlignments)))
		}
		sb.WriteRune('\n')
	}

	fmt.Print(sb.String())
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
