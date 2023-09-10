package configmap

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"text/template"
)

func PrepareEmbeddedTemplate(files embed.FS, fileglob []string) (*template.Template, error) {
	basename := path.Base(fileglob[0])
	fmt.Fprintf(os.Stderr, "%s %s\n", fileglob, basename)
	//t := files.New(basename).Funcs(GetTemplateFuncMap())
	t := template.New(basename)
	tmpl, err := t.ParseFS(files, fileglob...)
	return tmpl, err
}
func PrepareFileTemplate(filename string) (*template.Template, error) {
	basename := path.Base(filename)
	//tmpl, err := template.New(basename).Funcs(GetTemplateFuncMap()).ParseFiles(filename)
	tmpl, err := template.New(basename).ParseFiles(filename)
	return tmpl, err
}
func GetEnvWithDefault(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}
func PrepareTemplate(files embed.FS, fileglob []string, varname string) (*template.Template, error) {
	filename := GetEnvWithDefault(varname, fileglob[0])
	tmpl, err := PrepareFileTemplate(filename)
	if err == nil {
		return tmpl, err
	}
	tmpl, err = PrepareEmbeddedTemplate(files, fileglob)
	return tmpl, err

}
func RenderTemplate(outStream io.Writer, files embed.FS, fileglob []string, varname string, data map[string]interface{}) bool {
	tmpl, err := PrepareTemplate(files, fileglob, varname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "RenderTemplate() error: %v\n", err)
		return false
	}
	err = tmpl.Execute(outStream, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}
	return true
}
