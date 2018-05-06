package main

import (
	"fmt"
	"github.com/yanc0/tamarin/lib"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Panic("error get working directory:", err.Error())
	}

	baseDir = filepath.Join(baseDir, "example-app")

	fileList := []string{}
	err = filepath.Walk(baseDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		log.Panic("error lookup files:", err.Error())
	}

	// Every file is a key, the paths where the file have
	// been seen if referenced
	var m = make(map[string][]string)
	for _, file := range fileList {
		dir, file := filepath.Split(file)
		m[file] = append(m[file], dir)
	}

	for file := range m {
		sort.Sort(stringsByLen(m[file]))
	}

	data0, err := ioutil.ReadFile(filepath.Join(m["deployment.yml"][0], "deployment.yml"))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	data1, err := ioutil.ReadFile(filepath.Join(m["deployment.yml"][1], "deployment.yml"))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var y0 map[interface{}]interface{}
	err = yaml.Unmarshal([]byte(data0), &y0)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var y1 map[interface{}]interface{}
	err = yaml.Unmarshal([]byte(data1), &y1)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	t0 := lib.MapToTree(y0)
	t1 := lib.MapToTree(y1)

	t0.MergeTree(t1)

	y, err := yaml.Marshal(t0.ToMap())
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(string(y))

}

type stringsByLen []string

func (a stringsByLen) Len() int {
	return len(a)
}

func (a stringsByLen) Less(i, j int) bool {
	return len(a[i]) < len(a[j])
}

func (a stringsByLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
