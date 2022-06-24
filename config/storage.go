package config

import (
	"bytes"
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

type Host struct {
	Name     string
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func LoadHosts() ([]Host, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(pathTo(homedir, ".config", "lazysql"), os.ModePerm)
	if err != nil {
		return nil, err
	}
	filepath := pathTo(homedir, ".config", "lazysql", "hosts.yaml")
	filecontent, err := os.ReadFile(filepath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Host{
				{
					Host: "localhost",
					Port: 3306,
				},
			}, nil
		}
		return nil, err
	}
	return unmarshalHosts([]byte(filecontent))
}

func SaveHosts(hosts []Host) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(pathTo(homedir, ".config", "lazysql"), os.ModePerm)
	if err != nil {
		return err
	}
	filepath := pathTo(homedir, ".config", "lazysql", "hosts.yaml")
	marshalledHosts := marshalHosts(hosts)
	err = os.WriteFile(filepath, []byte(marshalledHosts), os.ModePerm)
	return err
}

func marshalHosts(hosts []Host) string {
	node := hostsToYamlNode(hosts)
	buffer := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	err := encoder.Encode(node)

	if err != nil {
		panic(err)
	}
	return buffer.String()
}

func hostsToYamlNode(hosts []Host) *yaml.Node {
	hostNodes := make([]*yaml.Node, len(hosts)*2)
	for i, host := range hosts {
		hostNodes[i*2] = &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: host.Name,
		}
		hostNodes[(i*2)+1] = &yaml.Node{}
		hostNodes[(i*2)+1].Encode(host)
		hostNodes[(i*2)+1].Content = hostNodes[(i*2)+1].Content[2:]
	}
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "hosts",
			},
			{
				Kind:    yaml.MappingNode,
				Content: hostNodes,
			},
		},
	}
}

// TODO check that host names are unique
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

func pathTo(pathComponents ...string) (result string) {
	separator := string(os.PathSeparator)
	for i, component := range pathComponents {
		if i != 0 {
			result += separator
		}
		result += component
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
