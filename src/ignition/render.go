package ignition

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"text/template"
)

func LoadEmbeddedYamlFile(filename string) map[string]interface{} {
	data := map[string]interface{}{}
	buffer, err := defaults.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return data
	}
	err = yaml.Unmarshal(buffer, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return data
	}
	return data
}

func RenderEmbeddedTemplate(outStream io.Writer, filename string, data map[string]interface{}) bool {
	basename := path.Base(filename)

	//fmt.Fprintf(os.Stderr, "RenderEmbeddedTemplate() %s %s\n", filename, basename)

	t := template.New(basename).Funcs(GetTemplateFuncMap())
	tmpl, err := t.ParseFS(templates, filename, "templates/*.gotext")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}
	err = tmpl.Execute(outStream, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}
	return true
}

func RenderFileTemplate(outStream io.Writer, filename string, data interface{}) bool {

	basename := path.Base(filename)
	tmpl, err := template.New(basename).Funcs(GetTemplateFuncMap()).ParseFiles(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}
	tmpl.Execute(outStream, data)

	return true
}
func RenderIgnitionYaml(outStream io.Writer, data map[string]interface{}) {
	value, _ := data["template"]
	if value != "" {
		filename := value.(string)
		RenderFileTemplate(outStream, filename, data)
	} else {
		filename := "templates/ignition.yaml"
		RenderEmbeddedTemplate(outStream, filename, data)
	}
}
func RenderString(outStream io.Writer, input string, data map[string]interface{}) bool {
	input = "{{.Count}} items are made of {{.Material}}"
	tmpl, err := template.New("test").Parse(input)
	if err != nil {
		return false
	}
	err = tmpl.Execute(outStream, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}
	return true
}
