package world

import "sync"

// DefaultRegistry is the package-level world registry populated by world
// init() functions. Worlds register themselves by importing the worlds package.
var DefaultRegistry = NewWorldRegistry()

type WorldRegistry struct {
	mu     sync.RWMutex
	worlds []World
	byID   map[string]World
}

func NewWorldRegistry() *WorldRegistry {
	return &WorldRegistry{
		byID: make(map[string]World),
	}
}

func (r *WorldRegistry) Register(w World) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[w.ID()]; exists {
		panic("world: duplicate world ID: " + w.ID())
	}
	r.worlds = append(r.worlds, w)
	r.byID[w.ID()] = w
}

func (r *WorldRegistry) Get(id string) (World, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.byID[id]
	return w, ok
}

func (r *WorldRegistry) List() []World {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]World, len(r.worlds))
	copy(out, r.worlds)
	return out
}

func (r *WorldRegistry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, len(r.worlds))
	for i, w := range r.worlds {
		ids[i] = w.ID()
	}
	return ids
}
