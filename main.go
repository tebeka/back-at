// Show a progress bar up to a given time.
// Based on https://github.com/charmbracelet/bubbletea/blob/master/examples/progress-static/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxWidth = 80
)

var (
	prefix  = "☕ "
	padding = utf8.RuneCountInString(prefix)
)

func main() {
	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "usage: %s HH:MM (or HH:MMpm)\n", name)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "error: wrong number of arguments\n")
		os.Exit(1)
	}

	end, err := parseTime(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	start := time.Now()
	duration := end.Sub(start)
	if duration < 0 {
		fmt.Fprintf(os.Stderr, "error: %s is in the past\n", flag.Arg(0))
		os.Exit(1)
	}

	m := model{
		duration: float64(duration),
		start:    start,
		progress: progress.New(),
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}

type model struct {
	start    time.Time
	duration float64
	percent  float64
	progress progress.Model
}

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		m.percent = float64(time.Since(m.start)) / m.duration

		if m.percent > 1.0 {
			m.percent = 1.0
			return m, tea.Quit
		}
		return m, tickCmd()

	default:
		return m, nil
	}
}

func (m model) View() string {
	return "☕ " + m.progress.ViewAs(m.percent) + "\n"
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var layouts = []string{
	"15:04",
	time.Kitchen,
}

func parseTime(s string) (time.Time, error) {
	us := strings.ToUpper(s)
	for _, l := range layouts {
		t, err := time.Parse(l, us)
		if err != nil {
			continue
		}

		now := time.Now()
		t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unknown time format: %q", s)
}
