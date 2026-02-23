package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/theme"
)

// WorldVisual is the display-only data required by the overview carousel.
type WorldVisual struct {
	ID          string
	Name        string
	AccentColor string
	Completion  float64
	Coins       float64
	CPS         float64
	Prestige    int
}

// GalaxyMap renders an overview galaxy node map with wrap-around navigation.
type GalaxyMap struct {
	Width        int
	Height       int
	FocusedIndex int
}

func (g *GalaxyMap) normalize(max int) {
	if max <= 0 {
		g.FocusedIndex = 0
		return
	}
	g.FocusedIndex %= max
	if g.FocusedIndex < 0 {
		g.FocusedIndex += max
	}
}

// MoveLeft moves focus counter-clockwise.
func (g *GalaxyMap) MoveLeft(max int) {
	if max <= 0 {
		return
	}
	g.FocusedIndex--
	g.normalize(max)
}

// MoveRight moves focus clockwise.
func (g *GalaxyMap) MoveRight(max int) {
	if max <= 0 {
		return
	}
	g.FocusedIndex++
	g.normalize(max)
}

// MoveUp jumps toward upper arc nodes.
func (g *GalaxyMap) MoveUp(max int) {
	if max <= 0 {
		return
	}
	step := max / 4
	if step < 1 {
		step = 1
	}
	g.FocusedIndex -= step
	g.normalize(max)
}

// MoveDown jumps toward lower arc nodes.
func (g *GalaxyMap) MoveDown(max int) {
	if max <= 0 {
		return
	}
	step := max / 4
	if step < 1 {
		step = 1
	}
	g.FocusedIndex += step
	g.normalize(max)
}

// FocusedWorldID returns the ID of the currently focused world, or empty string.
func (g *GalaxyMap) FocusedWorldID(worlds []WorldVisual) string {
	if len(worlds) == 0 {
		return ""
	}
	g.normalize(len(worlds))
	return worlds[g.FocusedIndex].ID
}

type point struct {
	x int
	y int
}

type canvas struct {
	w     int
	h     int
	cells [][]rune
}

func newCanvas(w, h int) canvas {
	c := canvas{
		w:     w,
		h:     h,
		cells: make([][]rune, h),
	}
	for y := 0; y < h; y++ {
		c.cells[y] = make([]rune, w)
		for x := 0; x < w; x++ {
			c.cells[y][x] = ' '
		}
	}
	return c
}

func (c *canvas) set(x, y int, r rune) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	c.cells[y][x] = r
}

func (c *canvas) drawText(x, y int, s string) {
	if y < 0 || y >= c.h {
		return
	}
	runes := []rune(s)
	for i, r := range runes {
		xi := x + i
		if xi < 0 || xi >= c.w {
			continue
		}
		c.cells[y][xi] = r
	}
}

func lineRune(dx, dy int) rune {
	if dx == 0 {
		return '|'
	}
	if dy == 0 {
		return '-'
	}
	if (dx > 0 && dy > 0) || (dx < 0 && dy < 0) {
		return '\\'
	}
	return '/'
}

func (c *canvas) drawLine(a, b point, r rune) {
	x0, y0 := a.x, a.y
	x1, y1 := b.x, b.y
	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy
	for {
		c.set(x0, y0, r)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (c canvas) String() string {
	lines := make([]string, c.h)
	for y := 0; y < c.h; y++ {
		lines[y] = string(c.cells[y])
	}
	return strings.Join(lines, "\n")
}

func shortName(name string, limit int) string {
	if lipgloss.Width(name) <= limit {
		return name
	}
	if limit <= 1 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(limit-1).Render(name) + "…"
}

func (g GalaxyMap) nodePositions(width, height, n int) []point {
	pos := make([]point, n)
	if n == 0 {
		return pos
	}
	cx := width / 2
	cy := height / 2
	rx := width/2 - 8
	ry := height/2 - 3
	if rx < 10 {
		rx = 10
	}
	if ry < 4 {
		ry = 4
	}
	for i := 0; i < n; i++ {
		theta := (2*math.Pi*float64(i))/float64(n) - math.Pi/2
		x := cx + int(float64(rx)*math.Cos(theta))
		y := cy + int(float64(ry)*math.Sin(theta)*0.62)
		pos[i] = point{x: x, y: y}
	}
	return pos
}

func (g GalaxyMap) renderCompact(worlds []WorldVisual, t theme.Theme) string {
	g.normalize(len(worlds))
	bg := lipgloss.Color(t.Background())
	if len(worlds) == 0 {
		return lipgloss.NewStyle().Width(g.Width).Background(bg).Render("(no worlds registered)")
	}

	list := make([]string, 0, len(worlds))
	for i, w := range worlds {
		prefix := "  "
		if i == g.FocusedIndex {
			prefix = "> "
		}
		list = append(list, fmt.Sprintf("%s%s", prefix, shortName(w.Name, 20)))
	}

	w := worlds[g.FocusedIndex]
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(w.AccentColor)).
		Bold(true).
		Render(w.Name)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.DimText()))
	details := strings.Join([]string{
		title,
		dimStyle.Render(fmt.Sprintf("Completion: %.1f%%", w.Completion)),
		dimStyle.Render(fmt.Sprintf("CPS: %.2f   Prestige: %d", w.CPS, w.Prestige)),
		dimStyle.Render(fmt.Sprintf("Coins: %.2f", w.Coins)),
	}, "\n")

	return lipgloss.NewStyle().
		Width(g.Width).
		Background(bg).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			strings.Join(list, "\n"),
			"",
			details,
		))
}

