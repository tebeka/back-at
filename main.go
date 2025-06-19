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
	maxWidth      = 80
	flashDuration = 200 * time.Millisecond
	totalFlashes  = 6 // for 3 full on-off cycles
)

type flashMsg struct{}

var (
	version = "0.3.0"

	options struct {
		showVersion bool
		prefix      string
	}
)

const (
	atHelp = "usage: %s HH:MM (or HH:MMpm)\n\nOptions:\n"
	inHelp = "usage: %s DURATION (e.g. 15m)\n\nOptions:\n"
)

func main() {
	progName := path.Base(os.Args[0])
	usage := atHelp
	switch progName {
	case "back-at":
		usage = atHelp
	case "back-in":
		usage = inHelp
	}

	flag.Usage = func() {
		name := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, usage, name)
		flag.PrintDefaults()
	}
	flag.BoolVar(&options.showVersion, "version", false, "show version and exit")
	flag.StringVar(&options.prefix, "prefix", "☕ ", "progress bar prefix")
	flag.Parse()

	if options.showVersion {
		fmt.Printf("%s version %s\n", path.Base(os.Args[0]), version)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "error: wrong number of arguments\n")
		os.Exit(1)
	}

	var (
		end time.Time
		err error
	)

	if usage == atHelp {
		end, err = parseTime(flag.Arg(0))
	} else {
		end, err = parseDuartion(flag.Arg(0))
	}

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
		progress: progress.New(progress.WithoutPercentage()),
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}

type model struct {
	start      time.Time
	duration   float64
	percent    float64
	progress   progress.Model
	isFlashing bool
	flashCount int
	showBar    bool
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
		padding := utf8.RuneCountInString(options.prefix)
		m.progress.Width = msg.Width - padding*2 - 6
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		m.percent = float64(time.Since(m.start)) / m.duration

		// Check if timer is complete
		if m.percent >= 1.0 && !m.isFlashing {
			m.percent = 1.0
			m.isFlashing = true
			m.flashCount = 0
			m.showBar = true // Start with bar visible
			// Start flashing
			return m, tea.Tick(flashDuration, func(t time.Time) tea.Msg { return flashMsg{} })
		}

		if m.isFlashing {
			// If flashing, don't update progress or schedule normal ticks
			return m, nil
		}

		// If still running and not yet flashing
		if m.percent < 1.0 {
			return m, tickCmd()
		}
		// This case should ideally not be reached if logic is correct,
		// but as a fallback, quit if percent is >= 1.0 and not caught by flashing logic.
		return m, tea.Quit

	case flashMsg:
		if !m.isFlashing { // Should not happen if logic is correct
			return m, nil
		}
		m.flashCount++
		m.showBar = !m.showBar // Toggle visibility
		if m.flashCount >= totalFlashes {
			return m, tea.Quit // Quit after enough flashes
		}
		// Schedule next flash
		return m, tea.Tick(flashDuration, func(t time.Time) tea.Msg { return flashMsg{} })

	default:
		return m, nil
	}
}

func (m model) left() string {
	d := time.Duration(m.duration) - time.Since(m.start)
	return fmt.Sprintf("%02d:%02d", int(d.Minutes())%60, int(d.Seconds())%60)
}

func (m model) View() string {
	barView := m.progress.ViewAs(m.percent)
	if m.isFlashing && !m.showBar {
		// Replace bar with spaces of the same width
		spaceCount := utf8.RuneCountInString(barView)
		return options.prefix + strings.Repeat(" ", spaceCount) + " " + m.left() + "\n"
	}
	return options.prefix + barView + " " + m.left() + "\n"
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var layouts = []string{
	"15:04",
	"15PM",
	time.Kitchen,
}

func parseTime(s string) (time.Time, error) {
	us := strings.ToUpper(s)
	for _, l := range layouts {
		t, err := time.ParseInLocation(l, us, time.Local)
		if err != nil {
			continue
		}

		now := time.Now()
		t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unknown time format: %q", s)
}

func parseDuartion(s string) (time.Time, error) {
	d, err := time.ParseDuration(s)
	var zt time.Time
	if err != nil {
		return zt, err
	}
	if d <= 0 {
		return zt, fmt.Errorf("%v: bad duration", d)
	}

	return time.Now().Add(d), nil
}
