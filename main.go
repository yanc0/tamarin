package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
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

	printTree(recurseMapTree(y0), "-")
	//fmt.Println(branchesToMap(mergeData(y0, y1)))
}

func branchesToMap(branches []branch) map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	for _, b := range branches {
		key := b.path[0]
		m[key] = branchToMap(b)[key]
	}
	return m
}

func branchToMap(b branch) map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	if len(b.path) == 1 {
		key := b.path[0]
		m[key] = b.leaf
	} else {
		key := b.path[0]
		b.path = append(b.path[:0], b.path[1:]...)
		m[key] = branchToMap(b)
	}
	return m
}

func mergeData(src map[interface{}]interface{}, merge map[interface{}]interface{}) []branch {
	srcBranches := recurseMap(src)
	mergeBranches := recurseMap(merge)
	for i, srcB := range srcBranches {
		for _, mergeB := range mergeBranches {
			if isSameBranchPath(srcB, mergeB) {
				srcBranches[i].leaf = mergeB.leaf
			}
		}
	}

	for _, mergeB := range mergeBranches {
		if !isBranchPresent(mergeB, srcBranches) {
			srcBranches = append(srcBranches, mergeB)
		}
	}

	return srcBranches
}

func isBranchPresent(a branch, branches []branch) bool {
	for _, b := range branches {
		if isSameBranchPath(a, b) {
			return true
		}
	}
	return false
}

func isSameBranchPath(a branch, b branch) bool {
	for i := range a.path {
		if a.path[i] != b.path[i] {
			return false
		}
	}
	return true
}

func recurseMap(m map[interface{}]interface{}) []branch {
	var branches []branch
	for key := range m {
		if reflect.ValueOf(m[key]).Kind() == reflect.Map {
			b := recurseMap(m[key].(map[interface{}]interface{}))
			for i, br := range b {
				b[i].path = append([]string{key.(string)}, br.path...)
			}
			branches = append(branches, b...)
		} else if reflect.ValueOf(m[key]).Kind() == reflect.Slice &&
			len(m[key].([]interface{})) > 0 &&
			reflect.ValueOf(m[key].([]interface{})[0]).Kind() == reflect.Map {
			for j, m := range m[key].([]interface{}) {
				b := recurseMap(m.(map[interface{}]interface{}))
				for i, br := range b {
					br.path = append([]string{strconv.Itoa(j)}, br.path...)
					b[i].path = append([]string{key.(string)}, br.path...)
				}
				branches = append(branches, b...)
			}
		} else {
			path := []string{key.(string)}
			b := branch{
				leaf: m[key],
				path: path,
			}
			branches = append(branches, b)
		}
	}
	return branches
}

func recurseMapTree(m map[interface{}]interface{}) *tree {
	t := &tree{}
	for key := range m {
		var keyTree *tree
		switch reflect.ValueOf(m[key]).Kind() {
		case reflect.Map:
			keyTree = recurseMapTree(m[key].(map[interface{}]interface{}))
			keyTree.name = key
		default:
			keyTree = &tree{
				name:  key,
				value: m[key],
			}
		}
		t.nodes = append(t.nodes, keyTree)
	}
	return t
}

type tree struct {
	name  interface{}
	value interface{}
	nodes []*tree
}

func printTree(t *tree, prefix string) {
	if len(t.nodes) == 0 {
		fmt.Println(prefix, t.name, t.value)
	} else {
		for _, st := range t.nodes {
			fmt.Println(prefix, st.name)
			printTree(st, prefix+prefix)
		}
	}
}

type branch struct {
	path []string
	leaf interface{}
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
