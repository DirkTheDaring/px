package aliases

import (
	"fmt"
	"os"
	"px/configmap"
)

func InitAliases() {
	pxAliases = InitAliases2()
}
func InitAliases2() map[string]interface{} {
	filename := "files/aliases.yaml"
	aliases := map[string]interface{}{}
	configmap.LoadEmbeddedYamlFile(aliases, px, filename)
	return aliases
}
func LookupAlias(name string, available_storages []string) string {
	if pxAliases == nil {
		InitAliases()
	}
	fmt.Fprintf(os.Stderr, "LookupAlias() %v %v\n", name, available_storages)
	if len(available_storages) == 0 {
		return name
	}

	matches := configmap.GetStringSliceWithDefault(pxAliases, name, []string{})
	fmt.Fprintf(os.Stderr, "LookupAlias() matches %v\n", matches)
	if len(matches) == 0 {
		return name
	}
	for _, available_storage := range available_storages {
		for _, match := range matches {
			if available_storage == match {
				return available_storage
			}
		}
	}
	return name

}
