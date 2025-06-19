package main

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func Test_parseTime(t *testing.T) {
	testCases := []struct {
		s      string
		err    bool
		hour   int
		minute int
	}{
		{"2:45PM", false, 14, 45},
		{"2:45pm", false, 14, 45},
		{"13:37", false, 13, 37},
		{"4pm", false, 16, 00},
		{"6", true, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			tm, err := parseTime(tc.s)
			if tc.err {
				require.Error(t, err)
				return
			}

			now := time.Now()
			require.Equal(t, tc.hour, tm.Hour())
			require.Equal(t, tc.minute, tm.Minute())

			require.Equal(t, now.Year(), tm.Year())
			require.Equal(t, now.Month(), tm.Month())
			require.Equal(t, now.Day(), tm.Day())
		})
	}
}

func TestUpdate_FlashingLogic(t *testing.T) {
	initialModel := func() model {
		return model{
			start:    time.Now(),
			duration: float64(10 * time.Second), // Example duration
			percent:  0,
			progress: progress.New(progress.WithoutPercentage()),
			// isFlashing, flashCount, showBar are zero/false by default
		}
	}

	t.Run("starts_flashing_when_timer_completes", func(t *testing.T) {
		m := initialModel()
		// Adjust start time so that time.Since(m.start) / m.duration will be >= 1.0
		// This ensures the internal recalculation of m.percent in Update() hits the condition.
		m.start = time.Now().Add(-time.Duration(m.duration) - time.Second) // Set start to be more than 'duration' ago

		// Send a tickMsg to trigger the percent check and flashing logic
		updatedModel, cmd := m.Update(tickMsg(time.Now()))
		um := updatedModel.(model)

		require.True(t, um.isFlashing, "isFlashing should be true")
		require.Equal(t, 0, um.flashCount, "flashCount should be 0 initially")
		require.True(t, um.showBar, "showBar should be true initially")
		require.NotNil(t, cmd, "A command to trigger flashMsg should be returned")
		// Also check that percent is capped at 1.0 after flashing starts
		require.Equal(t, 1.0, um.percent, "percent should be capped at 1.0")
	})

	t.Run("toggles_showBar_and_increments_flashCount_on_flashMsg", func(t *testing.T) {
		m := initialModel()
		m.isFlashing = true
		m.showBar = true
		m.flashCount = 0

		// Simulate first flash
		updatedModel, cmd := m.Update(flashMsg{})
		um := updatedModel.(model)

		require.Equal(t, 1, um.flashCount, "flashCount should increment")
		require.False(t, um.showBar, "showBar should toggle to false")
		require.NotNil(t, cmd, "A command for the next flash should be returned")

		// Simulate second flash
		updatedModel2, cmd2 := um.Update(flashMsg{})
		um2 := updatedModel2.(model)

		require.Equal(t, 2, um2.flashCount, "flashCount should increment again")
		require.True(t, um2.showBar, "showBar should toggle back to true")
		require.NotNil(t, cmd2, "A command for the next flash should be returned")
	})

	t.Run("quits_after_totalFlashes", func(t *testing.T) {
		m := initialModel()
		m.isFlashing = true
		m.showBar = false // Current state before the last flash
		m.flashCount = totalFlashes - 1 // One flash away from quitting

		updatedModel, cmd := m.Update(flashMsg{})
		um := updatedModel.(model)

		require.Equal(t, totalFlashes, um.flashCount, "flashCount should reach totalFlashes")
		require.True(t, um.showBar, "showBar should toggle one last time")

		// Check if tea.Quit is signaled
		// For this specific setup, tea.Quit is a special command that, when returned by Update,
		// signals the Program to terminate. We can check if the returned cmd is nil,
		// as tea.Quit itself is often represented as a nil command in some contexts or
		// a specific quit command type if the library offers it.
		// More robustly, we'd check if the command causes a quit,
		// but here we assume tea.Quit is returned directly.
		// If tea.Quit is a specific type, we'd assert its type.
		// Given the current code, Update returns `m, tea.Quit`, and tea.Quit is a special marker.
		// The bubbletea library handles `tea.Quit` which is a sentinel error value.
		// Let's assume the command itself will be tea.Quit.
		// isQuitCmd := false // This variable was declared but not used.
		if cmd != nil {
			// This is a bit of a hack. In Bubble Tea, tea.Quit is a function that returns a Cmd.
			// To properly test this, we'd need to see if cmd() == tea.Quit().
			// Or, if tea.Quit is an exported sentinel value, compare directly.
			// For now, we'll assume that if flashCount >= totalFlashes, the cmd is tea.Quit.
			// The actual tea.Quit command is a function, so direct comparison is tricky.
			// Let's verify the state that LEADS to tea.Quit.
			// The subtask actually returns `m, tea.Quit`
			// A more direct test of `cmd == tea.Quit` is not straightforward as tea.Quit is a function.
			// We'll rely on the logic: if flashCount >= totalFlashes, the *next* message processing loop will quit.
			// The test setup here means that cmd *is* tea.Quit.
		}
		// The most reliable way to test tea.Quit is to check if the command is the tea.Quit command.
		// tea.Quit is of type tea.Cmd. It's a function that returns itself.
		// We test it by executing the command and checking if the message is what tea.Quit() produces.
		require.NotNil(t, cmd, "Command should not be nil when expecting tea.Quit")
		actualMsg := cmd()
		expectedMsg := tea.Quit() // tea.Quit() returns a tea.quitMsg
		require.Equal(t, expectedMsg, actualMsg, "Command should effectively be tea.Quit by producing the same message")

	})

	t.Run("percent_is_capped_at_1_and_flashing_starts_correctly_from_tickMsg", func(t *testing.T) {
		m := initialModel()
		m.duration = float64(1 * time.Second) // Short duration
		m.start = time.Now().Add(-2 * time.Second) // Ensure time has passed

		// Simulate a tick that would push percent over 1.0
		updatedModel, cmd := m.Update(tickMsg(time.Now()))
		um := updatedModel.(model)

		require.Equal(t, 1.0, um.percent, "percent should be capped at 1.0")
		require.True(t, um.isFlashing, "isFlashing should be true")
		require.Equal(t, 0, um.flashCount, "flashCount should be 0")
		require.True(t, um.showBar, "showBar should be true")
		require.NotNil(t, cmd, "A command to trigger flashMsg should be returned")

		// Check that the command is for flashMsg, not another tickMsg for the main timer
		// This requires being ableto inspect the command, which is tricky with tea.Tick.
		// For now, we trust the logic implemented.
	})
}
