package claude

import (
	"encoding/json"
	"regexp"
	"strings"
)

// RawMessage represents a single line from a Claude Code JSONL conversation file.
type RawMessage struct {
	Type        string          `json:"type"`
	UUID        string          `json:"uuid"`
	ParentUUID  string          `json:"parentUuid"`
	Timestamp   string          `json:"timestamp"`
	SessionID   string          `json:"sessionId"`
	CWD         string          `json:"cwd"`
	GitBranch   string          `json:"gitBranch"`
	IsSidechain bool            `json:"isSidechain"`
	Message     *MessageContent `json:"message,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}

// MessageContent holds the role, content blocks, and model metadata.
type MessageContent struct {
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model,omitempty"`
	Usage      json.RawMessage `json:"usage,omitempty"`
	StopReason string         `json:"stop_reason,omitempty"`
}

// UnmarshalJSON handles the case where "content" is either a string or an array.
// Claude Code JSONL files use a string for plain user input and an array for
// structured messages (tool_use, tool_result, etc.).
func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion.
	type Alias struct {
		Role       string          `json:"role"`
		RawContent json.RawMessage `json:"content"`
		Model      string          `json:"model,omitempty"`
		Usage      json.RawMessage `json:"usage,omitempty"`
		StopReason string          `json:"stop_reason,omitempty"`
	}

	var a Alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	mc.Role = a.Role
	mc.Model = a.Model
	mc.Usage = a.Usage
	mc.StopReason = a.StopReason

	if len(a.RawContent) == 0 {
		return nil
	}

	// Try as string first.
	var s string
	if err := json.Unmarshal(a.RawContent, &s); err == nil {
		mc.Content = []ContentBlock{{Type: "text", Text: s}}
		return nil
	}

	// Otherwise parse as array of ContentBlock.
	return json.Unmarshal(a.RawContent, &mc.Content)
}

// ContentBlock is a single block inside a message (text, tool_use, tool_result, thinking).
type ContentBlock struct {
	Type    string          `json:"type"`
	Text    string          `json:"text,omitempty"`
	Name    string          `json:"name,omitempty"`
	ID      string          `json:"id,omitempty"`
	Input   json.RawMessage `json:"input,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

var (
	reCommandName = regexp.MustCompile(`<command-name>(.+?)</command-name>`)
	reCommandArgs = regexp.MustCompile(`<command-args>([\s\S]*?)</command-args>`)
	reSkillDir    = regexp.MustCompile(`^Base directory for this skill:\s*(.+)`)
)

// ParseLine parses a single JSONL line into a RawMessage.
// It also normalizes system-injected user messages into human-readable form.
func ParseLine(line []byte) (*RawMessage, error) {
	var msg RawMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil, err
	}
	if msg.Type == "user" && msg.Message != nil && len(msg.Message.Content) > 0 {
		normalizeUserContent(msg.Message)
	}
	return &msg, nil
}

// normalizeUserContent transforms system-injected user messages into readable form.
func normalizeUserContent(mc *MessageContent) {
	if len(mc.Content) == 0 || mc.Content[0].Type != "text" {
		return
	}
	text := mc.Content[0].Text

	// Slash command: extract /command-name args
	if strings.Contains(text, "<command-name>") {
		name := ""
		args := ""
		if m := reCommandName.FindStringSubmatch(text); len(m) > 1 {
			name = m[1]
		}
		if m := reCommandArgs.FindStringSubmatch(text); len(m) > 1 {
			args = strings.TrimSpace(m[1])
		}
		if name != "" {
			display := name
			if args != "" {
				display += " " + args
			}
			mc.Content = []ContentBlock{{Type: "text", Text: display}}
		}
		return
	}

	// Skill expansion: "Base directory for this skill: ..."
	// These are duplicates of the slash command message — hide them.
	if reSkillDir.MatchString(text) {
		mc.Content = nil // Mark as empty so IsRenderable filters it out.
		return
	}
}

// IsRenderable reports whether the message should be displayed in a conversation view.
func (m *RawMessage) IsRenderable() bool {
	if m.Type != "user" && m.Type != "assistant" {
		return false
	}
	if m.Message == nil || len(m.Message.Content) == 0 {
		return false
	}
	// Skip local command caveats — pure system noise.
	if m.Type == "user" && len(m.Message.Content) == 1 && m.Message.Content[0].Type == "text" {
		text := m.Message.Content[0].Text
		if strings.HasPrefix(text, "<local-command-caveat>") {
			return false
		}
	}
	return true
}
