package config

import (
	"errors"
	"os"

	// "fmt"

	"gopkg.in/yaml.v3"
)

type Host struct {
	Name     string
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func LoadHosts() ([]Host, error) {
	var filecontent, err = os.ReadFile("./config.yaml")
	if err != nil {
		return nil, err
	}
	return unmarshalHosts([]byte(filecontent))
}

func unmarshalHosts(filecontent []byte) (result []Host, err error) {
	var node yaml.Node
	err = yaml.Unmarshal(filecontent, &node)
	if err != nil {
		return
	}

	if len(node.Content) != 1 ||
		node.Content[0].Kind != yaml.MappingNode ||
		node.Content[0].Content[0].Kind != yaml.ScalarNode ||
		node.Content[0].Content[0].Value != "hosts" ||
		node.Content[0].Content[1].Kind != yaml.MappingNode {
		err = errors.New("no top level `hosts` key in config.yaml")
		return
	}

	hostsNode, err := parsYamlMappingNode(node.Content[0].Content[1])
	result = make([]Host, len(hostsNode))
	for i, node := range hostsNode {
		if node.value.Kind != yaml.MappingNode {
			err = errors.New("hosts should be maps")
			return
		}
		err = node.value.Decode(&result[i])
		if err != nil {
			return
		}
		result[i].Name = node.key
		if len(result[i].Host) == 0 {
			result[i].Host = "localhost"
		}
	}

	return
}

func parsYamlMappingNode(node *yaml.Node) ([]yamlMappingNode, error) {
	if node.Kind != yaml.MappingNode {
		panic("Cannot parse non mapping node as mapping node!")
	}
	length := len(node.Content) / 2
	result := make([]yamlMappingNode, length)
	for i := 0; i < length; i++ {
		err := node.Content[i*2].Decode(&result[i].key)
		if err != nil {
			return nil, err
		}
		result[i].value = node.Content[(i*2)+1]
	}
	return result, nil
}

type yamlMappingNode struct {
	key   string
	value *yaml.Node
}
