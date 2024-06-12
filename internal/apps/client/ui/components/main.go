package components

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

var (
	_             Component = (*Main)(nil)
	focusedButton           = focusedStyle.Render("[ Submit ]")
	blurredButton           = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type Main struct {
	title          string
	next           Component
	passInput      textinput.Model
	focusIndex     int
	submittedCount int
}

func NewMain(title string, next Component) *Main {
	return &Main{
		title:      title,
		next:       next,
		passInput:  passwordInput("Master-password"),
		focusIndex: 0,
	}
}

func (c *Main) Title() string {
	return c.title
}

func (c *Main) Init() tea.Cmd {
	c.focusIndex = 0
	c.passInput.PromptStyle = focusedStyle
	c.passInput.TextStyle = focusedStyle
	c.passInput.SetValue("")
	return c.passInput.Focus()
}

func (c *Main) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "esc", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "tab", "shift+tab", "up", "down", "enter":
			if k == "enter" && c.focusIndex == 1 {
				result.Next = c.next
				return result, nil
			}

			// Cycle indexes
			if k == "up" || k == "shift+tab" {
				c.focusIndex--
			} else {
				c.focusIndex++
			}
			if c.focusIndex > 1 {
				c.focusIndex = 0
			} else if c.focusIndex < 0 {
				c.focusIndex = 1
			}
			if c.focusIndex == 0 {
				cmd = c.passInput.Focus()
				c.passInput.PromptStyle = focusedStyle
				c.passInput.TextStyle = focusedStyle
			} else {
				c.passInput.Blur()
				c.passInput.PromptStyle = noStyle
				c.passInput.TextStyle = noStyle
			}
			return result, cmd
		}
	}
	c.passInput, cmd = c.passInput.Update(msg)
	return result, cmd
}

func (c *Main) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.passInput.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	if c.focusIndex == 1 {
		sb.WriteString(focusedButton)
	} else {
		sb.WriteString(blurredButton)
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(subtleStyle.Render("q, esc: quit"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}

func (c *Main) SetNext(next Component) {
	c.next = next
}
