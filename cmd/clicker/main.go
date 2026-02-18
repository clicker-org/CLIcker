package main

import (
	"fmt"
	"os"
	"time"

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

	// compute offline income.
	var (
		offlineWorldCoins float64
		offlineGC         float64
		offlineDuration   time.Duration
	)
	if !sf.SavedAt.IsZero() {
		elapsed := time.Since(sf.SavedAt).Seconds()
		if elapsed > 0 {
			offlineDuration = time.Duration(elapsed * float64(time.Second))
			if sf.LastScreen == "world" && sf.LastWorldID != "" {
				if ws, ok := gs.Worlds[sf.LastWorldID]; ok {
					offlineWorldCoins = offline.CalculateOfflineIncome(
						ws.CPS, 0.10, elapsed, 8.0,
					)
					if offlineWorldCoins > 0 {
						ws.Coins += offlineWorldCoins
						ws.TotalCoinsEarned += offlineWorldCoins
					}
				}
			} else {
				offlineGC = offline.CalculateOverviewOfflineIncome(0.001, elapsed, 10.0)
				if offlineGC > 0 {
					gs.Player.GeneralCoins += offlineGC
				}
			}
		}
	}

	// create engine.
	eng := engine.New(gs, worldReg, achievReg)
	eng.Earned = sf.Achievements
	if eng.Earned == nil {
		eng.Earned = make(map[string]bool)
	}

	// build offline report.
	// Show whenever the player has been away for at least 1 minute, regardless
	// of whether any income was generated (CPS may be 0 early in the game).
	const minOfflineToReport = 60 * time.Second
	showReport := offlineDuration >= minOfflineToReport
	offlineReport := screens.NewOfflineReportModel(
		activeTheme,
		sf.LastWorldID,
		offlineDuration,
		offlineWorldCoins,
		offlineGC,
		showReport,
	)

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
