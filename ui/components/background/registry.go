package background

// AnimationRegistry holds factories for background animations keyed by name.
type AnimationRegistry struct {
	factories map[string]func() BackgroundAnimation
}

// NewAnimationRegistry creates an empty registry.
func NewAnimationRegistry() *AnimationRegistry {
	return &AnimationRegistry{factories: make(map[string]func() BackgroundAnimation)}
}

// Register adds an animation factory. Panics on duplicate key.
func (r *AnimationRegistry) Register(name string, factory func() BackgroundAnimation) {
	if _, exists := r.factories[name]; exists {
		panic("background: duplicate animation name: " + name)
	}
	r.factories[name] = factory
}

// New creates a new instance of the named animation. Returns nil if not found.
func (r *AnimationRegistry) New(name string) BackgroundAnimation {
	factory, ok := r.factories[name]
	if !ok {
		return nil
	}
	return factory()
}

// Keys returns all registered animation names.
func (r *AnimationRegistry) Keys() []string {
	keys := make([]string, 0, len(r.factories))
	for k := range r.factories {
		keys = append(keys, k)
	}
	return keys
}
