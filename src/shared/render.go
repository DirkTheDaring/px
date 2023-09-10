package shared

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func RenderOnConsole(outputs []map[string]interface{}, headers []string, filterColumn string, filterString string) {

	if len(headers) == 0 && len(outputs) > 0 {
		list := []string{}
		item := outputs[0]
		for key, _ := range item {
			list = append(list, key)
		}
		headers = list
	}

	rows := [][]any{}

	maxColSizes := make([]int, len(headers))
	for i, _ := range maxColSizes {
		maxColSizes[i] = len(headers[i])
	}
	for _, output := range outputs {
		if filterString != "" {
			value, _ := output[filterColumn].(string)
			if !strings.HasPrefix(value, filterString) {
				continue
			}
			//fmt.Fprintf(os.Stderr, "MATCH: >%v<\n", filterString)
			/*
						// hasPrefix which was used previously
				// had the weird effect that
				// on a list of
				// VM 111
				// VM 112
				// VM9999
				// if selected "VM " with a space, it matched
				// even VM9999 even if there is no space
				pos := strings.Index(value, filterString)
				if pos == -1 || pos != 0 {
					continue
				}
				fmt.Fprintf(os.Stderr, "MATCH: %v\n", value)
			*/
		}

		cols := []any{}
		for i, header := range headers {
			value, ok := output[header]
			if !ok {
				value = ""
			}
			valueString, ok := value.(string)
			if ok {
				if len(valueString) > maxColSizes[i] {
					maxColSizes[i] = len(valueString)
				}
				cols = append(cols, valueString)
				continue
			}

			valueFloat64, ok := value.(float64)
			if !ok {
				cols = append(cols, "")
				continue
			}
			valueInt := int(valueFloat64)
			valueString = strconv.Itoa(valueInt)
			if len(valueString) > maxColSizes[i] {
				maxColSizes[i] = len(valueString)
			}
			cols = append(cols, valueString)

		}
		//fmt.Fprintf(os.Stderr, "%v\n", cols)
		rows = append(rows, cols)
	}

	format := "%-" + strconv.Itoa(maxColSizes[0]) + "s"
	for i := 1; i < len(maxColSizes); i++ {
		format = format + " %-" + strconv.Itoa(maxColSizes[i]) + "s"
	}
	format = format + "\n"

	headers2 := []any{}
	for _, header := range headers {
		headers2 = append(headers2, strings.ToUpper(header))
	}

	fmt.Fprintf(os.Stdout, format, headers2...)

	for _, cols := range rows {
		fmt.Fprintf(os.Stdout, format, cols...)
	}

	//colSize = len(headers)
	//rowSize = len(outputs)
}
