package shared

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"github.com/DirkTheDaring/px-api-client-go"
	"github.com/DirkTheDaring/px-api-client-internal-go"

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
		fmt.Fprintf(os.Stderr, "PtrToString() unknown kind: s\n", kind)
	}
	return value, false
}

func flatten(myval reflect.Value) string {
	if myval.Kind() != reflect.Struct {
		return ""
	}
	iface := myval.Interface()
	valueOf := reflect.ValueOf(iface)

	result := ""

	for i := 0; i < valueOf.NumField(); i++ {
		struct_field := valueOf.Type().Field(i)
		varname := struct_field.Name
		tagName := strings.Split(struct_field.Tag.Get("json"), ",")[0]
		//fmt.Printf("  tagName: %s\n", tagName)
		val := valueOf.FieldByName(varname)
		element := ""
		if val.Kind() == reflect.Ptr {
			value, ok := PtrToString(val, nil)
			if !ok {
				continue
			}
			element = tagName + "=" + value
			//fmt.Fprintf(os.Stderr, "elem: %s\n", element)

		} else if val.Kind() == reflect.String {
			element = tagName + "=" + val.String()
		} else if val.Kind() == reflect.Bool {
			if val.Bool() {
				element = tagName + "=1"

			} else {
				element = tagName + "=0"

			}
		}
		if result == "" {
			result = element
		} else {
			result = result + "," + element
		}

	}
	//fmt.Printf("  VALUES %s\n", result)
	return result
}

// https://pkg.go.dev/reflect#Zero
func copyGeneral(dst interface{}, src interface{}) {

	src_p := reflect.ValueOf(src)
	dst_p := reflect.ValueOf(dst)

	src_e := src_p.Elem()
	dst_e := dst_p.Elem()

	for i := 0; i < src_e.NumField(); i++ {
		src_field := src_e.Type().Field(i)
		dst_field := dst_e.Type().Field(i)

		method := src_p.MethodByName("Has" + src_field.Name)

		// Skip all valid methods which do not have a value set
		if method.IsValid() {
			hasValue := method.Call([]reflect.Value{})
			if !hasValue[0].Bool() {
				continue
			}
		}
		result := src_p.MethodByName("Get" + src_field.Name).Call(nil)

		if src_field.Type == dst_field.Type {
			//fmt.Printf("KEY %+v\n", src_field.Name)
			//fmt.Printf("VAL %+v\n", result)
			dst_p.MethodByName("Set" + src_field.Name).Call(result)
		} else {

			//fmt.Printf("XXX_KEY %+v\n", src_field.Name)
			//fmt.Printf("VAL %+v\n", result)
			//fmt.Printf("VAL %+v\n", result[0].CanInterface())
			//fmt.Printf("VAL %+v\n", result[0].Interface())
			val := flatten(result[0])
			params := []reflect.Value{reflect.ValueOf(val)}
			dst_p.MethodByName("Set" + src_field.Name).Call(params)
		}

	}
}
func CopyVM(dst *pxapiflat.CreateVMRequest, src *pxapiobject.CreateVMRequest) {
	copyGeneral(dst, src)
}
func CopyContainer(dst *pxapiflat.CreateContainerRequest, src *pxapiobject.CreateContainerRequest) {
	copyGeneral(dst, src)
}

func CopyUpdateVMConfigRequest(dst *pxapiflat.UpdateVMConfigRequest, src *pxapiobject.UpdateVMConfigRequest) {
	copyGeneral(dst, src)
}
