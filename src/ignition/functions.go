package ignition

import (
	"encoding/json"
	"strconv"
	"strings"
	"text/template"
)

func templateSum(a int, b int) string {
	result := strconv.Itoa(a + b)
	return result
}
func templateJoin(i []interface{}, sep string) string {
	array := []string{}
	for _, item := range i {
		array = append(array, item.(string))
	}
	result := strings.Join(array, sep)
	return result
}
func templateLen(i []interface{}) int {
	return len(i)
}
func templateEq(a string, b string) bool {
	return a == b
}
func templateToJson(a string) string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	s := string(b)
	return s[1 : len(s)-1]
}

func GetTemplateFuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"sum":    templateSum,
		"join":   templateJoin,
		"len":    templateLen,
		"eq":     templateEq,
		"toJson": templateToJson,
	}
	return funcMap
}
