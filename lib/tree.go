package lib


import (
	"reflect"
	"fmt"
)

// Tree is name with a value and subtrees
// called nodes. Slice of map are represented as
// map with nodes 0, 1, 2, etc.
type Tree struct {
	Name         interface{}
	Value        interface{}
	IsSliceOfMap bool
	Nodes        map[interface{}]*Tree
}

// MakeTree returns an initialized tree
func MakeTree() *Tree {
	nodes := make(map[interface{}]*Tree)
	return &Tree{
		Nodes: nodes,
	}
}

// MapToTree return tree representation
// of map[interface{}]interface{}
func MapToTree(m map[interface{}]interface{}) *Tree {
	t := MakeTree()
	for key := range m {
		var keyTree *Tree
		switch reflect.ValueOf(m[key]).Kind() {
		case reflect.Map:
			keyTree = MapToTree(m[key].(map[interface{}]interface{}))
			keyTree.Name = key

		case reflect.Slice:
			keyTree = MakeTree()
			keyTree.Name = key
			if isSliceContainsMap(m[key].([]interface{})) {
				for i, mp := range m[key].([]interface{}) {
					tmpTree := MapToTree(mp.(map[interface{}]interface{}))
					tmpTree.Name = i
					keyTree.Append(tmpTree)
				}
				keyTree.IsSliceOfMap = true
			} else {
				keyTree.Value = m[key]
			}

		default:
			keyTree = MakeTree()
			keyTree.Name = key
			keyTree.Value = m[key]
		}
		t.Append(keyTree)
	}
	return t
}

// IsRoot returns true if tree only
// contains nodes. The root tree.
func (t *Tree) IsRoot() bool {
	return t.Name == nil
}

// ToMap return the map[interface{}]interface{} of
// a tree. Useful for marshaling into yaml or json
func (t *Tree) ToMap() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})

	for _, n := range t.Nodes {
		if n.IsLeaf() {
			m[n.Name] = n.Value
		} else {
			if ! n.IsSliceOfMap {
				m[n.Name] = n.ToMap()
			} else {
				s := make([]interface{}, 0)
				for i := range n.Nodes {
					s = append(s, n.Nodes[i].ToMap())
				}
				m[n.Name] = s
			}
		}
	}
	return m
}

// MergeTree merge t2 into tree
func (t *Tree) MergeTree(t2 *Tree) {
	if t2.IsLeaf() {
		t = t2
	} else {
		for _, n := range t2.Nodes {
			if n.IsLeaf() {
				t.Append(n)
			} else {
				if _, ok := t.Nodes[n.Name]; ok {
					t.Nodes[n.Name].MergeTree(n)
				} else {
					t.Append(n)
				}
			}
		}
	}
}

// Append a tree as subtree
func (t *Tree) Append(add *Tree) {
	if t.Nodes == nil {
		t.Nodes = make(map[interface{}]*Tree)
	}
	t.Nodes[add.Name] = add
}

// Copy return the exact representation of
// a tree but with a different memory alloc
func (t *Tree) Copy() *Tree {
	tree := MakeTree()
	tree.Name = t.Name
	if t.IsLeaf() {
		tree.Value = t.Value
	} else {
		for _, n := range t.Nodes {
			tree.Append(n.Copy())
		}
	}
	return tree
}

// IsLeaf returns true if there are no
// subtree under this tree
func (t *Tree) IsLeaf() bool {
	return len(t.Nodes) == 0
}

// PrintTree uses fmt.Println and show
// a graphical representation of a tree
// useful for debuging purpose
func PrintTree(t *Tree, prefix string) {
	if len(t.Nodes) == 0 {
		fmt.Println(prefix, t.Name, t.Value)
	} else {
		if t.Name != nil {
			fmt.Println(prefix, t.Name)
		}
		for _, st := range t.Nodes {
			PrintTree(st, prefix+"-")
		}
	}
}

func isSliceContainsMap(m []interface{}) bool {
	return len(m) > 0 && reflect.ValueOf(m[0]).Kind() == reflect.Map
}
