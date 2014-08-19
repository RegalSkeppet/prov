package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"prov"
	"strings"
	"testing"
)

const copyFileYAML = `
tasks:
  - task: copy file
    source: src
    destination: DEST
    mode: 0664
`

const copyFileSource = "content"

func TestCopyFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	destFilename := filepath.Join(dir, "dest")
	yamlFilename := filepath.Join(dir, "test.yml")
	err = ioutil.WriteFile(yamlFilename, []byte(strings.Replace(copyFileYAML, "DEST", destFilename, -1)), 0644)
	if err != nil {
		t.Fatal(err)
	}
	srcFilename := filepath.Join(dir, "src")
	err = ioutil.WriteFile(srcFilename, []byte(copyFileSource), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = Provision(yamlFilename, make(prov.Vars), true, false)
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
	if string(content) != copyFileSource {
		t.Fatal("wrong content: ", string(content))
	}
	stat, err := os.Stat(destFilename)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Mode().Perm() != 0664 {
		t.Fatal("wrong mode: ", stat.Mode().Perm())
	}
}
