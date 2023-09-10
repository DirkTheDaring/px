package clusters

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"px/configmap"
	"px/proxmox/aliases"

	"gopkg.in/yaml.v3"
)

func FindConfigDir(path string) string {
	configPath := filepath.Join(path, pxConfigFileDir)
	fileInfo, err := os.Stat(configPath)
	if err == nil && fileInfo.IsDir() {
		return configPath
	}
	parent := filepath.Dir(path)
	if parent == path {
		return ""
	}
	return FindConfigDir(parent)
}

func FindConfigFile() string {
	configFile := os.Getenv(pxConfigFilenameVar)
	if configFile != "" {
		fileInfo, err := os.Stat(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "file referenced in %s does not exist: %s\n", pxConfigFilenameVar, configFile)
			os.Exit(1)
		}
		if fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "file referenced in %s is not a file\n", pxConfigFilenameVar)
			os.Exit(1)
		}
		return configFile
	}
	workingDirectory, err := os.Getwd()
	//fmt.Fprintf(os.Stderr, "workingDirectory: %s\n", workingDirectory)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	configDir := FindConfigDir(workingDirectory)
	if configDir == "" {
		fmt.Fprintf(os.Stderr, "config  dir not found.")
		os.Exit(1)
	}
	//fmt.Fprintf(os.Stderr, "configDir: %s\n", configDir)

	configFilePath := filepath.Join(configDir, pxConfigFileName)
	//fmt.Fprintf(os.Stderr, "configFilePath: %s\n", configFilePath)
	return configFilePath
}

func LoadYamlFile(data map[string]interface{}, filename string) bool {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return false
	}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return false
	}
	return true
}
func MergeDefaultWithSection(defaultData map[string]interface{}, configData map[string]interface{}, sectionName string) {
	section, ok := configmap.GetMapEntry(configData, sectionName)
	if !ok {
		return
	}
	result := configmap.MergeMapRecursive(defaultData, section)
	configData[sectionName] = result

}

func LoadClusterDefault(data map[string]interface{}) map[string]interface{} {
	machineData := map[string]interface{}{}
	filepath := "files/cluster.yaml"
	err := configmap.LoadEmbeddedYamlFile(machineData, files, filepath)
	fmt.Fprintf(os.Stderr, "error: %v\n", err)

	return machineData
}

func LoadConfigDefault() map[string]interface{} {
	configDefault := map[string]interface{}{}
	filepath := "files/config.yaml"
	err := configmap.LoadEmbeddedYamlFile(configDefault, files, filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return configDefault
}

func MergeDefaults(configData map[string]interface{}) map[string]interface{} {

	configDefault := LoadConfigDefault()
	configData = configmap.MergeMapRecursive(configDefault, configData)

	defaultAliases := aliases.InitAliases2()
	MergeDefaultWithSection(defaultAliases, configData, "aliases")

	// Defaults
	clusterDefault := map[string]interface{}{}
	filepath := "files/cluster.yaml"
	err := configmap.LoadEmbeddedYamlFile(clusterDefault, files, filepath)
	//fmt.Fprintf(os.Stderr, "error: %v\n", err)

	// Defaults
	nodeDefault := map[string]interface{}{}
	filepath = "files/node.yaml"
	err = configmap.LoadEmbeddedYamlFile(nodeDefault, files, filepath)
	//fmt.Fprintf(os.Stderr, "error: %v\n", err)

	clusters, err := configmap.GetInterfaceSliceValue(configData, "clusters")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get clusters: %v\n", err)
	}
	newClusters := []map[string]interface{}{}
	for _, cluster := range clusters {
		//fmt.Fprintf(os.Stderr, "%v\n", cluster["name"])
		cluster = configmap.MergeMapRecursive(clusterDefault, cluster)

		newNodes := []map[string]interface{}{}
		nodes, _ := configmap.GetInterfaceSliceValue(cluster, "nodes")
		for _, node := range nodes {
			node = configmap.MergeMapRecursive(nodeDefault, node)
			newNodes = append(newNodes, node)
		}
		cluster["nodes"] = newNodes
		newClusters = append(newClusters, cluster)
	}
	configData["clusters"] = newClusters

	// now the inheritance of aliases, ignition, selectors
	return configData
}
func InheritData(sectionName string, configData map[string]interface{}, rootVars []string) map[string]interface{} {

	newClusters := []map[string]interface{}{}
	clusters, err := configmap.GetInterfaceSliceValue(configData, sectionName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILED!!!! %s\n", err)
	}
	for _, cluster := range clusters {
		var content map[string]interface{}
		for _, keyName := range rootVars {
			parentContent, ok := configmap.GetMapEntry(configData, keyName)
			if !ok {
				fmt.Fprintf(os.Stderr, "InheritData() key not found in configData: %s\n", keyName)
				continue
			}
			childContent, ok := configmap.GetMapEntry(cluster, keyName)
			if !ok {
				fmt.Fprintf(os.Stderr, "InheritData() key not found in cluster: %s\n", keyName)
				continue
			}
			content = configmap.MergeMapRecursive(parentContent, childContent)

			cluster[keyName] = content
		}
		newClusters = append(newClusters, cluster)
	}
	configData[sectionName] = newClusters
	return configData
}

func GetSystemConfiguration() map[string]interface{} {
	configFile := FindConfigFile()
	//fmt.Fprintf(os.Stderr, "configFile: %s\n", configFile)
	configData := map[string]interface{}{}
	ok := LoadYamlFile(configData, configFile)
	if ok {

	}
	configData = MergeDefaults(configData)
	rootVars := []string{"aliases", "ignition", "selectors", "vars"}
	configData = InheritData("clusters", configData, rootVars)

	newClusters := []map[string]interface{}{}
	clusters, _ := configmap.GetInterfaceSliceValue(configData, "clusters")
	for _, cluster := range clusters {
		newNodes := []map[string]interface{}{}
		nodes, _ := configmap.GetInterfaceSliceValue(cluster, "nodes")
		for _, node := range nodes {
			var content map[string]interface{}
			for _, keyName := range rootVars {
				parentContent, ok := configmap.GetMapEntry(cluster, keyName)
				if !ok {
					fmt.Fprintf(os.Stderr, "key not found in cluster: %s\n", keyName)
					continue
				}
				childContent, ok := configmap.GetMapEntry(node, keyName)
				if !ok {
					fmt.Fprintf(os.Stderr, "key not found in node: %s\n", keyName)
					continue
				}
				content = configmap.MergeMapRecursive(parentContent, childContent)
				node[keyName] = content
			}
			newNodes = append(newNodes, node)
		}
		cluster["nodes"] = newNodes
		newClusters = append(newClusters, cluster)
	}
	configData["clusters"] = newClusters
	return configData
}
