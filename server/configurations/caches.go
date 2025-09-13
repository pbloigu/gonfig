package configurations

import "sync"

type actionCache struct {
	c map[string][]Action
	m *sync.RWMutex
}

type appCache struct {
	c map[string]bool
	m *sync.RWMutex
}

func (ac *actionCache) put(key string, value []Action) {
	ac.m.Lock()
	defer ac.m.Unlock()
	ac.c[key] = value
}
func (ac *actionCache) get(key string) []Action {
	ac.m.RLock()
	defer ac.m.RUnlock()
	a, ok := ac.c[key]
	if !ok {
		return nil
	} else {
		return a
	}
}
func (ac *actionCache) remove(key string) {
	ac.m.Lock()
	defer ac.m.Unlock()
	delete(ac.c, key)
}

func (ac *appCache) put(key string, value bool) {
	ac.m.Lock()
	defer ac.m.Unlock()
	ac.c[key] = value
}

func (ac *appCache) get(key string) bool {
	ac.m.RLock()
	defer ac.m.RUnlock()
	a, ok := ac.c[key]
	if !ok {
		return false
	} else {
		return a
	}
}

func (ac *appCache) remove(key string) {
	ac.m.Lock()
	defer ac.m.Unlock()
	delete(ac.c, key)
}

func (ac *appCache) values() []string {
	ac.m.RLock()
	defer ac.m.RUnlock()
	vals := make([]string, len(ac.c))
	for k := range ac.c {
		vals = append(vals, k)
	}
	return vals
}
