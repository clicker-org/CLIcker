package background

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	animFrameInterval = 150 * time.Millisecond
	starsPerCells     = 8
)

var starChars = []rune{'·', '✦', '*', '·', '·'}

type star struct {
	x, y  int
	speed int
	char  rune
}

// StarsAnimation is a drifting-stars background animation.
type StarsAnimation struct {
	stars []star
	rng   *rand.Rand
	w, h  int
}

// NewStarsAnimation creates a new StarsAnimation instance.
func NewStarsAnimation() BackgroundAnimation {
	return &StarsAnimation{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *StarsAnimation) Name() string { return "stars" }

func (s *StarsAnimation) Init() tea.Cmd {
	return AnimTickCmd("stars", animFrameInterval)
}

func (s *StarsAnimation) seed(width, height int) {
	if width == s.w && height == s.h && len(s.stars) > 0 {
		return
	}
	s.w = width
	s.h = height
	total := (width * height) / starsPerCells
	if total < 1 {
		total = 1
	}
	s.stars = make([]star, total)
	for i := range s.stars {
		s.stars[i] = star{
			x:     s.rng.Intn(width),
			y:     s.rng.Intn(height),
			speed: s.rng.Intn(3) + 1,
			char:  starChars[s.rng.Intn(len(starChars))],
		}
	}
}

func (s *StarsAnimation) Update(msg tea.Msg) (BackgroundAnimation, tea.Cmd) {
	tick, ok := msg.(AnimTickMsg)
	if !ok || tick.AnimationName != "stars" {
		return s, nil
	}
	if s.w > 0 {
		for i := range s.stars {
			s.stars[i].x -= s.stars[i].speed
			if s.stars[i].x < 0 {
				s.stars[i].x = s.w - 1
				s.stars[i].y = s.rng.Intn(s.h)
			}
		}
	}
	return s, AnimTickCmd("stars", animFrameInterval)
}

func (s *StarsAnimation) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	s.seed(width, height)

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	for _, st := range s.stars {
		if st.x >= 0 && st.x < width && st.y >= 0 && st.y < height {
			grid[st.y][st.x] = st.char
		}
	}

	var sb strings.Builder
	for row, line := range grid {
		sb.WriteString(string(line))
		if row < height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
