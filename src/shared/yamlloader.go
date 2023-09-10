package shared

import (
	"bytes"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func SplitYAML(resources []byte) ([][]byte, error) {

	dec := yaml.NewDecoder(bytes.NewReader(resources))
	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}

func ReadYamlWithDashDashDashSingle(filename string) ([]map[string]interface{}, error) {

	list := []map[string]interface{}{}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	datas, err := SplitYAML(yamlFile)
	if err != nil {
		return nil, err
	}
	for _, data := range datas {
		iMap := map[string]interface{}{}
		err = yaml.Unmarshal(data, &iMap)
		if err != nil {
			return nil, err
		}
		list = append(list, iMap)
	}
	return list, nil
}
