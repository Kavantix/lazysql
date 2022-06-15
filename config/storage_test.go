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
  `

	hosts, err := unmarshalHosts([]byte(yaml))
	if err != nil {
		t.Fatalf("Unexpected error while ummarshaling %s", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("Wrong number of hosts: %d", len(hosts))
	}
	host1 := hosts[0]
	if host1.Name != "name1" ||
		host1.Port != "3306" ||
		host1.User != "gggg" ||
		host1.Password != "doesntmatter" {
		t.Fatalf("Incorrect host parsed `%#v`", host1)
	}
}
