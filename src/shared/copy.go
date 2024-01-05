package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
	pxapiobject "github.com/DirkTheDaring/px-api-client-internal-go"
)

func PtrToString(val reflect.Value, boolValues []string) (string, bool) {
	value := ""
	if val.Kind() != reflect.Ptr {
		//fmt.Fprintf(os.Stderr, "PtrToString() not reflect.Ptr\n")
		return value, false
	}
	elem := val.Elem()
	if !elem.IsValid() {
		//fmt.Fprintf(os.Stderr, "PtrToString() not valid\n")
		return value, false
	}

	kind := elem.Kind()
	if kind == reflect.String {
		value := elem.String()
		return value, true
	} else if kind == reflect.Bool {
		if boolValues == nil {
			boolValues = []string{"0", "1"}
		}
		if elem.Bool() {
			value = boolValues[1]
		} else {
			value = boolValues[0]
		}
		return value, true
	} else if kind == reflect.Int32 || kind == reflect.Int64 {
		int_value := elem.Int()
		value := strconv.FormatInt(int_value, 10)
		return value, true
	} else {
		fmt.Fprintf(os.Stderr, "PtrToString() unknown kind: %v\n", kind)
	}
	return value, false
}

// getJSONTag extracts the JSON tag name from the struct field.
func getJSONTag(fieldType reflect.StructField) string {
	tagParts := strings.Split(fieldType.Tag.Get("json"), ",")
	return tagParts[0]
}

// isSupportedKind checks if the field kind is supported for flattening.
func isSupportedKind(kind reflect.Kind) bool {
	return kind == reflect.Ptr || kind == reflect.String || kind == reflect.Bool
}

// formatField returns the string representation of the field.
func formatField(field reflect.Value, tagName string) (string, error) {
	switch field.Kind() {
	case reflect.Ptr:
		if value, ok := PtrToString(field, nil); ok {
			return tagName + "=" + value, nil
		}
	case reflect.String:
		return tagName + "=" + field.String(), nil
	case reflect.Bool:
		if field.Bool() {
			return tagName + "=1", nil
		}
		return tagName + "=0", nil
	}
	msg := fmt.Sprintf("Wrong kind: %+v\n", field.Kind())
	myError := errors.New(msg)
	return "", myError
}

// flatten converts a struct to a string representation, concatenating json tag names and values.
func flatten(v reflect.Value) string {
	var result strings.Builder
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		jsonTag := getJSONTag(v.Type().Field(i))

		if jsonTag == "" || !isSupportedKind(field.Kind()) {
			continue
		}

		str, err := formatField(field, jsonTag)
		if err != nil {
			continue
		}
		if result.Len() > 0 {
			result.WriteString(",")
		}
		result.WriteString(str)
	}

	str := result.String()

	return str
}

// copyGeneral copies fields from src to dst, handling different field types.
func copyGeneral(dst, src interface{}) {

	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField, dstField := srcVal.Type().Field(i), dstVal.Type().Field(i)
		if !fieldHasValue(reflect.ValueOf(src), srcField.Name) {
			continue
		}
		copyField(reflect.ValueOf(src), reflect.ValueOf(dst), srcField, dstField)

	}
}

// fieldHasValue checks if the field has a value set using the "Has" method.
func fieldHasValue(obj reflect.Value, fieldName string) bool {
	method := obj.MethodByName("Has" + fieldName)
	if method.IsValid() {
		return method.Call(nil)[0].Bool()
	}
	return true
}

// copyField copies a field from src to dst, considering type conversions.
func copyField(src, dst reflect.Value, srcField, dstField reflect.StructField) {
	result := src.MethodByName("Get" + srcField.Name).Call(nil)
	if srcField.Type == dstField.Type {
		dst.MethodByName("Set" + srcField.Name).Call(result)
	} else {
		val, err := handleTypeConversion(result[0])
		if err != nil {
			return
		}
		dst.MethodByName("Set" + srcField.Name).Call([]reflect.Value{reflect.ValueOf(val)})

	}
}