func (g GalaxyMap) renderGalaxy(worlds []WorldVisual, t theme.Theme) string {
	g.normalize(len(worlds))
	bg := lipgloss.Color(t.Background())
	mapW := g.Width
	mapH := g.Height - 9
	if mapH < 10 {
		mapH = g.Height - 5
	}
	if mapH < 8 {
		mapH = 8
	}
	c := newCanvas(mapW, mapH)
	pos := g.nodePositions(mapW, mapH, len(worlds))

	orbit1 := g.nodePositions(mapW, mapH, 72)
	for i := 0; i < len(orbit1); i++ {
		p := orbit1[i]
		c.set(p.x, p.y, '.')
	}
	orbit2 := g.nodePositions(mapW-8, mapH-2, 56)
	for i := 0; i < len(orbit2); i++ {
		p := orbit2[i]
		c.set(p.x+4, p.y+1, ':')
	}

	for i := 0; i < len(pos); i++ {
		next := (i + 1) % len(pos)
		a := pos[i]
		b := pos[next]
		c.drawLine(a, b, lineRune(b.x-a.x, b.y-a.y))
	}

	for i, p := range pos {
		marker := 'o'
		if i == g.FocusedIndex {
			marker = '@'
		}
		c.set(p.x, p.y, marker)
		label := shortName(worlds[i].Name, 12)
		offset := -len([]rune(label)) / 2
		if p.y <= mapH/2 {
			c.drawText(p.x+offset, p.y-1, label)
		} else {
			c.drawText(p.x+offset, p.y+1, label)
		}
	}

	mapText := c.String()
	mapStyle := lipgloss.NewStyle().
		Width(mapW).
		Height(mapH).
		Background(bg).
		Foreground(lipgloss.Color(t.PrimaryText())).
		Render(mapText)

	w := worlds[g.FocusedIndex]
	accent := w.AccentColor
	if accent == "" {
		accent = t.AccentColor()
	}
	cardTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(accent)).
		Bold(true).
		Render(strings.ToUpper(w.Name))
	cardDim := lipgloss.NewStyle().Foreground(lipgloss.Color(t.DimText()))
	card := lipgloss.NewStyle().
		Background(bg).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(accent)).
		Padding(0, 1).
		Render(strings.Join([]string{
			cardTitle,
			cardDim.Render(fmt.Sprintf("Completion: %.1f%%", w.Completion)),
			cardDim.Render(fmt.Sprintf("CPS: %.2f   Prestige: %d", w.CPS, w.Prestige)),
			cardDim.Render(fmt.Sprintf("Coins: %.2f", w.Coins)),
		}, "\n"))

	cardRow := lipgloss.Place(
		g.Width,
		lipgloss.Height(card),
		lipgloss.Center,
		lipgloss.Top,
		card,
		lipgloss.WithWhitespaceBackground(bg),
	)

	mapRow := lipgloss.Place(
		g.Width,
		mapH,
		lipgloss.Left,
		lipgloss.Top,
		mapStyle,
		lipgloss.WithWhitespaceBackground(bg),
	)

	return lipgloss.NewStyle().
		Width(g.Width).
		Height(g.Height).
		Background(bg).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			mapRow,
			"",
			cardRow,
		))
}

// View renders a galaxy node map with all worlds navigable.
func (g GalaxyMap) View(worlds []WorldVisual, t theme.Theme) string {
	bg := lipgloss.Color(t.Background())

	var body string
	if len(worlds) == 0 {
		body = "(no worlds registered)"
	} else if g.Width < 90 || g.Height < 24 {
		body = g.renderCompact(worlds, t)
	} else {
		body = g.renderGalaxy(worlds, t)
	}

	content := body
	content = lipgloss.NewStyle().
		Background(bg).
		Width(g.Width).
		Height(g.Height).
		Render(content)
	h := g.Height
	if h <= 0 {
		h = 24
	}
	return lipgloss.Place(g.Width, h,
		lipgloss.Center, lipgloss.Top,
		content,
		lipgloss.WithWhitespaceBackground(bg))
}
