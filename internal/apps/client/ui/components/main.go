package components

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	core "github.com/dlomanov/gophkeeper/internal/core"
	"strings"
	"sync/atomic"
	"time"
)

var (
	_             Component = (*Main)(nil)
	focusedButton           = focusedStyle.Render("[ Submit ]")
	blurredButton           = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type (
	Main struct {
		title      string
		container  *deps.Container
		passInput  textinput.Model
		focusIndex int
		processing atomic.Bool
	}
	regDepsMsg struct {
		err error
	}
)

func NewMain(title string, container *deps.Container) *Main {
	c := &Main{
		title:      title,
		container:  container,
		focusIndex: 0,
		processing: atomic.Bool{},
	}
	c.passInput = c.passwordInput()
	return c
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
			if c.processing.Load() {
				result.Status = "🤔"
				return result, nil
			}
			result.Quitting = true
			return result, tea.Quit
		case "tab", "shift+tab", "up", "down", "enter":
			if c.processing.Load() {
				result.Status = "🤔"
				return result, nil
			}
			value := c.passInput.Value()
			if k == "enter" && c.focusIndex == 1 && value != "" {
				if len(value) < 8 {
					result.Status = "pass length should be at least 8 🤔"
					return result, nil
				}
				if !c.processing.CompareAndSwap(false, true) {
					result.Status = "🤔"
					return result, nil
				}
				return result, c.authCmd(value)
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
	if msg, ok := msg.(regDepsMsg); ok {
		c.processing.Store(false)
		switch {
		case errors.Is(msg.err, entities.ErrUserMasterPassInvalid):
			result.Status = "invalid password 😳"
			return result, nil
		case msg.err != nil:
			result.Status = "pls try again 🙃"
			return result, nil
		}
		result.PassAccepted = true
		return result, nil
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

func (c *Main) authCmd(pass string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Second)
		defer cancel()
		if err := c.container.Register(ctx, core.Pass(pass)); err != nil {
			return regDepsMsg{err: fmt.Errorf("main: failed to auth user and register dependencies: %w", err)}
		}
		return regDepsMsg{err: nil}
	}
}

func (*Main) passwordInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Master-password"
	ti.CharLimit = 32
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Cursor.SetMode(cursor.CursorBlink)
	return ti
}
