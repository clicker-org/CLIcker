package theme

// Theme defines the color palette for a visual theme.
type Theme interface {
	Background() string
	PrimaryText() string
	DimText() string
	AccentColor() string
	SecondaryAccent() string
	CoinColor() string
	SuccessColor() string
	WarningColor() string
	ErrorColor() string
	BorderColor() string
	Name() string
}

// ThemeRegistry holds registered themes.
type ThemeRegistry struct {
	themes map[string]Theme
	active string
}

// NewThemeRegistry creates an empty registry.
func NewThemeRegistry() *ThemeRegistry {
	return &ThemeRegistry{themes: make(map[string]Theme)}
}

// Register adds a theme to the registry. Panics on duplicate name.
func (r *ThemeRegistry) Register(t Theme) {
	if _, exists := r.themes[t.Name()]; exists {
		panic("theme: duplicate theme name: " + t.Name())
	}
	r.themes[t.Name()] = t
}

// Get returns the theme with the given name, or nil if not found.
func (r *ThemeRegistry) Get(name string) Theme {
	return r.themes[name]
}

// Active returns the currently active theme.
func (r *ThemeRegistry) Active() Theme {
	if t, ok := r.themes[r.active]; ok {
		return t
	}
	for _, v := range r.themes {
		return v
	}
	return nil
}

// SetActive sets the active theme by name. Does nothing if not found.
func (r *ThemeRegistry) SetActive(name string) {
	if _, ok := r.themes[name]; ok {
		r.active = name
	}
}

// List returns all registered theme names.
func (r *ThemeRegistry) List() []string {
	names := make([]string, 0, len(r.themes))
	for k := range r.themes {
		names = append(names, k)
	}
	return names
}
