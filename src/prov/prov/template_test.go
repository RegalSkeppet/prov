package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const templateFileYAML = `
- task: template
  template: temp
  destination: DEST
  mode: 0664
`

const templateFileSource = "{{ .content }}"

func TestTemplate(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	destFilename := filepath.Join(dir, "dest")
	yamlFilename := filepath.Join(dir, "test.yml")
	err = ioutil.WriteFile(yamlFilename, []byte(strings.Replace(templateFileYAML, "DEST", destFilename, -1)), 0644)
	if err != nil {
		t.Fatal(err)
	}
	templateFilename := filepath.Join(dir, "temp")
	err = ioutil.WriteFile(templateFilename, []byte(templateFileSource), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = Provision(yamlFilename, map[interface{}]interface{}{"content": "test"}, true)
	if err != nil {
		t.Fatal(err)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		fmt.Println(file.Name())
	}
	content, err := ioutil.ReadFile(destFilename)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test" {
		t.Fatal("wrong content: ", string(content))
	}
}
