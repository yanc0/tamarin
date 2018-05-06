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

	t0 := mapToTree(y0)
	t1 := mapToTree(y1)
	//printTree(t0, "-")
	//printTree(t1, "-")

	fmt.Println("***************")
	mergeTrees(t0, t1)
	//printTree(t0, "-")

	fmt.Println("***************")
	y, _ := yaml.Marshal(t0.ToMap())
	fmt.Println(string(y))

}

func mergeTrees(t1 *tree, t2 *tree) {
	if t2.isLeaf() {
		t1 = t2
	} else {
		for _, n := range t2.nodes {
			if n.isLeaf() {
				t1.Append(n)
			} else {
				if _, ok := t1.nodes[n.name]; ok {
					mergeTrees(t1.nodes[n.name], n)
				} else {
					t1.Append(n)
				}
			}
		}
	}
}

func mapToTree(m map[interface{}]interface{}) *tree {
	t := makeTree()
	for key := range m {
		var keyTree *tree
		switch reflect.ValueOf(m[key]).Kind() {
		case reflect.Map:
			keyTree = mapToTree(m[key].(map[interface{}]interface{}))
			keyTree.name = key

		case reflect.Slice:
			keyTree = &tree{
				name: key,
			}
			if isSliceContainsMap(m[key].([]interface{})) {
				for i, mp := range m[key].([]interface{}) {
					tmpTree := mapToTree(mp.(map[interface{}]interface{}))
					tmpTree.name = i
					keyTree.Append(tmpTree)
				}
				keyTree.isSliceOfMap = true
			} else {
				keyTree.value = m[key]
			}

		default:
			keyTree = &tree{
				name:  key,
				value: m[key],
			}
		}
		t.Append(keyTree)
	}
	return t
}

func makeTree() *tree {
	nodes := make(map[interface{}]*tree)
	return &tree{
		nodes: nodes,
	}
}

type tree struct {
	name         interface{}
	value        interface{}
	isSliceOfMap bool
	nodes        map[interface{}]*tree
}

func (t *tree) isRoot() bool {
	return t.name == nil
}

func (t *tree) ToMap() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})

	for _, n := range t.nodes {
		if n.isLeaf() {
			m[n.name] = n.value
		} else {
			if ! n.isSliceOfMap {
				m[n.name] = n.ToMap()
			} else {
				s := make([]interface{}, 0)
				for i := range n.nodes {
					s = append(s, n.nodes[i].ToMap())
				}
				m[n.name] = s
			}
		}
	}
	return m
}

func (t *tree) Append(add *tree) {
	if t.nodes == nil {
		t.nodes = make(map[interface{}]*tree)
	}
	t.nodes[add.name] = add
}

func (t *tree) Copy() *tree {
	tree := makeTree()
	tree.name = t.name
	if t.isLeaf() {
		tree.value = t.value
	} else {
		for _, n := range t.nodes {
			tree.Append(n.Copy())
		}
	}
	return tree
}

func (t *tree) isLeaf() bool {
	return len(t.nodes) == 0
}

func isSliceContainsMap(m []interface{}) bool {
	return len(m) > 0 && reflect.ValueOf(m[0]).Kind() == reflect.Map
}

func printTree(t *tree, prefix string) {
	if len(t.nodes) == 0 {
		fmt.Println(prefix, t.name, t.value)
	} else {
		if t.name != nil {
			fmt.Println(prefix, t.name)
		}
		for _, st := range t.nodes {
			printTree(st, prefix+"-")
		}
	}
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
