package lib

type FileContext struct {
	vars map[string]interface{}
}

func (fc *FileContext) AddVar(key string, value interface{}) {
	if fc.vars == nil {
		fc.vars = make(map[string]interface{})
	}
	fc.vars[key] = value
}
