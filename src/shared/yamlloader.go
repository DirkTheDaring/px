package shared

import (
	"bytes"
	"io"
	"io/ioutil"
	"fmt"

	"gopkg.in/yaml.v3"
)

// splitYAML takes a byte slice of a YAML and splits it into individual documents.
func splitYAML(resources []byte) ([][]byte, error) {
    dec := yaml.NewDecoder(bytes.NewReader(resources))
    var res [][]byte
    for {
        var value interface{}
        err := dec.Decode(&value)
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("error decoding YAML: %w", err)
        }
        valueBytes, err := yaml.Marshal(value)
        if err != nil {
            return nil, fmt.Errorf("error marshaling YAML: %w", err)
        }
        res = append(res, valueBytes)
    }
    return res, nil
}

// ReadYAMLWithDashDashDashSingle reads a YAML file specified by filename and
// returns a slice of maps. Each map represents a YAML document.
func ReadYAMLWithDashDashDashSingle(filename string) ([]map[string]interface{}, error) {
    yamlFile, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("error reading file %s: %w", filename, err)
    }

    datas, err := splitYAML(yamlFile)
    if err != nil {
        return nil, fmt.Errorf("error splitting YAML: %w", err)
    }

    var list []map[string]interface{}
    for _, data := range datas {
        var iMap map[string]interface{}
        err = yaml.Unmarshal(data, &iMap)
        if err != nil {
            return nil, fmt.Errorf("error unmarshaling YAML: %w", err)
        }
        list = append(list, iMap)
    }

    return list, nil
}

