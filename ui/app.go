package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/save"
	"github.com/clicker-org/clicker/ui/components"
	"github.com/clicker-org/clicker/ui/components/background"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/screens"
	"github.com/clicker-org/clicker/ui/theme"
)

// TickMsg is sent every 100ms to drive the game loop.
type TickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// App is the root Bubble Tea model.
type App struct {
	eng          *engine.Engine
	activeScreen engine.ScreenID
	width        int
	height       int
	t            theme.Theme
	animReg      *background.AnimationRegistry
	savePath     string
	saveSettings save.Settings

	overview      screens.OverviewModel
	dashboard     screens.DashboardModel
	worldScreen   screens.WorldModel
	offlineReport screens.OfflineReportModel
	notification  components.Notification
	statusBar     components.StatusBar
}

// NewApp creates the root App model.
func NewApp(
	eng *engine.Engine,
	t theme.Theme,
	animReg *background.AnimationRegistry,
	savePath string,
	settings save.Settings,
	offlineReport screens.OfflineReportModel,
	width, height int,
) App {
	initialScreen := engine.ScreenOverview
	if offlineReport.IsVisible() {
		initialScreen = engine.ScreenOfflineReport
	}

	overview := screens.NewOverviewModel(t, &eng.State, eng.WorldReg, width, height)
	dashboard := screens.NewDashboardModel(t, &eng.State, width, height)

	var worldScreen screens.WorldModel
	worldIDs := eng.WorldReg.IDs()
	if len(worldIDs) > 0 {
		wID := worldIDs[0]
		animKey := "stars"
		if w, ok := eng.WorldReg.Get(wID); ok {
			animKey = w.AmbientAnimation()
		}
		worldScreen = screens.NewWorldModel(t, eng, &eng.State, wID, animReg, animKey, width, height)
	}

	return App{
		eng:           eng,
		activeScreen:  initialScreen,
		width:         width,
		height:        height,
		t:             t,
		animReg:       animReg,
		savePath:      savePath,
		saveSettings:  settings,
		overview:      overview,
		dashboard:     dashboard,
		worldScreen:   worldScreen,
		offlineReport: offlineReport,
		notification:  components.NewNotification(t),
		statusBar:     components.NewStatusBar(t, width),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		a.overview.Init(),
		a.offlineReport.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.SetWidth(msg.Width)
		a.overview, _ = a.overview.Update(msg)
		a.dashboard, _ = a.dashboard.Update(msg)
		a.worldScreen, _ = a.worldScreen.Update(msg)
		return a, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "Q" {
			_ = save.Save(a.eng.State, a.eng.Earned, a.saveSettings, a.savePath)
			return a, tea.Quit
		}
		return a.routeKey(msg, cmds)

	case TickMsg:
		events := a.eng.Tick(float64(engine.TickIntervalMs) / 1000.0)
		cmds = append(cmds, tickCmd())
		for _, ev := range events {
			ev := ev
			switch ev.Type {
			case engine.EventAchievementUnlocked:
				cmds = append(cmds, func() tea.Msg {
					return messages.AchievementUnlockedMsg{ID: ev.AchievementID}
				})
			case engine.EventLevelUp:
				cmds = append(cmds, func() tea.Msg {
					return messages.LevelUpMsg{NewLevel: ev.NewLevel}
				})
			case engine.EventAutoSave:
				_ = save.Save(a.eng.State, a.eng.Earned, a.saveSettings, a.savePath)
			}
		}
		// Pass tick to active screen for animation updates.
		a = a.routeToActiveScreen(msg)
		return a, tea.Batch(cmds...)

	case messages.AchievementUnlockedMsg:
		cmd := a.notification.Show("Achievement: "+msg.ID, 3*time.Second)
		cmds = append(cmds, cmd)

	case messages.NavigateToOverviewMsg:
		a.activeScreen = engine.ScreenOverview
		a.eng.State.LastScreen = "overview"

	case messages.NavigateToDashboardMsg:
		a.activeScreen = engine.ScreenDashboard

	case messages.NavigateToWorldMsg:
		a.eng.State.ActiveWorldID = msg.WorldID
		a.eng.State.LastWorldID = msg.WorldID
		a.eng.State.LastScreen = "world"
		animKey := "stars"
		if w, ok := a.eng.WorldReg.Get(msg.WorldID); ok {
			animKey = w.AmbientAnimation()
		}
		a.worldScreen = screens.NewWorldModel(
			a.t, a.eng, &a.eng.State, msg.WorldID, a.animReg, animKey, a.width, a.height,
		)
		a.activeScreen = engine.ScreenWorld
		cmds = append(cmds, a.worldScreen.Init())

	case messages.OfflineReportDismissedMsg:
		a.offlineReport, _ = a.offlineReport.Update(msg)
		a.activeScreen = engine.ScreenOverview

	case components.NotificationDismissMsg:
		var notifCmd tea.Cmd
		a.notification, notifCmd = a.notification.Update(msg)
		if notifCmd != nil {
			cmds = append(cmds, notifCmd)
		}
	}

	// Route remaining messages to active screen.
	a = a.routeToActiveScreen(msg)

	return a, tea.Batch(cmds...)
}

func (a App) routeKey(msg tea.KeyMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.activeScreen {
	case engine.ScreenOverview:
		a.overview, cmd = a.overview.Update(msg)
	case engine.ScreenDashboard:
		a.dashboard, cmd = a.dashboard.Update(msg)
	case engine.ScreenWorld:
		a.worldScreen, cmd = a.worldScreen.Update(msg)
	case engine.ScreenOfflineReport:
		a.offlineReport, cmd = a.offlineReport.Update(msg)
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return a, tea.Batch(cmds...)
}

func (a App) routeToActiveScreen(msg tea.Msg) App {
	switch a.activeScreen {
	case engine.ScreenOverview:
		a.overview, _ = a.overview.Update(msg)
	case engine.ScreenDashboard:
		a.dashboard, _ = a.dashboard.Update(msg)
	case engine.ScreenWorld:
		a.worldScreen, _ = a.worldScreen.Update(msg)
	case engine.ScreenOfflineReport:
		a.offlineReport, _ = a.offlineReport.Update(msg)
	}
	return a
}

func (a App) View() string {
	bg := lipgloss.Color(a.t.Background())

	if a.activeScreen == engine.ScreenOfflineReport && a.offlineReport.IsVisible() {
		return lipgloss.Place(
			a.width, a.height,
			lipgloss.Center, lipgloss.Center,
			a.offlineReport.View(),
			lipgloss.WithWhitespaceBackground(bg),
		)
	}

	var content string
	switch a.activeScreen {
	case engine.ScreenOverview:
		content = a.overview.View()
	case engine.ScreenDashboard:
		content = a.dashboard.View()
	case engine.ScreenWorld:
		content = a.worldScreen.View()
	default:
		content = a.overview.View()
	}

	if notif := a.notification.View(); notif != "" {
		content = notif + "\n" + content
	}

	// Fill the entire terminal with the game background so nothing bleeds through.
	return lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Background(bg).
		Foreground(lipgloss.Color(a.t.PrimaryText())).
		Render(content)
}