// handleTypeConversion handles type conversion for field values.
func handleTypeConversion(value reflect.Value) (interface{}, error) {
	switch value.Kind() {
	case reflect.Bool:
		if value.Bool() {
			return int32(1), nil
		}
		return int32(0), nil
	case reflect.Int64, reflect.Int32:
		// FIXME not sure if this right
		var val int64 = value.Int()
		return []reflect.Value{reflect.ValueOf(val)}, nil
	case reflect.Struct:
		return flatten(value), nil
	default:
		return nil, fmt.Errorf("unsupported kind: %s", value.Kind())
	}
}

/*
	func dumpDest(dst any) {
		dstVal := reflect.ValueOf(dst).Elem()
		for i := 0; i < dstVal.NumField(); i++ {
			dstField := dstVal.Type().Field(i)
			if !fieldHasValue(reflect.ValueOf(dst), dstField.Name) {
				continue
			}
			result := reflect.ValueOf(dst).MethodByName("Get" + dstField.Name).Call(nil)

			name := strings.ToLower(dstField.Name)

			switch result[0].Kind() {
			case reflect.String:
				fmt.Fprintf(os.Stderr, "%s: %s\n", name, result[0])
			case reflect.Int64, reflect.Int32:
				fmt.Fprintf(os.Stderr, "%s: %v\n", name, result[0])
			default:
				fmt.Fprintf(os.Stderr, "%s: %v (default)\n", name, result[0])
			}
		}
	}
*/
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// Absolute QUIRK to satisfy proxmox API
func escapeForProxmoxPerl(input string) string {
	escaped := url.PathEscape(input)
	escaped = strings.Replace(escaped, "@", "%40", -1)
	escaped = strings.Replace(escaped, "+", "%2B", -1)
	escaped = strings.Replace(escaped, "=", "%3D", -1)
	//fmt.Fprintf(os.Stderr, "BEFORE: %s\n", input)
	//fmt.Fprintf(os.Stderr, "AFTER : %s\n", escaped)
	return escaped
}

// escape additional chars, which perl on the server side of proxmox cannot handle
func postProcessFieldsForEscaping(dst any, slice []string) {
	dstVal := reflect.ValueOf(dst).Elem()
	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Type().Field(i)
		if !fieldHasValue(reflect.ValueOf(dst), dstField.Name) {
			continue
		}
		name := strings.ToLower(dstField.Name)
		if !containsString(slice, name) {
			continue
		}
		result := reflect.ValueOf(dst).MethodByName("Get" + dstField.Name).Call(nil)
		escaped := escapeForProxmoxPerl(result[0].String())
		reflect.ValueOf(dst).MethodByName("Set" + dstField.Name).Call([]reflect.Value{reflect.ValueOf(escaped)})
	}
}

func CopyVM(dst *pxapiflat.CreateVMRequest, src *pxapiobject.CreateVMRequest) {
	copyGeneral(dst, src)
	// postproces the fields that need extra conversion (ssh)
	rottenFields := []string{"sshkeys"}
	postProcessFieldsForEscaping(dst, rottenFields)
	// FIXME add a debug dump function
	//dumpDest(dst)
}
func CopyContainer(dst *pxapiflat.CreateContainerRequest, src *pxapiobject.CreateContainerRequest) {
	copyGeneral(dst, src)
	rottenFields := []string{"ssh-public-keys"}
	postProcessFieldsForEscaping(dst, rottenFields)
}

func CopyUpdateVMConfigRequest(dst *pxapiflat.UpdateVMConfigRequest, src *pxapiobject.UpdateVMConfigRequest) {
	copyGeneral(dst, src)
	rottenFields := []string{"sshkeys"}
	postProcessFieldsForEscaping(dst, rottenFields)
	jsonData, _ := json.Marshal(dst)
	fmt.Println(string(jsonData))
}

func CopyUpdateContainerConfigSyncRequest(dst *pxapiflat.UpdateContainerConfigSyncRequest, src *pxapiobject.UpdateContainerConfigSyncRequest) {
	copyGeneral(dst, src)
	rottenFields := []string{"sshkeys"}
	postProcessFieldsForEscaping(dst, rottenFields)
	jsonData, _ := json.Marshal(dst)
	fmt.Println(string(jsonData))
}
