package lib

import (
	"log"
	"reflect"
	"regexp"
)

// Replacer is used for replacing variables
// with their contextual value
type Replacer struct {
	rgx *regexp.Regexp
}

// NewDefaultReplacer returns an initialized Replacer
// with default regex
func NewDefaultReplacer() *Replacer {
	rgx, err := regexp.Compile(`{{\s*([a-z]+)\s*}}`)
	if err != nil {
		log.Fatal(err)
	}
	return &Replacer{
		rgx: rgx,
	}
}

// Replace fill all var in tree with FileContext vars
func (r *Replacer) Replace(t *Tree, fc FileContext) {
	if t.IsLeaf() {
		if reflect.ValueOf(t.Value).Kind() == reflect.String {
			key := ""
			submatches := r.rgx.FindStringSubmatch(t.Value.(string))
			if len(submatches) > 0 {
				//submatches[0] is all the match string
				//submatches[1] is the submatch between spaces and {{}}
				key = submatches[1]
				log.Println(submatches)
				switch reflect.ValueOf(fc.vars[key]).Kind() {
				case reflect.String:
					t.Value = r.rgx.ReplaceAllString(t.Value.(string), fc.vars[key].(string))
				default:
					t.Value = fc.vars[key]
				}
			}
		}
	} else {
		for _, n := range t.Nodes {
			r.Replace(n, fc)
		}
	}
}
