package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

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
	// set up file logging. All log.Printf calls (including those in internal
	// packages) will write here. The file is created on first run and appended
	// to on subsequent runs, so crash context is preserved across sessions.
	logPath := save.LogPath()
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err == nil {
		if lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
			defer lf.Close()
			log.SetOutput(lf)
			log.SetFlags(log.LstdFlags | log.Lshortfile)
		}
	}

	// catch panics so they are written to the log file before the process exits.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v\n%s", r, debug.Stack())
			fmt.Fprintf(os.Stderr, "clicker crashed — see %s for details\n", logPath)
			os.Exit(2)
		}
	}()

	// load save file.
	savePath := save.SavePath()
	sf, err := save.Load(savePath)
	if err != nil {
		log.Printf("warning: could not load save: %v", err)
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
	gs := save.GameStateFromSave(sf, worldReg)

	// compute and apply offline income.
	offlineResult := offline.Apply(sf.LastScreen, sf.LastWorldID, sf.SavedAt, &gs, worldReg)

	// create engine.
	eng := engine.New(gs, worldReg, achievReg)
	eng.Earned = sf.Achievements
	if eng.Earned == nil {
		eng.Earned = make(map[string]bool)
	}

	// build offline report.
	offlineReport := screens.NewOfflineReportModel(activeTheme, offlineResult, worldReg)

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
		log.Printf("error: %v", err)
		fmt.Fprintf(os.Stderr, "clicker error — see %s for details\n", logPath)
		os.Exit(1)
	}
}
