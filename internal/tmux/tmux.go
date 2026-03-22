package tmux

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// run executes a tmux command with the given arguments.
// On error it returns the combined output for diagnostics.
func run(args ...string) error {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return nil
}

// NewSession creates a new detached tmux session with the given name and
// working directory.
func NewSession(name, cwd string) error {
	return run("new-session", "-d", "-s", name, "-c", cwd)
}

// SendMessage safely sends a message to a tmux session by piping the text
// through load-buffer (via stdin) and paste-buffer, then pressing Enter.
// This avoids all shell escaping issues that plague send-keys with literal text.
func SendMessage(session, message string) error {
	// Send an empty string first to ensure the TUI input is focused.
	// After long idle periods, the first keystroke may be consumed by
	// the TUI framework to re-focus, swallowing the actual message.
	_ = run("send-keys", "-t", session, "")

	// Load the message into the tmux paste buffer via stdin.
	cmd := exec.Command("tmux", "load-buffer", "-")
	cmd.Stdin = strings.NewReader(message)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux load-buffer: %w: %s", err, out)
	}

	// Paste the buffer into the target session.
	if err := run("paste-buffer", "-t", session); err != nil {
		return err
	}

	// Press Enter to submit.
	return run("send-keys", "-t", session, "Enter")
}

// SendKeys sends raw key sequences to a tmux session.
func SendKeys(session string, keys ...string) error {
	args := append([]string{"send-keys", "-t", session}, keys...)
	return run(args...)
}

// Interrupt sends an Escape key to the given tmux session.
func Interrupt(session string) error {
	// Send Escape twice — sometimes Claude Code needs a second press
	// to cancel out of a thinking/tool state.
	if err := SendKeys(session, "Escape"); err != nil {
		return err
	}
	return SendKeys(session, "Escape")
}

// StartClaude launches claude in the given tmux session.
// It auto-confirms the workspace trust prompt by polling tmux output.
func StartClaude(session string) error {
	if err := SendKeys(session, "claude --dangerously-skip-permissions", "Enter"); err != nil {
		return err
	}
	go autoConfirmTrust(session)
	return nil
}

// ResumeClaude launches claude with --resume to continue an existing
// conversation in the given tmux session.
func ResumeClaude(session, sessionID string) error {
	if err := SendKeys(session, fmt.Sprintf("claude --resume %s --dangerously-skip-permissions", sessionID), "Enter"); err != nil {
		return err
	}
	go autoConfirmTrust(session)
	return nil
}

// autoConfirmTrust polls tmux pane output and sends Enter when the trust prompt is detected.
func autoConfirmTrust(session string) {
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		cmd := exec.Command("tmux", "capture-pane", "-t", session, "-p")
		out, err := cmd.Output()
		if err != nil {
			log.Printf("[autoConfirmTrust] capture-pane %s error: %v", session, err)
			continue
		}
		content := string(out)
		if strings.Contains(content, "Yes, I trust this folder") {
			log.Printf("[autoConfirmTrust] trust prompt detected for %s, confirming", session)
			_ = SendKeys(session, "Enter")
			return
		}
	}
	log.Printf("[autoConfirmTrust] no trust prompt detected for %s within 15s", session)
}

// CapturePaneDynamic captures the current visible content of a tmux pane,
// trimming trailing empty lines.
func CapturePaneDynamic(session string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", session, "-p")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("capture-pane: %w", err)
	}
	// Trim trailing blank lines
	content := strings.TrimRight(string(out), "\n ")
	return content, nil
}

// HasSession reports whether a tmux session with the given name exists.
func HasSession(name string) bool {
	return run("has-session", "-t", name) == nil
}

// KillSession destroys the tmux session with the given name.
func KillSession(name string) error {
	return run("kill-session", "-t", name)
}

// CleanupStaleSessions kills all tmux sessions with the "pilot-" prefix.
// Intended to be called on server startup.
func CleanupStaleSessions() {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.HasPrefix(name, "pilot-") {
			_ = KillSession(name)
		}
	}
}
