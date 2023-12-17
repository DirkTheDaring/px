package configmap

import (
	"embed"
	"fmt"
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
