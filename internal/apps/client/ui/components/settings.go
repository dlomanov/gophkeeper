package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

var _ Component = (*Settings)(nil)

type Settings struct {
	title string
	back  Component
}

func NewSettings(title string, back Component) *Settings {
	return &Settings{
		title: title,
		back:  back,
	}
}

func (c *Settings) Title() string {
	return c.title
}

func (c *Settings) Init() tea.Cmd {
	return nil
}

func (c *Settings) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "esc":
			result.Prev = c.back
			return result, nil
		}
	}

	return result, nil
}

func (c *Settings) View() string {
	sb := strings.Builder{}
	sb.WriteString("Settings view")
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(subtleStyle.Render("esc: back"))
	sb.WriteString(dotStyle)
	sb.WriteString(subtleStyle.Render("q: quit"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}

func (c *Settings) SetPrev(back Component) {
	c.back = back
}
