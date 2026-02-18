package achievement

import "sync"

// AchievementRegistry holds all registered achievements.
type AchievementRegistry struct {
	mu           sync.RWMutex
	achievements []Achievement
	byID         map[string]*Achievement
}

// NewAchievementRegistry creates an empty registry.
func NewAchievementRegistry() *AchievementRegistry {
	return &AchievementRegistry{
		byID: make(map[string]*Achievement),
	}
}

// Register adds an Achievement to the registry. Panics if the ID is duplicate.
func (r *AchievementRegistry) Register(a Achievement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[a.ID]; exists {
		panic("achievement: duplicate achievement ID: " + a.ID)
	}
	r.achievements = append(r.achievements, a)
	r.byID[a.ID] = &r.achievements[len(r.achievements)-1]
}

// Get returns a pointer to the Achievement with the given ID and a found flag.
func (r *AchievementRegistry) Get(id string) (*Achievement, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.byID[id]
	return a, ok
}

// GetAll returns all achievements in insertion order.
func (r *AchievementRegistry) GetAll() []Achievement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Achievement, len(r.achievements))
	copy(out, r.achievements)
	return out
}

// Total returns the total number of registered achievements.
func (r *AchievementRegistry) Total() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.achievements)
}
