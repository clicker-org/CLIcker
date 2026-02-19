package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/offline"
	"github.com/clicker-org/clicker/internal/save"
	"github.com/clicker-org/clicker/internal/world"
	_ "github.com/clicker-org/clicker/internal/world/worlds"
	clui "github.com/clicker-org/clicker/ui"
	"github.com/clicker-org/clicker/ui/components/background"
	"github.com/clicker-org/clicker/ui/screens"
	"github.com/clicker-org/clicker/ui/theme"
	"github.com/clicker-org/clicker/ui/theme/themes"
)

func main() {
	// load save file.
	savePath := save.SavePath()
	sf, err := save.Load(savePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load save: %v\n", err)
		sf = save.DefaultSaveFile()
	}

	// worlds self-register into DefaultRegistry via their init() functions
	// (triggered by the blank import of internal/world/worlds above).
	worldReg := world.DefaultRegistry
	achievReg := achievement.NewAchievementRegistry()

	// set up theme registry.
	themeReg := theme.NewThemeRegistry()
	themeReg.Register(themes.SpaceTheme{})
	themeReg.SetActive(sf.Settings.ActiveTheme)
	activeTheme := themeReg.Active()

	// set up animation registry.
	animReg := background.NewAnimationRegistry()
	animReg.Register("stars", func() background.BackgroundAnimation {
		return background.NewStarsAnimation()
	})

	// reconstruct game state from save.
	gs := save.GameStateFromSave(sf, worldReg.IDs())

	// compute and apply offline income.
	offlineResult := offline.Apply(sf.LastScreen, sf.LastWorldID, sf.SavedAt, &gs)

	// create engine.
	eng := engine.New(gs, worldReg, achievReg)
	eng.Earned = sf.Achievements
	if eng.Earned == nil {
		eng.Earned = make(map[string]bool)
	}

	// build offline report.
	offlineReport := screens.NewOfflineReportModel(activeTheme, offlineResult)

	// build and run the app.
	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		w, h = 80, 24
	}

	app := clui.NewApp(
		eng,
		activeTheme,
		animReg,
		savePath,
		sf.Settings,
		offlineReport,
		w, h,
	)

	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
