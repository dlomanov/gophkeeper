package components

import (
	"context"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"strings"
	"time"
)

var _ Component = (*SignIn)(nil)

type (
	SignIn struct {
		userUC *usecases.UserUC
		cache  *mem.Cache

		title      string
		back       Component
		inputs     []textinput.Model
		focusIndex int
		result     string
	}
	signInMsg struct {
		token string
		err   error
	}
)

func NewSignIn(
	title string,
	back Component,
	userUC *usecases.UserUC,
	cache *mem.Cache,
) *SignIn {
	c := &SignIn{
		title:  title,
		back:   back,
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

func (c *SignIn) Init() tea.Cmd {
	c.reset()
	return nil
}

func (c *SignIn) Update(msg tea.Msg) (result UpdateResult, cmd tea.Cmd) {
	if msg, ok := msg.(signInMsg); ok {
		if msg.err != nil {
			c.result = msg.err.Error()
			return result, nil
		}
		result.Status = "user signed-in"
		result.Prev = c.back
		return result, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		switch k {
		case "q", "ctrl+c":
			result.Quitting = true
			return result, tea.Quit
		case "esc":
			result.Prev = c.back
			return result, nil

		case "tab", "shift+tab", "up", "down", "enter":
			c.result = ""

			if k == "enter" && c.focusIndex == len(c.inputs) && c.inputsValid() {
				return result, c.signInCmd(c.inputs[0].Value(), c.inputs[1].Value())
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
					c.inputs[i].PromptStyle = focusedStyle
					c.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				c.inputs[i].Blur()
				c.inputs[i].PromptStyle = noStyle
				c.inputs[i].TextStyle = noStyle
			}
			return result, tea.Batch(cmds...)
		}
	}

	cmds := make([]tea.Cmd, len(c.inputs))
	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range c.inputs {
		c.inputs[i], cmds[i] = c.inputs[i].Update(msg)
	}
	return result, tea.Batch(cmds...)
}

func (c *SignIn) View() string {
	sb := strings.Builder{}
	for i := 0; i < len(c.inputs); i++ {
		sb.WriteString(c.inputs[i].View())
		sb.WriteRune('\n')
	}
	sb.WriteByte('\n')
	if c.focusIndex == len(c.inputs) {
		sb.WriteString(focusedButton)
	} else {
		sb.WriteString(blurredButton)
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	if c.result != "nil" {
		sb.WriteString(errStyle.Render(c.result))
		sb.WriteByte('\n')
	}
	sb.WriteString(subtleStyle.Render("tab: next"))
	sb.WriteString(dotStyle)
	sb.WriteString(subtleStyle.Render("q: quit"))
	sb.WriteString(dotStyle)
	sb.WriteString(subtleStyle.Render("esc: back"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return sb.String()
}

func (c *SignIn) SetPrev(back Component) {
	c.back = back
}

func (c *SignIn) reset() {
	c.focusIndex = 0
	c.result = ""
	c.inputs[0].Focus()
	c.inputs[0].PromptStyle = focusedStyle
	c.inputs[0].TextStyle = focusedStyle
	c.inputs[0].SetValue("")
	for i := 1; i < len(c.inputs); i++ {
		c.inputs[i].Blur()
		c.inputs[i].PromptStyle = noStyle
		c.inputs[i].TextStyle = noStyle
		c.inputs[i].Cursor.Style = cursorStyle
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
