package components

import (
	"context"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/input"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"github.com/dlomanov/gophkeeper/internal/core"
	"strings"
	"time"
)

var (
	_ base.Component = (*Main)(nil)
)

type (
	Main struct {
		title      string
		container  *deps.Container
		passInput  *input.Text
		focusIndex int
		processing bool
	}
	authMsg struct {
		err error
	}
)

func NewMain(title string, container *deps.Container) *Main {
	c := &Main{
		title:      title,
		container:  container,
		focusIndex: 0,
		processing: false,
	}
	c.passInput = input.NewTextPassword("Master-password", 64)
	return c
}

func (c *Main) Title() string {
	return c.title
}

func (c *Main) Init() (result base.InitResult) {
	c.focusIndex = 0
	c.passInput.Reset()
	return result.AppendCmd(
		c.passInput.Focus(),
		base.UpdateStatusCmd("💡 len(password) >= 8"))
}

func (c *Main) Update(msg tea.Msg) (result base.UpdateResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if result = c.updateKeyMsg(msg, result); result.Cmd != nil {
			return result
		}
	case authMsg:
		c.processing = false
		switch {
		case errors.Is(msg.err, entities.ErrUserMasterPassInvalid):
			result.Status = "invalid password 😳"
			return result
		case msg.err != nil:
			result.Status = "pls try again 🙃"
			return result
		}
		result.PassAccepted = true
	}
	return result.AppendCmd(
		c.passInput.Update(msg))
}

func (c *Main) View() string {
	sb := strings.Builder{}
	sb.WriteString(c.passInput.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	if c.focusIndex == 1 {
		sb.WriteString(styles.FocusedButton)
	} else {
		sb.WriteString(styles.BlurredButton)
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	return sb.String()
}

func (c *Main) updateKeyMsg(
	msg tea.KeyMsg,
	result base.UpdateResult,
) base.UpdateResult {
	if c.processing {
		result.Status = "🤔"
		return result
	}

	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		result.Quitting = true
		return result.AppendCmd(tea.Quit)
	case "tab", "shift+tab", "up", "down", "enter":
		value := c.passInput.Value()
		if k == "enter" && c.focusIndex == 1 && value != "" {
			if len(value) < 8 {
				result.Status = "pass length should be at least 8 🤔"
				return result
			}
			if c.processing {
				result.Status = "🤔"
				return result
			}
			c.processing = true
			return result.AppendCmd(c.authCmd(value))
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
			result.AppendCmd(c.passInput.Focus())
		} else {
			c.passInput.Blur()
		}
	}
	return result
}

func (c *Main) quite(result base.UpdateResult) base.UpdateResult {
	if c.processing {
		result.Status = "🤔"
		return result
	}
	result.Quitting = true
	return result.AppendCmd(tea.Quit)
}

func (c *Main) authCmd(pass string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Second)
		defer cancel()
		if err := c.container.Register(ctx, core.Pass(pass)); err != nil {
			return authMsg{err: fmt.Errorf("main: failed to auth user and register dependencies: %w", err)}
		}
		return authMsg{err: nil}
	}
}
