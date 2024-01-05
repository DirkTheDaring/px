package configmap

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func DataToJSON(data map[string]interface{}) []byte {
	json, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return json
}
func DataToJSONByPtr(data *map[string]interface{}) []byte {
	if data == nil {
		return []byte{}
	}
	json, err := json.Marshal(*data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return json
}

func LoadEmbeddedYamlFile(data map[string]interface{}, files embed.FS, filename string) error {
	buffer, err := files.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return err
	}
	err = yaml.Unmarshal(buffer, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
	}
	return err
}

func GetInterfaceSliceValue(data map[string]interface{}, name string) ([]map[string]interface{}, error) {
	clusters, ok := data[name]
	if !ok {
		msg := fmt.Sprintf("could not find key: %s\n", name)
		err := errors.New(msg)
		return nil, err
	}
	switch v := clusters.(type) {
	case []interface{}:
		//fmt.Fprintf(os.Stderr, "FOUND %v\n", len(v))
		//l = len(v)
		result := []map[string]interface{}{}
		for _, item := range v {
			//fmt.Fprintf(os.Stderr, "%v (%T)\n", i, item)
			//fmt.Fprintf(os.Stderr, "%v\n", item)
			item2 := item.(map[string]interface{})
			result = append(result, item2)
		}
		return result, nil
	case []map[string]interface{}:
		return v, nil
	default:
		msg := fmt.Sprintf("could not cast: %T for %s\n", v, name)
		err := errors.New(msg)
		return nil, err
	}
}

func GetInterfaceSliceValueByPtr(data *map[string]interface{}, name string) ([]map[string]interface{}, error) {
	if data == nil {
		return nil, errors.New("data map is nil")
	}

	clusters, ok := (*data)[name]
	if !ok {
		return nil, fmt.Errorf("could not find key: %s", name)
	}

	switch v := clusters.(type) {
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("item is not a map[string]interface{}: %T", item)
			}
			result = append(result, itemMap)
		}
		return result, nil
	case []map[string]interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("could not cast: %T for %s", v, name)
	}
}
