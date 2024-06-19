package input

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
)

var _ Input = (*Text)(nil)

type Text struct {
	model        textinput.Model
	noStyle      lipgloss.Style
	focusedStyle lipgloss.Style
	readonly     bool
}

func NewText(
	placeholder string,
	charLimit int,
) *Text {
	model := textinput.New()
	model.Placeholder = placeholder
	model.CharLimit = charLimit
	model.Cursor.SetMode(cursor.CursorBlink)
	t := &Text{
		model:        model,
		noStyle:      styles.NoStyle,
		focusedStyle: styles.FocusedStyle,
	}
	t.Reset()
	return t
}

func NewTextPassword(
	placeholder string,
	charLimit int,
) *Text {
	model := textinput.New()
	model.Placeholder = placeholder
	model.CharLimit = charLimit
	model.EchoMode = textinput.EchoPassword
	model.EchoCharacter = '•'
	model.Cursor.SetMode(cursor.CursorBlink)
	t := &Text{
		model:        model,
		noStyle:      styles.NoStyle,
		focusedStyle: styles.FocusedStyle,
	}
	t.Reset()
	return t
}

func NewTextReadonly(
	placeholder string,
	charLimit int,
) *Text {
	model := textinput.New()
	model.Placeholder = placeholder
	model.CharLimit = charLimit
	model.Cursor.SetMode(cursor.CursorBlink)
	t := &Text{
		model:        model,
		noStyle:      styles.NoStyle,
		focusedStyle: styles.FocusedStyle,
		readonly:     true,
	}
	t.Reset()
	return t
}

func (t *Text) Value() string {
	return t.model.Value()
}

func (t *Text) SetValue(v string) {
	t.model.SetValue(v)
}

func (t *Text) Focus() tea.Cmd {
	t.model.PromptStyle = t.focusedStyle
	t.model.TextStyle = t.focusedStyle
	return t.model.Focus()
}

func (t *Text) Blur() {
	t.model.PromptStyle = t.noStyle
	t.model.TextStyle = t.noStyle
	t.model.Blur()
}

func (t *Text) Reset() {
	t.model.Reset()
	t.model.PromptStyle = t.noStyle
	t.model.TextStyle = t.noStyle
	t.model.Blur()
}

func (t *Text) View() string {
	return t.model.View()
}

func (t *Text) Update(msg tea.Msg) (cmd tea.Cmd) {
	if t.readonly {
		return cmd
	}
	t.model, cmd = t.model.Update(msg)
	return cmd
}
