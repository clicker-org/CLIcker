package themes

// SpaceTheme is the default space/cosmic visual theme.
type SpaceTheme struct{}

func (s SpaceTheme) Background() string      { return "#0d0d1a" }
func (s SpaceTheme) PrimaryText() string     { return "#e2e8f0" }
func (s SpaceTheme) DimText() string         { return "#64748b" }
func (s SpaceTheme) AccentColor() string     { return "#00d4ff" }
func (s SpaceTheme) SecondaryAccent() string { return "#8b5cf6" }
func (s SpaceTheme) CoinColor() string       { return "#f59e0b" }
func (s SpaceTheme) SuccessColor() string    { return "#10b981" }
func (s SpaceTheme) WarningColor() string    { return "#f97316" }
func (s SpaceTheme) ErrorColor() string      { return "#ef4444" }
func (s SpaceTheme) BorderColor() string     { return "#334155" }
func (s SpaceTheme) Name() string            { return "space" }
