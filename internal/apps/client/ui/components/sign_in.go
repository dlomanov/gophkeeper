package components

import (
	"context"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"strings"
	"time"
)

var _ base.Component = (*SignIn)(nil)

type (
	SignIn struct {
		userUC     *usecases.UserUC
		cache      *mem.Cache
		title      string
		back       base.Component
		inputs     []textinput.Model
		focusIndex int
	}
	signInMsg struct {
		token string
		err   error
	}
)

func NewSignIn(
	title string,
	userUC *usecases.UserUC,
	cache *mem.Cache,
) *SignIn {
	c := &SignIn{
		title:  title,
		back:   nil,
		userUC: userUC,
		cache:  cache,
	}
	c.inputs = []textinput.Model{
		c.loginInput(),
		c.passwordInput(),
	}
	return c
}

func (c SignIn) Title() string {
	return c.title
}

func (c *SignIn) Init() (result base.InitResult) {
	c.reset()
	return result
}

func (c *SignIn) Update(msg tea.Msg) (result base.UpdateResult) {
	switch msg := msg.(type) {
	case signInMsg:
		return c.updateSignInMsg(msg, result)
	case tea.KeyMsg:
		if result = c.updateKeyMsg(msg, result); result.Cmd != nil {
			return result
		}
	}
	cmds := make([]tea.Cmd, len(c.inputs))
	for i := range c.inputs {
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
	}
	return result.AppendCmd(cmds...)
}

func (c *SignIn) View() string {
	sb := strings.Builder{}
	for i := 0; i < len(c.inputs); i++ {
		sb.WriteString(c.inputs[i].View())
		sb.WriteRune('\n')
	}
	sb.WriteByte('\n')
	if c.focusIndex == len(c.inputs) {
		sb.WriteString(styles.FocusedButton)
	} else {
		sb.WriteString(styles.BlurredButton)
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(styles.SubtleStyle.Render("tab: next"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("q: quit"))
	sb.WriteString(styles.DotStyle)
	sb.WriteString(styles.SubtleStyle.Render("esc: back"))
	return sb.String()
}

func (c *SignIn) SetPrev(back base.Component) {
	c.back = back
}

func (c *SignIn) updateSignInMsg(
	msg signInMsg,
	result base.UpdateResult,
) base.UpdateResult {
	if msg.err != nil {
		result.Status = msg.err.Error()
		return result
	}
	result.Status = "signed-in 😎"
	result.Prev = c.back
	return result
}

func (c *SignIn) updateKeyMsg(
	msg tea.KeyMsg,
	result base.UpdateResult,
) base.UpdateResult {
	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		result.Quitting = true
		result.Cmd = tea.Quit
		return result
	case "esc":
		result.Prev = c.back
		return result

	case "tab", "shift+tab", "up", "down", "enter":
		if k == "enter" && c.focusIndex == len(c.inputs) && c.inputsValid() {
			result.Cmd = c.signInCmd(c.inputs[0].Value(), c.inputs[1].Value())
			return result
		}

		if k == "up" || k == "shift+tab" {
			c.focusIndex--
		} else {
			c.focusIndex++
		}
		if c.focusIndex > len(c.inputs) {
			c.focusIndex = 0
		} else if c.focusIndex < 0 {
			c.focusIndex = len(c.inputs)
		}

		cmds := make([]tea.Cmd, len(c.inputs))
		for i := 0; i <= len(c.inputs)-1; i++ {
			if i == c.focusIndex {
				// Set focused state
				cmds[i] = c.inputs[i].Focus()
				c.inputs[i].PromptStyle = styles.FocusedStyle
				c.inputs[i].TextStyle = styles.FocusedStyle
				continue
			}
			// Remove focused state
			c.inputs[i].Blur()
			c.inputs[i].PromptStyle = styles.NoStyle
			c.inputs[i].TextStyle = styles.NoStyle
		}
		return result.AppendCmd(cmds...)
	}
	return result
}

func (c *SignIn) reset() {
	c.focusIndex = 0
	c.inputs[0].Focus()
	c.inputs[0].PromptStyle = styles.FocusedStyle
	c.inputs[0].TextStyle = styles.FocusedStyle
	c.inputs[0].SetValue("")
	for i := 1; i < len(c.inputs); i++ {
		c.inputs[i].Blur()
		c.inputs[i].PromptStyle = styles.NoStyle
		c.inputs[i].TextStyle = styles.NoStyle
		c.inputs[i].Cursor.Style = styles.CursorStyle
		c.inputs[i].SetValue("")
	}
}

func (c *SignIn) signInCmd(login, pass string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := c.userUC.SignIn(ctx, entities.SignInUserRequest{
			Login:    login,
			Password: pass,
		})
		return signInMsg{err: err}
	}
}

func (c *SignIn) loginInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Login"
	ti.CharLimit = 32
	ti.Cursor.SetMode(cursor.CursorBlink)
	return ti
}

func (c *SignIn) passwordInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Password"
	ti.CharLimit = 32
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Cursor.SetMode(cursor.CursorBlink)
	return ti
}

func (c *SignIn) inputsValid() bool {
	for i := range c.inputs {
		if c.inputs[i].Value() == "" {
			return false
		}
	}
	return true
}
