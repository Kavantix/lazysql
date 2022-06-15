package config

import (
	"testing"
)

func TestUnmarshalHosts(t *testing.T) {
	yaml := `
hosts:
  name1:
    host: fwfw
    port: 3306
    user: gggg
    password: doesntmatter
  name2:
    host:
    port: 3306
    user: gggg
    password: doesntmatter
`

	hosts, err := unmarshalHosts([]byte(yaml))
	if err != nil {
		t.Fatalf("Unexpected error while ummarshaling %s", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("Wrong number of hosts: %d", len(hosts))
	}
	host1 := hosts[0]
	if host1.Name != "name1" ||
		host1.Host != "fwfw" ||
		host1.Port != "3306" ||
		host1.User != "gggg" ||
		host1.Password != "doesntmatter" {
		t.Fatalf("Incorrect host1 parsed `%#v`", host1)
	}
	host2 := hosts[1]
	if host2.Name != "name2" ||
		host2.Host != "localhost" ||
		host2.Port != "3306" ||
		host2.User != "gggg" ||
		host2.Password != "doesntmatter" {
		t.Fatalf("Incorrect host2 parsed `%#v`", host2)
	}

	outputYaml := marshalHosts(hosts)

	expectedYaml := `
hosts:
  name1:
    host: fwfw
    port: "3306"
    user: gggg
    password: doesntmatter
  name2:
    host: localhost
    port: "3306"
    user: gggg
    password: doesntmatter
`[1:] // remove starting newline
	if outputYaml != expectedYaml {
		t.Fatalf("Marshaling failed!\n output yaml:`\n%s`\ninstead of:`\n%s`", outputYaml, expectedYaml)
	}
}
