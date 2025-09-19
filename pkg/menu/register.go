package menu

import "sync"

type registered struct {
	V map[string]any
	m sync.Mutex
}

var r = registered{
	V: make(map[string]any),
}

func Register(name string, value any) {
	r.m.Lock()
	defer r.m.Unlock()
	r.V[name] = value
}
