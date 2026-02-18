package world

import "sync"

// WorldRegistry is the central registry for all world implementations.
type WorldRegistry struct {
	mu     sync.RWMutex
	worlds []World
	byID   map[string]World
}

// NewWorldRegistry creates an empty WorldRegistry.
func NewWorldRegistry() *WorldRegistry {
	return &WorldRegistry{
		byID: make(map[string]World),
	}
}

// Register adds a World to the registry. Panics if the ID is already registered.
func (r *WorldRegistry) Register(w World) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[w.ID()]; exists {
		panic("world: duplicate world ID: " + w.ID())
	}
	r.worlds = append(r.worlds, w)
	r.byID[w.ID()] = w
}

// Get returns the World with the given ID and a found flag.
func (r *WorldRegistry) Get(id string) (World, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.byID[id]
	return w, ok
}

// List returns all registered worlds in insertion order.
func (r *WorldRegistry) List() []World {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]World, len(r.worlds))
	copy(out, r.worlds)
	return out
}

// IDs returns all registered world IDs in insertion order.
func (r *WorldRegistry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, len(r.worlds))
	for i, w := range r.worlds {
		ids[i] = w.ID()
	}
	return ids
}
