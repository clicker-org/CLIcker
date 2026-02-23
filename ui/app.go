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
	return tea.Tick(time.Duration(engine.TickIntervalMs)*time.Millisecond, func(t time.Time) tea.Msg {
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

	quit quitDialog
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

	app := App{
		eng:           eng,
		activeScreen:  initialScreen,
		width:         width,
		height:        height,
		t:             t,
		animReg:       animReg,
		savePath:      savePath,
		saveSettings:  settings,
		overview:      screens.NewOverviewModel(t, &eng.State, eng.WorldReg, width, height),
		dashboard:     screens.NewDashboardModel(t, &eng.State, width, height),
		offlineReport: offlineReport,
		notification:  components.NewNotification(t),
		statusBar:     components.NewStatusBar(t, width, eng.WorldReg),
		quit:          newQuitDialog(t),
	}

	worldIDs := eng.WorldReg.IDs()
	if len(worldIDs) > 0 {
		app.worldScreen = app.buildWorldScreen(worldIDs[0])
	}

	return app
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		a.overview.Init(),
		a.offlineReport.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.SetWidth(msg.Width)
		a.overview, _ = a.overview.Update(msg)
		a.dashboard, _ = a.dashboard.Update(msg)
		a.worldScreen, _ = a.worldScreen.Update(msg)
		a.offlineReport, _ = a.offlineReport.Update(msg)
		return a, nil

	case tea.KeyMsg:
		return a.routeKey(msg)

	case TickMsg:
		return a.handleTick(msg)

	case messages.AchievementUnlockedMsg:
		return a, a.notification.Show("Achievement: "+msg.ID, 3*time.Second)

	case messages.NavigateToOverviewMsg:
		a.activeScreen = engine.ScreenOverview
		a.eng.State.LastScreen = "overview"
		return a, nil

	case messages.NavigateToDashboardMsg:
		a.activeScreen = engine.ScreenDashboard
		return a, nil

	case messages.NavigateToWorldMsg:
		a.eng.State.ActiveWorldID = msg.WorldID
		a.eng.State.LastWorldID = msg.WorldID
		a.eng.State.LastScreen = "world"
		a.worldScreen = a.buildWorldScreen(msg.WorldID)
		a.activeScreen = engine.ScreenWorld
		return a, a.worldScreen.Init()

	case messages.OfflineReportDismissedMsg:
		a.offlineReport, _ = a.offlineReport.Update(msg)
		a.activeScreen = engine.ScreenOverview
		return a, nil

	case components.NotificationDismissMsg:
		var cmd tea.Cmd
		a.notification, cmd = a.notification.Update(msg)
		return a, cmd

	case components.ConfirmMsg:
		a.quit = a.quit.close()
		if msg.Confirmed {
			_ = save.Save(a.eng.State, a.eng.Earned, a.saveSettings, a.savePath)
			return a, tea.Quit
		}
		return a, nil
	}

	// Forward unhandled messages to the active screen.
	var cmd tea.Cmd
	a, cmd = a.routeToActiveScreen(msg)
	return a, cmd
}

// handleTick advances the engine one tick, processes engine events, and
// forwards the tick to the active screen for animation updates.
func (a App) handleTick(msg TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	events := a.eng.Tick(float64(engine.TickIntervalMs) / 1000.0)
	cmds = append(cmds, tickCmd())

	for _, ev := range events {
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

	var animCmd tea.Cmd
	a, animCmd = a.routeToActiveScreen(msg)
	if animCmd != nil {
		cmds = append(cmds, animCmd)
	}

	return a, tea.Batch(cmds...)
}

// buildWorldScreen constructs a WorldModel for the given world ID.
func (a App) buildWorldScreen(worldID string) screens.WorldModel {
	animKey := "stars"
	if w, ok := a.eng.WorldReg.Get(worldID); ok {
		animKey = w.AmbientAnimation()
	}
	return screens.NewWorldModel(a.t, a.eng, &a.eng.State, worldID, a.animReg, animKey, a.width, a.height)
}

// routeKey is the single entry point for all key input.
//
// It translates the raw key to an abstract message (translateKey), then passes
// it through the quit-dialog gate.  If the gate consumes the message, routing
// stops.  Otherwise the (possibly translated) message reaches the active screen.
//
// routeKey never inspects dialog state — that is entirely the gate's concern.
func (a App) routeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	abstract := translateKey(msg)

	var cmds []tea.Cmd
	var cmd tea.Cmd
	var consumed bool
	a.quit, cmd, consumed = a.quit.handle(abstract)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if consumed {
		return a, tea.Batch(cmds...)
	}

	a, cmd = a.routeToActiveScreen(abstract)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return a, tea.Batch(cmds...)
}

// translateKey converts a raw KeyMsg to the abstract message it represents.
// Keys with no abstract equivalent are returned as-is so active screens can
// handle them as first-letter shortcuts without any extra switch cases here.
func translateKey(msg tea.KeyMsg) tea.Msg {
	switch msg.String() {
	case "q", "Q":
		return quitRequestedMsg{}
	case "up", "k":
		return messages.NavUpMsg{}
	case "down", "j":
		return messages.NavDownMsg{}
	case "left", "h":
		return messages.NavLeftMsg{}
	case "right", "l":
		return messages.NavRightMsg{}
	case "enter":
		return messages.NavConfirmMsg{}
	}
	return msg
}

func (a App) routeToActiveScreen(msg tea.Msg) (App, tea.Cmd) {
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
	return a, cmd
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

	// Pre-render the full-screen background before any modal overlay.
	fullContent := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Background(bg).
		Foreground(lipgloss.Color(a.t.PrimaryText())).
		Render(content)

	return a.quit.view(fullContent, a.width, a.height)
}
