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


// flatten converts a struct to a string representation.
// It processes each field of the struct and concatenates their json tag names
// and values into a comma-separated string.
func flatten(v reflect.Value) string {

	if v.Kind() != reflect.Struct {
		return ""
	}

	var result strings.Builder
	//fmt.Printf("number of fields: %+v\n", v.NumField())

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		tagName := strings.Split(fieldType.Tag.Get("json"), ",")[0]

		if tagName == "" {
			//fmt.Printf("no tagname\n")
			continue
		}


		element := ""
		//fmt.Printf("field kind: %v\n", field.Kind())
		switch field.Kind() {
		case reflect.Ptr:
			value, ok := PtrToString(field, nil)
			if !ok {
				continue
			}
			element = tagName + "=" + value

		case reflect.String:
			element = tagName + "=" + field.String()

		case reflect.Bool:
			if field.Bool() {
				element = tagName + "=1"
			} else {
				element = tagName + "=0"
			}
		default:
			//fmt.Printf("type not found: %v\n", field.Kind())
			continue
		
		}

		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(element)


	}

	return result.String()
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
			//fmt.Printf("XXX_VAL %+v\n", result[0].Kind())

			if result[0].Kind() == reflect.Bool {
				var val int32 = 0
				if result[0].Bool() {
					val = 1
				} else {
					val = 0
				}

				//fmt.Printf("YYY_VAL %+v\n", val)
				params := []reflect.Value{reflect.ValueOf(val)}
				dst_p.MethodByName("Set" + src_field.Name).Call(params)
			} else {
				val := flatten(result[0])
				params := []reflect.Value{reflect.ValueOf(val)}
				dst_p.MethodByName("Set" + src_field.Name).Call(params)
			}
			//fmt.Printf("VAL %+v\n", result[0].CanInterface())
			//fmt.Printf("VAL %+v\n", result[0].Interface())

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
