package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clicker-org/clicker/ui/components"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// quitRequestedMsg is produced by translateKey when the user presses Q.
// Routing it through the normal message pipeline keeps routeKey unaware
// of dialog state: it always dispatches Q as this message; the gate decides.
type quitRequestedMsg struct{}

// quitDialog is the self-contained quit-confirmation overlay.
//
// It acts as a gate in the key-routing chain.  handle returns consumed=true
// whenever it processed a message, which tells the router to stop routing.
// When the dialog is closed, handle is a no-op for every message except
// quitRequestedMsg, so all other messages fall through to the active screen.
type quitDialog struct {
	t     theme.Theme
	open  bool
	modal components.ConfirmModal
}

func newQuitDialog(t theme.Theme) quitDialog {
	return quitDialog{t: t}
}

// handle processes msg and reports whether it was consumed.
// A consumed message must not be routed further by the caller.
func (d quitDialog) handle(msg tea.Msg) (quitDialog, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case quitRequestedMsg:
		if !d.open {
			d.modal = components.NewConfirmModal(d.t, "Quit")
			d.open = true
		}
		return d, nil, true

	case tea.KeyMsg:
		if !d.open {
			return d, nil, false
		}
		if msg.String() == "esc" {
			d.open = false
		}
		// All raw keys are consumed while the dialog is open.
		return d, nil, true

	case messages.NavLeftMsg, messages.NavRightMsg, messages.NavConfirmMsg:
		if !d.open {
			return d, nil, false
		}
		var cmd tea.Cmd
		d.modal, cmd = d.modal.Update(msg)
		return d, cmd, true

	default:
		// NavUpMsg, NavDownMsg, and any future message type: consumed only when open.
		return d, nil, d.open
	}
}

// close resets the dialog after the user makes a choice.
// Called by App when it receives a components.ConfirmMsg.
func (d quitDialog) close() quitDialog {
	d.open = false
	return d
}

// view overlays the dialog on bgContent when open; returns bgContent unchanged otherwise.
func (d quitDialog) view(bgContent string, width, height int) string {
	if !d.open {
		return bgContent
	}
	return d.modal.View("Quit CLIcker?", "Your progress will be saved.", bgContent, width, height)
}
